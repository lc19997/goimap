# go-imap-notmuch

:bangbang: | This software is experimental, use at your own risk. Please make a backup of your Maildir and notmuch database before trying to use this.
:---: | :---


go-imap-notmuch creates an IMAP gateway to your [notmuch](https://notmuchmail.org/) database,
allowing you to use any client to search notmuch, including converting
IMAP search semantics to notmuch queries.

If you'd like to directly use notmuch's syntax, you can do a full-text
search with the notmuch query.

Most of IMAPv4 is implemented, but not everything works. Nor does
everything work perfectly but I've tested with Evolution and Roundcube
and it's usable.

Thank you very much to the authors of [notmuch](https://notmuchmail.org/),
[go-imap](https://github.com/emersion/go-imap) and [go-notmuch](https://github.com/zenhack/go.notmuch) for making this possible.

## Why?

A long time ago, I hosted my own mail server but it's just become too
difficult to stay out of Junk folders -- despite doing all the right
things like DKIM, SPF, etc.  I now host my mail on Fastmail, mirror it
to a Maildir, index with notmuch, and mostly use neomutt. The problem
with this setup is I can't access messages from my iOS devices easily.

This lets me use a webmail client like Roundcube on my iPhone when I'm
away from home.

I don't reccomend putting this on the internet, but rather use a VPN
to get to it.

## Configuration

```yaml
---
username: "notmuch"
# bcrypt password:
password: "$2y$10$ieWd7rkUs/PNz1Iy5wGuY.hmDjfq5toZApZJb9P7Eu36ew/1thYwK"
maildir: "/home/stbenjam/Mail"
mailboxes:
  - name: INBOX
    query: "folder:INBOX"
  - name: Sent
    query: "folder:Sent"
    attributes:
      - Sent
```

Note: Mailbox attributes are ones specified by [RFC6154](https://datatracker.ietf.org/doc/html/rfc6154), such as Drafts, Sent, etc.

