package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"
)

var (
	freqFile     string
	maxRuntime   time.Duration
	topN         int
	maxSolutions int
	maxUnknown   int
	allowMapSelf bool
	maxParallel  = runtime.NumCPU() * 2
	cpuprofile   string
	memprofile   string
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
	flag.StringVar(&freqFile, "f", "freqc.txt", "shortcut for -freqfile")
	flag.DurationVar(&maxRuntime, "max-runtime", 0, "Quit after this amount of time. Ex: 30s or 1m")
	flag.DurationVar(&maxRuntime, "r", 0, "shortcut for -max-runtime")
	flag.IntVar(&topN, "topn", 3, "Display top N solutions each time one is found")
	flag.IntVar(&maxSolutions, "max-solve", 0, "Stop searching after find this man solutions")
	flag.IntVar(&maxSolutions, "s", 0, "shortcut for -max-solve")
	flag.IntVar(&maxUnknown, "max-unknown", 0, "Maximum allowed unknown words")
	flag.IntVar(&maxUnknown, "u", 0, "shortcut for -max-unknown")
	flag.BoolVar(&allowMapSelf, "map-self", false, "Allow encrypted letter to map to itself")
	flag.IntVar(&maxParallel, "p", maxParallel, "Number of worker threads to run")
	flag.StringVar(&cpuprofile, "cpuprofile", "", "Write cpu profile to 'file'")
	flag.StringVar(&memprofile, "memprofile", "", "Write memory profile to 'file'")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			"Usage: %s [OPTIONS] [CRYPTOGRAM FILE]\n\n"+
				"Read cryptograms, one per line, from CRYPTOGRAM FILE or stdin\n\n",
			os.Args[0],
		)
		flag.PrintDefaults()
	}
	flag.Parse()

	if maxParallel < 1 {
		maxParallel = 1
	}

	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	words.readWordList(freqFile)

	var cgFile io.ReadCloser
	var err error
	if len(flag.Args()) > 0 {
		if cgFile, err = os.Open(flag.Args()[0]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		cgFile = os.Stdin
	}
	defer cgFile.Close()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)

	s := bufio.NewScanner(cgFile)
	lno := 0
	for s.Scan() {
		lno++
		cg, err := newCryptogram(s.Bytes())
		ss := newSolutionSet(topN, cg)

		if err != nil {
			fmt.Printf("skipping cryptogram on line %v: %v", lno, err)
			continue
		}
		if cg.nrWords() < 1 {
			continue
		}

		ctx, cancelFunc := context.WithCancel(context.Background())
		if maxRuntime > 0 {
			ctx, cancelFunc = context.WithTimeout(ctx, maxRuntime)
		}

		go func() {
			select {
			case <-sigs:
				cancelFunc()
			case <-ctx.Done():
				return
			}
		}()

		sch := make(chan solution)
		nrFound := 0
		go func(sch chan solution) {
			for s := range sch {
				nrFound++

				if ss.add(s) {
					fmt.Println("Found:", nrFound)
					ss.dump()
					fmt.Println()
				}

				if maxSolutions > 0 && nrFound >= maxSolutions {
					cancelFunc()
					return
				}
			}
		}(sch)

		start := time.Now()

		fmt.Printf("\n%v\n", cg)
		cg.solve(ctx, maxUnknown, sch)
		cancelFunc()
		close(sch)
		fmt.Println("Evaluated", nrFound, "solutions in", time.Now().Sub(start))

		if memprofile != "" {
			f, err := os.Create(memprofile)
			if err != nil {
				log.Fatal("could not create memory profile: ", err)
			}
			defer f.Close() // error handling omitted for example
			runtime.GC()    // get up-to-date statistics
			if err := pprof.WriteHeapProfile(f); err != nil {
				log.Fatal("could not write memory profile: ", err)
			}
		}
	}
}
