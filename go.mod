module github.com/stbenjam/go-imap-notmuch

go 1.21

require (
	github.com/emersion/go-imap v1.2.1
	github.com/emersion/go-message v0.17.0
	github.com/sirupsen/logrus v1.9.3
	github.com/zenhack/go.notmuch v0.0.0-20220918173508-0c918632c39e
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/emersion/go-sasl v0.0.0-20200509203442-7bfe0ed36a21 // indirect
	github.com/emersion/go-textwrapper v0.0.0-20200911093747-65d896831594 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

replace github.com/zenhack/go.notmuch => github.com/stbenjam/go.notmuch v0.0.0-20211020000856-ac412a4e5b67
