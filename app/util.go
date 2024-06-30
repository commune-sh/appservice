package app

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
)

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
