package utils

import (
	"context"
	"fmt"
)

type contextKey string

const (
	contextKeyRequestID contextKey = "REQUEST_ID"
)

// APIServerContext is a context.Context wrapper that
// enables custom logs with request ID.
type APIServerContext struct {
	context.Context
}

// ToAPIServerContext takes a context.Context instance
// and converts it to *ApiServerContext.
func ToAPIServerContext(ctx context.Context) *APIServerContext {
	if asCtx, ok := ctx.(*APIServerContext); ok {
		return asCtx
	}
	return &APIServerContext{Context: ctx}
}

// SetRequestID associates a request ID for the context.
func (ctx *APIServerContext) SetRequestID(requestID uint64) context.Context {
	context.WithValue(ctx, contextKeyRequestID, requestID)
	return ctx
}

func (ctx *APIServerContext) requestID() uint64 {
	id := ctx.Value(contextKeyRequestID)
	uint64ID, _ := id.(uint64)
	return uint64ID
}

func (ctx *APIServerContext) getLogString(format string, params ...interface{}) string {
	return fmt.Sprintf("RID %d: ", ctx.requestID()) + fmt.Sprintf(format, params...)
}

// Tracef writes a customized formatted context
// related log with log level 'Trace'.
func (ctx *APIServerContext) Tracef(format string, params ...interface{}) {
	log.Trace(ctx.getLogString(format, params...))
}

// Debugf writes a customized formatted context
// related log with log level 'Debug'.
func (ctx *APIServerContext) Debugf(format string, params ...interface{}) {
	log.Debug(ctx.getLogString(format, params...))
}

// Infof writes a customized formatted context
// related log with log level 'Info'.
func (ctx *APIServerContext) Infof(format string, params ...interface{}) {
	log.Info(ctx.getLogString(format, params...))
}

// Warnf writes a customized formatted context
// related log with log level 'Warn'.
func (ctx *APIServerContext) Warnf(format string, params ...interface{}) {
	log.Warn(ctx.getLogString(format, params...))
}

// Errorf writes a customized formatted context
// related log with log level 'Error'.
func (ctx *APIServerContext) Errorf(format string, params ...interface{}) {
	log.Error(ctx.getLogString(format, params...))
}

// Criticalf writes a customized formatted context
// related log with log level 'Critical'.
func (ctx *APIServerContext) Criticalf(format string, params ...interface{}) {
	log.Criticalf(ctx.getLogString(format, params...))
}
