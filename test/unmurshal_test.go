package test

import (
	"cicd-processor/domain"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestParseMessage_Success(t *testing.T) {
	data, err := os.ReadFile("testdata/input.json")
	assert.NoError(t, err, "failed to load JSON file")

	message, err := domain.ParseMessage(data)
	assert.NoError(t, err, "failed to unmarshal JSON")
	assert.NotNil(t, message, "message should not be nil")

	assert.Equal(t, "refs/tags/v0.0.2", message.Ref, "unexpected ref value")
	assert.Equal(t, "tg-dispatcher", message.Repository.Name, "unexpected repository name")
	assert.Equal(t, "vsevolodtachii@gmail.com", message.Repository.Owner.Email, "unexpected email")
	assert.Equal(t, "Vsevololod", message.Pusher.Name, "unexpected publisher name")
}

func TestParseMessageAllNotNull_Success(t *testing.T) {

	data, err := os.ReadFile("testdata/input.json")
	assert.NoError(t, err, "failed to load JSON file")

	message, err := domain.ParseMessage(data)
	// Проверка, что все поля заполнены
	assert.NotEmpty(t, message.Ref, "Ref should not be empty")
	assert.NotEmpty(t, message.Repository.Name, "Repository.Name should not be empty")
	assert.NotEmpty(t, message.Repository.Owner.Name, "Repository.Owner.Name should not be empty")
	assert.NotEmpty(t, message.Repository.Owner.Email, "Repository.Owner.Email should not be empty")
	assert.NotEmpty(t, message.Repository.Url, "Repository.Url should not be empty")
	assert.NotEmpty(t, message.Repository.GitUrl, "Repository.GitUrl should not be empty")
	assert.NotEmpty(t, message.Pusher.Name, "Pusher.Name should not be empty")
	assert.NotEmpty(t, message.Pusher.Email, "Pusher.Email should not be empty")
	assert.NotEmpty(t, message.BaseRef, "BaseRef should not be empty")
}

func TestParseMessage_InvalidJSON(t *testing.T) {
	invalidJSON := []byte(`{"ref": "refs/tags/v0.0.2", "repository": "invalid"}`)
	message, err := domain.ParseMessage(invalidJSON)

	assert.Error(t, err, "expected error for invalid JSON")
	assert.Nil(t, message, "message should be nil for invalid JSON")
}
