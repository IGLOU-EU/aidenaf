#  [WIP] ⚖️ AideNaf - Aide à la Nomenclature d’activités française

[![License: GPL 3.0](https://img.shields.io/badge/Pull_request-Open-green.svg?style=flat-square)](https://www.gnu.org/licenses/gpl-3.0.html)
[![License: GPL 3.0](https://img.shields.io/badge/License-GPL_3.0_or_later-blue.svg?style=flat-square)](https://www.gnu.org/licenses/gpl-3.0.html)

Ce projet a pour but d’offrir à toutes personnes la possibilité de trouver son code d’activités française, sans avoir à comprendre ou maitriser les complexités des nomenclatures. Ainsi que de lui indiquer à quel organisme s’adresser pour ces formalités.    
Dans un premier temps, l’indication des organismes à qui s’adresser sera uniquement disponibles pour les auto-entrepreneurs.    

Pour soutenir ce projet Opensource et à destination du bien commun :
- Partager le site du projet
- Le faire savoir si vous trouvez une erreur ou un bug
- Participer au code et a la correction des textes
- Offrir un soutien financier et assurer la pérennité du projet

## Nomenclatures 
### Nomenclature d’activités française (NAF / APE) 
**Publication**: l’ANSII     
**Licence**: Licence Ouverte / Open Licence     
**Lien**: https://www.data.gouv.fr/fr/datasets/nomenclature-dactivites-francaise-naf/   
**Data**: https://www.insee.fr/fr/statistiques/fichier/2120875/int_courts_naf_rev_2.xls    

#### Format du code NAF
Un code NAF se compose de 4 chiffres, un point (optionnel) et une lettre.  
Exemple de représentation possible `10.39B`.    
- `10` Les deux premiers chiffres représentent la division
- `3`  Le 3eme chiffre représentent l’identifiant du groupe de la division
- `9`  Le 4eme chiffre est l’identifiant de la classe du groupe
- `B`  La lettre représente la déclinaison de la classe, sans déclinaison cette lettre est un `Z`

### Nomenclature d’Activités Française de l’Artisanat (NAFA / APRM)
**Publication**: Opendata hauts-de-seine     
**Licence**: Licence Ouverte / Open Licence version 2.0     
**Lien**: https://www.data.gouv.fr/fr/datasets/entreprises-artisanales-par-code-nafa/       
**Data**: https://opendata.hauts-de-seine.fr/explore/dataset/entreprises-artisanales-par-code-nafa/download/?format=csv&timezone=Europe/Berlin&lang=fr&use_labels_for_header=true&csv_separator=%3B       

#### Format du code NAFA
Le code NAFA se base sur la nomenclature du code NAF, il en reprend les divisions, les groupe et classe.    
Deux différences existent dans le fichier fourni par le publiant, mais ne gênant pas l’interprétation du code.
- En aucun cas il ne comporte de `.`
- Il se finit par deux lettres et non pas une

Exemple de représentation possible `1039AP`.    

### Professions libérales      
**Publication**: Ministère de l’Economie, de l’Industrie et du Numérique  
**Licence**: Licence Ouverte / Open Licence version 2.0      
**Lien**: https://www.data.gouv.fr/fr/datasets/les-professions-liberales/       
**Data**: https://www.data.gouv.fr/fr/datasets/r/c98687e6-0d80-476f-9ccc-d66a379c5c7e       

Ce cas est particulier, ne disposant pas de nomenclature ou de listing. Il est une construction manuelle, faite depuis une liste etalab du `Ministère de l'Economie, de l'Industrie et du Numérique`. Chaque profession libérale a un code NAF générique qui lui est attribué [Voir section NAF](#nomenclature-dactivités-française-naf-ape).   

## Structures du projet
Les fichiers html pres-rendu portent l’extension courte `htm` 
et ce afin de les différencier des fichiers conçus manuellement.
```
📦AideNaf
 ┣ 📂public/                (point d'entrée internet)
 ┃ ┣ 📂css/                 (feuilles de style)
 ┃ ┃ ┗📄bootstrap.min.css
 ┃ ┣ 📂data/                (Donnee formate pretes a l'usage)
 ┃ ┃ ┣ 📂classes/           (Classes des codes NAF/NAFA)
 ┃ ┃ ┃ ┗ 📄xxxx.htm
 ┃ ┃ ┣ 📂divisions/         (divisions NAF/NAFA)
 ┃ ┃ ┃ ┗ 📄x.htm
 ┃ ┃ ┣ 📂groupes/           (groupes NAF/NAFA)
 ┃ ┃ ┃ ┗ 📄xxx.htm
 ┃ ┃ ┗ 📄secteurs.htm       (secteurs NAF/NAFA)
 ┃ ┣ 📂js/                  (chargement des data et resultat)
 ┃ ┃ ┗ 📄manager.js
 ┃ ┣ 📂media/               (les divers images et icon)
 ┃ ┗ 📄index.html
 ┣ 📂tools/                 (les utilitaires de construction)
 ┃ ┗ 📂extractor/           (extraction et mise en forme des data)
 ┃   ┣ 📄go.mod
 ┃   ┣ 📄go.sum
 ┃   ┗ 📄main.go
 ┣ 📄LICENSE                (license License-GPL 3.0 or later)
 ┗ 📄README.md              (fonctionnement et definition du projet)
```

## LICENCE du projet
- Code `GPL-3.0-or-later`
- Formatage des datas `CC BY-SA`

## Remerciements
Un grand merci à [Mylène Portal](https://www.linkedin.com/in/mylene-p-906113196/), Employée du  Greffe du Tribunal de Commerce de Paris, qui a pris sur son temps libre pour être consultante bénévole sur ce projet.