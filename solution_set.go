package main

import (
	"fmt"
	"sort"
)

type solutionSet struct {
	set []solution
	nr  int
	cg  cryptogram
}

func newSolutionSet(size int, cg cryptogram) solutionSet {
	return solutionSet{make([]solution, 0, size+1), size, cg}
}

// return true if we added s to the set
func (ss *solutionSet) add(s solution) bool {
	s.score(ss.cg)

	if len(ss.set) >= ss.nr {
		if s.wordScore <= ss.set[len(ss.set)-1].wordScore {
			return false
		}
	}

	ss.set = append(ss.set, s)
	sort.Slice(ss.set, func(i, j int) bool { return ss.set[i].wordScore > ss.set[j].wordScore })

	if len(ss.set) > ss.nr {
		ss.set = ss.set[:ss.nr]
	}

	return true
}

func (ss solutionSet) dump() {
	for _, s := range ss.set {
		fmt.Println(s.decodedString(ss.cg))
	}
}
