package main

import (
	"./markdown"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
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

var blogPosts map[string]string

var parser *markdown.MarkdownParser

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

	blogPosts = make(map[string]string)
	blogFiles, err := filepath.Glob(appConfig.BlogDir + "*.md")
	if err != nil {
		panic("Could not load blog markdown files")
	}

	parser = markdown.NewMarkdownParser()

	for _, mdfile := range blogFiles {
		s, _ := ioutil.ReadFile(mdfile)
		blogPosts[strings.TrimSuffix(filepath.Base(mdfile), ".md")] =
			parser.Parse(string(s))
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
