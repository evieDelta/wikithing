package web

import (
	"fmt"
	"html"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"git.lan/wikithing/wterr"
	"github.com/go-chi/chi"
)

func (r *Site) SetupChi() error {
	r.R = chi.NewMux()

	e, err := fs.ReadDir(r.StaticFS, ".")
	if err != nil {
		return err
	}
	for _, x := range e {
		fmt.Println(x.Name(), x.Type().String())
	}

	r.R.NotFound(theMostHorrible404)
	r.R.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(r.StaticFS))))
	r.R.Get("/page/*", r.Page)
	r.R.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(301)
		fmt.Fprint(w, `<!DOCTYPE html><html><head><meta http-equiv="Refresh" content="0; url='/page/index" /></head></html>`)
	})

	return nil
}

func (s *Site) WrapRun(w http.ResponseWriter, r *http.Request, n string, fn func() error) {
	err := fn()
	if err == nil {
		return
	}
	log.Println(n+":", err)

	if e, ok := err.(wterr.Err); ok {
		s.HandleWterr(w, r, e)
	} else {
		s.HandleGenericError(w, r, e)
	}
}

// i'm sorry
var theMostHorrible404 = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	fmt.Fprintf(w, `
		<html style="background: linear-gradient(to bottom right, #E673A7 0%%, #994c5f 100%%); height: 100%%">
		<head><link rel="stylesheet" href="/static/external/normalize.css"></head>
		<style>#xbutton:hover { background-color: #994C5F !important; box-shadow: 2px 2px 4px #FF99BA !important; }</style>
		<div style="max-width: 30em; margin: auto; margin-top: 5em; color: #383235; box-shadow: 20px 20px 30px #804060;">
		<div style="display: flex; clear: both; padding: 0em 1.25em; background-color: #FFABD2; width: max; font-size: 95%%; text-shadow: 2px 2px 4px #FFCCDC;">
		<h3 style="margin: 0em; padding: 0.15em 0em 0.12em; float: left">404 not found</h3>
		<a id="xbutton" style="font-family: monospace; margin: 0em 0em 0em auto; padding: 0.25em 0.33em; background-color: #D96C87; text-decoration: none; display: grid" href="/"><img src="/static/assets/svg/roundedXwhite.svg" style="width: 1.5em; height: 1.5em; margin: auto"/></a>
		</div>
		<div style="padding: 0.5em 1.25em 1.75em 1em; margin: 0em; background-color: #E6E1E3; color: #383235">
		<p style="margin: 0em 0em 0.75em">I'm sorry we could not find what you were looking for</p>
		<code style="border-left: 0.25em solid #804060; background-color: #CCCACB; padding: 0.25em 0.33em">%v : %v</code></div></div></html>`,
		r.Method, html.EscapeString(r.URL.String()))
}

func (s *Site) HandleWterr(w http.ResponseWriter, r *http.Request, e wterr.Err) {
	// TODO: kinda important that this needs to be able to categorise more errors lol
	switch e.Type {

	default:
		s.HandleGenericError(w, r, e)
	}
}

type ErrorContext struct {
	ErrorCode int

	ShowRawError bool
	RawError     string
}

func (s *Site) HandleGenericError(w http.ResponseWriter, r *http.Request, e error) {
	w.WriteHeader(http.StatusInternalServerError)
	buf := &strings.Builder{}

	log.Println(e)

	err := s.templates.ExecuteTemplate(buf, "error500", ErrorContext{
		ErrorCode: 500,

		ShowRawError: s.Opts.RevealRawErr,
		RawError:     e.Error(),
	})
	if err != nil {
		log.Println("!", err)
		return
	}

	err = s.ShowPage(w, r, template.HTML(buf.String()), Head{Title: "Error"}, nil)
	if err != nil {
		log.Println("!", err)
		return
	}
}
