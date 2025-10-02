package golog

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/lmittmann/tint"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Mode string

const (
	dev  Mode = "dev"
	prod Mode = "prod"
)

type LogConfig struct {
	Mode    Mode
	LogPath string
	Level   slog.Level
}

func InitLogger(cfg LogConfig) *slog.Logger {
	var logger *slog.Logger
	logFile := getRotatedLogFile(cfg.LogPath)

	replaceAttr := func(groups []string, a slog.Attr) slog.Attr {
		switch a.Key {
		case slog.TimeKey:
			t := a.Value.Time()
			a.Value = slog.StringValue(t.Format("2006-01-02 15:04:05"))

		case slog.SourceKey:
			src := a.Value.Any().(*slog.Source)
			funcParts := strings.Split(src.Function, ".")
			pkg := "unknown"
			if len(funcParts) > 1 {
				pkg = funcParts[0]
			}

			val := fmt.Sprintf("%s %s:%d", pkg, filepath.Base(src.File), src.Line)
			a.Value = slog.StringValue(val)
		}
		return a
	}

	localOptions := &tint.Options{
		AddSource:   true,
		Level:       cfg.Level,
		ReplaceAttr: replaceAttr,
	}

	prodOptions := &slog.HandlerOptions{
		AddSource:   true,
		Level:       cfg.Level,
		ReplaceAttr: replaceAttr,
	}

	switch cfg.Mode {
	case dev:
		logger = slog.New(tint.NewHandler(os.Stdout, localOptions))

	case prod:
		logger = slog.New(slog.NewJSONHandler(logFile, prodOptions))

	default:
		logger = slog.New(tint.NewHandler(os.Stdout, localOptions))
	}

	return logger
}

func getRotatedLogFile(logPath string) io.Writer {
	return &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   //days
		Compress:   true, // disabled by default
	}
}
