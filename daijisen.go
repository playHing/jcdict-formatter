/*
 * Modified by: playHing <https://github.com/playHing>

 * Copyright (c) 2016 Alex Yatskov <alex@foosoft.net>
 * Author: Alex Yatskov <alex@foosoft.net>
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
	"regexp"
	"strings"
)

type daijisenExtractor struct {
	partsExp     *regexp.Regexp
	expShapesExp *regexp.Regexp
	expVarExp    *regexp.Regexp
	readGroupExp *regexp.Regexp
	metaExp      *regexp.Regexp
	v5Exp        *regexp.Regexp
	v1Exp        *regexp.Regexp
}

func makeDaijisenExtractor() epwingExtractor {
	return &daijisenExtractor{
		partsExp:     regexp.MustCompile(`([^【]+)(?:【(.*)】)?`),
		expShapesExp: regexp.MustCompile(`[×△]+`),
		expVarExp:    regexp.MustCompile(`（([^）]*)）`),
		readGroupExp: regexp.MustCompile(`[‐・]+`),
		metaExp:      regexp.MustCompile(`［([^］]*)］`),
		v5Exp:        regexp.MustCompile(`(動.[四五](［[^］]+］)?)|(動..二)`),
		v1Exp:        regexp.MustCompile(`(動..一)`),
	}
}

func (e *daijisenExtractor) extractTerms(entry epwingEntry, sequence int) []dbTerm {
	matches := e.partsExp.FindStringSubmatch(entry.Heading)
	if matches == nil {
		return nil
	}

	var expressions []string
	if expression := matches[2]; len(expression) > 0 {
		expression = e.expShapesExp.ReplaceAllString(expression, "")
		for _, split := range strings.Split(expression, "・") {
			splitInc := e.expVarExp.ReplaceAllString(split, "$1")
			expressions = append(expressions, splitInc)
			if split != splitInc {
				splitExc := e.expVarExp.ReplaceAllLiteralString(split, "")
				expressions = append(expressions, splitExc)
			}
		}
	}

	var reading string
	if reading = matches[1]; len(reading) > 0 {
		reading = e.readGroupExp.ReplaceAllLiteralString(reading, "")
		reading = e.expVarExp.ReplaceAllLiteralString(reading, "")
	}

	var tags []string
	for _, split := range strings.Split(entry.Text, "\n") {
		if matches := e.metaExp.FindStringSubmatch(split); matches != nil {
			tags = append(tags, strings.Split(matches[1], "・")...)
		}
	}

	var terms []dbTerm
	if len(expressions) == 0 {
		term := dbTerm{
			Expression: reading,
			Glossary:   []string{entry.Text},
			Sequence:   sequence,
		}

		e.exportRules(&term, tags)
		terms = append(terms, term)

	} else {
		for _, expression := range expressions {
			term := dbTerm{
				Expression: expression,
				Reading:    reading,
				Glossary:   []string{entry.Text},
				Sequence:   sequence,
			}

			e.exportRules(&term, tags)
			terms = append(terms, term)
		}
	}

	return terms
}

func (e *daijisenExtractor) exportRules(term *dbTerm, tags []string) {
	for _, tag := range tags {
		term.addRules(tag)
	}
}

func (*daijisenExtractor) getRevision() string {
	return "daijisen"
}

func (*daijisenExtractor) getFontNarrow() map[int]string {
	return getKanji("daijisen", "narrow")
}

func (*daijisenExtractor) getFontWide() map[int]string {
	return getKanji("daijisen", "wide")
}
