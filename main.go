package main

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"runtime/pprof"
	"sync"
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
	moreParallel int
	cpuprofile   string
	memprofile   string
	initKey      string
	progress     bool
	partial      bool
	includeKey   bool
)

//go:embed freqc.txt
var freqData []byte

var words = wordMap{}

func main() {
	flag.StringVar(&freqFile, "f", "", "Word frequency list")
	flag.DurationVar(&maxRuntime, "r", 0, "Quit after this amount of time. Ex: 30s or 1m")
	flag.IntVar(&topN, "topn", 5, "Display top N solutions each time one is found")
	flag.IntVar(&maxSolutions, "s", 0, "Stop searching after finding this many solutions")
	flag.IntVar(&maxUnknown, "u", 0, "Maximum allowed unknown words")
	flag.BoolVar(&allowMapSelf, "map-self", false, "Allow encrypted letter to map to itself")
	flag.IntVar(&moreParallel, "m", moreParallel, "Use more parallelism, probably best to leave at 0 or maybe 1")
	flag.StringVar(&cpuprofile, "cpuprofile", "", "Write cpu profile to 'file'")
	flag.StringVar(&memprofile, "memprofile", "", "Write memory profile to 'file'")
	flag.StringVar(&initKey, "key", "", "Initial key mapping in the form 'ABC=XYZ[, ]...'")
	flag.BoolVar(&progress, "p", false, "Display solutions as they're found instead of only when complete")
	flag.BoolVar(&partial, "P", false, "Evaluate all partial solutions, much slower but may be useful with -p")
	flag.BoolVar(&includeKey, "v", false, "Include the key in the output for each solution")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			"Usage: %s [OPTIONS] [CRYPTOGRAM FILE]\n\n"+
				"Read cryptograms, one per line, from CRYPTOGRAM FILE or stdin\n\n",
			os.Args[0],
		)
		flag.PrintDefaults()
	}
	flag.Parse()

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
	if freqFile != "" {
		fmt.Println("Reading freqs from", freqFile)
		words.readWordListFile(freqFile)
	} else {
		fmt.Println("Using embedded word frequency data")
		words.readWordListData(freqData)
	}

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

	startingKey := newKeyMap()
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

		if bytes.Contains(line, []byte("=")) {
			if initKey != "" {
				continue
			}
			if k, ok := parseKeyMap(line); ok {
				key = k
			} else {
				fmt.Printf("Skipping invalid key map on line %v\n", lno)
			}
			continue
		}

		cg, err := newCryptogram(line)
		if err != nil {
			fmt.Printf("skipping cryptogram on line %v: %v\n", lno, err)
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
				fmt.Println("Stopping solve of current cryptogram...")
				cancelFunc()
			case <-ctx.Done():
				return
			}
		}()

		// This routine accepts solutions from the solver and adds them
		// to our solution set.
		sch := make(chan solution)
		nrFound := 0

		var wg sync.WaitGroup
		wg.Add(1)
		go func(sch chan solution) {
			defer wg.Done()
			dump := func() {
				fmt.Println("Found:", nrFound)
				ss.dump(includeKey)
				fmt.Println()
			}

			for s := range sch {
				nrFound++

				if ss.add(s) && progress {
					dump()
				}

				if maxSolutions > 0 && nrFound >= maxSolutions {
					cancelFunc()
					break
				}
			}

			if !progress {
				dump()
			}
		}(sch)

		// Now we can start the solver for this cryptogram
		start := time.Now()
		fmt.Printf("\nSolving: %v\n", cg)
		fmt.Printf("Using Key: %v\n", cg.initialMap)

		// Won't return until all workers have stopped, which means
		// nothing will write to sch after this returns
		cg.solve(ctx, maxUnknown, sch)

		// Tidy up
		cancelFunc() // Will shut down the signal handler if it's still running
		close(sch)   // Shuts down our solution reader routine
		wg.Wait()

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
