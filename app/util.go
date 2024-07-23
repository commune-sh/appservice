package app

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
)

func (c *App) IsLocalHomeserver(hs string) bool {
	parts := strings.Split(hs, ":")
	if parts[1] == c.Config.Matrix.ServerName {
		return true
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
