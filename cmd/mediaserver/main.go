package main

import (
	"flag"
	"net/http"
	"os"

	"git.lan/wikithing/wtmedia/wtmediaserv"
)

func main() {
	cfg := wtmediaserv.Config{
		WritePass: os.Getenv("MEDIA_PASS"),
	}
	flag.StringVar(&cfg.Dir, "data", "./data/", "the data directory")
	if cfg.WritePass == "" {
		flag.StringVar(&cfg.WritePass, "pass", "", "a password to protect adding and removing files (optionally can use the $MEDIA_PASS envvar")
	}
	flag.IntVar(&cfg.CacheSize, "cachesize", 512, "the size of the LRU cache")
	flag.Parse()

	s, err := wtmediaserv.New(cfg)
	if err != nil {
		panic(err)
	}

	s.Init()
	if err = http.ListenAndServe(":5555", s.R); err != nil {
		panic(err)
	}
}
