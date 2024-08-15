package config

import (
	"io"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	App struct {
		Domain string `toml:"domain" json:"domain"`
		Port   int    `toml:"port" json:"port"`
	} `toml:"app" json:"app"`
	AppService struct {
		ID              string `json:"id" toml:"id"`
		SenderLocalPart string `json:"sender_local_part" toml:"sender_local_part"`
		AccessToken     string `json:"access_token" toml:"access_token"`
		HSAccessToken   string `json:"hs_access_token" toml:"hs_access_token"`
		Rules           struct {
			AutoJoin                  bool     `json:"auto_join" toml:"auto_join"`
			InviteByLocalUser         bool     `json:"invite_by_local_user" toml:"invite_by_local_user"`
			FederationDomainWhitelist []string `json:"federation_domain_whitelist" toml:"federation_domain_whitelist"`
		} `json:"rules" toml:"rules"`
	} `json:"appservice" toml:"appservice"`
	Log struct {
		File       string `toml:"file" json:"file"`
		MaxSize    int    `toml:"max_size" json:"max_size"`
		MaxBackups int    `toml:"max_backups" json:"max_backups"`
		MaxAge     int    `toml:"max_age" json:"max_age"`
		Compress   bool   `toml:"compress" json:"compress"`
	} `json:"log" toml:"log"`
	Matrix struct {
		Homeserver string `toml:"homeserver" json:"homeserver"`
		ServerName string `toml:"server_name" json:"server_name"`
	} `json:"matrix" toml:"matrix"`
	Redis struct {
		Address    string `toml:"address" json:"address"`
		Password   string `toml:"password" json:"password"`
		RoomsDB    int    `toml:"rooms_db" json:"rooms_db"`
		MessagesDB int    `toml:"messages_db" json:"messages_db"`
		EventsDB   int    `toml:"events_db" json:"events_db"`
		StateDB    int    `toml:"state_db" json:"state_db"`
	} `json:"redis" toml:"redis"`
	Security struct {
		AllowedOrigins []string `toml:"allowed_origins" json:"allowed_origins"`
	} `json:"security" toml:"security"`
}

var conf Config

// Read reads the config file and returns the Values struct
func Read(s string) (*Config, error) {
	file, err := os.Open(s)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	b, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	if _, err := toml.Decode(string(b), &conf); err != nil {
		panic(err)
	}

	return &conf, err
}
