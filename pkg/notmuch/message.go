package notmuch

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/backend/backendutil"
	"github.com/emersion/go-message"
	"github.com/emersion/go-message/textproto"
)

type Message struct {
	ID       string
	Uid      uint32
	Date     time.Time
	Size     uint32
	Flags    []string
	Filename string
}

func (m *Message) headerAndBody() (*os.File, textproto.Header, io.Reader, error) {
	f, err := os.Open(m.Filename)
	if err != nil {
		return nil, textproto.Header{}, nil, err
	}
	body := bufio.NewReader(f)
	hdr, err := textproto.ReadHeader(body)
	if err != nil {
		return nil, textproto.Header{}, nil, err
	}

	return f, hdr, body, nil
}

func (m *Message) Fetch(seqNum uint32, items []imap.FetchItem) (*imap.Message, error) {
	fetched := imap.NewMessage(seqNum, items)
	for _, item := range items {
		switch item {
		case imap.FetchEnvelope:
			f, hdr, _, err := m.headerAndBody()
			if err != nil {
				continue
			}
			fetched.Envelope, _ = backendutil.FetchEnvelope(hdr)
			f.Close()
		case imap.FetchBody, imap.FetchBodyStructure:
			f, hdr, body, err := m.headerAndBody()
			if err != nil {
				continue
			}
			fetched.BodyStructure, err = backendutil.FetchBodyStructure(hdr, body, item == imap.FetchBodyStructure)
			if err != nil {
				fetched.BodyStructure = &imap.BodyStructure{}
			}
			f.Close()
		case imap.FetchFlags:
			fetched.Flags = m.Flags
		case imap.FetchInternalDate:
			fetched.InternalDate = m.Date
		case imap.FetchRFC822Size:
			fetched.Size = m.Size
		case imap.FetchUid:
			fetched.Uid = m.Uid
		default:
			f, hdr, body, err := m.headerAndBody()
			if err != nil {
				continue
			}
			section, err := imap.ParseBodySectionName(item)
			if err != nil {
				break
			}
			l, err := backendutil.FetchBodySection(hdr, body, section)
			if err == nil {
				fetched.Body[section] = l
			}
			f.Close()
		}
	}

	return fetched, nil
}

func (m *Message) Match(seqNum uint32, c *imap.SearchCriteria) (bool, error) {
	f, _, body, err := m.headerAndBody()
	if err != nil {
		return false, err
	}
	defer f.Close()
	e, err := message.Read(body)
	if err != nil {
		return false, err
	}

	return backendutil.Match(e, seqNum, m.Uid, m.Date, m.Flags, c)
}

func (m *Message) Tags() []string {
	tags := make([]string, 0)
	for _, flag := range m.Flags {
		switch flag {
		case imap.SeenFlag:
			tags = append(tags, "seen")
		case imap.FlaggedFlag:
			tags = append(tags, "flagged")
		case imap.DraftFlag:
			tags = append(tags, "draft")
		case imap.AnsweredFlag:
			tags = append(tags, "replied")
		case imap.DeletedFlag:
			tags = append(tags, "deleted")
		}
	}

	fmt.Println(tags)
	return tags
}
