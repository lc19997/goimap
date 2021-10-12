package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

const (
	defaultUsername = "notmuch"
	defaultPassword = "notmuch"
	inboxQuery = "folder:INBOX"
)

type Config struct {
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	Maildir    string `yaml:"maildir"`
	InboxQuery string `yaml:"inboxQuery"`
}

func New(path string) (*Config, error) {
	if path == "" {
		return &Config{
			Username: defaultUsername,
			Password: defaultPassword,
			InboxQuery: inboxQuery,
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
