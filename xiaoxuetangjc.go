package main

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

const xxtjcRevision = "XiaoXueTang JC"

func xxtjcExtractTerms(reader *os.File) (terms dbTermList, err error) {

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

func xxtjcExportDb(inputPath, outputPath string, params gloParams) error {
	reader, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	terms, err := xxtjcExtractTerms(reader)
	if err != nil {
		return err
	}

	if params.title == "" {
		params.title = "XiaoXueTang JC"
	}

	recordData := map[string]dbRecordList{
		"term": terms.crush(),
	}

	return writeDb(
		outputPath,
		params.title,
		xxtjcRevision,
		true,
		recordData,
		params.stride,
		params.pretty,
	)
}
