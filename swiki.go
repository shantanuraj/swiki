package swiki

import (
	"appengine"
	"appengine/datastore"
	"html/template"
	"net/http"
	"regexp"
)

var (
	templates = template.Must(template.ParseFiles("tmpl/edit.html", "tmpl/view.html"))
	validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
)

type Page struct {
	Title string
	Body  []byte
}

func renderTemplate(w http.ResponseWriter, tmpl string, page *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", page)
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

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/view/Main", http.StatusFound)
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	page, err := getWiki(title, r)
	if err != nil || page.Body == nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}

	renderTemplate(w, "view", page)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	page, _ := getWiki(title, r)
	renderTemplate(w, "edit", page)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	page := &Page{Title: title, Body: []byte(body)}

	c := appengine.NewContext(r)
	key := datastore.NewIncompleteKey(c, "Wiki", wikiStoreKey(c))

	_, err := datastore.Put(c, key, page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func wikiStoreKey(c appengine.Context) *datastore.Key {
	return datastore.NewKey(c, "Wikistore", "default_wikis", 0, nil)
}

func getWiki(title string, r *http.Request) (*Page, error) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("Wiki").Ancestor(wikiStoreKey(c)).Filter("Title =", title)
	wikis := make([]Page, 0, 5)
	page := new(Page)

	_, err := q.GetAll(c, &wikis)

	if err != nil || len(wikis) == 0 {
		page = &Page{Title: title}
	} else {
		page = &wikis[0]
	}

	return page, err
}

func init() {
	http.HandleFunc("/", rootHandler)

	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

}
