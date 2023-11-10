package notmuch

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/stbenjam/go-imap-notmuch/pkg/uid"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/backend"
	notmuch "github.com/zenhack/go.notmuch"
	"golang.org/x/crypto/bcrypt"

	"github.com/stbenjam/go-imap-notmuch/pkg/config"
)

type Backend struct {
	user      *User
	uidMapper *uid.Mapper
}

func (b *Backend) Login(_ *imap.ConnInfo, username, password string) (backend.User, error) {
	err := bcrypt.CompareHashAndPassword([]byte(b.user.password), []byte(password))
	if err != nil {
		logrus.WithError(err).Errorf("couldn't compare hash and password")
	}
	if err == nil && username == b.user.username {
		return b.user, nil
	}

	return nil, fmt.Errorf("invalid username or password")
}

func New(cfg *config.Config) (*Backend, error) {
	db, err := notmuch.Open(cfg.Maildir, notmuch.DBReadWrite)
	if err != nil {
		return nil, err
	}
	db.Close()

	uidMapper, err := uid.New(cfg.UidDatabase)
	if err != nil {
		return nil, fmt.Errorf("couldn't open uid database: %w", err)
	}

	user := &User{
		username: cfg.Username,
		password: cfg.Password,
	}

	// Parse mailbox list from config file
	mailboxes := make(map[string]*Mailbox)
	for _, mailbox := range cfg.Mailboxes {
		attrs := make([]string, 0)
		for _, attr := range mailbox.Attributes {
			if attr[0] != '\\' {
				attrs = append(attrs, "\\"+attr)
			} else {
				attrs = append(attrs, attr)
			}
		}

		mailboxes[mailbox.Name] = &Mailbox{
			name:       mailbox.Name,
			query:      mailbox.Query,
			maildir:    cfg.Maildir,
			attributes: attrs,
			user:       user,
			lock:       &sync.RWMutex{},
			uidMapper:  uidMapper,
		}
	}

	user.mailboxes = mailboxes

	return &Backend{
		user:      user,
		uidMapper: uidMapper,
	}, nil
}
