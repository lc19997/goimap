package main

import (
	"fmt"
	"github.com/stbenjam/go-imap-notmuch/pkg/config"
	"github.com/stbenjam/go-imap-notmuch/pkg/notmuch"
	"log"
	"os"

	"github.com/emersion/go-imap/server"
)

func main() {
	var cfg *config.Config
	var err error

	if len(os.Args) > 1 {
		cfg, err = config.New(os.Args[1])
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Fprintf(os.Stderr, "usage: %s <config file>\n", os.Args[0])
		os.Exit(1)
	}


	be, err := notmuch.New(cfg)
	if err != nil {
		panic(err)
	}

	s := server.New(be)
	s.Debug = os.Stderr

	s.Addr = ":9943"
	s.AllowInsecureAuth = true
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}