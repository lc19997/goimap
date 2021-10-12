package notmuch

import (
	"fmt"
	"github.com/emersion/go-imap/backend"
)

type User struct {
	username string
}

func (u *User) Username() string {
	return u.username
}

func (u *User) ListMailboxes(subscribed bool) (mailboxes []backend.Mailbox, err error) {
	return []backend.Mailbox{
		&Mailbox{
			name: "INBOX",
		},
	}, nil
}

func (u *User) GetMailbox(name string) (mailbox backend.Mailbox, err error) {
	return
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