package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var dataPath = filepath.Join(".", "data")
var templatePath = filepath.Join(".", "tmpl")
var pathSep = string(os.PathSeparator)

var templates = template.Must(template.ParseFiles(
	filepath.Join(templatePath, "view.html"),
	filepath.Join(templatePath, "edit.html"),
))

var validPath = regexp.MustCompile(`^/(view|edit|save)/([a-zA-Z0-9/\-_]+)$`)

// Page data structure representing a wiki page
type Page struct {
	Title string
	Body  []byte
}

func getFilename(title string) string {
	ext := ".txt"
	return filepath.Join(dataPath, title+ext)
}

func (p *Page) save() error {
	filename := getFilename(p.Title)

	// Create a directory if it does not exist
	segments := strings.Split(filename, pathSep)
	filepath := strings.Join(segments[:len(segments)-1], pathSep)
	os.MkdirAll(filepath, os.ModePerm)

	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := getFilename(title)
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	port := ":8080"
	fmt.Println("Starting server on port " + port)

	log.Fatal(http.ListenAndServe(port, nil))
}
