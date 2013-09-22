package main

import (
	"bytes"
	"sort"
	"text/template"
)

type Project struct {
	Assets []Asset
	Tag    string
}

type Asset struct {
	Url      string
	Content  string
	Tag      string
	Mime     string
	Filename string
	Order    int64
	Image    bool
}

type Page struct {
	Title    string
	Projects map[string]Project
	Js       []Asset
	Css      []Asset
}

var templates = template.Must(template.ParseFiles("files/page.html"))

// Renders the assets
func renderPage(p Page) (string, error) {
	buf := new(bytes.Buffer)
	err := templates.ExecuteTemplate(buf, "page.html", p)
	return buf.String(), err
}

type byOrder []Asset

func (v byOrder) Len() int {
	return len(v)
}

func (v byOrder) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
	return
}

func (v byOrder) Less(i, j int) bool {
	return v[i].Order < v[j].Order
}

// Evaluates the mimetype of the assets and makes a page
func evalAssets(assets []Asset) Page {
	var page = Page{
		"",
		make(map[string]Project),
		make([]Asset, 0),
		make([]Asset, 0),
	}

	// Loop over all the assets
	for _, a := range assets {
		proj := page.Projects[a.Tag]
		if proj.Tag == "" {
			proj = Project{
				make([]Asset, 0),
				a.Tag,
			}
		}

		switch a.Mime {
		case "image/png", "image/jpeg", "image/jpg", "image/gif":
			a.Image = true
			proj.Assets = append(proj.Assets, a)
		case "text/inject":
			proj.Assets = append(proj.Assets, a)
		case "text/css":
			page.Css = append(page.Css, a)
		case "application/javascript":
			page.Js = append(page.Js, a)
		}

		page.Projects[proj.Tag] = proj
	}

	for _, p := range page.Projects {
		sort.Sort(byOrder(p.Assets))
		if len(p.Assets) == 0 {
			delete(page.Projects, p.Tag)
		}
	}

	return page
}
