package then

import (
	"context"
	"log/slog"

	"github.com/byuoitav/common/nerr"
	shipwrightThen "github.com/byuoitav/shipwright/actions/then"
	"go.uber.org/zap"
)

// toThenFunc takes a plain error handler and returns a then.Func
func toThenFunc(
	fn func(ctx context.Context, with []byte, log *zap.SugaredLogger) error,
) shipwrightThen.Func {
	return func(ctx context.Context, with []byte, log *zap.SugaredLogger) *nerr.E {
		if err := fn(ctx, with, log); err != nil {
			// translate any error into an *nerr.E and annotate
			return nerr.Translate(err).Addf("running %T", fn)
		}
		return nil
	}
}

// Zap to Slog section

// zapSugaredHandler implements slog.Handler by forwarding into zap.SugaredLogger.
type zapSugaredHandler struct {
	sug *zap.SugaredLogger
}

func (h *zapSugaredHandler) Enabled(_ context.Context, level slog.Level) bool {
	// You can tighten this by inspecting h.sug.Desugar().Core().Enabled(...)
	return true
}

func (h *zapSugaredHandler) Handle(_ context.Context, rec slog.Record) error {
	// collect all attrs into a slice [key1, val1, key2, val2, …]
	var args []interface{}
	rec.Attrs(func(a slog.Attr) bool {
		args = append(args, a.Key, a.Value.Any())
		return true
	})

	msg := rec.Message

	// route by level
	switch rec.Level {
	case slog.LevelDebug:
		h.sug.Debugw(msg, args...)
	case slog.LevelInfo:
		h.sug.Infow(msg, args...)
	case slog.LevelWarn:
		h.sug.Warnw(msg, args...)
	case slog.LevelError, slog.LevelError + 1:
		h.sug.Errorw(msg, args...)
	default:
		h.sug.Infow(msg, args...)
	}
	return nil
}

func (h *zapSugaredHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// translate attrs into zap fields, then create a child sugared logger
	var fields []interface{}
	for _, a := range attrs {
		fields = append(fields, a.Key, a.Value.Any())
	}
	return &zapSugaredHandler{sug: h.sug.With(fields...)}
}

func (h *zapSugaredHandler) WithGroup(name string) slog.Handler {
	// slog groups would nest keys like "group.key"; zap sug doesn’t have groups
	return h
}

// ZapSugaredToSlog wraps a zap.SugaredLogger in an slog.Logger.
func ZapSugaredToSlog(sug *zap.SugaredLogger) *slog.Logger {
	return slog.New(&zapSugaredHandler{sug: sug})
}
