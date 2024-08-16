package config

import (
	"io"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	App struct {
		Domain string `toml:"domain"`
		Port   int    `toml:"port"`
	} `toml:"app" json:"app"`
	AppService struct {
		ID              string `toml:"id"`
		SenderLocalPart string `toml:"sender_local_part"`
		AccessToken     string `toml:"access_token"`
		HSAccessToken   string `toml:"hs_access_token"`
		Rules           struct {
			AutoJoin                  bool     `toml:"auto_join"`
			InviteByLocalUser         bool     `toml:"invite_by_local_user"`
			FederationDomainWhitelist []string `toml:"federation_domain_whitelist"`
		} `toml:"rules"`
	} `toml:"appservice"`
	Log struct {
		File       string `toml:"file"`
		MaxSize    int    `toml:"max_size"`
		MaxBackups int    `toml:"max_backups"`
		MaxAge     int    `toml:"max_age"`
		Compress   bool   `toml:"compress"`
	} `json:"log" toml:"log"`
	Matrix struct {
		Homeserver string `toml:"homeserver"`
		ServerName string `toml:"server_name"`
	} `json:"matrix" toml:"matrix"`
	Redis struct {
		Address    string `toml:"address"`
		Password   string `toml:"password"`
		RoomsDB    int    `toml:"rooms_db"`
		MessagesDB int    `toml:"messages_db"`
		EventsDB   int    `toml:"events_db"`
		StateDB    int    `toml:"state_db"`
	} `toml:"redis"`
	Cache struct {
		PublicRooms struct {
			Enabled     bool  `toml:"enabled"`
			ExpireAfter int64 `toml:"expire_after"`
		} `toml:"public_rooms"`
		RoomState struct {
			Enabled     bool  `toml:"enabled"`
			ExpireAfter int64 `toml:"expire_after"`
		} `toml:"room_state"`
		Messages struct {
			Enabled     bool  `toml:"enabled"`
			ExpireAfter int64 `toml:"expire_after"`
		} `toml:"messages"`
	} `toml:"cache"`
	Security struct {
		AllowedOrigins []string `toml:"allowed_origins"`
	} `toml:"security"`
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
