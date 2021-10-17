package notmuch

import (
	"fmt"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/backend"
	"github.com/stbenjam/go-imap-notmuch/pkg/config"
	notmuch "github.com/zenhack/go.notmuch"
	"golang.org/x/crypto/bcrypt"
	"sync"
)

type Backend struct {
	user *User
}

func (b *Backend) Login(_ *imap.ConnInfo, username, password string) (backend.User, error) {
	err := bcrypt.CompareHashAndPassword([]byte(b.user.password), []byte(password))
	if err != nil {
		fmt.Println(err)
	}
	if err == nil && username == b.user.username {
		return b.user, nil
	}

	return nil, fmt.Errorf("invalid username or password")
}

func New(cfg *config.Config) (*Backend, error) {
	// Open the DB just to make sure all is OK
	db, err := notmuch.Open(cfg.Maildir, notmuch.DBReadWrite)
	if err != nil {
		return nil, err
	}
	db.Close()

	user := &User{
		username: cfg.Username,
		password: cfg.Password,
	}

	// Parse mailbox list from config file
	mailboxes := make(map[string]*Mailbox)
	for _, mailbox := range cfg.Mailboxes {
		mailboxes[mailbox.Name] = &Mailbox{
			name:    mailbox.Name,
			query:   mailbox.Query,
			maildir: cfg.Maildir,
			user:    user,
			lock:    &sync.RWMutex{},
		}
	}

	user.mailboxes = mailboxes

	return &Backend{
		user: user,
	}, nil
}
