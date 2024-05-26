// Copyright 2024 by Chris Palmer, https://noncombatant.org/
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"os"

	lint "github.com/noncombatant/html_lint"
	"golang.org/x/net/html"
)

const (
	helpMessage = `Analyzes HTML files for style, completeness, and overall deliciousness. 😋

Usage:

  html-lint [file [...]]

If no files are given, analyzes the standard input.`
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), helpMessage)
	}
	flag.Parse()

	report := lint.Report{Writer: os.Stderr, ErrorCount: 0}

	for _, pathname := range flag.Args() {
		reader, e := os.Open(pathname)
		if e != nil {
			report.Println(e)
			continue
		}
		defer reader.Close()

		document, e := html.Parse(reader)
		if e != nil {
			report.Println(e)
			continue
		}
		lint.Lint(&report, document, pathname)
		if _, e := reader.Seek(0, 0); e != nil {
			report.Println(e)
			continue
		}
		lint.LintNesting(&report, reader, pathname)
	}
	if len(flag.Args()) == 0 {
		document, e := html.Parse(os.Stdin)
		if e != nil {
			report.Println(e)
			os.Exit(report.ErrorCount)
		}
		lint.Lint(&report, document, "<stdin>")
	}
	os.Exit(report.ErrorCount)
}
