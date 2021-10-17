package notmuch

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/emersion/go-imap/backend/backendutil"
	"github.com/stbenjam/go-imap-notmuch/pkg/maildir"
	notmuch "github.com/zenhack/go.notmuch"

	"github.com/emersion/go-imap"
)

type Mailbox struct {
	name     string
	query    string
	uidNext  uint32
	Messages []*Message
	maildir  string
}

func (mbox *Mailbox) Name() string {
	return mbox.name
}

func (mbox *Mailbox) Info() (*imap.MailboxInfo, error) {
	info := &imap.MailboxInfo{
		Delimiter: "/",
		Name:      mbox.name,
	}
	return info, nil
}

func (mbox *Mailbox) unseenSeqNum() uint32 {
	for i, msg := range mbox.Messages {
		seqNum := uint32(i + 1)

		seen := false
		for _, flag := range msg.Flags {
			if flag == imap.SeenFlag {
				seen = true
				break
			}
		}

		if !seen {
			return seqNum
		}
	}
	return 0
}

func (mbox *Mailbox) Status(items []imap.StatusItem) (*imap.MailboxStatus, error) {
	status := imap.NewMailboxStatus(mbox.name, items)
	status.PermanentFlags = []string{"\\*"}
	status.UnseenSeqNum = mbox.unseenSeqNum()

	for _, name := range items {
		switch name {
		case imap.StatusMessages:
			status.Messages = uint32(len(mbox.Messages))
		case imap.StatusUidNext:
			status.UidNext = mbox.uidNext
		case imap.StatusUidValidity:
			status.UidValidity = 1
		case imap.StatusRecent:
			status.Recent = 0 // TODO
		case imap.StatusUnseen:
			status.Unseen = 0 // TODO
		}
	}

	return status, nil
}

func (mbox *Mailbox) SetSubscribed(subscribed bool) error {
	return fmt.Errorf("unsupported operation")
}

func (mbox *Mailbox) Check() error {
	return fmt.Errorf("unsupported operation")
}

func (mbox *Mailbox) ListMessages(uid bool, seqSet *imap.SeqSet, items []imap.FetchItem, ch chan<- *imap.Message) error {
	defer close(ch)

	for i, msg := range mbox.Messages {
		seqNum := uint32(i + 1)

		var id uint32
		if uid {
			id = msg.Uid
		} else {
			id = seqNum
		}
		if !seqSet.Contains(id) {
			continue
		}

		m, err := msg.Fetch(seqNum, items)
		if err != nil {
			continue
		}

		ch <- m
	}

	return nil
}

func (mbox *Mailbox) SearchMessages(uid bool, criteria *imap.SearchCriteria) ([]uint32, error) {
	notmuchQuery := mbox.query
	notmuchQuery = fmt.Sprintf("%s %s", notmuchQuery, imapSearchToNotmuch(criteria))
	fmt.Fprintf(os.Stderr, "query is: %s\n", notmuchQuery)

	db, err := notmuch.Open(mbox.maildir, notmuch.DBReadOnly)
	if err != nil {
		return nil, fmt.Errorf("could not open mailbox: %s", err.Error())
	}
	defer db.Close()

	results, err := db.NewQuery(notmuchQuery).Messages()
	if err != nil {
		return nil, fmt.Errorf("could not search: %s", err.Error())
	}

	ids := make([]uint32, 0)
	for _, message := range mbox.Messages {
		if criteria.Uid != nil && !criteria.Uid.Contains(message.Uid) {
			continue
		}

		var m *notmuch.Message
		for results.Next(&m) {
			if message.ID == m.ID() {
				ids = append(ids, message.Uid)
			}
		}
	}

	return ids, nil
}

func (mbox *Mailbox) CreateMessage(flags []string, date time.Time, body imap.Literal) error {
	return nil
}

func (mbox *Mailbox) UpdateMessagesFlags(uid bool, seqset *imap.SeqSet, op imap.FlagsOp, flags []string) error {
	for i, msg := range mbox.Messages {
		var id uint32
		if uid {
			id = msg.Uid
		} else {
			id = uint32(i + 1)
		}
		if !seqset.Contains(id) {
			continue
		}

		db, err := notmuch.Open(mbox.maildir, notmuch.DBReadWrite)
		if err != nil {
			return fmt.Errorf("could not open mailbox: %s", err.Error())
		}
		defer db.Close()

		msg.Flags = backendutil.UpdateFlags(msg.Flags, op, flags)
		notMuchMessage, err := db.FindMessage(msg.ID)
		if err != nil {
			return err
		}

		if err := notMuchMessage.Atomic(func(m *notmuch.Message) {
			if err := m.RemoveAllTags(); err != nil {
				fmt.Fprintf(os.Stderr, "failed to remove tags from message %s: %s", m.ID(), err.Error())
				return
			}
			for _, tag := range msg.Tags() {
				if err := notMuchMessage.AddTag(tag); err != nil {
					fmt.Fprintf(os.Stderr, "failed to add tag to message %s: %s\n", m.ID(), err.Error())
					return
				}
			}
		}); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			continue
		}


		if err := notMuchMessage.TagsToMaildirFlags(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to convert tag to mail dir flags: %s\n", err.Error())
			continue
		}

		newFile := notMuchMessage.Filename()
		if err := notMuchMessage.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to close notmuch message %s: %s\n", msg.ID, err.Error())
		}

		msg.Filename = newFile
	}

	return nil
}

func (mbox *Mailbox) CopyMessages(uid bool, seqset *imap.SeqSet, destName string) error {
	return nil
}

func (mbox *Mailbox) Expunge() error {
	return nil
}

func (mbox *Mailbox) loadMessages() {
	db, err := notmuch.Open(mbox.maildir, notmuch.DBReadOnly)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not open mailbox: %s", err.Error())
		return
	}
	defer db.Close()

	if len(mbox.Messages) == 0 {
		messages := make([]*Message, 0)
		query := db.NewQuery(mbox.query)
		results, err := query.Messages()
		if err != nil {
			panic(err)
		}
		var message *notmuch.Message
		for results.Next(&message) {
			mbox.uidNext++

			f := message.Filename()
			s, err := os.Stat(f)
			if err != nil {
				fmt.Println(err)
				continue
			}

			imapFlags := make([]string, 0)
			maildirFlags := maildir.FlagFromFilename(f)
			for _, flag := range maildirFlags {
				if imapFlag := maildir.ImapFlagFromMaildir(flag); imapFlag != "" {
					imapFlags = append(imapFlags, imapFlag)
				}
			}

			messages = append(messages, &Message{
				ID:       message.ID(),
				Uid:      mbox.uidNext,
				Date:     message.Date(),
				Filename: f,
				Flags:    imapFlags,
				Size:     uint32(s.Size()),
			})
		}

		mbox.Messages = messages
	}
}

func imapSearchToNotmuch(criteria *imap.SearchCriteria) string {
	notmuchQuery := ""

	if len(criteria.Text) > 0 {
		notmuchQuery = strings.Join(criteria.Text, " ")
	}

	if values := criteria.Header.Values("From"); len(values) > 0 {
		for _, value := range values {
			notmuchQuery = fmt.Sprintf("%s from:%s", notmuchQuery, value)
		}
	}

	if values := criteria.Header.Values("Subject"); len(values) > 0 {
		for _, value := range values {
			notmuchQuery = fmt.Sprintf("%s subject:%s", notmuchQuery, value)
		}
	}

	for _, pair := range criteria.Or {
		notmuchQuery = fmt.Sprintf("%s and (%s or %s)", notmuchQuery, imapSearchToNotmuch(pair[0]), imapSearchToNotmuch(pair[1]))
	}

	return notmuchQuery
}
