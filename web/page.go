package web

import (
	"net/http"
)

func (s *Site) Page(w http.ResponseWriter, r *http.Request) {
	s.WrapRun(w, r, "Serve-Page", func() error {
		return s.ShowPage(w, r, "a", TitleHead("Page"), nil)
	})
}
