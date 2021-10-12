package notmuch

import (
	"bufio"
	"bytes"
	"github.com/emersion/go-imap"
	notmuch "github.com/zenhack/go.notmuch"
	"io"
	"time"
)

type Message struct {
/*	Uid   uint32
	Date  time.Time
	Size  uint32
	Flags []string
	Body  []byte */
	notmuch *notmuch.Message
}

func (m *Message) Fetch(seqNum uint32, items []imap.FetchItem) (*imap.Message, error) {
	fetched := imap.NewMessage(seqNum, items)
	for _, item := range items {
		switch item {
		case imap.FetchEnvelope:
			fetched.Envelope = &imap.Envelope{
				Date: m.notmuch.Date(),
				Subject: m.notmuch.Header("Subject"),
				From: &imap.Address{
				},


			}
		case imap.FetchBody, imap.FetchBodyStructure:
			m.notmuch.

			hdr, body, _ := m.headerAndBody()
			fetched.BodyStructure, _ = backendutil.FetchBodyStructure(hdr, body, item == imap.FetchBodyStructure)
		case imap.FetchFlags:
			fetched.Flags = m.Flags
		case imap.FetchInternalDate:
			fetched.InternalDate = m.Date
		case imap.FetchRFC822Size:
			fetched.Size = m.Size
		case imap.FetchUid:
			fetched.Uid = m.Uid
		default:
			section, err := imap.ParseBodySectionName(item)
			if err != nil {
				break
			}

			body := bufio.NewReader(bytes.NewReader(m.Body))
			hdr, err := textproto.ReadHeader(body)
			if err != nil {
				return nil, err
			}

			l, _ := backendutil.FetchBodySection(hdr, body, section)
			fetched.Body[section] = l
		}
	}

	return fetched, nil
}

func (m *Message) Match(seqNum uint32, c *imap.SearchCriteria) (bool, error) {
	e, _ := m.entity()
	return backendutil.Match(e, seqNum, m.Uid, m.Date, m.Flags, c)
}
