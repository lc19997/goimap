package main

import (
	"net/textproto"
	"os"
	"path"
	"runtime"
	"testing"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-imap/server"
	"golang.org/x/crypto/bcrypt"

	"github.com/stbenjam/go-imap-notmuch/pkg/config"
	"github.com/stbenjam/go-imap-notmuch/pkg/notmuch"
)

func TestNotmuchIMAP(t *testing.T) {
	var c *client.Client
	var err error
	t.Run("can connect to imap server", func(t *testing.T) {
		setupIMAPServer(t)
		c, err = client.Dial("localhost:6143")
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("notmuch user can login", func(t *testing.T) {
		// Check login
		err = c.Login("notmuch", "notmuch")
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("can find inbox", func(*testing.T) {
		// Check mailboxes
		mailboxes := make(chan *imap.MailboxInfo, 1)
		done := make(chan error, 1)
		done <- c.List("", "*", mailboxes)
		found := false
		for m := range mailboxes {
			if m.Name == "INBOX" {
				found = true
				break
			}
		}
		if !found {
			t.Fatal("Did not find INBOX")
		}
	})

	t.Run("can get last message from mailbox", func(t *testing.T) {
		mbox, err := c.Select("INBOX", false)
		if err != nil {
			t.Fatal(err)
		}
		i := mbox.UidNext - 1
		seqset := new(imap.SeqSet)
		seqset.AddRange(i, i)
		messages := make(chan *imap.Message, 10)
		done := make(chan error, 1)
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
		for msg := range messages {
			if msg.Envelope.Subject == "[notmuch] [PATCH 1/2] Close message file after parsing message headers" {
				return
			}
		}

		t.Error("couldn't load last message from mailbox")
	})

	t.Run("can search mailbox", func(t *testing.T) {
		_, err := c.Select("INBOX", false)
		if err != nil {
			t.Fatal(err)
		}

		after, _ := time.Parse("2006-01-02T15:04:05.000Z", "2009-11-18T00:00:00.000Z")
		before, _ := time.Parse("2006-01-02T15:04:05.000Z", "2010-12-30T00:00:00.000Z")
		seqNums, err := c.Search(&imap.SearchCriteria{
			Since:        after,
			Before:       before,
			WithoutFlags: []string{imap.SeenFlag},
			Or: [][2]*imap.SearchCriteria{
				{
					{
						Body: []string{"Makefile"},
					},
					{
						Header: textproto.MIMEHeader{
							"subject": []string{"[PATCH]"},
						},
					},
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(seqNums) != 15 {
			t.Fatalf("expected 15 messages to match search, found %d", len(seqNums))
		}

	})

	t.Run("can logout", func(*testing.T) {
		if err := c.Logout(); err != nil {
			t.Error(err.Error())
		}
	})
}

func setupIMAPServer(t *testing.T) *server.Server {
	_, currentPath, _, ok := runtime.Caller(1)
	if !ok {
		t.Fatalf("could not get current path")
	}
	dir := path.Join(path.Dir(currentPath), "fixtures", "database-v1")
	hash, err := bcrypt.GenerateFromPassword([]byte("notmuch"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		Maildir:  dir,
		Username: "notmuch",
		Password: string(hash),
		Mailboxes: []config.Mailbox{
			{
				Name:  "INBOX",
				Query: "is:unread",
			},
		},
		Debug:       true,
		UidDatabase: path.Join(dir, "uid.dat"),
	}
	cfg.SetDefaults()

	backend, err := notmuch.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	s := server.New(backend)
	s.AllowInsecureAuth = true
	s.Addr = ":6143"
	s.Debug = os.Stderr
	go func() {
		if err := s.ListenAndServe(); err != nil {
			panic(err)
		}
	}()
	return s
}
