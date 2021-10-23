package notmuch

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/emersion/go-imap"
	"github.com/stbenjam/go-imap-notmuch/pkg/maildir"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

// IMAPSearchToNotmuch converts an *imap.SearchCriteria to a notmuch query string. As
// notmuch doesn't index sizes, we can't search by them, and an error will be returned.
func IMAPSearchToNotmuch(s *imap.SearchCriteria, topLevel bool) (string, error) {
	notmuchQuery := ""

	for _, text := range s.Text {
		notmuchQuery += fmt.Sprintf(" %q ", text)
		notmuchQuery = strings.TrimSpace(notmuchQuery)
	}

	for _, text := range s.Body {
		notmuchQuery += fmt.Sprintf(" body:%q ", text)
		notmuchQuery = strings.TrimSpace(notmuchQuery)
	}

	for _, header := range []string{"from", "subject", "to", "cc"} {
		if values := s.Header.Values(header); len(values) > 0 {
			for _, value := range values {
				notmuchQuery = fmt.Sprintf("%s %s:%q", notmuchQuery, header, value)
				notmuchQuery = strings.TrimSpace(notmuchQuery)
			}
		}
	}

	if flags := s.WithFlags; len(flags) > 0 {
		for _, flag := range flags {
			notmuchQuery = fmt.Sprintf("%s %s", notmuchQuery, maildir.NotmuchFlagFromImap(false, flag))
			notmuchQuery = strings.TrimSpace(notmuchQuery)
		}
	}

	if flags := s.WithoutFlags; len(flags) > 0 {
		for _, flag := range flags {
			notmuchQuery = fmt.Sprintf("%s %s", notmuchQuery, maildir.NotmuchFlagFromImap(true, flag))
			notmuchQuery = strings.TrimSpace(notmuchQuery)
		}
	}

	before := s.Before
	since := s.Since

	if !before.IsZero() && !since.IsZero() {
		notmuchQuery = fmt.Sprintf("%s date:@%d..@%d", notmuchQuery, since.Unix(), before.Unix())
		notmuchQuery = strings.TrimSpace(notmuchQuery)
	} else if !before.IsZero() {
		notmuchQuery = fmt.Sprintf("%s date:..@%d", notmuchQuery, before.Unix())
		notmuchQuery = strings.TrimSpace(notmuchQuery)
	} else if !since.IsZero() {
		notmuchQuery = fmt.Sprintf("%s date:@%d..", notmuchQuery, since.Unix())
		notmuchQuery = strings.TrimSpace(notmuchQuery)
	}

	sentBefore := s.SentBefore
	sentSince := s.SentSince

	if !sentBefore.IsZero() && !sentSince.IsZero() {
		notmuchQuery = fmt.Sprintf("%s date:@%d..@%d", notmuchQuery, sentSince.Unix(), sentBefore.Unix())
		notmuchQuery = strings.TrimSpace(notmuchQuery)
	} else if !sentBefore.IsZero() {
		notmuchQuery = fmt.Sprintf("%s date:..@%d", notmuchQuery, sentBefore.Unix())
		notmuchQuery = strings.TrimSpace(notmuchQuery)
	} else if !sentSince.IsZero() {
		notmuchQuery = fmt.Sprintf("%s date:@%d..", notmuchQuery, sentSince.Unix())
		notmuchQuery = strings.TrimSpace(notmuchQuery)
	}

	if topLevel && len(s.Or) > 0 {
		notmuchQuery = notmuchQuery + " ("
	}

	for _, pair := range s.Or {
		first, err := IMAPSearchToNotmuch(pair[0], false)
		if err != nil {
			return "", err
		}

		second, err := IMAPSearchToNotmuch(pair[1], false)
		if err != nil {
			return "", err
		}

		notmuchQuery = fmt.Sprintf("%s %s or %s", notmuchQuery, strings.TrimSpace(first), strings.TrimSpace(second))
		notmuchQuery = strings.TrimSpace(notmuchQuery)
	}

	if topLevel && len(s.Or) > 0 {
		notmuchQuery = notmuchQuery + " ) "
	}

	for _, not := range s.Not {
		result, err := IMAPSearchToNotmuch(not, false)
		if err != nil {
			return "", err
		}
		notmuchQuery = fmt.Sprintf("%s not %s", notmuchQuery, result)
		notmuchQuery = strings.TrimSpace(notmuchQuery)
	}

	if s.Larger > 0 || s.Smaller > 0 {
		return "", fmt.Errorf("size parameter was requested; this is currently unsupported by the notmuch backend")
	}

	return strings.TrimSpace(notmuchQuery), nil
}

// https://github.com/emersion/go-maildir/blob/ced1977bfb902354b2a8bfbf7517b8d6ba07cccb/maildir.go#L311
func (mbox *Mailbox) newMessageKey() (string, error) {
	var key string
	key += strconv.FormatInt(time.Now().Unix(), 10)
	key += "."
	host, err := os.Hostname()
	if err != nil {
		return "", err
	}
	host = strings.Replace(host, "/", "\057", -1)
	key += host
	key += "."
	key += strconv.FormatInt(int64(os.Getpid()), 10)
	bs := make([]byte, 10)
	_, err = io.ReadFull(rand.Reader, bs)
	if err != nil {
		return "", err
	}
	key += hex.EncodeToString(bs)
	return key, nil
}
