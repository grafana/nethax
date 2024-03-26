package pkg

import (
	"log/slog"
)

type logger struct {
	log           slog.Logger
	podFrom       string
	namespaceFrom string
	podTo         string
	namespaceTo   string
	remoteURI     string
}

func (l logger) appendLoggerDetails(extra []any) []any {
	if l.podFrom != "" {
		extra = append(extra, l.podFrom)
	}
	if l.namespaceFrom != "" {
		extra = append(extra, l.namespaceFrom)
	}
	if l.podTo != "" {
		extra = append(extra, l.podTo)
	}
	if l.namespaceFrom != "" {
		extra = append(extra, l.namespaceFrom)
	}
	if l.remoteURI != "" {
		extra = append(extra, l.remoteURI)
	}

	return extra
}

func (l logger) Info(msg string, extra ...any) {
	extra = l.appendLoggerDetails(extra)

	l.log.Info(msg, extra...)
}

// TODO - use codegen to build similar methods for other log levels
func (l logger) Error(msg string, extra ...any) {
	extra = l.appendLoggerDetails(extra)

	l.log.Error(msg, extra...)
}
