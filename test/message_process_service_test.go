package test

import (
	"cicd-processor/config"
	"cicd-processor/domain"
	"cicd-processor/service"
	"context"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"os"
	"testing"
)

func getPathConfigs(baseDir string, t *testing.T) (config.PathsConf, func()) {
	tempDir := baseDir + "/temp"
	err := os.Mkdir(tempDir, os.ModePerm)
	assert.NoError(t, err, "failed to create temp dir")

	tempLogsDir := baseDir + "/logs"
	err = os.MkdirAll(tempLogsDir, os.ModePerm)
	assert.NoError(t, err, "failed to create temp dir")

	configDir := baseDir + "/configs"
	err = os.MkdirAll(configDir, os.ModePerm)
	assert.NoError(t, err, "failed to create temp dir")

	return config.PathsConf{
			Projects:     tempDir,
			CommandsLogs: tempLogsDir,
			Configs:      configDir,
		}, func() {
			_ = os.RemoveAll(tempDir)
			_ = os.RemoveAll(tempLogsDir)
			_ = os.RemoveAll(configDir)
		}

}

func TestProcessMessage_Success(t *testing.T) {
	data, err := os.ReadFile("testdata/input.json")
	assert.NoError(t, err, "failed to load JSON file")

	baseDir, err := os.Getwd()
	assert.NoError(t, err, "failed to get working directory")

	configs, f := getPathConfigs(baseDir, t)
	defer f()

	message, err := domain.ParseMessage(data)
	assert.NoError(t, err, "failed to parse message")

	inputMessageChannel := make(chan domain.UpdateMessageWithContext, 5)
	defer close(inputMessageChannel)

	log := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)

	processService := service.NewMessageProcessService(inputMessageChannel, configs, log)
	processService.ProcessMessage(1, domain.UpdateMessageWithContext{
		UpdateMessage: message,
		UUID:          "dac0394b-9edf-46f0-b295-d990963aef2b",
		Context:       context.Background(),
	})

}
