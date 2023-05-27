// Copyright 2023 by Chris Palmer, https://noncombatant.org/
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func runTest(t *testing.T, text string, expected []string, expectedErrorCount int) {
	reader := strings.NewReader(text)
	document, e := html.Parse(reader)
	if e != nil {
		t.Error(e)
	}

	var builder strings.Builder
	report := Report{Writer: &builder, ErrorCount: 0}
	Lint(&report, document, "")

	received := builder.String()
	for _, e := range expected {
		if !strings.Contains(received, e) {
			t.Errorf("received %q, expected %q", received, e)
		}
	}
	if report.ErrorCount != expectedErrorCount {
		t.Errorf("received ErrorCount %d, expected %d", report.ErrorCount, expectedErrorCount)
	}
}

func TestLintLazyLoading(t *testing.T) {
	document := `
<figure><img src="goat" alt="goat" width="0" height="0"/>
<figcaption>goat</figcaption></figure>
<iframe width="0" height="0"></iframe>
`
	expected := []string{
		"<img>/<iframe> missing loading=lazy",
	}
	runTest(t, document, expected, 2)
}

func TestLintWidthAndHeight(t *testing.T) {
	document := `
<figure><img src="goat" alt="goat" height="0" loading="lazy"/>
<figcaption>goat</figcaption></figure>
<figure><img src="goat" alt="goat" width="0" loading="lazy"/>
<figcaption>goat</figcaption></figure>
`
	expected := []string{
		"<img> missing width",
	  "<img> missing height",
	}
	runTest(t, document, expected, 2)
}

func TestLintAltText(t *testing.T) {
	document := `
<figure><img src="goat" width="0" height="0" loading="lazy"/>
<figcaption>goat</figcaption></figure>
`
	expected := []string {
		"<img> missing alt",
	}
	runTest(t, document, expected, 1)
}

func TestLintAName(t *testing.T) {
	document := `<a name="florb"></a>`
	expected := []string{
		"<a> has name; should use id",
	}
	runTest(t, document, expected, 1)
}

func TestLintImgNestedInFigure(t *testing.T) {
	document := `<img src="goat" width="0" height="0" alt="goat" loading="lazy"/>`
	expected := []string{
		"<img> not inside <figure>",
	}
	runTest(t, document, expected, 1)
}

func TestLintTimeFormatting(t *testing.T) {
	document := `
<time></time>
<time>June 99th, 12 BCE</time>
`
	expected := []string{
		"<time> needs exactly 1 text child",
		"does not have correct format",
	}
	runTest(t, document, expected, 2)
}

func TestLintFigureHasFigcaption(t *testing.T) {
	document := `<figure>hello</figure>`
	expected := []string{
		"<figure> missing <figcaption> child",
	}
	runTest(t, document, expected, 1)
}

func TestLintCurlyQuotes(t *testing.T) {
	document := `
<p>Hello, "World"</p>
<figure><img src="goat" width="0" height="0" alt="Hello, 'World'" loading="lazy"/>
<figcaption>hi</figcaption></figure>
<figure><img src="goat" width="0" height="0" alt="Hello, ‘World’" title="'wow'" loading="lazy"/>
<figcaption>hi</figcaption></figure>
`
	expected := []string{
		"contains non-curly quotes text node",
		"<img> alt or title contains non-curly quotes",
	}
	runTest(t, document, expected, 3)
}

func TestLintNesting(t *testing.T) {
	// TODO
}
