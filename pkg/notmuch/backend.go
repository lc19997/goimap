package notmuch

import (
	"fmt"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/backend"
	"github.com/stbenjam/go-imap-notmuch/pkg/config"
	"github.com/zenhack/go.notmuch"
	"golang.org/x/crypto/bcrypt"
)

type Backend struct {
	user User
	db   *notmuch.DB
}

func (b *Backend) Login(_ *imap.ConnInfo, username, password string) (backend.User, error) {
	err := bcrypt.CompareHashAndPassword([]byte(b.user.password), []byte(password))
	if err != nil {
		fmt.Println(err)
	}
	if err == nil && username == b.user.username {
		return &b.user, nil
	}

	return nil, fmt.Errorf("invalid username or password")
}

func New(cfg *config.Config) (*Backend, error) {
	db, err := notmuch.Open(cfg.Maildir, notmuch.DBReadWrite)
	if err != nil {
		return nil, err
	}

	mailboxes := make(map[string]*Mailbox)
	for _, mailbox := range cfg.Mailboxes {
		mailboxes[mailbox.Name] = &Mailbox{
			name:  mailbox.Name,
			query: mailbox.Query,
			db: db,
			maildir: cfg.Maildir,
		}
		mailboxes[mailbox.Name].loadMessages()
	}
	return &Backend{
		user: User{
			db:        db,
			username:  cfg.Username,
			password:  cfg.Password,
			mailboxes: mailboxes,
		},
		db: db,
	}, nil
}
