package config

import (
	"os"
	"path"

	"golang.org/x/crypto/bcrypt"

	"gopkg.in/yaml.v2"
)

const (
	defaultUsername = "notmuch"
	defaultPassword = "notmuch"
)

type Config struct {
	// Username contains the IMAPv4 user that your client will use to login.
	Username string `yaml:"username"`

	// Password contains the password for the user above. It must be bcrypted,
	// e.g. with `htpasswd -bnBC 10 "" 'potato'`
	Password string `yaml:"password"`

	// Maildir is the path to the directory holding your mail notmuch
	// database.
	Maildir string `yaml:"maildir"`

	// Mailboxes is a list of mailbox names and notmuch search querie.
	Mailboxes []Mailbox `yaml:"mailboxes"`

	// UidDatabase is used to persist the assigned UID for a message, to
	// keep clients from getting confused.
	UidDatabase string `yaml:"uidDataabase"`

	// If wanting to use encryption, path to the TLS certificate and private
	// key.
	TLSCertificate string `yaml:"tlsCertificate"`
	TLSKey         string `yaml:"tlsKey"`

	// Enable debug logging in go-imap
	Debug bool `yaml:"debug"`
}

type Mailbox struct {
	Name       string   `yaml:"name"`
	Query      string   `yaml:"query"`
	Attributes []string `yaml:"attributes"`
}

func (c *Config) SetDefaults() {
	if c.Username == "" {
		c.Username = defaultUsername
	}

	if c.Password == "" {
		hash, _ := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
		c.Password = string(hash)
	}

	if c.Maildir == "" {
		c.Maildir = path.Join(os.Getenv("HOME"), "/Mail")
	}

	if c.UidDatabase == "" {
		c.UidDatabase = path.Join(c.Maildir, "/.uid.dat")
	}

	if len(c.Mailboxes) == 0 {
		c.Mailboxes = []Mailbox{
			{
				Name:  "INBOX",
				Query: "folder:INBOX",
			},
		}
	}

}

func New(path string) (*Config, error) {
	config := Config{}

	if path != "" {
		cfg, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(cfg, &config)
		if err != nil {
			return nil, err
		}
	}

	config.SetDefaults()

	return &config, nil
}
