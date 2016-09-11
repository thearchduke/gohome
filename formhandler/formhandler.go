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

func HandleEmailForm(r *http.Request) (map[string]Handler, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}

	form := map[string]Handler{
		"name": &NameField{
			input: r.Form.Get("name"),
		},
		"email": &EmailField{
			input: r.Form.Get("email"),
		},
		"subject": &SubjectField{
			input: r.Form.Get("subject"),
		},
		"message": &MessageField{
			input: r.Form.Get("message"),
		},
	}
	if err_str := HandleForm(&form); err_str != "" {
		return nil, errors.New(err_str)
	}
	return form, nil
}

func HandleForm(form *map[string]Handler) string {
	errs := ""
	for _, field := range *form {
		if err := field.Handle(); err != nil {
			errs += fmt.Sprintf("Form Error: %v<br/>", err)
		}
	}
	return errs
}

//////////////////////////////////////////////

type Handler interface {
	Handle() error
	Guts() map[string]interface{}
}

//////////////////////////////////////////////

type NameField struct {
	input  string
	output string
}

func (f *NameField) Handle() error {
	f.output = f.input
	if f.output == "" {
		return errors.New("Name is required")
	}
	return nil
}

func (f NameField) Guts() map[string]interface{} {
	return map[string]interface{}{"input": f.input, "output": f.output}
}

//////////////////////////////////////////////

type EmailField struct {
	input  string
	output string
}

func (f *EmailField) Handle() error {
	match, _ := regexp.MatchString(`(?i)\b[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}\b`, f.input)
	if match == false {
		return errors.New("Invalid email address")
	}
	f.output = f.input
	return nil
}

func (f EmailField) Guts() map[string]interface{} {
	return map[string]interface{}{"input": f.input, "output": f.output}
}

//////////////////////////////////////////////

type SubjectField struct {
	input  string
	output string
}

func (f *SubjectField) Handle() error {
	f.output = f.input
	if f.output == "" {
		return errors.New("Subject is required")
	}
	return nil
}

func (f SubjectField) Guts() map[string]interface{} {
	return map[string]interface{}{"input": f.input, "output": f.output}
}

//////////////////////////////////////////////

type MessageField struct {
	input  string
	output string
}

func (f *MessageField) Handle() error {
	f.output = f.input
	if f.output == "" {
		return errors.New("Message is required")
	}
	return nil
}

func (f MessageField) Guts() map[string]interface{} {
	return map[string]interface{}{"input": f.input, "output": f.output}
}
