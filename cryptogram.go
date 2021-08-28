package main

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"sync"
)

type cryptogramWord struct {
	word
	whitespace bool
}

func (w cryptogramWord) String() string {
	return fmt.Sprintf("%v %v", w.word, w.whitespace)
}

type cryptogram struct {
	words       []*cryptogramWord
	uniqueWords map[string]*cryptogramWord
	nrLetters   int // number of non-whitespace characters
}

func getNextCGWord(cgLine []byte) *cryptogramWord {
	w := &cryptogramWord{}

	if len(cgLine) < 1 {
		return nil
	}

	want := validLetter[cgLine[0]]
	func() {
		for i, c := range cgLine {
			if validLetter[c] != want {
				w.letters = append(w.letters, cgLine[:i]...)
				return
			}
		}
		w.letters = append(w.letters, cgLine...)
	}()
	w.pattern = wordPattern(w.letters)
	w.whitespace = !want
	w.freq = 1

	return w
}

func newCryptogram(cgLine []byte) (cryptogram, error) {
	cg := cryptogram{uniqueWords: make(map[string]*cryptogramWord)}

	if len(cgLine) < 1 {
		return cg, nil
	}

	i := 0
	// Skip any leading 'whitespace' (non word characters)
	for ; i < len(cgLine) && !validLetter[cgLine[i]]; i++ {
		// Commented line, ignore
		if cgLine[i] == '#' {
			return cg, nil
		}
	}

	for i < len(cgLine) {
		w := getNextCGWord(cgLine[i:len(cgLine)])
		if w == nil || len(w.letters) < 1 {
			break
		}

		if x, ok := cg.uniqueWords[string(w.letters)]; ok {
			cg.words = append(cg.words, x)
			x.freq++
			w = x
		} else {
			cg.words = append(cg.words, w)
			if !w.whitespace {
				cg.uniqueWords[string(w.letters)] = w
			}
		}

		if !w.whitespace {
			cg.nrLetters += len(w.letters)
		}

		//fmt.Println(w)
		i += len(w.letters)
	}

	return cg, nil
}

func (c cryptogram) String() string {
	s := ""
	for _, w := range c.words {
		s += string(w.letters)
	}
	return s
}

func (cg cryptogram) nrWords() int {
	return len(cg.words)
}

func (cg cryptogram) solve(ctx context.Context, maxUnsolved int, sch chan solution) {
	// Get a list of the unique code words to solve.
	cgWords := make([]*cryptogramWord, 0, len(cg.uniqueWords))
	for _, x := range cg.uniqueWords {
		cgWords = append(cgWords, x)
	}

	// We want to start with the coded words that have the least possible
	// solutions, i.e. the least words in our word list with the same
	// pattern.
	sort.Slice(cgWords, func(i, j int) bool {
		return len(words.forPattern(cgWords[i].pattern)) < len(words.forPattern(cgWords[j].pattern))
	})

	// Now we can attempt a solve
	s := solution{}
	s.key['\''] = '\''
	s.letterUsed['\''] = true

	cg.solveR(ctx, maxUnsolved, sch, s, cgWords)
}

var bucket chan interface{}

func init() {
	bucket = make(chan interface{}, runtime.NumCPU()*2)
}

func (cg cryptogram) solveR(ctx context.Context, maxUnsolved int, sch chan solution, s solution, cgWords []*cryptogramWord) {
	cw := cgWords[0]
	var wg sync.WaitGroup

	solve := func(s solution, w word) {
		if !s.tryWord(cw, w) {
			s.nrUnsolved++
			// No words worked with the current solution
			if s.nrUnsolved > maxUnsolved {
				return
			}
		}

		if len(cgWords) > 1 {
			cg.solveR(ctx, maxUnsolved, sch, s, cgWords[1:])
		} else {
			// We're on the last word in cgWords[] so here we can report our solution
			sch <- s
		}
	}

	for _, w := range words.forPattern(cw.pattern) {
		if ctx.Err() != nil {
			break
		}

		select {
		case bucket <- nil:
			wg.Add(1)
			go func(s solution, w word) {
				solve(s, w)
				<-bucket
				wg.Done()
			}(s, w)
		default:
			solve(s, w)
		}
	}

	/*
		if len(cgWords) > 1 && s.nrUnsolved < maxUnsolved {
			s.nrUnsolved++
			cg.solveR(ctx, maxUnsolved, sch, s, cgWords[1:])
		}
	*/

	wg.Wait()
}
