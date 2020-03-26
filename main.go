package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] input-path\n", path.Base(os.Args[0]))
	fmt.Fprint(os.Stderr, "https://github.com/playhing/yomichan-import/\n\n")
	fmt.Fprint(os.Stderr, "Parameters:\n")
	flag.PrintDefaults()
}

func main() {
	var (
		outputPath  = flag.String("outputpath", "", "dictionary output path")
		supportdict = flag.String("supportdict", "default", "path to support dictionary")
		pretty      = flag.Bool("pretty", false, "output prettified dictionary JSON")
		stride      = flag.Int("stride", 10000, "dictionary bank stride")
	)

	flag.Usage = usage
	flag.Parse()

	if flag.NArg() == 0 {
		usage()
		os.Exit(2)
	}

	inputPath := flag.Arg(0)

	log.Println("inputpath:", inputPath)
	log.Println("outputpath:", *outputPath)
	log.Println("supportdict:", *supportdict)
	log.Println("pretty:", *pretty)
	log.Println("stride:", *stride)

	if _, err := os.Stat(inputPath); err != nil {
		log.Fatalf("dictionary path '%s' does not exist", inputPath)
	}

	fnBase := filepath.Base(inputPath)
	if *outputPath == "" {
		fnDir := filepath.Dir(inputPath)
		outfn := strings.Split(fnBase, ".")[0] + "-import.zip"
		*outputPath = filepath.Join(fnDir, outfn)
	}

	params := gloParams{*stride, *pretty, *supportdict, ""}

	var convertor dictConv
	if format, err := detectFormat(fnBase); err == nil {
		switch format {
		case "shougakukanJC":
			convertor = &shougakukanJCConv{"20.03.26.0", inputPath, *outputPath, params}
		case "epwing":
			epwingExtractors := map[string]epwingExtractor{
				"大辞泉": makeDaijisenExtractor(),
			}
			convertor = &epwingConv{"epwing", inputPath, *outputPath, params, epwingExtractors}
		}
	} else {
		log.Fatal(err)
	}

	if err := convertor.Export(); err != nil {
		log.Fatalf("conversion process failed: %s", err.Error())
	}
	log.Println("success.")
}
