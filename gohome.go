package main

import (
	"encoding/json"
	"fmt"
	"github.com/thearchduke/gohome/formhandler"
	"github.com/thearchduke/gohome/markdown"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

//////FORM VALIDATION: Good place to use a Checker interface to make
// so that I can make a slice of things that are checkable and just run
// check() on them!
//////

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
	Message   template.HTML
	Title     string
	Date      string
	Previous  map[string]string
	Next      map[string]string
}

func NewWebPage(msg string) *WebPage {
	return &WebPage{
		Urls:    &appUrls,
		Message: template.HTML(msg),
	}
}

func NewBlogPage(num, msg string) *WebPage {
	if num == "main" {
		return &WebPage{
			Urls:      &appUrls,
			Message:   template.HTML(msg),
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
		Message:  template.HTML(msg),
		BlogPost: template.HTML(blogPosts[num]["body"]),
		Previous: prev,
		Next:     next,
		Title:    blogPosts[num]["title"],
		Date:     blogPosts[num]["date"],
	}
}

/*//////
Global-y stuff
/////*/

var appConfig Config

var appUrls map[string]string

var templates map[string]*template.Template

var blogPosts map[string]map[string]string

var blogIndex []map[string]string

var mdParser markdown.MarkdownParser

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

	mdParser = markdown.NewMarkdownParser()
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
	_, is_page := templates[path]
	switch {
	case is_page:
		renderTemplate(w, path, NewWebPage(""))
	case r.URL.Path == "/":
		if r.Method == "POST" {
			if _, err := formhandler.HandleEmailForm(r); err != nil {
				renderTemplate(w, "home", NewWebPage(err.Error()))
			} else {
				renderTemplate(w, "home", NewWebPage("Thanks for your submission!"))
			}
		} else {
			renderTemplate(w, "home", NewWebPage(""))
		}
	case !is_page:
		renderTemplate(w, "home", NewWebPage("Looks like we couldn't find your page, sorry."))
	}
}

//-/-/-/-/-/-/-/-/
// Here we go!
//-/-/-/-/-/-/-/-/

// for heroku
func GetPort() string {
	port := os.Getenv("PORT")

	if port == "" {
		port = appConfig.Port
		fmt.Println("INFO: No PORT environment variable detected, defaulting to " + port)
	}
	return ":" + port
}

func main() {
	initApp()
	http.HandleFunc(appUrls["blog"], blogHandler)
	http.Handle(appUrls["static"],
		http.StripPrefix("/static/",
			http.FileServer(http.Dir(appUrls["staticRoot"]))))
	http.HandleFunc("/", genericHandler)
	err := http.ListenAndServe(GetPort(), nil)
	if err != nil {
		fmt.Println(err)
	}
}
