package main

import (
	"flag"
	"log"

	"git.lan/wikithing/web"
)

func main() {
	var templd, staticd, datad, url string
	flag.StringVar(&templd, "templ", "", "override the page generation templates")
	flag.StringVar(&staticd, "static", "", "override the static resources dir")
	flag.StringVar(&datad, "data", "./data/", "override the data directory")
	flag.StringVar(&url, "url", ":7380", "the url and port to run off of")
	flag.Parse()

	s := web.Site{}
	err := s.Initialise(web.Options{
		TemplateDir: templd,
		StaticDir:   staticd,
		DataDir:     datad,

		RevealRawErr: true,
	})
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Running on", url)
	err = s.Run(url)
	log.Println(err)
}
