package notmuch

import (
	"fmt"
	"github.com/emersion/go-imap/backend"
	notmuch "github.com/zenhack/go.notmuch"
)

type User struct {
	username  string
	password  string
	db        *notmuch.DB
	mailboxes map[string]*Mailbox
}

func (u *User) Username() string {
	return u.username
}

func (u *User) ListMailboxes(subscribed bool) (mailboxes []backend.Mailbox, err error) {
	mailboxes = make([]backend.Mailbox, 0)
	for _, v := range u.mailboxes {
		mailboxes = append(mailboxes, v)
	}

	return mailboxes, nil
}

func (u *User) GetMailbox(name string) (mailbox backend.Mailbox, err error) {
	if mbox, ok := u.mailboxes[name]; ok {
		return mbox, nil
	}

	return nil, fmt.Errorf("not found")
}

func (u *User) CreateMailbox(name string) error {
	return fmt.Errorf("unsupported operation")
}

func (u *User) DeleteMailbox(name string) error {
	return fmt.Errorf("unsupported operation")
}

func (u *User) RenameMailbox(existingName, newName string) error {
	return fmt.Errorf("unsupported operation")
}

func (u *User) Logout() error {
	return nil
}
