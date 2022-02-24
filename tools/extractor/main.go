// Copyright 2022 Iglou.eu.
// Written by Adrien Kara <adrien@iglou.eu>
// Use of this source code is governed by GPL-3.0-or-later
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
	"unicode"

	xlslib "github.com/extrame/xls"
)

type OFile struct {
	ID   string
	File *os.File
}

type Naf struct {
	Secteur  OFile
	Division OFile
	Groupe   OFile
	Classe   OFile
}

type NafRows []NafRow
type NafRow struct {
	Code string
	Name string
}

type SelectOption struct {
	Value string
	NomID string
	Text  string
	Type  string
}

const (
	SourceNaf  = "https://www.insee.fr/fr/statistiques/fichier/2120875/int_courts_naf_rev_2.xls"
	SourceNafa = "https://opendata.hauts-de-seine.fr/explore/dataset/entreprises-artisanales-par-code-nafa/download/?format=csv&timezone=Europe/Berlin&lang=fr&use_labels_for_header=true&csv_separator=%3B  "
)

var (
	output string

	NafFileSecteurs  = "secteurs.htm"
	NafFoldDivisions = "divisions"
	NafFoldGroupes   = "groupes"
	NafFoldClasses   = "classes"
)

func main() {
	flag.StringVar(&output, "o", "", "Output dir option (required)")
	flag.Parse()

	if output == "" {
		log.Panicln("You must to provide an output directory")
	}

	// Initialise data folder var
	NafFileSecteurs = path.Join(output, NafFileSecteurs)
	NafFoldDivisions = path.Join(output, NafFoldDivisions)
	NafFoldGroupes = path.Join(output, NafFoldGroupes)
	NafFoldClasses = path.Join(output, NafFoldClasses)

	// Clear old data
	if err := os.Remove(NafFileSecteurs); err != nil {
		log.Panicln(err)
	}
	if err := os.RemoveAll(NafFoldDivisions); err != nil {
		log.Panicln(err)
	}
	if err := os.RemoveAll(NafFoldGroupes); err != nil {
		log.Panicln(err)
	}
	if err := os.RemoveAll(NafFoldClasses); err != nil {
		log.Panicln(err)
	}

	// create new data folder
	if err := os.Mkdir(NafFoldDivisions, os.ModePerm); err != nil {
		log.Panicln(err)
	}
	if err := os.Mkdir(NafFoldGroupes, os.ModePerm); err != nil {
		log.Panicln(err)
	}
	if err := os.Mkdir(NafFoldClasses, os.ModePerm); err != nil {
		log.Panicln(err)
	}

	// Liste des codes
	// Pour eviter les doublons, peut servir a une futur static api
	var extracted NafRows

	// Nafa process
	// A executer en premier pour eviter les doublons fournis par la liste Naf
	if err := nafaExtractor(extracted); err != nil {
		log.Panicln(err)
	}

	// Naf process
	// La fiche naf donne aussi la structure des secteurs, divisions et groupes
	if err := nafExtractor(extracted); err != nil {
		log.Panicln(err)
	}
}

func nafExtractor(extracted NafRows) error {
	// ID des col xls a exploiter
	const (
		NafID = int(iota)
		NafCode
		NafName
	)

	// Telechargement de la nomenclature NAF
	file, err := requestData(SourceNaf)
	if err != nil {
		return err
	}

	// Traitement du XLS ... Merci l'INSEE
	xls, err := xlslib.OpenReader(bytes.NewReader(file), "")
	if err != nil {
		return err
	}

	sheet := xls.GetSheet(0)
	if sheet == nil {
		return errors.New("no sheet found in xls file")
	}

	// Enregistrement des data au format htm
	var out Naf

	// Fichier specifique au secteurs
	if err := out.Secteur.New(NafFileSecteurs, "", true); err != nil {
		return err
	}
	out.Secteur.File.WriteString("<option disabled selected>Sélectionnez votre Secteur</option>")

	var secteur string
	for i := 0; i <= (int(sheet.MaxRow)); i++ {
		row := sheet.Row(i)
		naf := NafRow{
			Code: row.Col(NafCode),
			Name: row.Col(NafName),
		}

		// Si le code est vide on skip pour eviter tout check
		if naf.Code == "" {
			continue
		}

		// Cas specifique des secteurs
		if strings.HasPrefix(naf.Code, "SECTION") {
			secteur = naf.Code[8:]
			_, err := out.Secteur.File.WriteString(
				SelectOption{
					Value: naf.Code[8:],
					NomID: "all",
					Text:  naf.Name,
					Type:  "secteur",
				}.String(),
			)

			if err != nil {
				return err
			}

			continue
		}

		// Verification du format du code naf
		if !naf.Normalize() {
			continue
		}

		// Ajout du code a son fichier
		line := SelectOption{Value: naf.Code, NomID: "all", Text: naf.Name}
		switch len(naf.Code) {
		// Divisions
		case 2:
			if new, err := out.Division.NeedNew(NafFoldDivisions, secteur); err != nil {
				return err
			} else if new {
				out.Division.File.WriteString("<option disabled selected>Sélectionnez votre Division</option>")
			}

			line.Type = "division"
			if _, err := out.Division.File.WriteString(line.String()); err != nil {
				return err
			}
		// Groupes
		case 3:
			if new, err := out.Groupe.NeedNew(NafFoldGroupes, naf.Code[:2]); err != nil {
				return err
			} else if new {
				out.Groupe.File.WriteString("<option disabled selected>Sélectionnez votre Groupe</option>")
			}

			line.Type = "groupe"
			if _, err := out.Groupe.File.WriteString(line.String()); err != nil {
				return err
			}
		// Classes
		case 5:
			// Ajout du code a la liste `extracted`
			if !extracted.NotExist(naf) {
				extracted.Add(naf)
			}

			if new, err := out.Classe.NeedNew(NafFoldClasses, naf.Code[:3]); err != nil {
				return err
			} else if new {
				out.Classe.File.WriteString("<option disabled selected>Sélectionnez votre Classe</option>")
			}

			line.Type = "classe"
			line.NomID = "naf"
			if _, err := out.Classe.File.WriteString(line.String()); err != nil {
				return err
			}
		default:
			continue
		}
	}

	// Fermer les fichiers ouverts
	out.Secteur.Close()
	out.Division.Close()
	out.Groupe.Close()
	out.Classe.Close()

	return nil
}

func nafaExtractor(extracted NafRows) error {
	// ID de la col du code NAFA
	const NafaCode = 1

	// Telechargement de la nomenclature NAFA
	file, err := requestData(SourceNafa)
	if err != nil {
		return err
	}

	// Traitement du csv
	csvr := csv.NewReader(bytes.NewReader(file))
	csvr.Comma = ';'

	// Enregistrement des data au format htm
	var out Naf
	for i := 0; ; i++ {
		row, err := csvr.Read()
		if err == io.EOF {
			break
		}

		// Skip la premiere ligne ou code invalide
		if i == 0 || len(row) < 3 || len(row[NafaCode]) < 8 {
			continue
		}

		naf := NafRow{
			Code: row[NafaCode][:6],
			Name: row[NafaCode][7:],
		}

		// Verification du format du code naf
		if !naf.Normalize() {
			continue
		}

		// Ajout du code a la liste `extracted`
		if !extracted.NotExist(naf) {
			extracted.Add(naf)
		}

		// Ajout du code a son fichier
		line := SelectOption{Value: naf.Code, Type: "classe", NomID: "nafa", Text: naf.Name}
		if _, err := out.Classe.NeedNew(NafFoldClasses, naf.Code[:3]); err != nil {
			return err
		}

		if _, err := out.Classe.File.WriteString(line.String()); err != nil {
			return err
		}
	}

	// Fermer le dernier fichier ouvert
	out.Classe.Close()

	return nil
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

func (o *OFile) New(p, n string, isFile ...bool) (err error) {
	// Si un fichier est deja ouvert on le ferme
	if o.File != nil {
		o.File.Close()
	}

	// Si le path n'est pas un fichier
	if len(isFile) == 0 {
		p = path.Join(p, n) + ".htm"
	}

	o.ID = n
	o.File, err = os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)

	return
}

func (o *OFile) NeedNew(p, id string) (new bool, err error) {
	if o.ID == id {
		return
	}

	return true, o.New(p, id)
}

func (o *OFile) Close() {
	if o.File == nil {
		return
	}

	o.File.Close()
}

func (n *NafRow) Normalize() bool {
	// Les codes NAF et NAFA sont compris entre 2 et 6 char
	if l := len(n.Code); l < 2 || l > 6 {
		return false
	}

	// Suprimer le point ou les espaces possibles
	n.Code = strings.Replace(n.Code, ".", "", 1)
	n.Code = strings.TrimSpace(n.Code)

	for i, r := range n.Code {
		// Les 4 premiers char sont des chiffres
		if i < 4 && unicode.IsDigit(r) {
			continue
		}

		// Char 5 et 6 doivent etre des lettres
		if i >= 4 && unicode.IsLetter(r) {
			continue
		}

		return false
	}

	return true
}

func (s SelectOption) String() string {
	return strings.Join(
		[]string{
			`<option value="`,
			s.Value,
			`" data-nomenclature="`,
			s.NomID,
			`" data-type="`,
			s.Type,
			`">`,
			strings.Title(strings.ToLower(s.Text)),
			`</option>`,
		},
		"",
	)
}

func (n *NafRows) NotExist(naf NafRow) bool {
	for _, nr := range *n {
		if naf.Code[:5] == nr.Code[:5] {
			return false
		}
	}

	return true
}

func (n *NafRows) Add(naf NafRow) {
	*n = append(*n, naf)
}
