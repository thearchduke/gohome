package markdown

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

var markdownPatterns map[string]*regexp.Regexp

type MarkdownParser struct {
	patterns map[string]*regexp.Regexp
}

func NewMarkdownParser() *MarkdownParser {
	p := &MarkdownParser{patterns: make(map[string]*regexp.Regexp)}

	p.patterns["h"] = regexp.MustCompile("(?m)\n|^#+ *[^#][^\n]*")
	p.patterns["p"] = regexp.MustCompile("(?m)\n|^[^<][^#][^\n]+")
	p.patterns["hr"] = regexp.MustCompile("(?m)\n|^---+")
	p.patterns["em"] = regexp.MustCompile(`(?U)[\*_]+(.*)[\*_]+`)
	p.patterns["inline"] = regexp.MustCompile("(?U)`(.*)`")
	p.patterns["a"] = regexp.MustCompile(`(?U)\[(.*)\]\((.*)\)`)
	p.patterns["img"] = regexp.MustCompile(`(?U)!\[(.*)\]\((.*)\)`)

	return p
}

func markdownMakeH(src string) string {
	octothorps := strings.Count(src, "#")
	s := strings.Replace(src, "#", "", -1)
	out := fmt.Sprintf("<h%[1]v>%[2]v</h%[1]v>", octothorps, s)
	if s != "\n" {
		return out
	}
	return src
}

func markdownMakeP(src string) string {
	if src != "\n" {
		return "<p>" + src + "</p>"
	}
	return src
}

func markdownMakeHr(src string) string {
	if src != "\n" {
		return "<hr/>"
	}
	return src
}

func (p *MarkdownParser) ParseFile(fname string) string {
	b, _ := ioutil.ReadFile(fname)
	return p.Parse(string(b))
}

func (p *MarkdownParser) Parse(src string) string {
	src = p.patterns["h"].ReplaceAllStringFunc(src, markdownMakeH)
	src = p.patterns["em"].ReplaceAllString(src, "<em>$1</em>")
	src = p.patterns["inline"].ReplaceAllString(src, "<code>$1</code>")
	src = p.patterns["img"].ReplaceAllString(src, "<img src=\"$2\" alt=\"$1\">")
	src = p.patterns["a"].ReplaceAllString(src, "<a href=\"$2\">$1</a>")
	src = p.patterns["hr"].ReplaceAllStringFunc(src, markdownMakeHr)
	src = p.patterns["p"].ReplaceAllStringFunc(src, markdownMakeP)
	return src
}
