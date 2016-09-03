package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
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

var appConfig Config

/*//////
Global-y stuff
/////*/

var appUrls map[string]string

var templates map[string]*template.Template

var markdownPatterns map[string]*regexp.Regexp

var blogPosts map[string]string

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

	markdownPatterns = make(map[string]*regexp.Regexp)
	markdownPatterns["h"] = regexp.MustCompile("(?m)\n|^#+ *[^#][^\n]*")
	markdownPatterns["p"] = regexp.MustCompile("(?m)\n|^[^<][^#][^\n]+")
	markdownPatterns["hr"] = regexp.MustCompile("(?m)\n|^---+")
	markdownPatterns["em"] = regexp.MustCompile(`(?U)[\*_]+(.*)[\*_]+`)
	markdownPatterns["inline"] = regexp.MustCompile("(?U)`(.*)`")
	markdownPatterns["a"] = regexp.MustCompile(`(?U)\[(.*)\]\((.*)\)`)
	markdownPatterns["img"] = regexp.MustCompile(`(?U)!\[(.*)\]\((.*)\)`)

	blogPosts = make(map[string]string)
	blogFiles, err := filepath.Glob(appConfig.BlogDir + "*.md")
	if err != nil {
		panic("Could not load blog markdown files")
	}
	for _, mdfile := range blogFiles {
		blogPosts[strings.TrimSuffix(filepath.Base(mdfile), ".md")] =
			parseMarkdownFile(mdfile)
	}
}

/*//////
Types
/////*/

type WebPage struct {
	Urls     map[string]string
	BlogPost template.HTML
}

func NewWebPage() *WebPage {
	return &WebPage{
		Urls: appUrls,
	}
}

func NewBlogPage(n string) *WebPage {
	return &WebPage{
		Urls:     appUrls,
		BlogPost: template.HTML(blogPosts[n]),
	}
}

/*//////
Functions & Helpers
/////*/

// markdown
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

func parseMarkdownFile(fname string) string {
	s, _ := ioutil.ReadFile(fname)
	return parseMarkdown(string(s))
}

func parseMarkdown(src string) string {
	src = markdownPatterns["h"].ReplaceAllStringFunc(src, markdownMakeH)
	src = markdownPatterns["em"].ReplaceAllString(src, "<em>$1</em>")
	src = markdownPatterns["inline"].ReplaceAllString(src, "<code>$1</code>")
	src = markdownPatterns["img"].ReplaceAllString(src, "<img src=\"$2\" alt=\"$1\">")
	src = markdownPatterns["a"].ReplaceAllString(src, "<a href=\"$2\">$1</a>")
	src = markdownPatterns["hr"].ReplaceAllStringFunc(src, markdownMakeHr)
	src = markdownPatterns["p"].ReplaceAllStringFunc(src, markdownMakeP)
	return src
}

func renderTemplate(w http.ResponseWriter, name string, p *WebPage) error {
	tmpl, ok := templates[name]
	if !ok {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(
			"<h2>Better tell the administrator we couldn't find the template.</h2>"))
		return fmt.Errorf("Could not locate template")
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.ExecuteTemplate(w, "base", p)
}

/*//////
Handlers
/////*/

func homeHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index", NewWebPage())
}

func blogTestHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "blog", NewBlogPage("test"))
}

//-/-/-/-/-/-/-/-/
// Here we go!
//-/-/-/-/-/-/-/-/
func main() {
	initApp()
	http.HandleFunc(appUrls["home"], homeHandler)
	http.HandleFunc(appUrls["blog"], blogTestHandler)
	http.Handle(appUrls["static"],
		http.StripPrefix(appUrls["static"],
			http.FileServer(http.Dir(appUrls["staticRoot"]))))
	//http.HandleFunc("/", homeHandler)
	err := http.ListenAndServe(":"+appConfig.Port, nil)
	if err != nil {
		fmt.Println(err)
	}
}
