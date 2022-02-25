// Recuperation des DOM
// Secteurs
var domSecteurs = document.getElementById('data_secteurs').children[1];
// Divisions
var domDivisions = document.getElementById('data_divisions').children[1];
// Groupes
var domGroupes = document.getElementById('data_groupes').children[1];
// Classes
var domClasses = document.getElementById('data_classes').children[1];

// Listes des URL's
var secteursFile = '/data/secteurs.htm';
var divisionsFolder = '/data/divisions/';
var groupesFolder = '/data/groupes/';
var classesFolder = '/data/classes/';

// Mise a jours des selects
function updateSelect(s) {
    var option = s.options[s.selectedIndex];

    switch (option.dataset.type) {
        case 'secteur':
            disableSelect(false, false, true, true)
            setElemFromUrl(divisionsFolder + option.value + '.htm', domDivisions);
            break;

        case 'division':
            disableSelect(false, false, false, true)
            setElemFromUrl(groupesFolder + option.value + '.htm', domGroupes);
            break;

        case 'groupe':
            disableSelect(false, false, false, false)
            setElemFromUrl(classesFolder + option.value + '.htm', domClasses);
            break;

        default:
            break;
    }
}

// Affichage du resultat pour la classe
function codeResult(e) {
    var id = e.id;
    var value = e.value;
    var text = e.options[e.selectedIndex].text;

    if ('' !== value) {
        alert('Votre code NAF/NAFA ' + text + ' (' + value + ')');
    }
}

// Activarion et desactivation des select
function disableSelect(sec, div, grp, cls) {
    disable(domSecteurs, sec)
    disable(domDivisions, div)
    disable(domGroupes, grp)
    disable(domClasses, cls)
}

function disable(e, b) {
    var dis = e.hasAttribute('disabled')

    if (!dis && b) {
        e.setAttribute('disabled', '')
    } else if (dis && !b) {
        e.removeAttribute('disabled');
    }
}

// Ajout du chargement
function loadingOn(e, b) {
    console.log(e.classList)
    if (b) {
        e.classList.add('loading');
    } else {
        e.classList.remove('loading');
    }
}

// Telechargement d'un set et assignation a un select
function setElemFromUrl(url, elem) {
    var parent = elem.parentNode

    loadingOn(parent, true)
    fetch(url)
        .then(function(response) {
            if (!response.ok) {
                return `<div class="alert alert-danger"><strong>Error!</strong> Une erreur est survenue, veuillez réessayer plus tard.</div>`
            }
            return response.text();
        })
        .then(function(data) {
            elem.innerHTML = data;
            loadingOn(parent, false)
        });
}

// Attente du dom
document.addEventListener('DOMContentLoaded', function() {
    // Chargement du premier set
    setElemFromUrl('/data/secteurs.htm', domSecteurs);
    disableSelect(false, true, true, true)
});