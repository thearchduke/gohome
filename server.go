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
}

/*//////
Types
/////*/

type WebPage struct {
	Urls map[string]string
}

func NewWebPage() *WebPage {
	return &WebPage{
		Urls: appUrls,
	}
}

/*//////
Functions & Helpers
/////*/

// markdown
func markdownMakeH(src []byte) []byte {
	str_src := string(src)
	octothorps := strings.Count(str_src, "#")
	s := strings.Replace(str_src, "#", "", -1)
	out := fmt.Sprintf("<h%[1]v>%[2]v</h%[1]v>", octothorps, s)
	if s != "\n" {
		return []byte(out)
	}
	return src
}

func markdownMakeP(src []byte) []byte {
	s := string(src)
	if s != "\n" {
		return []byte("<p>" + s + "</p>")
	}
	return src
}

func markdownMakeHr(src []byte) []byte {
	s := string(src)
	if s != "\n" {
		return []byte("<hr/>")
	}
	return src
}

func parseMarkdownFile(fname string) []byte {
	s, _ := ioutil.ReadFile(fname)
	return parseMarkdown(s)
}

func parseMarkdown(src []byte) []byte {
	src = markdownPatterns["h"].ReplaceAllFunc(src, markdownMakeH)
	src = markdownPatterns["em"].ReplaceAll(src, []byte("<em>$1</em>"))
	src = markdownPatterns["inline"].ReplaceAll(src, []byte("<code>$1</code>"))
	src = markdownPatterns["img"].ReplaceAll(src, []byte("<img src=\"$2\" alt=\"$1\">"))
	src = markdownPatterns["a"].ReplaceAll(src, []byte("<a href=\"$2\">$1</a>"))
	src = markdownPatterns["hr"].ReplaceAllFunc(src, markdownMakeHr)
	src = markdownPatterns["p"].ReplaceAllFunc(src, markdownMakeP)
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

//-/-/-/-/-/-/-/-/
// Here we go!
//-/-/-/-/-/-/-/-/
func main() {
	initApp()
	//fmt.Println(string(parseMarkdownFile("reg.md")))
	http.HandleFunc(appUrls["home"], homeHandler)
	http.Handle(appUrls["static"],
		http.StripPrefix(appUrls["static"],
			http.FileServer(http.Dir(appUrls["staticRoot"]))))
	//http.HandleFunc("/", homeHandler)
	err := http.ListenAndServe(":"+appConfig.Port, nil)
	if err != nil {
		fmt.Println(err)
	}
}
