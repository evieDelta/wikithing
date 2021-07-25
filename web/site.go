package web

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"os"

	"git.lan/wikithing/wtfs"
	"github.com/go-chi/chi"
)

//go:embed static/*
var static embed.FS

//go:embed templates
var templates embed.FS

type Options struct {
	TemplateDir string
	StaticDir   string
	DataDir     string

	RevealRawErr bool
}

type Site struct {
	Opts Options

	Wiki *wtfs.Filesystem
	R    *chi.Mux

	templates *template.Template

	TemplFS  fs.FS
	StaticFS fs.FS
}

func (r *Site) Initialise(opts Options) error {
	r.Opts = opts

	if opts.StaticDir == "" {
		sfs, err := fs.Sub(static, "static")
		if err != nil {
			return err
		}
		r.StaticFS = sfs
	} else {
		r.StaticFS = os.DirFS(opts.StaticDir)
	}

	if opts.TemplateDir == "" {
		sfs, err := fs.Sub(templates, "templates")
		if err != nil {
			return err
		}
		r.TemplFS = sfs
	} else {
		r.TemplFS = os.DirFS(opts.TemplateDir)
	}
	err := r.loadTemplates()
	if err != nil {
		return err
	}

	if opts.StaticDir == "" {
		opts.StaticDir = "./data/"
	}
	r.Wiki, err = wtfs.New(opts.StaticDir)
	if err != nil {
		return err
	}

	return r.SetupChi()
}

func (r *Site) Run(url string) error {
	return http.ListenAndServe(url, r.R)
}

func (r *Site) loadTemplates() error {
	t, err := template.New("root").Funcs(templFuncs).ParseFS(r.TemplFS, "*.html")
	if err != nil {
		return err
	}
	r.templates = t

	return nil
}

var templFuncs = template.FuncMap{
	"uwu": func(a string) string { return "uwuth doth be doth" },
}
