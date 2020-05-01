// groupe.go - Gentes

package main

// signets
// vaMorph

import (
	"github.com/ycollatin/gocol"
	"strings"
)

type Groupe struct {
	id              string
	noyaux, noyexcl []*Noy   // pos du noyau
	morph           []string // traits morpho du noyau
	lexsynt         []string // étiquettes lexicosyntaxiques du noyau
	ante            []*Sub   // éléments précédant le noyau
	post            []*Sub   // éléments suivant le noyau
	multi           bool     // le groupe possède au moins un sub multi
	nbsubs          int      // nombre de subs
}

var grpTerm, grp []*Groupe

// créateur de Groupe
func creeGroupe(ll []string) *Groupe {
	if len(ll) == 0 {
		return nil
	}
	g := new(Groupe)
	for _, l := range ll {
		kv := strings.Split(l, ":")
		k := kv[0]
		v := kv[1]
		switch k {
		case "ter", "grp":
			g.id = v
		case "pos":
			g.noyaux, g.noyexcl = creeNoy(v)
		case "morph":
			g.morph = strings.Split(v, ",")
		case "lexsynt":
			lecl := strings.Split(v, " ")
			for _, ecl := range lecl {
				g.lexsynt = append(g.lexsynt, ecl)
			}
		case "a":
			g.ante = append(g.ante, creeSub(v, g, true))
		case "ag":
			g.ante = append(g.ante, creeSub(v, g, false))
		case "p":
			g.post = append(g.post, creeSub(v, g, true))
		case "pg":
			g.post = append(g.post, creeSub(v, g, false))
		}
	}
	// multi
	for _, sa := range g.ante {
		if sa.multi {
			g.multi = true
		}
	}
	for _, sp := range g.post {
		if sp.multi {
			g.multi = true
		}
	}
	g.nbsubs = len(g.ante) + len(g.post)
	return g
}

// vrai si l'élément de pos ou id id est exclu par le groupe
func (g *Groupe) estExclu(id string) bool {
	for _, ne := range g.noyexcl {
		if ne.id == id {
			return true
		}
	}
	return false
}

// lecture des groupes dans le fichier nf
func lisGroupes(nf string) {
	llin := gocol.Lignes(nf)
	var ll []string
	for _, l := range llin {
		deb := l[:4]
		if deb == "ter:" || deb == "grp:" {
			g := creeGroupe(ll)
			if g != nil {
				if ll[0][:4] == "grp:" {
					grp = append(grp, g)
				} else {
					grpTerm = append(grpTerm, g)
				}
				ll = nil
			}
		}
		ll = append(ll, l)
	}
	grp = append(grp, creeGroupe(ll))
}

// le pos pos est-il compatible avec le noyau du groupe g ?
func (g *Groupe) vaPos(pos string) bool {
	for _, ne := range g.noyexcl {
		if ne.vaPos(pos) {
			return false
		}
	}
	for _, noy := range g.noyaux {
		if noy.vaPos(pos) {
			return true
		}
	}
	return false
}

// la morpho morf est-elle compatible avec le noyau du groupe g ?
func (g *Groupe) vaMorph(morf string) bool {
	for _, gmorf := range g.morph {
		va := true
		ecl := strings.Split(gmorf, " ")
		for _, trait := range ecl {
			va = va && strings.Contains(morf, trait)
		}
		if va {
			return true
		}
	}
	return false
}
