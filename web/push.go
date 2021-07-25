package web

import (
	"html/template"
	"net/http"
)

type pushTempl struct {
	Content template.HTML
	Head    Head
	Sidebar []Sidebar

	Config Config
}

func (s *Site) ShowPage(w http.ResponseWriter, r *http.Request, content template.HTML, head Head, sidebar []Sidebar) error {
	if sidebar == nil {
		sidebar = []Sidebar{s.DefaultSidebar()}
	}

	if content == "" {
		content = "<code>No content.</code>"
	}

	return s.templates.ExecuteTemplate(w, "base", pushTempl{
		Content: content,
		Head:    head,
		Sidebar: sidebar,
	})
}

type Sidebar struct {
	Title string

	Content template.HTML
}

func (s *Site) DefaultSidebar() Sidebar {
	return Sidebar{
		Title:   "TODO",
		Content: "todo, make this do stuff",
	}
}

type Head struct {
	Title string

	Extras []template.HTML
}

func TitleHead(t string) Head { return Head{Title: t} }
