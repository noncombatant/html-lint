// Copyright 2023 by Chris Palmer, https://noncombatant.org/
// SPDX-License-Identifier: Apache-2.0

package html_lint

import (
	"fmt"
	"io"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const (
	timeFormat = "_2 January 2006"
)

type Report struct {
	io.Writer
	ErrorCount int
}

func (r *Report) Println(objects ...interface{}) {
	r.ErrorCount += 1
	fmt.Fprintln(r.Writer, objects...)
}

func hasAttribute(as []html.Attribute, key, value string) bool {
	for _, a := range as {
		if a.Key == key {
			if value == "*" {
				return a.Val != ""
			}
			return a.Val == value
		}
	}
	return false
}

func isElement(node *html.Node, tag string) bool {
	return node.Type == html.ElementNode && node.Data == tag
}

func hasParent(node *html.Node, tag string) bool {
	for p := node.Parent; p != nil; p = p.Parent {
		if p.Type == html.ElementNode && p.Data == tag {
			return true
		}
	}
	return false
}

func hasChild(node *html.Node, tag string) bool {
	if node == nil {
		return false
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == tag {
			return true
		}
		if hasChild(c.FirstChild, tag) {
			return true
		}
	}
	return false
}

// LintLazyLoading ensures that <img> and <iframe> have loading=lazy and that
// <script> has type=module. These attributes improve loading and rendering
// performance; see
// https://developer.mozilla.org/en-US/docs/Web/Performance/Lazy_loading.
func LintLazyLoading(report *Report, node *html.Node, pathname string) {
	if isElement(node, "img") || isElement(node, "iframe") {
		if !hasAttribute(node.Attr, "loading", "lazy") {
			report.Println(pathname, "<img>/<iframe> missing loading=lazy")
		}
	} else if isElement(node, "script") {
		if !hasAttribute(node.Attr, "type", "module") {
			report.Println(pathname, "<script> missing type=module")
		}
	}
}

// LintWidthAndHeight ensures that <img> has width and height attributes. This
// improves rendering performance by avoiding janky reflows.
func LintWidthAndHeight(report *Report, node *html.Node, pathname string) {
	if isElement(node, "img") {
		if !hasAttribute(node.Attr, "width", "*") {
			report.Println(pathname, "<img> missing width")
		}
		if !hasAttribute(node.Attr, "height", "*") {
			report.Println(pathname, "<img> missing height")
		}
	}
}

// LintAltText ensures that <img> has an alt attribute for accessibility.
func LintAltText(report *Report, node *html.Node, pathname string) {
	if isElement(node, "img") && !hasAttribute(node.Attr, "alt", "*") {
		report.Println(pathname, "<img> missing alt")
	}
}

// LintAName ensures that <a> does not have the name attribute (which is
// deprecated in favor of id).
func LintAName(report *Report, node *html.Node, pathname string) {
	if isElement(node, "a") && hasAttribute(node.Attr, "name", "*") {
		report.Println(pathname, "<a> has name; should use id")
	}
}

// LintImgNestedInFigure ensures that <img> is nested inside a <figure> parent.
func LintImgNestedInFigure(report *Report, node *html.Node, pathname string) {
	if isElement(node, "img") && !hasParent(node, "figure") {
		report.Println(pathname, "<img> not inside <figure>")
	}
}

// LintTimeFormatting ensures that <time> elements are correctly formatted.
func LintTimeFormatting(report *Report, node *html.Node, pathname string) {
	if isElement(node, "time") {
		c := node.FirstChild
		if c == nil || c.Type != html.TextNode {
			report.Println(pathname, "<time> needs exactly 1 text child")
		} else {
			_, e := time.Parse(timeFormat, c.Data)
			if e != nil {
				report.Println(pathname, "<time> child", c.Data, "does not have correct format", timeFormat)
			}
		}
	}
}

// LintFigureHasFigcaption ensures that <figure> has a <figcaption> child.
func LintFigureHasFigcaption(report *Report, node *html.Node, pathname string) {
	if isElement(node, "figure") && !hasChild(node, "figcaption") {
		report.Println(pathname, "<figure> missing <figcaption> child")
	}
}

// LintCurlyQuotes ensures that non-code text nodes, alt attributes, and title
// attributes use curly quotes.
func LintCurlyQuotes(report *Report, node *html.Node, pathname string) {
	if node.Type == html.TextNode && !hasParent(node, "pre") && !hasParent(node, "code") && !hasParent(node, "script") && !hasParent(node, "style") {
		if strings.ContainsAny(node.Data, "'\"") {
			report.Println(pathname, "contains non-curly quotes text node", node.Data)
		}
	}
	if isElement(node, "img") {
		for _, a := range node.Attr {
			if a.Key == "alt" || a.Key == "title" {
				if strings.ContainsAny(a.Val, "'\"") {
					report.Println(pathname, "<img> alt or title contains non-curly quotes")
				}
			}
		}
	}
}

// Lint applies all the Lint* functions and then recurses down the tree.
func Lint(report *Report, node *html.Node, pathname string) {
	LintLazyLoading(report, node, pathname)
	LintWidthAndHeight(report, node, pathname)
	LintAltText(report, node, pathname)
	LintAName(report, node, pathname)
	LintImgNestedInFigure(report, node, pathname)
	LintTimeFormatting(report, node, pathname)
	LintFigureHasFigcaption(report, node, pathname)
	LintCurlyQuotes(report, node, pathname)

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		Lint(report, c, pathname)
	}
}

// LintNesting ensures that all tags are properly closed.
func LintNesting(report *Report, reader io.Reader, pathname string) {
	z := html.NewTokenizer(reader)
	var stack []string

	for {
		token := z.Next()
		if token == html.ErrorToken {
			break
		}
		tagBytes, _ := z.TagName()
		tag := string(tagBytes)
		if token == html.StartTagToken {
			stack = append(stack, tag)
		} else if token == html.EndTagToken {
			if len(stack) == 0 {
				report.Println(pathname, "tag stack underflow")
			}
			last := len(stack) - 1
			previous := stack[last]
			if tag != previous {
				report.Println(pathname, "Unmatched pair", string(tag), string(previous))
			}
			stack = stack[:last]
		}
	}

	if len(stack) != 0 {
		report.Println(pathname, "Unclosed tags", stack)
	}
}
