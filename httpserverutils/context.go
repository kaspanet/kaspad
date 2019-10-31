package httpserverutils

import (
	"context"
	"fmt"
)

type contextKey string

const (
	contextKeyRequestID contextKey = "REQUEST_ID"
)

// ServerContext is a context.Context wrapper that
// enables custom logs with request ID.
type ServerContext struct {
	context.Context
}

// ToServerContext takes a context.Context instance
// and converts it to *ServerContext.
func ToServerContext(ctx context.Context) *ServerContext {
	if asCtx, ok := ctx.(*ServerContext); ok {
		return asCtx
	}
	return &ServerContext{Context: ctx}
}

// SetRequestID associates a request ID for the context.
func (ctx *ServerContext) SetRequestID(requestID uint64) context.Context {
	context.WithValue(ctx, contextKeyRequestID, requestID)
	return ctx
}

func (ctx *ServerContext) requestID() uint64 {
	id := ctx.Value(contextKeyRequestID)
	uint64ID, _ := id.(uint64)
	return uint64ID
}

func (ctx *ServerContext) getLogString(format string, params ...interface{}) string {
	return fmt.Sprintf("RID %d: ", ctx.requestID()) + fmt.Sprintf(format, params...)
}

// Tracef writes a customized formatted context
// related log with log level 'Trace'.
func (ctx *ServerContext) Tracef(format string, params ...interface{}) {
	log.Trace(ctx.getLogString(format, params...))
}

// Debugf writes a customized formatted context
// related log with log level 'Debug'.
func (ctx *ServerContext) Debugf(format string, params ...interface{}) {
	log.Debug(ctx.getLogString(format, params...))
}

// Infof writes a customized formatted context
// related log with log level 'Info'.
func (ctx *ServerContext) Infof(format string, params ...interface{}) {
	log.Info(ctx.getLogString(format, params...))
}

// Warnf writes a customized formatted context
// related log with log level 'Warn'.
func (ctx *ServerContext) Warnf(format string, params ...interface{}) {
	log.Warn(ctx.getLogString(format, params...))
}

// Errorf writes a customized formatted context
// related log with log level 'Error'.
func (ctx *ServerContext) Errorf(format string, params ...interface{}) {
	log.Error(ctx.getLogString(format, params...))
}

// Criticalf writes a customized formatted context
// related log with log level 'Critical'.
func (ctx *ServerContext) Criticalf(format string, params ...interface{}) {
	log.Criticalf(ctx.getLogString(format, params...))
}
