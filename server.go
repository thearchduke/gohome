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
	markdownPatterns["h1"] = regexp.MustCompile("(?m)\n|^# *[^#][^\n]*")
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
Functions
/////*/

// markdown
func md_makeH1(src []byte) []byte {
	s := strings.Replace(string(src), "#", "", 1)
	if s != "\n" {
		out := "<h1>" + s + "</h1>"
		return []byte(out)
	}
	return nil
}

/*//////
Helpers
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

//-/-/-/-/-/-/-/-/
// Here we go!
//-/-/-/-/-/-/-/-/
func main() {
	initApp()
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
