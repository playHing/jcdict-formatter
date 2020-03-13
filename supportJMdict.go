/*
 * Copyright (c) 2016 Alex Yatskov <alex@foosoft.net>
 * Author: Alex Yatskov <alex@foosoft.net>
 * Copyright (c) 2020 playHing <https://github.com/playHing>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in
 * the Software without restriction, including without limitation the rights to
 * use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
 * the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package main

import (
	"log"
	"os"
	"strings"

	"github.com/FooSoft/jmdict"
)

const jmdictRevision = "jmdict4"

func jmdictBuildRules(term *dbTerm) {
	for _, tag := range term.DefinitionTags {
		switch tag {
		case "adj-i", "v1", "vk":
			term.addRules(tag)
		default:
			if strings.HasPrefix(tag, "v5") {
				term.addRules("v5")
			} else if strings.HasPrefix(tag, "vs") {
				term.addRules("vs")
			}
		}
	}
}

func jmdictBuildScore(term *dbTerm) {
	for _, tag := range term.DefinitionTags {
		switch tag {
		case "arch":
			term.Score -= 100
		}
	}
	for _, tag := range term.TermTags {
		switch tag {
		case "news", "ichi", "spec", "gai1":
			term.Score += 100
		case "P":
			term.Score += 500
		case "iK", "ik", "ok", "oK", "io", "oik":
			term.Score -= 100
		}
	}
}

func jmdictAddPriorities(term *dbTerm, priorities ...string) {
	for _, priority := range priorities {
		switch priority {
		case "news1", "ichi1", "spec1", "gai1":
			term.addTermTags("P")
			fallthrough
		case "news2", "ichi2", "spec2", "gai2":
			term.addTermTags(priority[:len(priority)-1])
		}
	}
}

func jmdictExtractTerms(edictEntry jmdict.JmdictEntry, language string) []dbTerm {
	var terms []dbTerm

	convert := func(reading jmdict.JmdictReading, kanji *jmdict.JmdictKanji) {
		if kanji != nil && reading.Restrictions != nil && !hasString(kanji.Expression, reading.Restrictions) {
			return
		}

		var termBase dbTerm
		termBase.addTermTags(reading.Information...)

		if kanji == nil {
			termBase.Expression = reading.Reading
			jmdictAddPriorities(&termBase, reading.Priorities...)
		} else {
			termBase.Expression = kanji.Expression
			termBase.Reading = reading.Reading
			termBase.addTermTags(kanji.Information...)

			for _, priority := range kanji.Priorities {
				if hasString(priority, reading.Priorities) {
					jmdictAddPriorities(&termBase, priority)
				}
			}
		}

		var partsOfSpeech []string
		for index, sense := range edictEntry.Sense {

			if len(sense.PartsOfSpeech) != 0 {
				partsOfSpeech = sense.PartsOfSpeech
			}

			if sense.RestrictedReadings != nil && !hasString(reading.Reading, sense.RestrictedReadings) {
				continue
			}

			if kanji != nil && sense.RestrictedKanji != nil && !hasString(kanji.Expression, sense.RestrictedKanji) {
				continue
			}

			term := dbTerm{
				Reading:    termBase.Reading,
				Expression: termBase.Expression,
				Score:      len(edictEntry.Sense) - index,
				Sequence:   edictEntry.Sequence,
			}

			for _, glossary := range sense.Glossary {
				if glossary.Language == nil && language == "" || glossary.Language != nil && language == *glossary.Language {
					term.Glossary = append(term.Glossary, glossary.Content)
				}
			}

			if len(term.Glossary) == 0 {
				continue
			}

			term.addDefinitionTags(termBase.DefinitionTags...)
			term.addTermTags(termBase.TermTags...)
			term.addDefinitionTags(partsOfSpeech...)
			term.addDefinitionTags(sense.Fields...)
			term.addDefinitionTags(sense.Misc...)
			term.addDefinitionTags(sense.Dialects...)

			jmdictBuildRules(&term)
			jmdictBuildScore(&term)

			terms = append(terms, term)
		}
	}

	if len(edictEntry.Kanji) > 0 {
		for _, kanji := range edictEntry.Kanji {
			for _, reading := range edictEntry.Readings {
				if reading.NoKanji == nil {
					convert(reading, &kanji)
				}
			}
		}
		for _, reading := range edictEntry.Readings {
			if reading.NoKanji != nil {
				convert(reading, nil)
			}
		}
	} else {
		for _, reading := range edictEntry.Readings {
			convert(reading, nil)
		}
	}

	return terms
}

func SupportJMdict(inputPath string) map[string][]string {
	reader, err := os.Open(inputPath)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	defer reader.Close()

	dict, _, err := jmdict.LoadJmdictNoTransform(reader)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	definitionTagsDict := make(map[string][]string)
	for _, entry := range dict.Entries {
		terms := jmdictExtractTerms(entry, "")
		for _, term := range terms {
			if tags, b := definitionTagsDict[term.Expression]; b {
				definitionTagsDict[term.Expression] = appendStringUnique(tags, term.DefinitionTags...)
			} else {
				definitionTagsDict[term.Expression] = term.DefinitionTags
			}
		}
	}
	log.Println("loaded support dict.")
	return definitionTagsDict
}
