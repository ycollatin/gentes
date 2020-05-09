//     Branche.go - Gentes

/*
Signets
exploregrou
scopie
sexplore
snoeud
srecolte
sresEl
*/

package main

import (
	"fmt"
	"github.com/ycollatin/gocol"
	"sort"
	"strings"
)

var (
	mots    []*Mot
	nbmots  int
	journal []string
)

type Branche struct {
	gr     string            // texte de la phrase
	imot   int               // rang du mot courant
	nods   []Nod             // noeuds validés
	niveau int               // n° de la branche par rapport au tronc
	veto   map[int][]Nod     // index : rang du mot; valeur : liste des liens interdits
	photos map[int]gocol.Res // lemmatisations et appartenance de groupe propres à la branche
	filles []*Branche        // liste des branches filles
}

// le tronc est la branche de départ. Il
// est initialisé avec les mots lemmatisés
// par Collatinus. var tronc est déclarée dans texte.go
func creeTronc(t string) *Branche {
	mots = nil
	br := new(Branche)
	br.gr = t
	mm := gocol.Mots(t)
	for i, m := range mm {
		nm := creeMot(m)
		nm.rang = i
		mots = append(mots, nm)
	}
	nbmots = len(mots)
	br.photos = make(map[int]gocol.Res) // l'index de la map est le numéro des mots
	br.veto = make(map[int][]Nod)
	// peuplement des photos
	for _, m := range mots {
		phm := m.ans
		br.photos[m.rang] = phm
	}
	journal = nil
	return br
}

func (b *Branche) adeja(noyau *Mot, lien string) bool {
	for _, nod := range b.nods {
		if nod.nucl == noyau {
			for i, _ := range nod.mma {
				if nod.groupe.ante[i].lien == lien {
					return true
				}
			}
			for i, _ := range nod.mmp {
				if nod.groupe.post[i].lien == lien {
					return true
				}
			}
		}
	}
	return false
}

// copie branche mère  - branche fille
func (b *Branche) copie() *Branche {
	// signet scopie
	nb := new(Branche)
	nb.gr = b.gr
	nb.niveau = b.niveau + 1
	//nb.nods = b.nods
	for _, nbm := range b.nods {
		nb.nods = append(nb.nods, nbm.copie())
	}
	nb.photos = make(map[int]gocol.Res)
	for i, r := range b.photos {
		for _, sr := range r {
			var nr gocol.Sr
			nr.Lem = sr.Lem
			for _, morf := range sr.Morphos {
				nr.Morphos = append(nr.Morphos, morf)
			}
			nb.photos[i] = append(nb.photos[i], nr)
		}
	}
	nb.veto = make(map[int][]Nod)
	for i, v := range b.veto {
		for _, nod := range v {
			nb.veto[i] = append(nb.veto[i], nod.copie())
		}
	}
	return nb
}

// Faux si m є à un groupe
func (b *Branche) dejasub(m *Mot) bool {
	for _, n := range b.nods {
		for _, ma := range n.mma {
			if ma == m {
				return true
			}
		}
		for _, mp := range n.mmp {
			if mp == m {
				return true
			}
		}
	}
	return false
}

// mb є au groupe ma, directement ou indirectement
func (b *Branche) domine(ma, mb *Mot) bool {
	mnoy := b.noyau(mb)
	for mnoy != nil {
		if mnoy == ma {
			return true
		}
		mnoy = b.noyau(mnoy)
	}
	return false
}

// texte de la Branche, le mot courant surligné en rouge
func (b *Branche) enClair() string {
	var lm []string
	for i := 0; i < len(mots); i++ {
		m := mots[i].gr
		if i == b.imot {
			m = rouge(m)
		}
		lm = append(lm, m)
	}
	return strings.Join(lm, " ") + "."
}

// explore toutes les possibilités d'une branche
func (b *Branche) explore() {
	// signet sexplore
	for _, m := range mots {
		// 1. groupes terminaux
		b.exploreGroupes(m, grpTerm)
		// 2. groupes non terminaux
		b.exploreGroupes(m, grp)
	}
}

// essaye toutes les règles de groupes de grps où m pourrait
// être noyau
func (bm *Branche) exploreGroupes(m *Mot, grps []*Groupe) {
	// signet exploregrou
	for _, g := range grps {
		// tester la possibilité de création noeud de type g
		// dont le noyau est m
		n := bm.noeud(m, g)
		if n.valide {
			// Si le groupe a été exploré pour m dans une
			// autre branche, passer
			va := true
			for _, veto := range bm.veto[m.rang] {
				va = va && !n.egale(veto)
				if !va {
					break
				}
			}
			if !va {
				continue
			}
			//créer une branche fille (bf)
			// copiée de la mère (bm)
			// où le noeud sera obligatoire. Dans la branche mère,
			// il sera interdit.
			bf := bm.copie()
			bf.nods = append(bf.nods, n)
			for _, mph := range mots {
				if mph == m {
					// photo du noyau
					bf.photos[mph.rang] = m.restmp
					n.rnucl = mph.restmp
				}
				for _, ma := range n.mma {
					// photos des éléments antéposés
					if mph == ma {
						bf.photos[ma.rang] = ma.restmp
						n.rra[ma.rang] = mph.restmp
					}
				}
				for _, mp := range n.mmp {
					// photos des éléments postposés
					if mph == mp {
						bf.photos[mp.rang] = mp.restmp
						n.rrp[mp.rang] = mph.restmp
					}
				}
			}
			// màj journal
			indent := strings.Repeat("  ", bm.niveau)
			journal = append(journal, fmt.Sprintf("%s %d. %s", indent, bm.niveau, n.doc()))
			// la fille est explorée récursivement
			journal = append(journal, fmt.Sprintf("%s %d montée au niveau %d", indent, bm.niveau, bf.niveau))
			bf.explore()
			if len(bf.filles) == 0 {
				journal = append(journal, fmt.Sprintf("%s   %d branche terminale, %d arc(s)",
					indent, bf.niveau, len(bf.nods)))
			}
			journal = append(journal, fmt.Sprintf("%s retour niveau %d. niveau %d, %d fille(s)",
				indent, bm.niveau, bf.niveau, len(bf.filles)))
			// ajout à la liste des filles
			bm.filles = append(bm.filles, bf)
			// le noeud est désormais interdit chez
			// la mère et ses futures filles
			bm.veto[m.rang] = append(bm.veto[m.rang], n)
			// réinitialiser les photos
			for _, mot := range mots {
				mot.restmp = bm.photos[mot.rang]
			}
		}
	}
}

// affiche la Branche en colorant n mots à partir
// du mot n° d
func (b *Branche) exr(d, n int) (e string) {
	var gab string = "%s"
	for i := 0; i < len(mots); i++ {
		if e != "" {
			gab = " %s"
		}
		if i >= d && i < d+n {
			e += fmt.Sprintf(gab, rouge(mots[i].gr))
		} else {
			e += fmt.Sprintf(gab, mots[i].gr)
		}
	}
	return
}

// id des groupes dont m est noyau
func (b *Branche) ids(m *Mot) (lids []string) {
	for _, nod := range b.nods {
		if nod.nucl == m {
			lids = append(lids, nod.groupe.id)
		}
	}
	return
}

func (b *Branche) motCourant() *Mot {
	return mots[b.imot]
}

// si m peut être noyau d'un gourpe g, un Nod est renvoyé, sinon nil.
func (b *Branche) noeud(m *Mot, g *Groupe) Nod {
	// snoeud, signet

	// noeud nnul pour le retour d'échec
	var nnul Nod

	// vérification de rang
	rang := m.rang
	lante := len(g.ante)
	// mot de rang trop faible
	if rang-lante < 0 {
		return nnul
	}
	// ou trop élevé
	if rang+len(g.post)-1 >= nbmots {
		return nnul
	}

	// validation du noyau
	m.restmp = b.photos[m.rang]
	m.restmp = b.resEl(m, g.nucl, m, m.restmp)
	if m.restmp == nil {
		return nnul
	}

	// création du noeud de retour
	var nod Nod
	nod.rra = make(map[int]gocol.Res)
	nod.rrp = make(map[int]gocol.Res)
	nod.groupe = g
	nod.nucl = m
	nod.rang = rang

	// reгcherche rétrograde des subs ante
	r := rang - 1
	for ia := lante - 1; ia > -1; ia-- {
		if r < 0 {
			// le rang du mot est < 0 : impossible
			return nnul
		}
		ma := mots[r]
		// passer les mots déjà subordonnés | b branche.go:336 cond 2 m.gr=="rapuit" && g.id=="v.conjet"
		for b.dejasub(ma) {
			r--
			if r < 0 {
				return nnul
			}
			ma = mots[r]
		}
		// vérification de réciprocité, puis du lien lui-même
		if b.domine(ma, m) && g.ante[ia].lien > "" {
			return nnul
		}
		sub := g.ante[ia]
		// réinitialisation des lemmatisations de test
		ma.restmp = b.photos[ma.rang]
		// validation du noyau
		ma.restmp = b.resEl(ma, sub, m, ma.restmp)
		if ma.restmp == nil {
			return nnul
		}
		// si le lien est muet, c'est qu'il est étranger au
		// groupe. Il y a hyperbate.
		if sub.lien > "" {
			nod.mma = append(nod.mma, ma)
		}
		r--
	}

	// reгcherche des subs post
	r = rang + 1
	for ip, sub := range g.post {
		if r >= nbmots {
			break
		}
		mp := mots[r]
		for b.dejasub(mp) {
			r++
			if r >= nbmots {
				return nnul
			}
			mpn := b.noyau(mp)
			if mpn != nil && mpn.rang < m.rang {
				return nnul
			}
			mp = mots[r]
		}
		// réciprocité
		if b.domine(mp, m) {
			return nnul
		}
		mp.restmp = b.photos[mp.rang]
		mp.restmp = b.resEl(mp, sub, m, mp.restmp)
		if mp.restmp == nil {
			return nnul
		}
		// cf. supra nod.mma
		if g.post[ip].lien > "" {
			nod.mmp = append(nod.mmp, mp)
		}
		r++

	}

	if len(nod.mma)+len(nod.mmp) > 0 {
		nod.valide = true
		return nod
	}
	return nnul
}

func (b *Branche) noyau(m *Mot) *Mot {
	for _, n := range b.nods {
		for i, msub := range n.mma {
			if n.groupe.ante[i].lien > "" && msub == m {
				return n.nucl
			}
		}
		for i, msub := range n.mmp {
			if n.groupe.post[i].lien > "" && msub == m {
				return n.nucl
			}
		}
	}
	return nil
}

// récolte tous les noeuds terminaux d'un arbre
func (b *Branche) recolte() (rec [][]Nod) {
	// signet srecolte
	if b.terminale() {
		rec = append(rec, b.nods)
		return rec
	}
	for _, f := range b.filles {
		nrec := f.recolte()
		rec = append(rec, nrec...)
	}
	sort.Slice(rec, func(i, j int) bool {
		return len(rec[i]) > len(rec[j])
	})
	return rec
}

// vrai si m est compatible avec Sub et le noyau mn
func (b *Branche) resEl(m *Mot, el *El, mn *Mot, res gocol.Res) gocol.Res {
	// signet sresEl

	// contraintes de groupe
	if !el.multi && b.adeja(mn, el.lien) {
		return nil
	}

	// vérification du pos : id du noyau, ou pos du mot
	ids := b.ids(m)
	var va bool
	if len(ids) > 0 {
		for _, id := range ids {
			// familles
			pel := PrimEl(id, ".")
			if len(el.famexcl) > 0 && contient(el.famexcl, pel) {
				return nil
			}
			if len(el.idsexcl) > 0 && contient(el.idsexcl, id) {
				return nil
			}
		}
		if len(el.familles) > 0 {
			var vafam bool
			for _, id := range ids {
				for _, elf := range el.familles {
					if elf == PrimEl(id, ".") {
						vafam = true
					}
				}
			}
			if !vafam {
				return nil
			}
		}
		// id des groupes dont m est noyau
		if len(el.ids) > 0 {
			var vaids bool
			for _, id := range ids {
				for _, idel := range el.ids {
					if idel == id {
						vaids = true
					}
				}
			}
			if !vaids {
				return nil
			}
		}
		va = true
	}
	// si l'élément n'a aucune des propriétés requises pour un mot isolé,
	if !va && len(el.poss)+len(el.cles)+len(el.morpho)+len(el.lexsynt) == 0 {
		// il ne peut appartenir au groupe
		return nil
	}
	var nres gocol.Res

	// contraintes de lemmatisation
	// lexicosyntaxe, exclus
	if len(el.lsexcl) > 0 {
		nres = nil
		for _, excl := range el.lsexcl {
			for _, an := range res {
				if !lexsynt(an.Lem, excl) {
					nres = append(nres, an)
				}
			}
		}
		res = nres
	}

	// lexicosyntaxe, possibles
	if len(el.lexsynt) > 0 {
		nres = nil
		for _, an := range res {
			va := true
			for _, ls := range el.lexsynt {
				va = va && lexsynt(an.Lem, ls)
			}
			if va {
				nres = append(nres, an)
			}
		}
		if len(nres) == 0 {
			return nil
		}
		res = nres
	}

	// canons exclus et possibles
	if len(el.clesexcl) > 0 {
		nres = nil
		for _, an := range res {
			if contient(el.clesexcl, an.Lem.Cle) {
				continue
			}
			nres = append(nres, an)
		}
		if len(nres) == 0 {
			return nil
		}
		res = nres
	}

	if len(el.cles) > 0 {
		nres = nil
		for _, an := range res {
			for _, cle := range el.cles {
				if an.Lem.Cle == cle {
					nres = append(nres, an)
				}
			}
		}
		if len(nres) == 0 {
			return nil
		}
		res = nres
	}

	// pos
	if len(el.poss) > 0 {
		nres = nil
		for _, an := range res {
			if contient(el.posexcl, an.Lem.Pos) {
				continue
			}
			if contient(el.poss, an.Lem.Pos) {
				nres = append(nres, an)
			}
		}
		if len(nres) == 0 {
			return nil
		}
		res = nres
	}

	//morphologie
	if len(el.morpho) > 0 {
		var nres gocol.Res
		for _, an := range res {
			for _, morfs := range an.Morphos {
				// pour toutes les morphos valides de m
				if el.vaMorpho(morfs) {
					nres = gocol.AddRes(nres, an.Lem, morfs, 0)
				}
			}
		}
		if len(nres) == 0 {
			return nil
		}
		res = nres
	}

	// accord
	if el.accord > "" {
		var nres gocol.Res
		for _, an := range res {
			for _, anoy := range mn.restmp {
				for _, morfs := range an.Morphos {
					for _, morfn := range anoy.Morphos {
						if accord(morfn, morfs, el.accord) {
							//nres = append(nres, an)
							nres = gocol.AddRes(nres, an.Lem, morfs, 0)
						}
					}
				}
			}
		}
		if len(nres) == 0 {
			return nil
		}
		res = nres
	}
	//res = nres
	return res
}

// Une branche est terminale si elle n'a pas de filles
// la récolte récupère alors tous les liens des noeuds
// dont elle a hérité.
func (b *Branche) terminale() bool {
	return len(b.filles) == 0
}
