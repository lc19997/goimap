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

func (m *Message) headerAndBody() (textproto.Header, io.Reader, error) {
	f, err := os.Open(m.Filename)
	if err != nil {
		return textproto.Header{}, nil, err
	}
	body := bufio.NewReader(f)
	hdr, err := textproto.ReadHeader(body)
	if err != nil {
		return textproto.Header{}, nil, err
	}

	return hdr, body, nil
}

func (m *Message) Fetch(seqNum uint32, items []imap.FetchItem) (*imap.Message, error) {
	fetched := imap.NewMessage(seqNum, items)
	for _, item := range items {
		switch item {
		case imap.FetchEnvelope:
			hdr, _, err := m.headerAndBody()
			if err == nil {
				fetched.Envelope, _ = backendutil.FetchEnvelope(hdr)
			}
		case imap.FetchBody, imap.FetchBodyStructure:
			hdr, body, err := m.headerAndBody()
			if err == nil {
				fetched.BodyStructure, err = backendutil.FetchBodyStructure(hdr, body, item == imap.FetchBodyStructure)
				if err != nil {
					fetched.BodyStructure = &imap.BodyStructure{}
				}
			}
		case imap.FetchFlags:
			fetched.Flags = m.Flags
		case imap.FetchInternalDate:
			fetched.InternalDate = m.Date
		case imap.FetchRFC822Size:
			fetched.Size = m.Size
		case imap.FetchUid:
			fetched.Uid = m.Uid
		default:
			hdr, body, err := m.headerAndBody()
			if err != nil {
				break
			}
			section, err := imap.ParseBodySectionName(item)
			if err != nil {
				break
			}
			l, err := backendutil.FetchBodySection(hdr, body, section)
			if err == nil {
				fetched.Body[section] = l
			}
		}
	}

	return fetched, nil
}

func (m *Message) Match(seqNum uint32, c *imap.SearchCriteria) (bool, error) {
	_, body, err := m.headerAndBody()
	if err != nil {
		return false, err
	}
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
