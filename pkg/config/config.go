package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

const (
	defaultUsername = "notmuch"
	defaultPassword = "notmuch"
)

type Config struct {
	Username  string    `yaml:"username"`
	Password  string    `yaml:"password"`
	Maildir   string    `yaml:"maildir"`
	Mailboxes []Mailbox `yaml:"mailboxes"`
}

type Mailbox struct {
	Name  string `yaml:"name"`
	Query string `yaml:"query"`
}

func New(path string) (*Config, error) {
	if path == "" {
		return &Config{
			Username: defaultUsername,
			Password: defaultPassword,
			Mailboxes: []Mailbox{
				{
					Name:  "INBOX",
					Query: "folder:INBOX",
				},
			},
		}, nil
	}

	cfg, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	c := Config{}
	err = yaml.Unmarshal(cfg, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
