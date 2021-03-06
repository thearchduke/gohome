package main

import (
	"encoding/json"
	"fmt"
	"github.com/thearchduke/gohome/formhandler"
	"github.com/thearchduke/gohome/markdown"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/smtp"
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

func NewBlogPage(num int, msg string) *WebPage {
	if num == -1 {
		return &WebPage{
			Urls:      &appUrls,
			Message:   template.HTML(msg),
			BlogIndex: &blogIndex,
		}

	}
	prev_a := num - 1
	next_a := num + 1
	prev, next := make(map[string]string), make(map[string]string)
	if prev_a >= 0 {
		prev = map[string]string{
			"a":     strconv.Itoa(prev_a),
			"title": blogIndex[prev_a]["title"],
		}
	}
	if next_a < len(blogIndex) {
		next = map[string]string{
			"a":     strconv.Itoa(next_a),
			"title": blogIndex[next_a]["title"],
		}
	}

	return &WebPage{
		Urls:     &appUrls,
		Message:  template.HTML(msg),
		BlogPost: template.HTML(blogIndex[num]["body"]),
		Previous: prev,
		Next:     next,
		Title:    blogIndex[num]["title"],
		Date:     blogIndex[num]["date"],
	}
}

/*//////
Global-y stuff
/////*/

var appConfig Config

var appUrls map[string]string

var templates map[string]*template.Template

var blogIndex []map[string]string

var mdParser markdown.MarkdownParser

var mailAuth smtp.Auth

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
	blogIndex = make([]map[string]string, len(blogFiles))
	metaMatcher := regexp.MustCompile("<META>::=<(.*)>::=\"(.*)\"")

	for _, mdfile := range blogFiles {
		s, _ := ioutil.ReadFile(mdfile)
		whichPost := strings.TrimSuffix(filepath.Base(mdfile), ".md")
		n, _ := strconv.Atoi(whichPost)
		blogIndex[n] = make(map[string]string)
		blogIndex[n]["body"] = mdParser.Parse(string(s))
		metas := metaMatcher.FindAllStringSubmatch(string(s), -1)
		for _, match := range metas {
			blogIndex[n][match[1]] = match[2]
		}
	}

	mailAuth = smtp.PlainAuth("",
		appConfig.Mail["username"],
		appConfig.Mail["password"],
		appConfig.Mail["server"])
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

func sendContactFormEmail(form map[string]formhandler.FormHandler) error {
	msg := fmt.Sprintf(`From: %v
To: %v
Subject: %v

The following message was submitted via the www.tynanburke.com comment form: %v`,
		form["email"].Output().(string),
		"tynanburke@gmail.com",
		form["subject"].Output().(string),
		form["message"].Output().(string))

	err := smtp.SendMail(appConfig.Mail["server"]+":"+appConfig.Mail["port"],
		mailAuth,
		form["email"].Output().(string),
		[]string{"tynanburke@gmail.com"},
		[]byte(msg))

	if err != nil {
		return err
	}
	return nil
}

/*//////
Handlers
/////*/

func blogHandler(w http.ResponseWriter, r *http.Request) {
	splitPath := strings.Split(r.URL.Path, "/")
	whichPost, err := strconv.Atoi(splitPath[2])
	switch {
	case r.URL.Path == appUrls["blog"]:
		renderTemplate(w, "blog_main", NewBlogPage(-1, ""))
	case err != nil || whichPost > len(blogIndex)-1 || whichPost < 0:
		renderTemplate(w, "blog_main", NewBlogPage(-1,
			"I'm sorry, I couldn't find that blog post. Here are some others."))
	default:
		renderTemplate(w, "blog_post", NewBlogPage(whichPost, ""))
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
			if form, err := formhandler.HandleEmailForm(r); err != nil {
				renderTemplate(w, "home", NewWebPage(err.Error()))
			} else {
				if mailError := sendContactFormEmail(form); mailError != nil {
					renderTemplate(w, "home", NewWebPage(mailError.Error()))
				} else {
					renderTemplate(w, "home", NewWebPage("Thanks for your submission!"))
				}
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
