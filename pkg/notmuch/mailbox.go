package notmuch

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"

	"github.com/emersion/go-imap/backend/backendutil"
	notmuch "github.com/zenhack/go.notmuch"

	"github.com/stbenjam/go-imap-notmuch/pkg/maildir"
	"github.com/stbenjam/go-imap-notmuch/pkg/uid"

	"github.com/emersion/go-imap"
)

type Mailbox struct {
	lock *sync.RWMutex

	Messages []*Message

	name       string
	maildir    string
	query      string
	user       *User
	attributes []string

	// Counts
	total       uint32
	recent      uint32
	unseen      uint32
	lastUpdated time.Time

	// All messages in a mailbox must be identified by
	// an unchanging UID (unless UID validity changes)
	uidMapper uid.Mapper
}

func (mbox *Mailbox) Name() string {
	return mbox.name
}

func (mbox *Mailbox) Info() (*imap.MailboxInfo, error) {
	info := &imap.MailboxInfo{
		Attributes: mbox.attributes,
		Delimiter:  "/",
		Name:       mbox.name,
	}
	return info, nil
}

func (mbox *Mailbox) Expire() {
	mbox.lock.Lock()
	defer mbox.lock.Unlock()
	mbox.Messages = nil
	mbox.total = 0
	mbox.unseen = 0
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
	mbox.loadCounts()

	status := imap.NewMailboxStatus(mbox.name, items)
	status.PermanentFlags = []string{"\\*"}
	status.UnseenSeqNum = mbox.unseenSeqNum()

	for _, name := range items {
		switch name {
		case imap.StatusMessages:
			status.Messages = mbox.total
		case imap.StatusUidNext:
			status.UidNext = 0 // don't support this
		case imap.StatusUidValidity:
			status.UidValidity = mbox.uidMapper.Validity()
		case imap.StatusRecent:
			status.Recent = mbox.recent
		case imap.StatusUnseen:
			status.Unseen = mbox.unseen
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
	mbox.loadMessages()
	mbox.lock.Lock()
	defer mbox.lock.Unlock()

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
	mbox.loadMessages()

	notmuchQuery, err := IMAPSearchToNotmuch(criteria, true)
	if err != nil {
		return nil, err
	}
	notmuchQuery = mbox.query + " " + notmuchQuery
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

	var m *notmuch.Message
	resultIDs := make(map[string]struct{})
	for results.Next(&m) {
		resultIDs[m.ID()] = struct{}{}
	}

	ids := make([]uint32, 0)
	for i, message := range mbox.Messages {
		if criteria.SeqNum != nil && !criteria.SeqNum.Contains(uint32(i)) {
			continue
		}

		if criteria.Uid != nil && !criteria.Uid.Contains(message.Uid) {
			continue
		}

		if _, ok := resultIDs[message.ID]; !ok {
			continue
		}

		ids = append(ids, message.Uid)
	}

	return ids, nil
}

func (mbox *Mailbox) CreateMessage(flags []string, date time.Time, body imap.Literal) error {
	mbox.lock.Lock()
	defer mbox.lock.Unlock()

	if date.IsZero() {
		date = time.Now()
	}

	filename, err := mbox.newMessageKey()
	if err != nil {
		return err
	}

	maildirFlags := make([]maildir.Flag, 0)
	for _, flag := range flags {
		if mdt := maildir.MaildirFlagFromImap(flag); mdt != 0 {
			maildirFlags = append(maildirFlags, mdt)
		}
	}

	if len(maildirFlags) > 0 {
		filename += ":2,"
		for _, mdt := range maildirFlags {
			filename += string(mdt)
		}
		filename = path.Join(mbox.maildir, mbox.name, "cur", filename)
	} else {
		filename = path.Join(mbox.maildir, mbox.name, "new", filename)
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	if _, err := f.Write(b); err != nil {
		return err
	}

	message := &Message{
		Date:     date,
		Size:     uint32(len(b)),
		Flags:    flags,
		Filename: filename,
	}

	// Add message
	db, err := notmuch.Open(mbox.maildir, notmuch.DBReadWrite)
	if err != nil {
		return err
	}
	defer db.Close()
	nmm, err := db.AddMessage(filename)
	if err != nil {
		return err
	}
	message.Uid = mbox.uidMapper.FindOrAdd(nmm.ID())
	db.Close()

	mbox.Messages = append(mbox.Messages, message)
	return nil
}

func (mbox *Mailbox) UpdateMessagesFlags(uid bool, seqset *imap.SeqSet, op imap.FlagsOp, flags []string) error {
	mbox.loadMessages()
	mbox.lock.Lock()
	defer mbox.lock.Unlock()

	db, err := notmuch.Open(mbox.maildir, notmuch.DBReadWrite)
	if err != nil {
		return fmt.Errorf("could not open mailbox: %s", err.Error())
	}
	defer db.Close()

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
				fmt.Printf("Adding tag %s\n", tag)
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
	return fmt.Errorf("not implemented")
}

func (mbox *Mailbox) MoveMessages(uid bool, seqset *imap.SeqSet, dest string) error {
	mbox.loadMessages()

	mbox.lock.Lock()
	defer mbox.lock.Unlock()
	db, err := notmuch.Open(mbox.maildir, notmuch.DBReadWrite)
	if err != nil {
		return fmt.Errorf("could not open mailbox: %s", err.Error())
	}
	defer db.Close()

	if _, err := os.Stat(path.Join(mbox.maildir, dest)); os.IsNotExist(err) {
		return fmt.Errorf("could not find destination: %s", err.Error())
	}

	newMessages := mbox.Messages[:0]

	for i, msg := range mbox.Messages {
		var id uint32
		if uid {
			id = msg.Uid
		} else {
			id = uint32(i + 1)
		}
		if !seqset.Contains(id) {
			newMessages = append(newMessages, msg)
			continue
		}

		message, err := db.FindMessageByFilename(msg.Filename)
		if err != nil {
			fmt.Fprint(os.Stderr, err.Error())
			continue
		}

		unread := false
		var t *notmuch.Tag
		tags := message.Tags()
		for tags.Next(&t) {
			if t.Value == "unread" {
				unread = true
			}
		}

		destPath := path.Join(mbox.maildir, dest, "cur", path.Base(message.Filename()))
		if unread {
			destPath = path.Join(mbox.maildir, dest, "new", path.Base(message.Filename()))
		}

		if err := os.Rename(message.Filename(), destPath); err != nil {
			fmt.Fprint(os.Stderr, err.Error())
			continue
		}

		if err := db.RemoveMessage(message.Filename()); err != nil {
			return err
		}
		if _, err = db.AddMessage(destPath); err != nil {
			return err
		}
		if destBox, ok := mbox.user.mailboxes[dest]; ok {
			destBox.Expire() // Expire any cached messages
		}
	}

	mbox.Messages = newMessages
	return nil
}

func (mbox *Mailbox) Expunge() error {
	mbox.lock.Lock()
	defer mbox.lock.Unlock()

	db, err := notmuch.Open(mbox.maildir, notmuch.DBReadWrite)
	if err != nil {
		return err
	}
	defer db.Close()

	newMessages := mbox.Messages[:0]
	for _, message := range mbox.Messages {
		deleted := false
		for _, flag := range message.Flags {
			if flag == imap.DeletedFlag {
				deleted = true
				break
			}
		}

		if deleted {
			fmt.Printf("Removing %s\n", message.ID)
			mbox.uidMapper.Remove(message.ID)
			if err := db.RemoveMessage(message.Filename); err != nil {
				return err
			}
		} else {
			newMessages = append(newMessages, message)
		}
	}

	mbox.Messages = newMessages
	return nil
}

func (mbox *Mailbox) loadCounts() {
	db, err := notmuch.Open(mbox.maildir, notmuch.DBReadOnly)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not open mailbox: %s", err.Error())
		return
	}
	defer db.Close()

	mbox.lock.Lock()
	defer mbox.lock.Unlock()
	mbox.total = uint32(db.NewQuery(mbox.query).CountMessages())
	mbox.recent = uint32(db.NewQuery(fmt.Sprintf("%s tag:new", mbox.query)).CountMessages())
	mbox.unseen = uint32(db.NewQuery(fmt.Sprintf("%s tag:unread", mbox.query)).CountMessages())
}

func (mbox *Mailbox) loadMessages() {
	db, err := notmuch.Open(mbox.maildir, notmuch.DBReadOnly)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not open mailbox: %s", err.Error())
		return
	}
	defer db.Close()

	mbox.lock.Lock()
	defer mbox.lock.Unlock()
	stat, err := os.Stat(path.Join(db.Path(), ".notmuch", "xapian", "flintlock"))
	needsUpdate := err != nil || mbox.lastUpdated.Before(stat.ModTime())
	mbox.lastUpdated = stat.ModTime()
	mbox.recent = uint32(db.NewQuery(fmt.Sprintf("%s tag:new", mbox.query)).CountMessages())
	mbox.unseen = uint32(db.NewQuery(fmt.Sprintf("%s tag:unread", mbox.query)).CountMessages())

	if len(mbox.Messages) > 0 && !needsUpdate {
		return
	}

	messages := make([]*Message, 0)
	query := db.NewQuery(mbox.query)
	results, err := query.Messages()
	if err != nil {
		panic(err)
	}
	var message *notmuch.Message
	for results.Next(&message) {
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
			Uid:      mbox.uidMapper.FindOrAdd(message.ID()),
			Date:     message.Date(),
			Filename: f,
			Flags:    imapFlags,
			Size:     uint32(s.Size()),
		})
	}

	mbox.Messages = messages
}
