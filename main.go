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
	"regexp"
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
	maxParallel  = runtime.NumCPU()
	cpuprofile   string
	memprofile   string
	initKey      string
)

var words = wordMap{}

func main() {
	flag.StringVar(&freqFile, "f", "freqc.txt", "Word frequency list")
	flag.DurationVar(&maxRuntime, "r", 0, "Quit after this amount of time. Ex: 30s or 1m")
	flag.IntVar(&topN, "topn", 5, "Display top N solutions each time one is found")
	flag.IntVar(&maxSolutions, "s", 0, "Stop searching after finding this many solutions")
	flag.IntVar(&maxUnknown, "u", 0, "Maximum allowed unknown words")
	flag.BoolVar(&allowMapSelf, "map-self", false, "Allow encrypted letter to map to itself")
	flag.IntVar(&maxParallel, "p", maxParallel, "Number of worker threads to run")
	flag.StringVar(&cpuprofile, "cpuprofile", "", "Write cpu profile to 'file'")
	flag.StringVar(&memprofile, "memprofile", "", "Write memory profile to 'file'")
	flag.StringVar(&initKey, "key", "", "Initial key mapping in the form 'ABC=XYZ[, ]...'")

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

	// Read and parse word frequency list
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

	// Set up signal handler to stop a search on SIGINT
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)

	// Start looking for cryptograms in input
	s := bufio.NewScanner(cgFile)
	lno := 0
	ctx, cancelFunc := context.WithCancel(context.Background())

	rxComment := regexp.MustCompile(`^\s*#`)

	var startingKey keyMap
	if k, ok := parseKeyMap([]byte(initKey)); ok {
		startingKey = k
	}
	key := startingKey

	for s.Scan() {
		lno++
		line := s.Bytes()

		if rxComment.Match(line) {
			continue
		}

		if k, ok := parseKeyMap(line); ok {
			key = k
			continue
		}

		cg, err := newCryptogram(line)
		if err != nil {
			fmt.Printf("skipping cryptogram on line %v: %v", lno, err)
			continue
		}
		if cg.nrWords() < 1 {
			continue
		}

		cg.initialMap = key
		ss := newSolutionSet(topN, cg)

		if maxRuntime > 0 {
			ctx, cancelFunc = context.WithTimeout(ctx, maxRuntime)
		}

		// Signal handler, INT (Ctrl-C) aborts current cryptogram
		go func() {
			select {
			case <-sigs:
				cancelFunc()
			case <-ctx.Done():
				return
			}
		}()

		// This routine accepts solutions from the solver and adds them
		// to our solution set.
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

		// Now we can start the solver for this cryptogram
		start := time.Now()
		fmt.Printf("\n%v\n", cg)
		fmt.Printf("Using Key: %v\n", cg.initialMap)

		// Won't return until all workers have stopped, which means
		// nothing will write to sch after this returns
		cg.solve(ctx, maxUnknown, sch)

		// Tidy up
		cancelFunc() // Will shut down the signal handler if it's still running
		close(sch)   // Shuts down our solution reader routine

		fmt.Println("Evaluated", nrFound, "solutions in", time.Now().Sub(start))

		// Reset for next cryptogram
		ctx, cancelFunc = context.WithCancel(context.Background())
		key = startingKey
	}

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
