package notmuch

import (
	"fmt"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/backend"
	"github.com/stbenjam/go-imap-notmuch/pkg/config"
	"github.com/zenhack/go.notmuch"
)

type Backend struct {
	user User
	db *notmuch.DB
}

func (b *Backend) Login(_ *imap.ConnInfo, username, password string) (backend.User, error) {
	if username == "notmuch" && password == "notmuch" {
		return &b.user, nil
	}

	return nil, fmt.Errorf("invalid username or password")
}

func New(cfg *config.Config) (*Backend, error) {
	db, err := notmuch.Open(cfg.Maildir, notmuch.DBReadOnly)
	if err != nil {
		return nil, err
	}

	return &Backend{
		user: User{
			username: cfg.Username,
		},
		db: db,
	}, nil
}