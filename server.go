package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

/*//////
Constants
/////*/

/*//////
Config
/////*/

type Config struct {
	TemplateDir string
	BlogDir     string
	Port        string
	Mail        map[string]string
	BasicAuth   map[string]string
}

func loadConfig(fname string, cf *Config) {
	body, err := ioutil.ReadFile(fname)
	if err != nil {
		panic("Could not locate/read config file")
	}
	err = json.Unmarshal(body, cf)
	if err != nil {
		panic("Could not parse config into Config struct")
	}
}

/*//////
Pages
/////*/

type WebPage struct {
	Urls      *map[string]string
	BlogPost  template.HTML
	BlogIndex *[]map[string]string
	Message   string
	Title     string
	Date      string
	Previous  map[string]string
	Next      map[string]string
}

func NewWebPage(msg string) *WebPage {
	return &WebPage{
		Urls:    &appUrls,
		Message: msg,
	}
}

func NewBlogPage(num, msg string) *WebPage {
	if num == "main" {
		return &WebPage{
			Urls:      &appUrls,
			Message:   msg,
			BlogIndex: &blogIndex,
		}

	}
	num_int, _ := strconv.Atoi(num)
	prev_a := strconv.Itoa(num_int - 1)
	next_a := strconv.Itoa(num_int + 1)
	prev := make(map[string]string)
	next := make(map[string]string)
	if _, ok := blogPosts[prev_a]; ok {
		prev = map[string]string{
			"a":     prev_a,
			"title": blogPosts[prev_a]["title"],
		}
	}
	if _, ok := blogPosts[next_a]; ok {
		next = map[string]string{
			"a":     next_a,
			"title": blogPosts[next_a]["title"],
		}
	}

	return &WebPage{
		Urls:     &appUrls,
		Message:  msg,
		BlogPost: template.HTML(blogPosts[num]["body"]),
		Previous: prev,
		Next:     next,
		Title:    blogPosts[num]["title"],
		Date:     blogPosts[num]["date"],
	}
}

/*//////
Markdown
//////*/

var markdownPatterns map[string]*regexp.Regexp

type MarkdownParser struct {
	patterns map[string]*regexp.Regexp
}

func NewMarkdownParser() *MarkdownParser {
	p := MarkdownParser{patterns: make(map[string]*regexp.Regexp)}

	p.patterns["h"] = regexp.MustCompile("(?m)\n|^#+ *[^#][^\n]*")
	p.patterns["p"] = regexp.MustCompile("(?m)\n|^[^<][^#][^\n]+")
	p.patterns["hr"] = regexp.MustCompile("(?m)\n|^---+")
	p.patterns["a"] = regexp.MustCompile(`(?U)\[(.*)\]\((.*)\)`)
	p.patterns["em"] = regexp.MustCompile(`(?U)[\*]+(.*)[\*]+`)
	p.patterns["inline"] = regexp.MustCompile("(?U)`(.*)`")
	p.patterns["img"] = regexp.MustCompile(`(?U)!\[(.*)\]\((.*)\)`)
	p.patterns["meta"] = regexp.MustCompile("<META>.*")
	return &p
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
	src = p.patterns["meta"].ReplaceAllString(src, "")
	src = p.patterns["p"].ReplaceAllStringFunc(src, markdownMakeP)
	return src
}

/*//////
Global-y stuff
/////*/

var appConfig Config

var appUrls map[string]string

var templates map[string]*template.Template

var blogPosts map[string]map[string]string

var blogIndex []map[string]string

var mdParser *MarkdownParser

func initApp() {
	loadConfig("config.json", &appConfig)
	appUrls = make(map[string]string)
	appUrls["home"] = "/home/"
	appUrls["photos"] = "/photos/"
	appUrls["blog"] = "/blog/"
	appUrls["about"] = "/about/"
	appUrls["code"] = "/code/"
	appUrls["writing"] = "/writing/"
	appUrls["static"] = "/static/"
	appUrls["staticRoot"] = "./static"

	templates = make(map[string]*template.Template)

	templateFiles, err := filepath.Glob(appConfig.TemplateDir + "*.tmpl")
	if err != nil {
		panic("Could not load files in templateDir")
	}

	for _, tmpl := range templateFiles {
		templates[strings.TrimSuffix(filepath.Base(tmpl), ".tmpl")] =
			template.Must(template.ParseFiles(
				tmpl, appConfig.TemplateDir+"base.tmpl"))
	}

	blogFiles, err := filepath.Glob(appConfig.BlogDir + "*.md")
	if err != nil {
		panic("Could not load blog markdown files")
	}

	mdParser = NewMarkdownParser()
	blogPosts = make(map[string]map[string]string)
	metaMatcher := regexp.MustCompile("<META>::=<(.*)>::=\"(.*)\"")

	for _, mdfile := range blogFiles {
		s, _ := ioutil.ReadFile(mdfile)
		whichPost := strings.TrimSuffix(filepath.Base(mdfile), ".md")
		blogPosts[whichPost] = make(map[string]string)
		blogPosts[whichPost]["body"] = mdParser.Parse(string(s))
		metas := metaMatcher.FindAllStringSubmatch(string(s), -1)
		for _, match := range metas {
			blogPosts[whichPost][match[1]] = match[2]
		}
	}
	blogIndex = make([]map[string]string, len(blogPosts))
	for i, _ := range blogIndex {
		blogIndex[i] = blogPosts[strconv.Itoa(i)]
	}
}

/*//////
Helpers
/////*/

func renderTemplate(w http.ResponseWriter, name string, p *WebPage) error {
	tmpl, ok := templates[name]
	if !ok {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(
			"<h2>Better tell the administrator something went wrong with the template.</h2>"))
		return fmt.Errorf("Could not locate template")
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.ExecuteTemplate(w, "base", p)
}

/*//////
Handlers
/////*/

func blogHandler(w http.ResponseWriter, r *http.Request) {
	splitPath := strings.Split(r.URL.Path, "/")
	_, is_post := blogPosts[splitPath[2]]
	switch {
	case r.URL.Path == appUrls["blog"]:
		renderTemplate(w, "blog_main", NewBlogPage("main", ""))
	case is_post:
		renderTemplate(w, "blog_post", NewBlogPage(splitPath[2], ""))
	case !is_post:
		renderTemplate(w, "blog_main",
			NewBlogPage("main",
				"I'm sorry, I couldn't find that blog post. Here are some others."))
	}
}

func genericHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.Replace(r.URL.Path, "/", "", -1)
	_, ok := templates[path]
	switch {
	case ok:
		renderTemplate(w, path, NewWebPage(""))
	case r.URL.Path == "/":
		renderTemplate(w, "home", NewWebPage(""))
	case !ok:
		renderTemplate(w, "home", NewWebPage("Looks like we couldn't find your page, sorry."))
	}
}

//-/-/-/-/-/-/-/-/
// Here we go!
//-/-/-/-/-/-/-/-/
func main() {
	initApp()
	http.HandleFunc(appUrls["blog"], blogHandler)
	http.Handle(appUrls["static"],
		http.StripPrefix("/static/",
			http.FileServer(http.Dir(appUrls["staticRoot"]))))
	http.HandleFunc("/", genericHandler)
	err := http.ListenAndServe(":"+appConfig.Port, nil)
	if err != nil {
		fmt.Println(err)
	}
}
