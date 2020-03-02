package main

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

type xxtJcConv struct {
	revision   string
	inputPath  string
	outputPath string
	gloParams
}

func (*xxtJcConv) extractTerms(reader *os.File) (terms dbTermList, err error) {

	for scanner := bufio.NewScanner(reader); scanner.Scan(); {
		line := scanner.Text()
		if strings.HasPrefix(line, "##") {
			continue
		}
		reg := regexp.MustCompile(`[\p{Ll}A-Z']+`)
		parts := strings.Split(line, "\\n")
		expparts := strings.Split(strings.Split(parts[0], "\t")[0], "|")

		var term dbTerm

		term.Expression = expparts[0]
		if len(expparts) >= 3 {
			term.Reading = expparts[2]
		}

		var glossary string
		for i := 1; i < len(parts); i++ {

			parts[i] = reg.ReplaceAllString(parts[i], "")

			if strings.HasPrefix(parts[i], "$") {
				glossary += "\n$ " + parts[i][1:]
			} else {
				if glossary != "" {
					term.Glossary = append(term.Glossary, glossary)
				}
				glossary = parts[i]
			}
		}
		if glossary != "" {
			term.Glossary = append(term.Glossary, glossary)
		}

		terms = append(terms, term)
	}

	return
}

func (x *xxtJcConv) Export() error {
	reader, err := os.Open(x.inputPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	terms, err := x.extractTerms(reader)
	if err != nil {
		return err
	}

	if x.title == "" {
		x.title = "XiaoXueTang JC"
	}

	recordData := map[string]dbRecordList{
		"term": terms.crush(),
	}

	return writeDb(
		x.outputPath,
		x.title,
		x.revision,
		true,
		recordData,
		x.stride,
		x.pretty,
	)
}
