package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"

	"github.com/stbenjam/go-imap-notmuch/pkg/config"
	"github.com/stbenjam/go-imap-notmuch/pkg/notmuch"

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

	s.Addr = ":9143"

	if cfg.TLSCertificate != "" && cfg.TLSKey != "" {
		certs, err := tls.LoadX509KeyPair(cfg.TLSCertificate, cfg.TLSKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not load certificates: %s", err.Error())
			os.Exit(1)
		}

		s.TLSConfig = &tls.Config{Certificates: []tls.Certificate{certs}}
	} else {
		s.AllowInsecureAuth = true
	}
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
