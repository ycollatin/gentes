//       mot.go - Gentes

// signets :
//
// motnoeud
// motresnoyau
// motestNoyauDeGroupe
// motresSub

// rappel de la lemmatisation dans gocol :
// type Sr struct {
//	Lem     *Lemme
//	Nmorph  []int
//	Morphos []string
// }
//
// type Res []Sr

package main

import (
	//"fmt"
	"github.com/ycollatin/gocol"
	"strings"
)

type Lm struct {
	l	*gocol.Lemme
	m	string
}

type Mot struct {
	gr			string		// graphie du mot
	rang		int			// rang du mot dans la phrase à partir de 0
	ans, ans2	gocol.Res	// ensemble des lemmatisations, ans provisoire
	dejasub		bool		// le mot est déjà l'élément d'n nœud
	llm			[]Lm		// liste des lemmes ٍ+ morpho possibles
	tmpl, tmpm	int			// n°s provisoires de Sr et morpho
	pos			string		// id du groupe dont le mot est noyau
							// ou à défaut pos du mot, si elle est décidée
	lexsynt		[]string	// propriétés lexicosyntaxiques
}

func creeMot(m string) *Mot {
	mot := new(Mot)
	mot.gr = m
	var echec bool
	mot.ans, echec = gocol.Lemmatise(m)
	if echec {
		mot.ans, echec = gocol.Lemmatise(gocol.Majminmaj(m))
	}
	// ajout du genre pour les noms
	if !echec {
		for i, a := range mot.ans {
			mot.ans[i] = genus(a)
		}
	}
	return mot
}

/*
// TODO À commenter avant déploiement
func (m *Mot) doc() string {
	return fmt.Sprintf("%d. %s %s\n---\n%s", m.rang, m.gr,
	gocol.Restostring(m.ans), gocol.Restostring(m.ans2))
}
*/

func accord(lma, lmb, cgn string) bool {
va := true
for i:=0; i<len(cgn); i++ {
	switch cgn[i] {
	case 'c':
		k := cas(lma)
		va = va && strings.Contains(lmb, k)
	case 'g':
		g := genre(lma)
		va = va && strings.Contains(lmb, g)
	case 'n':
		n := nombre(lma)
			va = va && strings.Contains(lmb, n)
		}
	}
	return va
}

func (m *Mot) dejaNoy() bool {
	for _, n := range texte.phrase.nods {
		if n.nucl == m {
			return true
		}
	}
	return false
}

func (ma *Mot) domine(mb *Mot) bool {
	//mnoy := noyDe(mb)
	mnoy := mb.noyau()
	for mnoy != nil {
		if mnoy == ma {
			return true
		}
		mnoy = mnoy.noyau()
		//noyDe(mnoy)
	}
	return false
}

// id des Nod dont m est déjà le noyau
func (m *Mot) estNuclDe() []string {
	var ret []string
	for _, nod := range texte.phrase.nods {
		if nod.nucl == m {
			ret = append(ret, nod.grp.id)
		}
	}
	return ret
}

// vrai si m est compatible avec Sub et le noyau mn
func (m *Mot) resSub(sub *Sub, mn *Mot) gocol.Res {
	//debog := sub.groupe.id=="v.cui" && m.gr == "cui" && mn.gr=="dedit"
	//if debog {fmt.Println(" -resSub m",m.gr,"pos",m.pos,"sub",sub.groupe.id,"mn",mn.gr,"pos",m.pos)}
	// signet motresSub
	var ans2 gocol.Res
	// vérification des pos
	if m.pos != "" {
		// 1. La pos du mot est définitive
		// noyaux exclus
		veto := false
		lgr := m.estNuclDe()
		for _, noy := range sub.noyexcl {
			//if debog {fmt.Println("   .resSub, excl",noy.id,"pos",m.pos)}
			veto = veto || contient(lgr, noy.id)
		}
		//if debog {fmt.Println("   .resSub, !=, excl",len(sub.noyexcl),"veto",veto)}
		if veto {
			return ans2
		}
		// noyaux possibles
		va := false
		for _, noy := range sub.noyaux {
			va = va || noy.vaPos(m.pos)
			//if debog {fmt.Println("   .resSub, noy",noy.id,"pos",m.pos,"va",va)}
		}
		if !va {
			//if debog {fmt.Println("   .resSub, vapa!")}
			return ans2
		}
		ans2 = m.ans2
		//if debog {fmt.Println("   .resSub, noy, len ans2",len(ans2))} //,ans2[0].Lem.Gr)}
	} else {
		//if debog {fmt.Println("   .resSub, else")}
		// 2. La pos définitif n'est pas encore fixée
		for _, an := range m.ans {
			//if debog {fmt.Println("   .resSub,lemme",an.Lem.Gr,"pos",an.Lem.Pos)}
			for _, noy := range sub.noyaux {
				if noy.canon > "" {
					//if debog {fmt.Println("   .resSub, noy.canon",noy.canon,"sr.Lem",an.Lem.Gr[0])}
					if noy.vaSr(an) {
						ans2 = append(ans2, an)
						break
					}
				} else {
					//if debog {fmt.Println("   .resSub, noy.vaPos",an.Lem.Pos)}
					if noy.vaPos(an.Lem.Pos) {
						//if debog {fmt.Println("   .resSub, vaPos",an.Lem.Gr,an.Lem.Pos)}
						ans2 = append(ans2, an)
						break
					}
				}
			}
		}
	}
	//if debog {fmt.Println("   .resSub, lien",sub.lien,"len(ans2)",len(ans2))} //,"lemme",ans2[0].Lem.Gr)}
	if len(ans2) == 0 {
		return ans2
	}

	//morphologie
	var ans3 gocol.Res
	for _, an := range ans2 {
		//if debog {fmt.Println("   .resSub2",an.Lem.Gr,"morphos",len(an.Morphos),"sub.morpho",len(sub.morpho))}
		var lmorf []string
		for _, morfs := range an.Morphos {
			//if debog {fmt.Println("   .resSub3 morfs",morfs,"sub",sub.morpho,sub.lien)}
			// pour toutes les morphos valides de m
			if strings.Contains(morfs, "inv.") || sub.vaMorpho(morfs) {
				lmorf = append(lmorf, morfs)
			}
		}
		if len(lmorf) > 0 {
			an.Morphos = lmorf
			ans3 = append(ans3, an)
		}
	}
	//if debog {fmt.Println("   .resSub1, oka, len ans3",len(ans3))}
	// accord
	var ans4 gocol.Res
	// pour toutes les morphos valides de mn
	for _, ann := range mn.ans2 {
		for _, morfn := range ann.Morphos {
			// pour toutes les morphos valides de m
			for _, an := range ans3 {
				var lmorf []string
				for _, morfs := range an.Morphos {
					if accord(morfn, morfs, sub.accord) {
						lmorf = append(lmorf, morfs)
					}
				}
				if len(lmorf) > 0 {
					an.Morphos = lmorf
					ans4 = append(ans4, an)
				}
			}
		}
	}
	//if debog {fmt.Println("   .resSub3, sortie ans3",len(ans3))}
	return ans4
}

// ajoute le genre à la morpho d'un nom
func genus(sr gocol.Sr) gocol.Sr {
	if sr.Lem.Pos != "n" && sr.Lem.Pos != "NP" {
		return sr
	}
	inc := 12
	switch sr.Lem.Genre {
	case "féminin":
		inc += 12
	case "neutre":
		inc += 24
	}
	for i, _ := range sr.Nmorph {
		sr.Nmorph[i] += inc
	}
	return sr
}

// si m peut être noyau d'un gourpe g, un Nod est renvoyé, sinon nil.
func (m *Mot) noeud(g *Groupe) *Nod {
	// signet motnoeud
	//debog := m.gr == "dedit" && g.id== "v.cui"
	//if debog {fmt.Println("-noeud",g.id,m.rang,m.gr,"pos=\""+m.pos+"\"")}
	rang := m.rang
	lante := len(g.ante)
	// mot de rang trop faible
	//if debog {fmt.Println("  .noeud,rang",rang,"lante",lante)}
	if rang - lante < 0 {
		return nil
	}
	//if debog {fmt.Println("  .noeud, lante ok")}
	// ou trop élevé
	//if texte.phrase.nbmots - rang < len(g.post) {
	if rang + len(g.post) - 1 >= texte.phrase.nbmots {
		return nil
	}
	//if debog {fmt.Println("  .noeud, nmbots ok")}
	// m peut-il être noyau du groupe g ?
	res2 := m.resNoyau(g)
	if len(res2) == 0 {
		return nil
	}
	//if debog {fmt.Println(" .noeud oka, res2", len(res2),res2[0].Lem.Gr)}
	m.ans2 = res2
	// création du noeud de retour
	nod := new(Nod)
	nod.grp = g
	nod.nucl = m
	nod.rang = rang
	// vérif des subs
	// ante
	r := rang - 1
	//if debog {fmt.Println(" .noeud okb",lante,"lante, r",r,nod.doc())}
	// reгcherche rétrograde des subs ante
	for ia := lante-1; ia > -1; ia-- {
		if r < 0 {
			// le rang du mot est < 0 : impossible
			return nil
		}
		ma := texte.phrase.mots[r]
		//if debog {fmt.Println("  .noeud, avant dejasub",ma.gr,"ia",ia,"r",r)}
		// passer les mots déjà subordonnés
		for ma.dejasub {
			//if debog {fmt.Println("  .noeud, ma1", ma.gr,"dejasub",ma.dejasub,"r",r)}
			r--
			if r < 0 {
				return nil
			}
			ma = texte.phrase.mots[r]
			//if debog {fmt.Println("  .noeud, ma2", ma.gr,"dejasub",ma.dejasub,"r",r)}
		}
		//if debog {fmt.Println("  .noeud ma",ma.gr,"dejasub",ma.dejasub,"pos",ma.pos,"grup",g.ante[ia].groupe.id)}
		// vérification de réciprocité, puis du lien lui-même
		if ma.domine(m) {
			return nil
		}
		sub := g.ante[ia]
		res3 := ma.resSub(sub, m)
		if len(res3) == 0 {
			//if debog {fmt.Println("  .noeud ma",ma.gr,"n'est pas sub",sub.lien,"de",m.gr)}
			return nil
		}
		ma.ans2 = res3
		nod.mma = append(nod.mma, ma)
		r--
		//if debog {fmt.Println("    vu",ma.gr)}
	}
	//if debog {fmt.Println(" .noeud okd",len(g.post),"g.post, rang",rang,"nbmots",texte.phrase.nbmots)}
	// post
	for ip, sub := range g.post {
		r := rang + ip + 1
		//if debog {fmt.Println(" .noeud, r",r,", nbmots",texte.phrase.nbmots)}
		if r >= texte.phrase.nbmots {
			break
		}
		mp := texte.phrase.mots[r]
		//if debog {fmt.Println("  .noeud avant dejasub, post, mp",mp.gr,"dejasub",mp.dejasub)}
		for mp.dejasub {
			r++
			if r >= texte.phrase.nbmots {
				return nil
			}
			mpn := mp.noyau()
			if mpn != nil && mpn.rang < m.rang {
				return nil
			}
			//if debog {fmt.Println("  .noeud, r++",r)}
			mp = texte.phrase.mots[r]
		}
		//if debog {fmt.Println("  .noeud après dejasub mp", mp.gr)}
		// réciprocité
		if mp.domine(m) {
			return nil
		}
		res4 := mp.resSub(sub, m)
		if len(res4) == 0 {
			//if debog {fmt.Println("   .noeud, return res4")}
			return nil
		}
		mp.ans2 = res4
		nod.mmp = append(nod.mmp, mp)
		r++
	}
	//if debog {fmt.Println("   .noeud, après les subs, mma",len(nod.mma),"mmp",len(nod.mmp))}
	// fixer les pos et sub des mots du noeud
	if len(nod.mma) + len(nod.mmp) > 0 {
		m.pos = g.id
		//if debog {fmt.Println("   .noeud, noy",m.gr,"pos",m.pos)}
		for _, m := range nod.mma {
			m.dejasub = true
		}
		for _, m := range nod.mmp {
			m.dejasub = true
		}
		//if debog {fmt.Println("   .noeud m.ans2", len(m.ans2),m.ans2[0].Lem.Gr)}
		return nod
	}
	return nil
}

func (m *Mot) noyau() *Mot {
	for _, n := range texte.phrase.nods {
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
func (m *Mot) resNoyau(g *Groupe) gocol.Res {
	//signet motresnoyau
	//debog := m.gr=="Admeto" && g.id=="n.genp"
	//if debog {fmt.Println(" -resNoyau",m.gr,g.id,"ans",len(m.ans),"pos=\""+m.pos+"\"")}

	var ans3 gocol.Res
	// vérif du pos
	if m.pos != "" {
		// 1. La pos définitif n'est pas encore fixé
		va := false
		for _, noy := range g.noyaux {
			va = va || noy.vaPos(m.pos)
		}
		if !va {
			return ans3
		}
		va = false
		for _, noy := range g.noyaux {
			for _, an := range m.ans2 {
				va = va || noy.vaPos(an.Lem.Pos)
			}
		}
		if !va {
			return ans3
		}
		ans3 = m.ans2
	} else {
		for _, a := range m.ans {
			for _, noy := range g.noyaux {
				//if debog {fmt.Println("  .estNoyau, noy",noy,"a.Lem",a.Lem.Pos)}
				if noy.canon > "" {
					if noy.vaSr(a) {
						ans3 = append(ans3, a)
						break
					}
				} else {
					//if debog {fmt.Println("  .resnoyau, a.Lem",a.Lem.Gr[0],"pos",a.Lem.Pos,"Noy",noy.id)}
					if noy.vaPos(a.Lem.Pos) {
						ans3 = append(ans3, a)
						break
					}
				}
			}
		}
	}
	//if debog {fmt.Println("  .estNoyau, oka, len ans3", len(ans3))}

	// vérif lexicosyntaxique
	var ans4 gocol.Res
	for _, a := range ans3 {
		va := true
		for _, ls := range g.lexsynt {
			va = va && lexsynt(a.Lem.Gr[0], ls)
		}
		for _, ls := range g.exclls {
			va = va && !lexsynt(a.Lem.Gr[0], ls)
		}
		if va {
			ans4 = append(ans4, a)
		}
	}
	//if debog {fmt.Println("  .estNoyau, okb, len ans4",len(ans4),"len gmorph",len(g.morph))}

	// vérif morpho.
	// Si aucune n'est requise, renvoyer true
	if len(g.morph) == 0 {
		return ans4
	}

	var ans5 gocol.Res
	for _, sr := range ans4 {
		var morfos []string  // morphos de sr acceptées par g
		for _, morf := range sr.Morphos {
			//if debog {fmt.Println("  .estNoyau,lem",sr.Lem.Gr[0],"morf",morf,"g.morph",g.morph)}
			if g.vaMorph(morf) {
				morfos = append(morfos, morf)
			}
		}
		//if debog {fmt.Println("  .estNoyau, morfos",len(morfos))}
		if len(morfos) > 0 {
			sr.Morphos = morfos
			ans5 = append(ans5, sr)
		}
	}
	//if debog {fmt.Println("  .estNoyau, len ans5",len(ans5))}
	return ans5
}
