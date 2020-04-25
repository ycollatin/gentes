//     Branche.go - Publicola

/*
Signets
sexplgrps
scopie
sexplore
snoeud
snoyau
srecolte
sresub

*/

package main

import (
	"fmt"
	"github.com/ycollatin/gocol"
	"strings"
)

// Chaque branche modifie trois propriétés attachées
// au mot. Leur liste est donc enregistrée dans
// Branche.photos.
type PhotoMot struct {
	mot     *Mot      // liaison avec le mot
	res     gocol.Res // lemmatisations réduites du mot
	dejasub bool      // appartenance du mot à un groupe
	pos     string    // nom du groupe dont le mot est noyau
}

type Branche struct {
	gr     string            // texte de la phrase
	imot   int               // rang du mot courant
	nbmots int               // nomb de mots de la phrase
	mots   []*Mot            // mots de la phrase XXX inutile ?
	nods   []*Nod            // noeuds validés
	niveau int               // n° de la branche par rapport au tronc
	veto   map[int][]*Groupe // index : rang du mot; valeur : liste des groupes interdits
	photos map[int]*PhotoMot // lemmatisations et appartenance de groupe propres à la branche
	mere   *Branche          // pointeur branche mère
	filles []*Branche        // liste des branches filles
}

func creeTronc(t string) *Branche {
	br := new(Branche)
	br.gr = t
	mm := gocol.Mots(t)
	for i, m := range mm {
		nm := creeMot(m)
		nm.rang = i
		br.mots = append(br.mots, nm)
	}
	br.nbmots = len(br.mots)
	br.photos = make(map[int]*PhotoMot)
	br.veto = make(map[int][]*Groupe)
	// peuplement des photos
	for _, m := range br.mots {
		phm := new(PhotoMot)
		phm.mot = m
		phm.res = m.ans
		phm.dejasub = false
		br.photos[m.rang] = phm
	}
	return br
}

func (b *Branche) copie() *Branche {
	// signet scopie
	nb := new(Branche)
	nb.gr = b.gr
	nb.nbmots = b.nbmots
	nb.mots = b.mots
	copy(nb.nods, b.nods)
	nb.mere = b
	nb.niveau = b.niveau + 1
	// les noeuds et les photos sont copiées
	// après création du noeud à l'origine de la copie
	nb.photos = make(map[int]*PhotoMot)
	nb.veto = b.veto
	return nb
}

func (b *Branche) dejasub(m *Mot) bool {
	return b.photos[m.rang].dejasub
}

func (b *Branche) domine(ma, mb *Mot) bool {
	//func (b *Branche) noyau(m *Mot) *Mot {
	mnoy := b.noyau(mb)
	for mnoy != nil {
		if mnoy == ma {
			return true
		}
		mnoy = b.noyau(mnoy)
	}
	return false
}

func (bm *Branche) explGrps(m *Mot, grps []*Groupe) {
	// signet sexplgrps
	for _, g := range grps {
		cont := false
		// Si le groupe a été exploré pour m dans une
		// autre branche, passer
		for _, gv := range bm.veto[m.rang] {
			if g == gv {
				cont = true
				break
			}
		}
		if cont {
			continue
		}
		n := bm.noeud(m, g)
		if n != nil {
			bf := bm.copie()
			for _, mph := range bm.mots {
				if n.inclut(mph) {
					if mph == n.nucl {
						// noyau
						ph := new(PhotoMot)
						ph.mot = mph
						ph.res = mph.restmp
						ph.pos = n.grp.id
						bf.photos[mph.rang] = ph
						// interdire le groupe au noyau
						bf.veto[mph.rang] = append(bm.veto[mph.rang], g)
					}
					for _, ma := range n.mma {
						// éléments ante
						ph := new(PhotoMot)
						ph.mot = ma
						ph.res = ma.restmp
						ph.dejasub = true
						ph.pos = bm.photos[ma.rang].pos
						bf.photos[ma.rang] = ph
					}
					for _, mp := range n.mmp {
						// éléments post
						ph := new(PhotoMot)
						ph.mot = mp
						ph.res = mp.restmp
						ph.dejasub = true
						ph.pos = bm.photos[mp.rang].pos
						bf.photos[mp.rang] = ph
					}
				} else {
					bf.photos[mph.rang] = bm.photos[mph.rang]
				}
			}
			//bm.nods = append(bm.nods, n)
			bf.nods = append(bf.nods, n)
			bm.nods = append(bm.nods, n)
			bm.filles = append(bm.filles, bf)
			bf.explore()
		}
	}
}

func (bm *Branche) explore() {
	// signet sexplore
	var nbf, nnbf int
	debut := true
	for debut || nnbf > nbf {
		nbf = len(bm.filles)
		for _, m := range bm.mots {
			bm.explGrps(m, grpTerm)
		}
		// 2. groupes non terminaux
		for _, m := range bm.mots {
			bm.explGrps(m, grp)
		}
		nnbf = len(bm.filles)
		debut = false
	}
}

// texte de la Branche, le mot courant surligné en rouge
func (b *Branche) enClair() string {
	var lm []string
	for i := 0; i < len(b.mots); i++ {
		m := b.mots[i].gr
		if i == b.imot {
			m = rouge(m)
		}
		lm = append(lm, m)
	}
	return strings.Join(lm, " ") + "."
}

// affiche la Branche en colorant n mots à partir
// du mot n° d
func (b *Branche) exr(d, n int) (e string) {
	var gab string = "%s"
	for i := 0; i < len(b.mots); i++ {
		if e != "" {
			gab = " %s"
		}
		if i >= d && i < d+n {
			e += fmt.Sprintf(gab, rouge(b.mots[i].gr))
		} else {
			e += fmt.Sprintf(gab, b.mots[i].gr)
		}
	}
	return
}

func (b *Branche) motCourant() *Mot {
	return b.mots[b.imot]
}

// si m peut être noyau d'un gourpe g, un Nod est renvoyé, sinon nil.
func (b *Branche) noeud(m *Mot, g *Groupe) *Nod {
	// signet snoeud

	// vérification de rang
	rang := m.rang
	lante := len(g.ante)
	// mot de rang trop faible
	if rang-lante < 0 {
		return nil
	}
	// ou trop élevé
	if rang+len(g.post)-1 >= b.nbmots {
		return nil
	}

	// m peut-il être noyau du groupe g ?
	m.restmp = cloneRes(m.ans)
	res := b.resNoyau(m, g, m.restmp)
	if res == nil {
		return nil
	}
	res = cloneRes(m.restmp)

	// création du noeud de retour
	nod := new(Nod)
	nod.grp = g
	nod.nucl = m
	nod.rang = rang

	// reгcherche rétrograde des subs ante
	r := rang - 1
	for ia := lante - 1; ia > -1; ia-- {
		if r < 0 {
			// le rang du mot est < 0 : impossible
			return nil
		}
		ma := b.mots[r]
		// passer les mots déjà subordonnés
		for b.dejasub(ma) {
			r--
			if r < 0 {
				return nil
			}
			ma = b.mots[r]
		}
		// vérification de réciprocité, puis du lien lui-même
		if b.domine(ma, m) {
			return nil
		}
		sub := g.ante[ia]
		res := b.resSub(ma, sub, m, ma.restmp)
		if res == nil {
			return nil
		}
		ma.restmp = cloneRes(res)
		nod.mma = append(nod.mma, ma)
		r--
	}

	// reгcherche des subs post
	for ip, sub := range g.post {
		r := rang + ip + 1
		if r >= b.nbmots {
			break
		}
		if sub.lien == "" {
			continue
		}
		mp := b.mots[r]
		for b.dejasub(mp) {
			r++
			if r >= b.nbmots {
				return nil
			}
			mpn := b.noyau(mp)
			if mpn != nil && mpn.rang < m.rang {
				return nil
			}
			mp = b.mots[r]
		}
		// réciprocité
		if b.domine(mp, m) {
			return nil
		}
		mp.restmp = cloneRes(mp.ans)
		res := b.resSub(mp, sub, m, mp.restmp)
		if res == nil {
			return nil
		}
		mp.restmp = cloneRes(res)
		nod.mmp = append(nod.mmp, mp)
		r++
	}

	if len(nod.mma)+len(nod.mmp) > 0 {
		return nod
	}
	return nil
}

func (b *Branche) noyau(m *Mot) *Mot {
	for _, n := range b.nods {
		for _, msub := range n.mma {
			if msub == m {
				return n.nucl
			}
		}
		for _, msub := range n.mmp {
			if msub == m {
				return n.nucl
			}
		}
	}
	return nil
}

// renvoie quelles lemmatisations de m lui permettent d'être le noyau du groupe g
func (b *Branche) resNoyau(m *Mot, g *Groupe, res gocol.Res) gocol.Res {
	// signet snoyau
	// valeurs variable de m pour la branche
	photom := b.photos[m.rang]
	// vérif du pos
	if photom.pos != "" {
		// 1. La pos définitif est fixée
		va := false
		for _, noy := range g.noyaux {
			if noy.canon > "" {
				for _, a := range res {
					va = va || noy.vaSr(a)
				}
			} else {
				va = va || noy.vaPos(photom.pos)
			}
		}
		if !va {
			return nil
		}
	} else {
		// Le mot est encore isolé
		var nres gocol.Res
		for _, a := range res {
			va := false
			for _, noy := range g.noyaux {
				if noy.canon > "" {
					va = va || noy.vaSr(a)
				} else {
					va = va || noy.vaPos(a.Lem.Pos)
				}
			}
			if va {
				nres = append(nres, a)
			}
		}
		if len(nres) == 0 {
			return nil
		}
		res = nres
	}

	// vérif lexicosyntaxique
	var nres gocol.Res
	for _, a := range res {
		va := true
		for _, ls := range g.lexsynt {
			va = va && lexsynt(a.Lem, ls)
		}
		if va {
			nres = append(nres, a)
		}
	}
	if len(nres) == 0 {
		return nil
	}
	res = nres

	// verif des exclusions lexicosyntaxiques
	nres = nil
	for _, a := range res {
		va := true
		for _, ls := range g.exclls {
			va = va && !lexsynt(a.Lem, ls)
		}
		if va {
			nres = append(nres, a)
		}
	}
	if len(nres) == 0 {
		return nil
	}
	res = nres

	// vérif morpho.
	// Si aucune n'est requise, renvoyer true
	if len(g.morph) == 0 {
		return res
	}

	nres = nil
	for _, sr := range res {
		var morfos []string // morphos de sr acceptées par g
		for _, morf := range sr.Morphos {
			if g.vaMorph(morf) {
				morfos = append(morfos, morf)
			}
		}
		if len(morfos) > 0 {
			sr.Morphos = morfos
			nres = append(nres, sr)
		}
	}
	// pour faire comme pour les autres vérifs :
	res = nres
	return res
}

// récolte tous les noeuds terminaux d'un arbre
func (b *Branche) recolte() (rec [][]*Nod) {
	// signet srecolte
	if len(b.filles) == 0 {
		rec = append(rec, b.nods)
		return rec
	}
	for _, f := range b.filles {
		rec = append(rec, b.nods)
		nrec := f.recolte()
		rec = append(rec, nrec...)
	}
	return rec
}

// vrai si m est compatible avec Sub et le noyau mn
func (b *Branche) resSub(m *Mot, sub *Sub, mn *Mot, res gocol.Res) (vares gocol.Res) {
	// signet sresub
	// si la fonction est déjà prise, renvoyer nil
	if !sub.multi && mn.adeja(sub) {
		return nil
	}

	// photo m et mn pour la branche
	photom := b.photos[m.rang]
	// vérification des pos
	// FIXME legatos decernis : avec v.obj, seul legatos pp est sélectionné par vaPos
	if photom.pos != "" {
		// 1. La pos du mot est définitive
		// noyaux exclus
		veto := false
		lgr := m.estNuclDe()
		for _, noy := range sub.noyexcl {
			veto = veto || contient(lgr, noy.id)
		}
		if veto {
			return nil
		}
		// noyaux possibles
		va := false
		for _, noy := range sub.noyaux {
			va = va || noy.vaPos(photom.pos)
		}
		if !va {
			return nil
		}
	} else {
		// 2. La pos définitif n'est pas encore fixée
		var nres gocol.Res
		// lexicosyntaxe
		for _, an := range res {
			va := true
			for _, ls := range sub.lexsynt {
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

		// canon et POS
		nres = nil
		for _, an := range res {
			va := false
			for _, noy := range sub.noyaux {
				if noy.canon > "" {
					va = va || noy.vaSr(an)
				} else {
					va = va || noy.vaPos(an.Lem.Pos)
				}
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

	//morphologie
	// si aucune morpho n'est requise, passer
	if len(sub.morpho) > 0 {
		var nres gocol.Res
		for _, an := range res {
			var lmorf []string
			for _, morfs := range an.Morphos {
				// pour toutes les morphos valides de m
				if strings.Contains(morfs, "inv.") || sub.vaMorpho(morfs) {
					lmorf = append(lmorf, morfs)
				}
			}
			if len(lmorf) > 0 {
				an.Morphos = lmorf
				nres = append(nres, an)
			}
		}
		if len(nres) == 0 {
			return nil
		}
		res = nres
	}

	// accord
	// pour chaque an.
	if sub.accord > "" {
		var nres gocol.Res
		for _, an := range res {
			va := false
			for _, anoy := range mn.restmp {
				// pour toutes les morphos valides de m
				var lmorf []string
				for _, morfn := range anoy.Morphos {
					for _, morfs := range an.Morphos {
						if accord(morfn, morfs, sub.accord) {
							lmorf = append(lmorf, morfs)
							va = true
						}
					}
				}
				if len(lmorf) > 0 {
					an.Morphos = lmorf
				}
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
	return res
}
