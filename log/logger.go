package log

import (
	"context"
	"log/slog"
	"os"
)

var Log *slog.Logger

func init() {
	Log = slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

func Info(msg string, args ...interface{}) {
	Log.Info(msg, args...)
}

func Error(msg string, args ...interface{}) {
	Log.Error(msg, args...)
}

func Debug(msg string, args ...interface{}) {
	Log.Debug(msg, args...)
}

func LogIt(ctx context.Context, level slog.Level, msg string, args ...interface{}) {
	Log.Log(ctx, level, msg, args...)
}
