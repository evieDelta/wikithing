package wtmediaserv

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"git.lan/wikithing/wterr"
	"git.lan/wikithing/wtmedia"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	// go:embed used
	_ "embed"
)

type Server struct {
	store *wtmedia.DefaultLocal

	R *chi.Mux

	writePass string
}

type Config struct {
	WritePass string
	Dir       string

	CacheSize int
}

func New(cfg Config) (*Server, error) {
	st, err := wtmedia.NewDefaultLocal(cfg.Dir, cfg.CacheSize)
	if err != nil {
		return nil, err
	}

	return &Server{
		store:     st,
		writePass: cfg.WritePass,
	}, nil
}

func (s *Server) Init() {
	s.R = chi.NewMux()
	s.R.Use(
		middleware.Logger,
		middleware.RequestID,
		middleware.Recoverer,
		middleware.Heartbeat("/ping"),
	)

	s.R.Get("/favicon.ico", s.favicon)
	s.R.Get("/{resource}", s.get)
	s.R.Route("/manage/", s.setupManage)
}

//go:embed CyubeA.png
var favicon []byte

func (s *Server) favicon(w http.ResponseWriter, r *http.Request) {
	w.Write(favicon)
}

func (s *Server) setupManage(r chi.Router) {
	r.Use(
		func(h http.Handler) http.Handler {
			fn := func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("x-auth") == s.writePass {
					h.ServeHTTP(w, r)
					return
				}
				w.WriteHeader(http.StatusForbidden)
				w.Header().Add("content-type", "text/plain")
				fmt.Fprint(w, "authorisation denied")
			}

			return http.HandlerFunc(fn)
		})

	r.Get("/{resource}", s.manageGet)
	r.Put("/{resource}", s.managePut)
	r.Delete("/{resource}", s.manageDelete)
}

type sendErr struct {
	Content string
	Code    int
}

func (s *Server) runWrap(w http.ResponseWriter, r *http.Request, fn func(w http.ResponseWriter, r *http.Request) error) {
	err := fn(w, r)
	if err != nil {
		log.Println(err)

		// yeah i'm using xml for errors but not the data
		// don't ask (makes it really easy to tell there was an error though when the errors are in a totally different language)
		w.Header().Set("content-type", "encoding/xml")

		if os.IsNotExist(err) {
			w.WriteHeader(404)
			xml.NewEncoder(w).Encode(sendErr{
				Content: "requested resource does not exist",
			})
			return
		}

		wer, ok := err.(wterr.Err)
		if !ok {
			w.WriteHeader(500)
			xml.NewEncoder(w).Encode(sendErr{
				Content: "Something went wrong",
			})
			return
		}
		msg := ""
		switch wer.Type {
		case wterr.ErrUnknown:
			w.WriteHeader(500)
			msg = "Something went wrong"
		case wterr.ErrError:
			w.WriteHeader(500)
			msg = "Something went wrong"
		case wterr.ErrAuthFailed:
			w.WriteHeader(http.StatusForbidden)
			msg = wer.Error()
		case wterr.ErrInvalidInput:
			w.WriteHeader(http.StatusBadRequest)
			msg = wer.Error()
		case wterr.ErrUnsupported:
			w.WriteHeader(http.StatusNotImplemented)
			msg = wer.Error()
		}

		xml.NewEncoder(w).Encode(sendErr{
			Content: msg,
			Code:    int(wer.Type),
		})
	}
}

func (s *Server) get(w http.ResponseWriter, r *http.Request) {
	s.runWrap(w, r, func(w http.ResponseWriter, r *http.Request) error {
		p := chi.URLParam(r, "resource")
		q := strings.Split(p, ".")
		h := q[0]
		e := ""
		if len(q) > 1 {
			e = q[1]
		}

		dat, mime, err := s.store.Get(h, wtmedia.QueryData{
			Extension: e,

			Values: r.URL.Query(),
		})
		if err != nil {
			return err
		}
		r.Header.Set("content-type", mime)
		_, err = io.Copy(w, dat)
		return err
	})
}

func (s *Server) manageGet(w http.ResponseWriter, r *http.Request) {
	s.runWrap(w, r, func(w http.ResponseWriter, r *http.Request) error {
		p := chi.URLParam(r, "resource")

		meta, err := s.store.GetMeta(p)
		if err != nil {
			return err
		}

		w.Header().Set("content-type", "encoding/json")
		return json.NewEncoder(w).Encode(meta)
	})
}

func (s *Server) managePut(w http.ResponseWriter, r *http.Request) {
	s.runWrap(w, r, func(w http.ResponseWriter, r *http.Request) error {
		p := chi.URLParam(r, "resource")

		kind := wtmedia.KindBinary
		switch r.Header.Get("content-type") {
		case "image/png", "image/jpeg", "image/tiff", "image/webp", "image/bmp":
			kind = wtmedia.KindImage
		}

		dat, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}

		meta := make(map[string]string, len(r.URL.Query()))
		bld := &strings.Builder{}
		for k, v := range r.URL.Query() {
			bld.Reset()
			uk, err := url.QueryUnescape(k)
			if err != nil {
				return err
			}

			c := csv.NewWriter(bld)
			err = c.Write(v)
			if err != nil {
				return err
			}
			c.Flush()
			err = c.Error()
			if err != nil {
				return err
			}

			uv, err := url.QueryUnescape(bld.String())
			if err != nil {
				return err
			}

			meta[uk] = uv
		}

		return s.store.Put(p, wtmedia.TypeMeta{
			Kind: kind,
			Mime: r.Header.Get("content-type"),

			Meta: meta,
		}, dat)
	})
}

func (s *Server) manageDelete(w http.ResponseWriter, r *http.Request) {
	s.runWrap(w, r, func(w http.ResponseWriter, r *http.Request) error {
		p := chi.URLParam(r, "resource")

		return s.store.Rem(p)
	})
}
