package main

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"sync"
	//"sync/atomic"
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
	initialMap  keyMap
	//nrSkips     uint64
	//nrTried     uint64
}

var validLetter [256]bool

func init() {
	validLetter['\''] = true
	for x := 'A'; x <= 'Z'; x++ {
		validLetter[x] = true
	}
}

func getNextCGWord(cgLine []byte) *cryptogramWord {
	w := &cryptogramWord{}

	if len(cgLine) < 1 {
		return nil
	}

	w.whitespace = !validLetter[cgLine[0]]
	func() {
		for i, c := range cgLine {
			if validLetter[c] == w.whitespace {
				w.letters = append(w.letters, cgLine[:i]...)
				return
			}
		}
		w.letters = append(w.letters, cgLine...)
	}()
	w.pattern = wordPattern(w.letters)
	w.freq = 1

	return w
}

func newCryptogram(cgLine []byte) (cryptogram, error) {
	cgLine = bytes.ToUpper(cgLine)
	cg := cryptogram{uniqueWords: make(map[string]*cryptogramWord)}
	cg.initialMap.key['\''] = '\''
	cg.initialMap.letterUsed['\''] = true

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

var solveWg sync.WaitGroup

func (cg *cryptogram) solveR(ctx context.Context, wg *sync.WaitGroup, spawn, maxUnsolved int, s solution, cgWords []*cryptogramWord, start word, sch chan solution) {
	triedUnknown := false

	if wg != nil {
		defer wg.Done()
	}

	//atomic.AddUint64(&cg.nrTried, 1)
	if !s.tryWord(cgWords[0], start) {
		s.nrUnsolved++
		if s.nrUnsolved > maxUnsolved {
			if partial {
				select {
				case sch <- s:
				case <-ctx.Done():
				}
			}
			return
		}
		triedUnknown = true
	}

	cgWords = cgWords[1:]
	i := 0
	wordSolved := true
	for ; i < len(cgWords) && wordSolved; i++ {
		x := make([]byte, 0, 64)
		for _, c := range cgWords[i].letters {
			if s.key[c] == 0 {
				wordSolved = false
				break
			}
			x = append(x, s.key[c])
		}

		if wordSolved {
			//atomic.AddUint64(&cg.nrSkips, uint64(len(words.forPattern(cgWords[i].pattern))))
			if _, ok := words.words[string(x)]; !ok {
				s.nrUnsolved++
				if s.nrUnsolved > maxUnsolved {
					//fmt.Println("Unknown word with solution", string(cgWords[i].letters), string(x))
					if partial {
						select {
						case sch <- s:
						case <-ctx.Done():
						}
					}
					return
				}
			}
		}
	}

	if wordSolved {
		cgWords = []*cryptogramWord{}
	} else if i > 0 {
		cgWords = cgWords[i-1:]
	}

	if len(cgWords) < 1 {
		// no encrypted words left to solve for, send our solution
		select {
		case sch <- s:
		case <-ctx.Done():
		}
		return
	}

	if ctx.Err() != nil {
		return
	}

	try := func(w word) {
		if spawn > 0 {
			solveWg.Add(1)
			go cg.solveR(ctx, &solveWg, spawn-1, maxUnsolved, s, cgWords, w, sch)
		} else {
			cg.solveR(ctx, nil, 0, maxUnsolved, s, cgWords, w, sch)
		}
	}

	for _, w := range words.forPattern(cgWords[0].pattern) {
		try(w)
	}

	if s.nrUnsolved < maxUnsolved && !triedUnknown {
		try(word{})
	}
}

func (cg *cryptogram) solve(ctx context.Context, maxUnsolved int, sch chan solution) {
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

	try := func(w word) {
		solveWg.Add(1)
		go cg.solveR(ctx, &solveWg, moreParallel, maxUnsolved, solution{keyMap: cg.initialMap}, cgWords, w, sch)
	}

	for _, w := range words.forPattern(cgWords[0].pattern) {
		try(w)
	}
	if maxUnknown > 0 {
		try(word{})
	}
	solveWg.Wait()

	//fmt.Println("nr tried", cg.nrTried)
	//fmt.Println("nr skips", cg.nrSkips)
}
