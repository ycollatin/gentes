//   main.go --	Gentes
//	analyseur syntaxique du latin

package main

// FIXME
// - Un groupe incomplet semble pouvoir être adopté
// - Iovis ut omne genus ... index out of range [1] with length 1

//
// TODO
// - Permettre l'omission du pos d'un sub dans groupes.la
// - La lemmatisation ne peut pas toujours être fixée.
// 	 par l'adoption d'un noeud. Cette adoption
//	 peut soit éliminer quelques lemmatisations,
//	 soit n'en éliminer aucune.
// - filtre lexsynt pour les subs
// - Le code graphviz devrait donner l'équivalence mot - rang

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
	aidePh =
	`l->mot suivant ; h->mot précédent ;
	j->phrase suivante ; k->phrase précédente ;
	c->lemmatisation du mot courant ;
	a->arbre de la phrase ; x->quitter`
	//s-> définir une suite morphosyntaxique ; x->Exit`
	//aideS =
	//`i-> id de la suite ; n-> n° du noyau ; 
	// l-> liens (n°départ.fonction.n°sub,n°etc.);
	// v-> valider ; r-> retour`
)

var (
	ch, chData	string	// chemins du binaire et des données
	chCorpus	string	// chemin du corpus
	dat			bool	// drapeau de chargement des données
	//module		string
	//modules		[]string
	rouge		func(...interface{}) string
	texte		*Texte
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
		strings.Join(ar,"\n"),
		"\n----------------")
	}
	fmt.Println(strings.Join(gr, "\n"))
}

// choix du texte latin
func chxTexte() {
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
		for i:= 0; i < len(files); i++ {
			fmt.Println(i+1, textes[i])
		}
		chx = InputInt("n° du texte")
	}
	ftexte := textes[chx-1]
	texte = CreeTexte(ftexte)
	texte.majPhrase()
}

func lemmatise() {
	texte.affiche(aidePh)
	texte.phrase.teste()
	fmt.Println("lemmatisation", rouge(texte.phrase.motCourant().gr))
	fmt.Println(gocol.Restostring(texte.phrase.motCourant().ans))
}

func motprec() {
	if texte == nil {
		txtNil()
		return
	}
	if texte.phrase.imot > 0 {
		texte.phrase.imot--
		texte.affiche(aidePh)
	}
}

func motsuiv() {
	if texte == nil {
		txtNil()
		return
	}
	if texte.phrase.imot < len(texte.phrase.mots)-1 {
		texte.phrase.imot++
		texte.affiche(aidePh)
	}
}

func txtNil() {
	fmt.Println("Il faut d'abord charger un texte : commande txt")
}

func main() {
	ClearScreen()
    fmt.Println("Suites, grammaire latine")
    fmt.Println("Yves Ouvrard, GPL3")
	// couleur
	rouge = color.New(color.FgRed, color.Bold).SprintFunc()
	// lecture des données Collatinus
	dir, _ := os.Executable()
	ch = path.Dir(dir)
	chData = ch + "/data/"
	chCorpus = ch + "/corpus/"
	go gocol.Data(chData)
	// lecture des données syntaxiques
	lisGroupes(chData+"groupes.la")
	lisLexsynt()
	// choix du texte
	chxTexte()
	texte.affiche(aidePh)
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
		case "x":
			fmt.Println("\nVale")
			os.Exit(0)
		}
	}
	return
}
