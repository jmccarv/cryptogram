package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
)

type word struct {
	letters []byte
	pattern int // normalized letter pattern, e.g. queen => 12334 => unique integer pattern id

	// how frequent this word is used in English
	// also used by the cryptogram type to track the number of occurrences of a code word
	freq int
}

type wordMap struct {
	words      map[string]word
	pattern    map[int][]word
	letterFreq [256]int
	letterPct  [265]float64
}

func (w word) String() string {
	return fmt.Sprintf("%v %d %d", string(w.letters), w.freq, w.pattern)
}

var (
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
func wordPattern(w []byte) int {
	i := 0
	var m [91]int
	var pat []int

	if len(w) < 1 {
		return 0
	}

	for _, c := range w {
		if n := m[c]; n > 0 {
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

// Given a line of word frequencies like:
// word xxx
// where word is a word consisting of only the letters A-Z or an apostrophe
// and xxx is a number representing the frequency that words appears in english text
//
// return a word value of the parsed data
func parseFreqLine(line []byte) (word, error) {
	l := bytes.SplitN(line, []byte(" "), 2)
	w := word{}

	if len(l) != 2 {
		return w, fmt.Errorf("Invalid input (not enough fields)")
	}

	if len(l[0]) == 0 {
		return w, fmt.Errorf("Invalid input (empty word)")
	}

	f, err := strconv.Atoi(string(l[1]))
	if err != nil {
		return w, fmt.Errorf("Invalid input (invalid number)")
	}
	w.freq = f

	w.letters = bytes.ToUpper(l[0])
	for _, x := range w.letters {
		if !validLetter[x] {
			return w, fmt.Errorf("Invalid input (invalid characters in word) on line")
		}
	}
	w.pattern = wordPattern(w.letters)
	return w, nil
}

func (m wordMap) forPattern(p int) []word {
	return m.pattern[p]
}

func (m wordMap) find(x []byte) word {
	return m.words[string(x)]
}

func (m wordMap) store(w word) {
	m.words[string(w.letters)] = w
	m.pattern[w.pattern] = append(m.pattern[w.pattern], w)
}

func (m wordMap) sort() {
	for _, wList := range m.pattern {
		sort.Slice(wList, func(i, j int) bool { return wList[j].freq < wList[i].freq })
	}
}

func (m *wordMap) readWordList(fn string) {
	*m = wordMap{
		words:   make(map[string]word),
		pattern: make(map[int][]word),
	}

	lfh, err := os.Open(fn)
	if err != nil {
		log.Fatal(err)
	}
	defer lfh.Close()

	s := bufio.NewScanner(lfh)
	lno := 0
	nrLetters := 0
	for s.Scan() {
		lno++
		w, err := parseFreqLine(s.Bytes())
		if err != nil {
			fmt.Printf("%v: %v on line %v", fn, err, lno)
			continue
		}

		m.store(w)
		nrLetters += len(w.letters) * w.freq
		for _, c := range w.letters {
			m.letterFreq[c] += w.freq
		}
	}

	for c := 'A'; c <= 'Z'; c++ {
		m.letterPct[c] = float64(m.letterFreq[c]) / float64(nrLetters)
	}

	// Sort each list in descending order of frequency
	m.sort()
	m.dispFreqs()
}

type freqs struct {
	letter byte
	freq   int
	pct    float64
}

func (m wordMap) dispFreqs() {
	freq := make([]freqs, 26)
	for c := 'A'; c <= 'Z'; c++ {
		i := c - 'A'

		freq[i].letter = byte(c)
		freq[i].freq = m.letterFreq[c]
		freq[i].pct = m.letterPct[c]
	}

	sort.Slice(freq, func(i, j int) bool { return freq[i].pct > freq[j].pct })
	//sort.Float64s(freq)

	for _, f := range freq {
		fmt.Printf("%c %4.2f  ", f.letter, f.pct*100)
	}
	fmt.Println()
}
