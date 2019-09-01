package server

import (
	"context"
	"fmt"
)

type contextKey string

const (
	contextKeyRequestID contextKey = "REQUEST_ID"
)

type customContext struct {
	context.Context
}

func newCustomContext(ctx context.Context) *customContext {
	return &customContext{Context: ctx}
}

func (ctx *customContext) setRequestID(requestID uint64) context.Context {
	context.WithValue(ctx, contextKeyRequestID, nextRequestID)
	return ctx
}

func (ctx *customContext) requestID() uint64 {
	id := ctx.Value(contextKeyRequestID)
	return id.(uint64)
}

func (ctx *customContext) getLogString(format string, params ...interface{}) string {
	params = append(params, ctx.requestID())
	return fmt.Sprintf("Request %d: "+format, params)
}

func (ctx *customContext) tracef(format string, params ...interface{}) {
	log.Tracef(ctx.getLogString(format, params))
}

func (ctx *customContext) debugf(format string, params ...interface{}) {
	log.Debugf(ctx.getLogString(format, params))
}

func (ctx *customContext) infof(format string, params ...interface{}) {
	log.Infof(ctx.getLogString(format, params))
}

func (ctx *customContext) warnf(format string, params ...interface{}) {
	log.Warnf(ctx.getLogString(format, params))
}

func (ctx *customContext) errorf(format string, params ...interface{}) {
	log.Errorf(ctx.getLogString(format, params))
}

func (ctx *customContext) criticalf(format string, params ...interface{}) {
	log.Criticalf(ctx.getLogString(format, params))
}
