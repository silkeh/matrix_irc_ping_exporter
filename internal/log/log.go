package log

import (
	"log/slog"
	"os"
)

// ParseLevel parses a log level representing a [slog.Level].
func ParseLevel(s string) (l slog.Level, err error) {
	err = l.UnmarshalText([]byte(s))

	return
}

// Fatal logs an error and exits with error code 1.
func Fatal(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}

func Setup(level string) error {
	lvl, err := ParseLevel(level)
	if err != nil {
		return err
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: lvl,
	})))

	return nil
}
