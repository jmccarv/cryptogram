package main

import (
	"fmt"
)

type solution struct {
	key        [91]byte
	letterUsed [91]bool
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

func (sorig *solution) tryWord(cw *cryptogramWord, w word) bool {
	s := *sorig

	for i, cc := range cw.letters {
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
		s.wordScore += float64(x.freq*w.freq) / float64(len(cg.words))

		for _, c := range w.letters {
			//fmt.Printf("%c %0.2f", c, words.letterPct[c])
			s.letterScore += words.letterPct[c] * float64(w.freq)
		}
	}
	s.letterScore = s.letterScore / float64(cg.nrLetters) / words.maxLetterPct * 100
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

	ret = fmt.Sprintf("Letter: %8.6f%%  Word: %f  ", s.letterScore, s.wordScore)
	for _, w := range s.decoded {
		ret += string(w.letters)
	}

	return ret
}
