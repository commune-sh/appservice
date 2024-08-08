package app

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

func (c *App) IsLocalHomeserver(hs string) bool {
	return strings.HasSuffix(hs, c.Config.Matrix.ServerName)
}

func (c *App) IsNotRestricted(hs string) bool {
	if len(c.Config.AppService.Rules.FederationDomainWhitelist) == 0 {
		return false
	}
	if c.Config.AppService.Rules.FederationDomainWhitelist[0] == "*" {
		return true
	}
	localpart := strings.Split(hs, ":")[1]
	for _, r := range c.Config.AppService.Rules.FederationDomainWhitelist {
		if r == localpart {
			return true
		}
	}
	return false
}

func Contains[T comparable](slice []T, element T) bool {
	for _, v := range slice {
		if v == element {
			return true
		}
	}
	return false
}

func IsValidRoomID(room_id string) bool {
	reg := `^(?:!)[\w-]+:(?:[\w.-]+|\[[\w:]+\])(?::\d+)?$`
	match, err := regexp.MatchString(reg, room_id)
	if err != nil {
		return false
	}
	return match
}

func ExtractAccessToken(req *http.Request) (*string, error) {
	authBearer := req.Header.Get("Authorization")

	if authBearer != "" {
		parts := strings.SplitN(authBearer, " ", 2)
		if len(parts) != 2 ||
			parts[0] != "Bearer" {
			return nil, errors.New("Invalid Authorization header.")
		}

		return &parts[1], nil

	}

	return nil, errors.New("Missing access token.")
}

func Slugify(s string) string {
	reg := regexp.MustCompile("[^a-zA-Z0-9]+")
	p := reg.ReplaceAllString(s, "-")
	return strings.ToLower(p)
}

func CouldBeBridge(state mautrix.RoomStateMap) bool {
	bridge_types := []string{
		"m.bridge",
		"m.room.bridged",
		"m.room.discord",
		"m.room.irc",
		"uk.half-shot.bridge",
	}

	for _, t := range bridge_types {
		ev := event.Type{t, 2}
		exists := state[ev]
		if len(exists) > 0 {
			return true
		}
	}

	return false
}
