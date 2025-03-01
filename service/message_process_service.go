package service

import (
	"cicd-processor/config"
	"cicd-processor/domain"
	"cicd-processor/lib/logger/sl"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// MessageProcessService — сервис обработки сообщений
type MessageProcessService struct {
	inputMessageChannel chan domain.UpdateMessageWithContext
	pathsConf           config.PathsConf
	log                 *slog.Logger
}

// NewMessageProcessService создает новый сервис и принимает канал для сообщений
func NewMessageProcessService(
	inputMessageChannel chan domain.UpdateMessageWithContext,
	pathsConf config.PathsConf,
	log *slog.Logger) *MessageProcessService {
	return &MessageProcessService{
		inputMessageChannel: inputMessageChannel,
		pathsConf:           pathsConf,
		log:                 log,
	}
}

// StartProcessing запускает обработку сообщений в отдельной горутине
func (s *MessageProcessService) StartProcessing(workerCount int) {
	for i := 0; i < workerCount; i++ {
		go func(workerID int) {
			for msg := range s.inputMessageChannel {
				s.ProcessMessage(workerID, msg)
			}
		}(i)
	}
}

// ProcessMessage выполняет обработку сообщения
func (s *MessageProcessService) ProcessMessage(workerID int, msg domain.UpdateMessageWithContext) {
	s.log.Info("Processing Message with id", slog.String("uuid", msg.UUID), slog.Int64("worker", int64(workerID)))

	message := msg.UpdateMessage
	if !strings.HasPrefix(message.Ref, "refs/tags") {
		s.log.Info("Ignoring message with invalid ref", slog.String("uuid", msg.UUID), slog.String("ref", message.Ref))
		return
	}

	projectPath := filepath.Join(s.pathsConf.Projects, message.Repository.Name)
	err := removeClonePathIfExists(projectPath)
	if err != nil {
		s.log.Error("Cannot remove existing repo", sl.Err(err))
		return
	}

	repoURL := message.Repository.Url
	tag := strings.Split(message.Ref, "/")[2]

	logsFile, err := os.OpenFile(getLogPath(s.pathsConf.CommandsLogs, msg.UUID), os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		s.log.Error("Cannot open internal logs file", sl.Err(err))
		return
	}
	defer logsFile.Close()

	// Клонируем репозиторий
	s.log.Info("Perform git clone", slog.String("destination", s.pathsConf.Projects))
	if err := s.runCommand(s.pathsConf.Projects, "git", logsFile, "clone", "--depth", "1", "--branch", tag, repoURL, projectPath); err != nil {
		_ = logsFile.Sync()
		s.log.Error("Ошибка при клонировании:", sl.Err(err))
		return
	}

	// Запускаем build.sh
	buildScript := filepath.Join(projectPath, "build.sh")
	s.log.Info("Start build.sh...", slog.String("script", buildScript))
	if err := s.runCommand(projectPath, "bash", logsFile, buildScript); err != nil {
		_ = logsFile.Sync()
		s.log.Error("Ошибка при выполнении build.sh:", sl.Err(err))
		return
	}

	sourceConfig := filepath.Join(projectPath, "config")
	destinationConfig := filepath.Join(s.pathsConf.Configs, message.Repository.Name)
	s.log.Info(
		"Copy configs",
		slog.String("destination", destinationConfig),
		slog.String("source", sourceConfig))
	err = copyDir(sourceConfig, destinationConfig)
	if err != nil {
		_ = logsFile.Sync()
		s.log.Error("cannot copy configs", sl.Err(err))
		return
	}

	s.log.Info("start docker compose")
	if err := s.runCommand(projectPath, "docker", logsFile, "compose", "up", "-d"); err != nil {
		_ = logsFile.Sync()
		s.log.Error("Error starting docker compose", sl.Err(err))
		return
	}

	s.log.Info("Скрипт выполнен успешно.")

}

func (s *MessageProcessService) StopProcessing() {

}

func getLogPath(commandsLogsPath string, uuid string) string {
	return filepath.Join(commandsLogsPath, uuid+".log")
}

// copyFile копирует один файл из src в dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// Копируем права доступа
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, srcInfo.Mode())
}

// copyDir копирует все файлы из srcDir в dstDir
func copyDir(srcDir, dstDir string) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}

	// Создаем целевую папку, если она не существует
	if err := os.MkdirAll(dstDir, os.ModePerm); err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		dstPath := filepath.Join(dstDir, entry.Name())

		if !entry.IsDir() {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func removeClonePathIfExists(cloneDir string) error {
	// Проверяем, существует ли директория
	if _, err := os.Stat(cloneDir); !os.IsNotExist(err) {
		if err := os.RemoveAll(cloneDir); err != nil {
			return err
		}
	}
	return nil
}

func (s *MessageProcessService) runCommand(dir, name string, file io.Writer, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = file
	cmd.Stderr = file
	s.log.Info("Executing:", slog.String("cmd", cmd.String()))
	return cmd.Run()
}
