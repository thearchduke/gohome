package main

import (
	"encoding/json"
	"fmt"
	//"html/template"
	"io/ioutil"
	"net/http"
)

/*//////
Constants
/////*/

/*//////
Config
/////*/

type Config struct {
	TemplateDir string
	Port        int
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

func initApp() {
	loadConfig("config.json", &appConfig)
	appUrls = make(map[string]string)
	appUrls["home"] = "/"
}

/*//////
Types
/////*/

type WebPage struct {
	Title   []byte
	Content []byte
	Urls    map[string]string
}

func NewWebPage(title []byte, content []byte) *WebPage {
	u := make(map[string]string)
	u["home"] = "/"
	return &WebPage{
		Title:   title,
		Content: content,
		Urls:    appUrls,
	}

}

/*//////
Functions
/////*/

/*//////
Helpers
/////*/

/*//////
Handlers
/////*/

//////////////////////
func main() {
	initApp()
	fmt.Println(NewWebPage([]byte("hello"), []byte("world!")))
	//	http.HandleFunc('/path/', aHandler)
	http.ListenAndServe(":8080", nil)
}
