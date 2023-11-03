package config

import (
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"go.mau.fi/util/dbutil"
	"go.mau.fi/zeroconfig"
	"maunium.net/go/mautrix/id"
)

type OAuth2Config struct {
	CredentialsJSON string `yaml:"credentials_json"`
	RedirectURL     string `yaml:"redirect_url"`
}

type Config struct {
	Listen string `yaml:"listen"`

	// Authentication settings
	Homeserver   string              `yaml:"homeserver"`
	Username     id.UserID           `yaml:"username"`
	PasswordFile string              `yaml:"password_file"`
	Displayname  string              `yaml:"displayname"`
	AvatarURL    id.ContentURIString `yaml:"avatar_url"`

	UserIDToEmail map[id.UserID]string `yaml:"user_id_to_email"`

	// Database settings
	Database dbutil.Config `yaml:"database"`

	// Logging configuration
	Logging zeroconfig.Config `yaml:"logging"`

	OAuth2 OAuth2Config `yaml:"oauth2"`
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

func (c *Config) GetCredentialsJSON() []byte {
	if creds := os.Getenv("MEETBOT_OAUTH2_CREDENTIALS_JSON"); creds != "" {
		return []byte(creds)
	}

	return []byte(c.OAuth2.CredentialsJSON)
}

func (c *Config) GetRedirectURL() string {
	if url := os.Getenv("MEETBOT_OAUTH2_REDIRECT_URL"); url != "" {
		return url
	}

	return c.OAuth2.RedirectURL
}
