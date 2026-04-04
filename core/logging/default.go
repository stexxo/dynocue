package logging

import (
	"log/slog"
	"os"
)

func NewDefaultLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}
