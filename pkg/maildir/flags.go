package maildir

import (
	"github.com/emersion/go-imap"
	"strings"
)

type Flag rune

// Reference https://cr.yp.to/proto/maildir.html
const (
	FlagPassed  Flag = 'P'
	FlagReplied Flag = 'R'
	FlagSeen    Flag = 'S'
	FlagTrashed Flag = 'T'
	FlagDraft   Flag = 'D'
	FlagFlagged Flag = 'F'
)

func ImapFlagFromMaildir(maildir Flag) string {
	switch(maildir) {
	case FlagReplied:
		return imap.AnsweredFlag
	case FlagSeen:
		return imap.SeenFlag
	case FlagTrashed:
		return imap.DeletedFlag
	case FlagDraft:
		return imap.DraftFlag
	case FlagFlagged:
		return imap.FlaggedFlag
	default:
		return ""
	}
}

func NotmuchFlagFromImap(imapFlag string) string {
	switch(imapFlag) {
	case imap.AnsweredFlag:
		return "tag:replied"
	case imap.SeenFlag:
		return "not tag:read"
	case imap.DraftFlag:
		return "tag:draft"
	case imap.FlaggedFlag:
		return "tag:flagged"
	default:
		return ""
	}
}

func FlagFromFilename(filename string) []Flag {
	flags := make([]Flag, 0)
	parts := strings.Split(filename, ":2,")
	if len(parts) != 2 {
		return nil
	}

	for _, flag := range parts[1] {
		flags = append(flags, Flag(flag))
	}
	return flags
}