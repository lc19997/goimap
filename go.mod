module github.com/stbenjam/go-imap-notmuch

go 1.16

require (
	github.com/emersion/go-imap v1.2.1
	github.com/emersion/go-message v0.17.0
	github.com/zenhack/go.notmuch v0.0.0-20220918173508-0c918632c39e
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
	golang.org/x/text v0.14.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/zenhack/go.notmuch => github.com/stbenjam/go.notmuch v0.0.0-20211020000856-ac412a4e5b67
