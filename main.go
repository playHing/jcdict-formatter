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
		outputPath  = flag.String("output-path", "", "dictionary output path")
		stride      = flag.Int("stride", 10000, "dictionary bank stride")
		pretty      = flag.Bool("pretty", false, "output prettified dictionary JSON")
		supportdict = flag.String("support-dict", "", "path to support dictionary")
	)

	flag.Usage = usage
	flag.Parse()

	if flag.NArg() == 0 {
		usage()
		os.Exit(2)
	}

	inputPath := flag.Arg(0)

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
		case "xxtjc":
			convertor = &xxtJcConv{"xxtjc", inputPath, *outputPath, params}
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
