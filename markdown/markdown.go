package markdown

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

/*//////
Markdown
//////*/

type MarkdownParser struct {
	patterns map[string]*regexp.Regexp
}

func NewMarkdownParser() MarkdownParser {
	p := MarkdownParser{patterns: make(map[string]*regexp.Regexp)}

	p.patterns["h"] = regexp.MustCompile("(?m)\n|^#+ *[^#][^\n]*")
	p.patterns["p"] = regexp.MustCompile("(?m)\n|^[^<][^#][^\n]+")
	p.patterns["hr"] = regexp.MustCompile("(?m)\n|^---+")
	p.patterns["a"] = regexp.MustCompile(`(?U)\[(.*)\]\((.*)\)`)
	p.patterns["em"] = regexp.MustCompile(`(?U)[\*]+(.*)[\*]+`)
	p.patterns["inline"] = regexp.MustCompile("(?U)`(.*)`")
	p.patterns["img"] = regexp.MustCompile(`(?U)!\[(.*)\]\((.*)\)`)
	p.patterns["meta"] = regexp.MustCompile("<META>.*")
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
		return fmt.Sprintf("<p>%v</p>", src)
	}
	return src
}

func markdownMakeHr(src string) string {
	if src != "\n" {
		return "<hr/>"
	}
	return src
}

func (p MarkdownParser) ParseFile(fname string) string {
	b, _ := ioutil.ReadFile(fname)
	return p.Parse(string(b))
}

func (p MarkdownParser) Parse(src string) (parsed string) {
	parsed = p.patterns["h"].ReplaceAllStringFunc(src, markdownMakeH)
	parsed = p.patterns["em"].ReplaceAllString(parsed, "<em>$1</em>")
	parsed = p.patterns["inline"].ReplaceAllString(parsed, "<code>$1</code>")
	parsed = p.patterns["img"].ReplaceAllString(parsed, "<img src=\"$2\" alt=\"$1\">")
	parsed = p.patterns["a"].ReplaceAllString(parsed, "<a href=\"$2\">$1</a>")
	parsed = p.patterns["hr"].ReplaceAllStringFunc(parsed, markdownMakeHr)
	parsed = p.patterns["meta"].ReplaceAllString(parsed, "")
	parsed = p.patterns["p"].ReplaceAllStringFunc(parsed, markdownMakeP)
	return
}
