package formhandler

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
)

/*//////
Form handling
//////*/

func HandleEmailForm(r *http.Request) (map[string]FormHandler, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}

	form := map[string]FormHandler{
		"name":    NewTextField(r.Form.Get("name"), "Name", true),
		"email":   NewEmailField(r.Form.Get("email"), "Email", true),
		"subject": NewTextField(r.Form.Get("subject"), "Subject", true),
		"message": NewTextField(r.Form.Get("message"), "Message", true),
	}
	if err_str := HandleForm(&form); err_str != "" {
		return nil, errors.New(err_str)
	}
	return form, nil
}

func HandleForm(form *map[string]FormHandler) string {
	errs := ""
	for _, field := range *form {
		if err := field.Handle(); err != nil {
			errs += fmt.Sprintf("Form Error: %v<br/>", err)
		}
	}
	return errs
}

//////////////////////////////////////////////

type FormHandler interface {
	Handle() error
	Input() interface{}
	Output() interface{}
}

//////////////////////////////////////////////

type TextField struct {
	input    string
	output   string
	name     string
	required bool
}

func (f *TextField) Handle() error {
	f.output = f.input
	if f.output == "" && f.required == true {
		return errors.New(fmt.Sprintf("%v is required", f.name))
	}
	return nil
}

func (f *TextField) Input() interface{} {
	return f.input
}

func (f *TextField) Output() interface{} {
	return f.output
}

func NewTextField(i string, n string, req bool) *TextField {
	return &TextField{
		input:    i,
		name:     n,
		required: req,
	}
}

//////////////////////////////////////////////

type EmailField struct {
	input    string
	output   string
	name     string
	required bool
}

func (f *EmailField) Handle() error {
	match, _ := regexp.MatchString(`(?i)\b[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}\b`, f.input)
	// vv NOTE: this doesn't match empty string ^^
	if match == false && f.required == true {
		return errors.New("Invalid email address")
	}
	f.output = f.input
	return nil
}

func (f *EmailField) Input() interface{} {
	return f.input
}

func (f *EmailField) Output() interface{} {
	return f.output
}

func NewEmailField(i string, n string, req bool) *EmailField {
	return &EmailField{
		input:    i,
		name:     n,
		required: req,
	}
}
