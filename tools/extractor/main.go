package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode"

	xlslib "github.com/extrame/xls"
)

type Data struct {
	Naf      Sections `json:"naf"`
	Liberale Sections `json:"liberale"`

	// Nafa helper for find Secteur for a given Division
	Helper Helpers `json:"-"`
}

type Sections []Section
type Section struct {
	Reglemente   bool     `json:"reglemente"`
	Titre        string   `json:"titre"`
	Valeur       string   `json:"valeur"`
	Nomenclature string   `json:"nomenclature"`
	Selecteur    string   `json:"selecteur"`
	CfeMorale    string   `json:"cfe-morale"`
	CfePhysique  string   `json:"cfe-physique"`
	Subsec       Sections `json:"subsec"`
}

type Helpers []Helper
type Helper struct {
	Secteur   string
	Divisions []string
}

const (
	SourceNaf     = "https://www.insee.fr/fr/statistiques/fichier/2120875/int_courts_naf_rev_2.xls"
	SourceNafa    = "https://opendata.hauts-de-seine.fr/explore/dataset/entreprises-artisanales-par-code-nafa/download/?format=csv&timezone=Europe/Berlin&lang=fr&use_labels_for_header=true&csv_separator=%3B"
	SourceLiberal = "../../src/data/professions_liberales.csv"

	Output = "../../public/data"
)

func main() {
	// Clear old data if any
	os.RemoveAll(Output)
	if err := os.Mkdir(Output, os.ModePerm); err != nil && !os.IsExist(err) {
		log.Panicln(err)
	}

	// Data storage
	var data Data

	// Naf process
	// La fiche naf donne aussi la structure des secteurs, divisions et groupes
	if err := data.nafBuild(SourceNaf); err != nil {
		log.Panicln(err)
	}

	// Nafa process
	if err := data.nafaBuild(SourceNafa); err != nil {
		log.Panicln(err)
	}

	// Liberale process
	if err := data.liberaleBuild(SourceLiberal); err != nil {
		log.Panicln(err)
	}

	//
	data.Naf.toHTM(Output, "")
	data.Liberale.toHTM(Output, "")
}

// nafBuild is the naf data builder from xls file
func (data *Data) nafBuild(uri string) (err error) {
	// xls col index
	const (
		IndexCode = 1
		IndexName = 2
	)

	// Download xls file
	xlsFile, err := requestData(uri)
	if err != nil {
		return
	}

	// Read and parse xls file
	// Thanks INSEE for this proprietary file format
	xls, err := xlslib.OpenReader(bytes.NewReader(xlsFile), "")
	if err != nil {
		return err
	}

	// Only get the first sheet
	sheet := xls.GetSheet(0)
	if sheet == nil {
		return errors.New("ne sheet fount in xls file at: " + uri)
	}

	// Iterate over row's
	active := make([]*Section, 3, 3)
	for i := 0; i <= int(sheet.MaxRow); i++ {
		row := sheet.Row(i)
		naf := struct {
			code string
			name string
		}{
			row.Col(IndexCode),
			row.Col(IndexName),
		}

		// skip empty naf code
		if naf.code == "" {
			continue
		}

		// Special case for sector
		if len(naf.code) > 8 && naf.code[:7] == "SECTION" {
			// Add new section to Helper list
			data.Helper = append(data.Helper, Helper{
				Secteur: naf.code[8:],
			})
			// Create new section
			data.Naf = append(data.Naf, Section{
				Titre:        naf.name,
				Valeur:       naf.code[8:],
				Selecteur:    "naf",
				Nomenclature: "naf",
			})
			// Add new section to active
			active[0] = &data.Naf[len(data.Naf)-1]
			continue
		}

		// Check and normalize naf code
		if !isNafCode(&naf.code) {
			continue
		}

		// Set common section values
		section := Section{
			Titre:        naf.name,
			Valeur:       naf.code,
			Nomenclature: "naf",
		}
		// Set section type
		switch len(naf.code) {
		// Divisions
		case 2:
			// Add new division to Helper list
			help := &data.Helper[len(data.Helper)-1]
			help.Divisions = append(help.Divisions, naf.code)
			// Create new division
			section.Selecteur = "division"
			active[0].Subsec = append(active[0].Subsec, section)
			// Add new division to active
			active[1] = &active[0].Subsec[len(active[0].Subsec)-1]
		// Groupes
		case 3:
			// Create new groupe
			section.Selecteur = "groupe"
			active[1].Subsec = append(active[1].Subsec, section)
			// Add new groupe to active
			active[2] = &active[1].Subsec[len(active[1].Subsec)-1]
		// Classes
		case 5:
			// Create new class
			section.Selecteur = "classe"
			active[2].Subsec = append(active[2].Subsec, section)
		default:
			continue
		}
	}

	return
}

// nafaBuild is the nafa data builder from csv file
// The result update or add nafa to naf data structure
func (data *Data) nafaBuild(uri string) (err error) {
	// csv col index
	const (
		IndexCode = 1
	)

	// Download csv file
	csvFile, err := requestData(uri)
	if err != nil {
		return
	}

	// Read and parse csv file
	csvReader := csv.NewReader(bytes.NewReader(csvFile))
	csvReader.Comma = ';'

	// Iterate over row's
	// start at 1 to skip header
	for i := 0; ; i++ {
		row, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		// skip first line empty or invalid naf code
		if i == 0 || len(row) < 3 || len(row[IndexCode]) < 6 {
			continue
		}

		// set basic naf values
		naf := struct {
			code string
			name string
		}{
			row[IndexCode][:6],
			row[IndexCode][6:],
		}

		// Check and normalize naf code
		if !isNafCode(&naf.code) {
			continue
		}

		// Get pointer to group
		group := data.getGroup(naf.code[:5])
		if group == nil {
			log.Println("no group found for naf code: " + naf.code)
			continue
		}

		// Update or Set section values
		ii := group.Subsec.getIndex(naf.code[:5])

		if ii >= 0 {
			group.Subsec[ii].Titre = naf.name
			group.Subsec[ii].Valeur = naf.code
			continue
		}

		// Create new section
		group.Subsec = append(group.Subsec, Section{
			Titre:        naf.name,
			Valeur:       naf.code,
			Selecteur:    "classe",
			Nomenclature: "nafa",
		})
	}

	return
}

// liberaleBuild is the liberale data builder from csv file
func (data *Data) liberaleBuild(uri string) (err error) {
	// csv col index
	const (
		secteurName = iota
		denominationID
		regID
		codeID
		cfeID
	)
	// Open csv file
	csvFile, err := os.Open(uri)
	if err != nil {
		return
	}
	defer csvFile.Close()

	// Read and parse csv file
	csvReader := csv.NewReader(csvFile)
	csvReader.Comma = ';'

	// Var for current secteur
	secteur := struct {
		id   int
		name string
	}{}

	// Iterate over row's
	// start at 1 to skip header
	for i := 0; ; i++ {
		row, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		// skip first or empty line
		if i == 0 || len(row) < 5 {
			continue
		}

		// Check if secteur changed
		if row[secteurName] != secteur.name {
			secteur.name = row[secteurName]
			secteur.id = data.Liberale.getIndex(secteur.name)

			if secteur.id < 0 {
				data.Liberale = append(data.Liberale, Section{
					Titre:        secteur.name,
					Valeur:       secteur.name,
					Selecteur:    "liberale",
					Nomenclature: "liberale",
				})
				secteur.id = len(data.Liberale) - 1
			}
		}

		// Add new classe to secteur
		classe := Section{
			Titre:        row[denominationID],
			Valeur:       row[codeID],
			Selecteur:    "classe",
			Nomenclature: "liberale",
			CfePhysique:  row[cfeID],
		}

		if row[regID] == "R" {
			classe.Reglemente = true
		}

		data.Liberale[secteur.id].Subsec = append(data.Liberale[secteur.id].Subsec, classe)
	}

	return
}

// getGroup returns a pointer to a group
func (data *Data) getGroup(code string) *Section {
	// Get the section code
	scode := data.Helper.getSection(code[:2])
	if scode == "" {
		return nil
	}

	// Section index
	section := data.Naf.getIndex(scode)
	if section == -1 {
		return nil
	}

	// Division index
	division := data.Naf[section].Subsec.getIndex(code[:2])
	if division == -1 {
		return nil
	}

	// Groupe index
	groupe := data.Naf[section].Subsec[division].Subsec.getIndex(code[:3])
	if groupe == -1 {
		return nil
	}

	return &data.Naf[section].Subsec[division].Subsec[groupe]
}

// getIndex returns the index of a section with a given code
func (sec *Sections) getIndex(code string) (index int) {
	index = -1

	for i, s := range *sec {
		if s.Valeur == code {
			index = i
			break
		}
	}

	return
}

// toHTM write htm selection file from data
func (sec *Sections) toHTM(path, parent string) (err error) {
	// If empty return
	if len(*sec) < 1 {
		return
	}

	// Define the output file
	var out string
	if parent == "" {
		out = strings.Join([]string{path, (*sec)[0].Selecteur + ".htm"}, "/")
	} else {
		// If not exist create parent folder
		out = path + `/` + (*sec)[0].Selecteur
		if err := os.Mkdir(out, os.ModePerm); err != nil && !os.IsExist(err) {
			return err
		}

		out = out + `/` + parent + ".htm"
	}

	// Open output file in write mode
	file, err := os.OpenFile(out, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return
	}
	defer file.Close()

	// Write option
	// default select option
	file.WriteString("<option disabled selected>Sélectionnez une option</option>")
	for i := range *sec {
		s := &(*sec)[i]
		file.WriteString(s.optionHTM())

		// Write sub section if exist
		if s.Subsec != nil {
			s.Subsec.toHTM(path, s.Valeur)
		}
	}

	return
}

// optionHTM returns the html option for a section
func (s *Section) optionHTM() string {
	var fields []string

	fields = append(fields, `value="`+s.Valeur+`"`)
	fields = append(fields, `data-nomenclature="`+s.Nomenclature+`"`)
	fields = append(fields, `data-selecteur="`+s.Selecteur+`"`)

	if s.Reglemente {
		fields = append(fields, `data-reglemente="true"`)
	}

	if s.CfeMorale != "" {
		fields = append(fields, `data-cfemorale="`+s.CfeMorale+`"`)
	}

	if s.CfePhysique != "" {
		fields = append(fields, `data-cfephysique="`+s.CfePhysique+`"`)
	}

	return `<option ` + strings.Join(fields, " ") + `>` + s.Titre + `</option>`
}

// getSection returns the section code for a given division code
func (h *Helpers) getSection(code string) string {
	for i := range *h {
		for _, d := range (*h)[i].Divisions {
			if d == code {
				return (*h)[i].Secteur
			}
		}
	}

	return ""
}

func isNafCode(n *string) bool {
	// NAF and NAFA codes are between 2 and 6 char
	if l := len(*n); l < 2 || l > 6 {
		return false
	}

	// Clean given code
	*n = strings.Replace(*n, ".", "", 1)
	*n = strings.TrimSpace(*n)

	for i, r := range *n {
		// 4 first char must be a number
		if i < 4 && unicode.IsDigit(r) {
			continue
		}

		// The 5 and 6 char must be a letter
		if i >= 4 && unicode.IsLetter(r) {
			continue
		}

		return false
	}

	return true
}

func requestData(url string) (file []byte, err error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return
	}

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = errors.New("status code error: " + resp.Status)
		return
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}
