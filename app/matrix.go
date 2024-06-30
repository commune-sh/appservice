package app

import (
	"commune/config"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

func NewMatrixClient() (*mautrix.Client, error) {

	c, err := config.Read(CONFIG_FILE)
	if err != nil {
		panic(err)
	}

	user_id := id.NewUserID(c.AppService.SenderLocalPart, c.Matrix.ServerName)

	client, err := mautrix.NewClient(c.Matrix.Homeserver, user_id, c.AppService.AccessToken)
	if err != nil {
		return nil, err
	}
	return client, nil
}
