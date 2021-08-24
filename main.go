package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
)

var (
	freqFile string
)

var validLetter [256]bool

func init() {
	validLetter['\''] = true
	for x := 'A'; x <= 'Z'; x++ {
		validLetter[x] = true
	}
}

var words = wordMap{}

func main() {
	flag.StringVar(&freqFile, "freqfile", "freqc.txt", "Word frequency list")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			"Usage: %s [OPTIONS] [CRYPTOGRAM FILE]\n\n"+
				"Read cryptograms, one per line, from CRYPTOGRAM FILE or stdin\n\n",
			os.Args[0],
		)
		flag.PrintDefaults()
	}
	flag.Parse()

	words.readWordList(freqFile)

	var cgFile io.ReadCloser
	var err error
	if len(os.Args) > 1 {
		if cgFile, err = os.Open(os.Args[1]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		cgFile = os.Stdin
	}
	defer cgFile.Close()

	s := bufio.NewScanner(cgFile)
	lno := 0
	for s.Scan() {
		lno++
		cg, err := newCryptogram(bytes.ToUpper(s.Bytes()))
		ss := newSolutionSet(3, cg)

		if err != nil {
			fmt.Printf("skipping cryptogram on line %v: %v", lno, err)
			continue
		}
		if cg.nrWords() < 1 {
			continue
		}

		sch := make(chan solution)
		go func(sch chan solution) {
			for s := range sch {
				if ss.add(s) {
					ss.dump()
					fmt.Println()
				}
			}
		}(sch)

		cg.solve(context.Background(), 0, sch)
		close(sch)
	}
}
