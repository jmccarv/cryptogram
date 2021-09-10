package main

import (
	"fmt"
)

type solution struct {
	keyMap
	nrUnsolved int

	// percentage (1-100), 100% would be if every word in the solution
	// contained only the most common letter, like 'EEE EE...', which can
	// obviously never happen, but that's how this score is calculated.
	letterScore float64

	// average of all word frequencies in the solution
	wordScore float64

	decoded    []cryptogramWord
	decodeDone bool
	scoreDone  bool
}

func (s solution) String() string {
	ret := make([]byte, 26)
	i := 0
	for _, c := range s.key['A' : 'Z'+1] {
		switch c {
		case 0:
			ret[i] = ' '
		default:
			ret[i] = c
		}
		i++
	}
	return string(ret)
}

func (sorig *solution) tryWord(cw *cryptogramWord, w word) bool {
	if len(w.letters) == 0 {
		return false
	}

	for i, cc := range cw.letters {
		//fmt.Println("cw:", string(cw.letters), "  w:", string(w.letters))
		wc := w.letters[i]

		// By default cryptograms never map a letter to itself
		if !allowMapSelf && wc == cc && wc != '\'' {
			return false
		}
	}

	s := *sorig

	for i, cc := range cw.letters {
		//fmt.Println("cw:", string(cw.letters), "  w:", string(w.letters))
		wc := w.letters[i]

		if s.key[cc] == 0 && !s.letterUsed[wc] {
			s.key[cc] = wc
			s.letterUsed[wc] = true

		} else if s.key[cc] != wc {
			return false
		}
	}

	*sorig = s
	return true
}

func (s *solution) score(cg cryptogram) {
	if s.scoreDone {
		return
	}
	s.scoreDone = true
	s.decode(cg)

	for _, w := range s.decoded {
		if w.whitespace {
			continue
		}

		x := words.find(w.letters)
		s.wordScore += float64(x.freq) / float64(words.maxPatternFreq[w.pattern])

		for _, c := range w.letters {
			s.letterScore += words.letterPct[c]
		}
	}
	s.letterScore = s.letterScore / float64(cg.nrLetters) / words.maxLetterPct * 100
	s.wordScore = s.wordScore / float64(len(cg.words)) * 100
}

func (s *solution) decode(cg cryptogram) {
	if s.decodeDone {
		return
	}
	s.decodeDone = true

	for _, w := range cg.words {
		w := *w
		if !w.whitespace {
			let := []byte{}
			for _, c := range w.letters {
				if s.key[c] > 0 {
					let = append(let, s.key[c])
				} else {
					let = append(let, '_')
				}
			}
			w.letters = let
		}
		s.decoded = append(s.decoded, w)
	}
}

func (s solution) decodedString(cg cryptogram) string {
	var ret string

	s.score(cg)

	ret = fmt.Sprintf("Letter: %8.6f%%  Word: %f%%  ", s.letterScore, s.wordScore)
	for _, w := range s.decoded {
		ret += string(w.letters)
	}

	return ret
}
