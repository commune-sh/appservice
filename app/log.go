package app

import (
	"commune/config"
	"fmt"
	"os"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/rs/zerolog"
)

func SetupLogger() (*zerolog.Logger, error) {

	config, err := config.Read(CONFIG_FILE)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	filename := fmt.Sprintf("%s%s.log", "./logs/", now.Format("2006-01-02"))

	lumberjackWriter := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    config.Log.MaxSize,
		MaxBackups: config.Log.MaxBackups,
		MaxAge:     config.Log.MaxAge,
		Compress:   config.Log.Compress,
	}

	mw := zerolog.MultiLevelWriter(lumberjackWriter, zerolog.ConsoleWriter{Out: os.Stdout})

	logger := zerolog.New(mw).With().Timestamp().Logger()

	return &logger, nil
}
