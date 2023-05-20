package config

import (
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"go.mau.fi/zeroconfig"
	"maunium.net/go/mautrix/id"
	"maunium.net/go/mautrix/util/dbutil"
)

type Config struct {
	Listen string `yaml:"listen"`

	// Authentication settings
	Homeserver   string              `yaml:"homeserver"`
	Username     id.UserID           `yaml:"username"`
	PasswordFile string              `yaml:"password_file"`
	Displayname  string              `yaml:"displayname"`
	AvatarURL    id.ContentURIString `yaml:"avatar_url"`

	// Database settings
	Database dbutil.Config `yaml:"database"`

	// Logging configuration
	Logging zeroconfig.Config `yaml:"logging"`

	GoogleCredentialsJSON string `yaml:"google_credentials_json"`
}

func (c *Config) GetPassword() (string, error) {
	if password := os.Getenv("MEETBOT_PASSWORD"); password != "" {
		return password, nil
	}

	log.Debug().Str("password_file", c.PasswordFile).Msg("reading password from file")
	buf, err := os.ReadFile(c.PasswordFile)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(buf)), nil
}

func (c *Config) GetCredentialsJSON() ([]byte, error) {
	if creds := os.Getenv("MEETBOT_GOOGLE_CREDENTIALS_JSON"); creds != "" {
		return []byte(creds), nil
	}

	return []byte(c.GoogleCredentialsJSON), nil
}
