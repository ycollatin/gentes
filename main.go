//   main.go --	Gentes
//	analyseur syntaxique du latin

package main

// FIXME
// - le surlignage des lemmatisations choisies ne marche plus
// - subiciunt veribus prunas et viscera torrent :
//   AmbiguÏté entre la coord prunas et viscera    (faux)
//				  et la coord subiciunt et torrent (juste)

// TODO
// - traiter la coordination par -que
// - traiter de la même manière le noyau et les subs, aussi bien dans le code
//   que dans les données ?
//   tenir compte de la morpho unique (voluptatem. acc. sing.)
// - un champ groupe.anrel - analyses du relatif ?
// - donner une POS distincte aux verbes intransitifs. v. gocol.indMorph
// - accord de personne sujet-verbe ?
// - saisie d'une phrase ?
// - fonction de sortie au format GraphViz
// - parasitage de /sum/ par /edo/ : comment supprimer "excl" dans lexsynt
// - parasitage de /do/ par /dato/ :   "

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/fatih/color"
	"github.com/ycollatin/gocol"
)

const (
	version = "Alpha"
	aidePh  = `    l->mot suivant ; h->mot précédent ;
    j->phrase suivante ; k->phrase précédente ;
    c->lemmatisation du mot courant ;
    a->arbre de la phrase ; r->retour; x->quitter`
	//s-> définir une suite morphosyntaxique ; x->Exit`
	//aideS =
	//`i-> id de la suite ; n-> n° du noyau ;
	// l-> liens (n°départ.fonction.n°sub,n°etc.);
	// v-> valider ; r-> retour`
)

var (
	ch, chData string // chemins du binaire et des données
	chCorpus   string // chemin du corpus
	dat        bool   // drapeau de chargement des données
	//module		string
	//modules		[]string
	rouge func(...interface{}) string
	texte *Texte
)

// affiche les arcs syntaxique de la phrase
func analyse(expl bool) {
	texte.phrase.teste()
	texte.affiche(aidePh)
	ar, _ := texte.phrase.arbre()
	gr := texte.phrase.src
	if expl {
		for _, n := range texte.phrase.nods {
			fmt.Println(n.doc())
		}
		fmt.Println("\n----- source ---\n",
			strings.Join(ar, "\n"),
			"\n----------------")
	}
	fmt.Println(strings.Join(gr, "\n"))
}

// choix du texte latin
func chxTexte() {
	ClearScreen()
	fmt.Println("Suites, grammaire latine")
	fmt.Println(" © Yves Ouvrard 2020, licence GPL3")
	texte = nil
	files, err := ioutil.ReadDir(ch + "/corpus/")
	if err != nil {
		fmt.Println("Répertoire", ch+"/corpus/", "introuvable")
		return
	}
	textes := []string{}
	for _, fileInfo := range files {
		textes = append(textes, fileInfo.Name())
	}
	nbf := len(files)
	chx := 1
	if nbf > 1 {
		for i := 0; i < len(files); i++ {
			fmt.Println(i+1, textes[i])
		}
		chx = InputInt("n° du texte")
	}
	if chx < 0 {
		main()
	}
	if chx > len(textes) {
		chx = len(textes)
	}
	ftexte := textes[chx-1]
	texte = CreeTexte(ftexte)
	texte.majPhrase()
	texte.affiche(aidePh)
}

func lemmatise() {
	texte.affiche(aidePh)
	texte.phrase.teste()
	fmt.Println("lemmatisation", rouge(texte.phrase.motCourant().gr))
	mc := texte.phrase.motCourant()
	if len(mc.ans2) > 0 {
		ll2 := gocol.Restostring(texte.phrase.motCourant().ans2)
		fmt.Println(rouge(ll2))
		ll3 := strings.Split(ll2, "\n")
		ll := strings.Split(gocol.Restostring(texte.phrase.motCourant().ans), "\n")
		for _, l := range ll {
			if !contient(ll3, l) {
				fmt.Println(l)
			}
		}
	} else {
		fmt.Println(gocol.Restostring(texte.phrase.motCourant().ans))
	}
}

func motprec() {
	if texte.phrase.imot > 0 {
		texte.phrase.imot--
		texte.affiche(aidePh)
	}
}

func motsuiv() {
	if texte.phrase.imot < len(texte.phrase.mots)-1 {
		texte.phrase.imot++
		texte.affiche(aidePh)
	}
}

func main() {
	// couleur
	rouge = color.New(color.FgRed, color.Bold).SprintFunc()
	// lecture des données Collatinus
	dir, _ := os.Executable()
	ch = path.Dir(dir)
	chData = ch + "/data/"
	chCorpus = ch + "/corpus/"
	gocol.Data(chData)
	// lecture des données syntaxiques
	lisGroupes(chData + "groupes.la")
	lisLexsynt()
	// choix du texte
	chxTexte()
	//texte.affiche(aidePh)
	for {
		k := GetKey()
		switch k {
		case "l":
			motsuiv()
		case "h":
			motprec()
		case "j":
			texte.porro()
		case "k":
			texte.retro()
		case "c":
			lemmatise()
		case "a":
			analyse(false)
		case "g":
			analyse(true)
		case "r":
			chxTexte()
		case "x":
			fmt.Println("\nVale")
			os.Exit(0)
		}
	}
	return
}
