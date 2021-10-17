# go-imap-notmuch

:bangbang: | This software is experimental, and use at your own risk. Please make a backup of your Maildir and notmuch database before trying to use this.
:---: | :---


Not everything works, but I've tested with Evolution and Roundcube and
it seems to mostly work. Searches work, along with moving between
folders, marking things read, flagged etc.

Thank you very much to the authors of [notmuch](https://notmuchmail.org/),
[go-imap](https://github.com/emersion/go-imap) and [go-notmuch](https://github.com/zenhack/go.notmuch) for making this possible.

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
```
