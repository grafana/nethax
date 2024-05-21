package logging

import (
	"log/slog"
	"os"
)

type Logger struct {
	log           *slog.Logger
	PodFrom       string
	NamespaceFrom string
	PodTo         string
	NamespaceTo   string
	RemoteURI     string
}

func (l Logger) appendLoggerDetails(extra []any) []any {
	if l.PodFrom != "" {
		extra = append(extra, l.PodFrom)
	}
	if l.NamespaceFrom != "" {
		extra = append(extra, l.NamespaceFrom)
	}
	if l.PodTo != "" {
		extra = append(extra, l.PodTo)
	}
	if l.NamespaceTo != "" {
		extra = append(extra, l.NamespaceTo)
	}
	if l.RemoteURI != "" {
		extra = append(extra, l.RemoteURI)
	}

	return extra
}

func (l Logger) slogInit() {
	l.log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))
}

func (l Logger) Info(msg string, extra ...any) {
	extra = l.appendLoggerDetails(extra)

	l.log.Info(msg, extra...)
}

// TODO - use codegen to build similar methods for other log levels
func (l Logger) Error(msg string, extra ...any) {
	extra = l.appendLoggerDetails(extra)

	l.log.Error(msg, extra...)
}
