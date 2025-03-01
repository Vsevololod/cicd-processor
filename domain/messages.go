package domain

import (
	"context"
	"encoding/json"
)

type UpdateMessage struct {
	Ref        string     `json:"ref"`
	Repository Repository `json:"repository"`
	Pusher     Pusher     `json:"pusher"`
	BaseRef    string     `json:"base_ref"`
}

type Repository struct {
	Name   string `json:"name"`
	Owner  Owner  `json:"owner"`
	Url    string `json:"url"`
	GitUrl string `json:"git_url"`
}

type Owner struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Pusher struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UpdateMessageWithContext struct {
	UpdateMessage *UpdateMessage
	UUID          string
	Context       context.Context
}

func ParseMessage(jsonData []byte) (*UpdateMessage, error) {
	message := UpdateMessage{}
	err := json.Unmarshal(jsonData, &message)
	if err != nil {
		return nil, err
	}

	return &message, nil
}
