package main

import (
	"bytes"
	"context"
	"fmt"
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
	initialMap  keyMap
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

type workerData struct {
	solution
	cgWords []*cryptogramWord
	start   word
}

func (cg cryptogram) solveR(ctx context.Context, maxUnsolved int, s solution, cgWords []*cryptogramWord, start word, sch chan solution) {
	triedUnknown := false
	if !s.tryWord(cgWords[0], start) {
		s.nrUnsolved++
		if s.nrUnsolved > maxUnsolved {
			//s.score(cg)
			//sch <- s
			return
		}
		triedUnknown = true
	}

	if len(cgWords[1:]) < 1 {
		// no encrypted words left to solve for, send our solution
		//s.score(cg)
		sch <- s
		return
	}

	cgWords = cgWords[1:]

	if ctx.Err() != nil {
		return
	}

	for _, w := range words.forPattern(cgWords[0].pattern) {
		cg.solveR(ctx, maxUnsolved, s, cgWords, w, sch)
	}

	if s.nrUnsolved < maxUnsolved && !triedUnknown {
		cg.solveR(ctx, maxUnsolved, s, cgWords, word{}, sch)
	}
}

func (cg cryptogram) solveWorker(ctx context.Context, maxUnsolved int, workCh chan workerData, sch chan solution) {
	for {
		var d workerData
		var ok bool

		select {
		case d, ok = <-workCh:
			if !ok {
				return
			}
			break
		case <-ctx.Done():
			return
		}

		cg.solveR(ctx, maxUnsolved, d.solution, d.cgWords, d.start, sch)
	}
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

	wch := make(chan workerData)
	// Spawn workers here
	// cgWords
	var wg sync.WaitGroup
	for i := 0; i < maxParallel; i++ {
		wg.Add(1)
		go func() {
			cg.solveWorker(ctx, maxUnsolved, wch, sch)
			wg.Done()
		}()
	}

	try := func(w word) {
		s := solution{keyMap: cg.initialMap}
		wch <- workerData{solution: s, cgWords: cgWords, start: w}
	}
	fmt.Printf("%s has %d possible solutions\n", string(cgWords[0].letters), len(words.forPattern(cgWords[0].pattern)))

	for _, w := range words.forPattern(cgWords[0].pattern) {
		if ctx.Err() != nil {
			break
		}

		try(w)
	}
	if maxUnknown > 0 {
		try(word{})
	}
	close(wch)
	wg.Wait()
}
