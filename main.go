package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

type word struct {
	letters string
	pattern int // normalized letter pattern, e.g. queen => 12334 => unique integer pattern id
	freq    int // how frequent this word is used in English
}

var (
	freqFile = "freqc.txt"
	words    = make(map[int][]word) // keyed on word length

	maxPatternID = 0
	patternID    = make(map[string]int) // map of patterns to unique pattern id
)

// Given a word, return an integer that is unique for the letter pattern
// of this word. These patterns are calculated by assigning a number for
// each unique letter in the word and then mapping that to a single integer.
//
// An example:
//  Q U E E N
//  1 2 3 3 4 is the pattern used. Any word with this letter pattern
//            will return the same integer from this routine.
//
// Another:
// M I S S I S S I P P I
// 1 2 3 3 2 3 3 2 4 4 2
func wordPattern(w string) int {
	i := 0
	m := make(map[rune]int)
	var pat []int

	for _, c := range w {
		if n, ok := m[c]; ok {
			pat = append(pat, n)
		} else {
			i++
			pat = append(pat, i)
			m[c] = i
		}
	}

	var p string
	for _, i := range pat {
		p += "." + strconv.Itoa(i)
	}

	if n, ok := patternID[p]; ok {
		// This pattern already exists, return its ID
		return n
	}

	// It's a new pattern, assign the next ID to this pattern
	maxPatternID++
	patternID[p] = maxPatternID
	return maxPatternID
}

func readWordList() {
	lfh, err := os.Open(freqFile)
	if err != nil {
		log.Fatal(err)
	}

	s := bufio.NewScanner(lfh)
	lno := 0
	for s.Scan() {
		lno++
		l := strings.SplitN(s.Text(), " ", 2)
		if len(l) != 2 {
			fmt.Printf("%v: Invalid input at line %v", freqFile, lno)
			continue
		}
		if len(l[0]) == 0 {
			fmt.Printf("%v: Invalid input (no word) on line %v", freqFile, lno)
			continue
		}
		f, err := strconv.Atoi(l[1])
		if err != nil {
			fmt.Printf("%v: Invalid input (invalid number) on line %v", freqFile, lno)
			continue
		}

		w := word{
			letters: l[0],
			freq:    f,
			pattern: wordPattern(l[0]),
		}

		wlen := len(w.letters)
		words[wlen] = append(words[wlen], w)
	}

	// Sort each list in descending order of frequency
	for _, wList := range words {
		sort.Slice(wList, func(i, j int) bool { return wList[j].freq < wList[i].freq })
		/*
			for _, w := range wList {
				fmt.Println(l, w.letters, w.freq, w.pattern)
			}
		*/
	}
}

func main() {
	readWordList()
}
