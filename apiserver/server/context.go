package server

import (
	"context"
	"fmt"
)

type contextKey string

const (
	contextKeyRequestID contextKey = "REQUEST_ID"
)

type apiServerContext struct {
	context.Context
}

func newAPIServerContext(ctx context.Context) *apiServerContext {
	if asCtx, ok := ctx.(*apiServerContext); ok {
		return asCtx
	}
	return &apiServerContext{Context: ctx}
}

func (ctx *apiServerContext) setRequestID(requestID uint64) context.Context {
	context.WithValue(ctx, contextKeyRequestID, nextRequestID)
	return ctx
}

func (ctx *apiServerContext) requestID() uint64 {
	id := ctx.Value(contextKeyRequestID)
	uint64ID, _ := id.(uint64)
	return uint64ID
}

func (ctx *apiServerContext) getLogString(format string, params ...interface{}) string {
	return fmt.Sprintf("RID %d: ", ctx.requestID()) + fmt.Sprintf(format, params...)
}

func (ctx *apiServerContext) tracef(format string, params ...interface{}) {
	log.Tracef(ctx.getLogString(format, params...))
}

func (ctx *apiServerContext) debugf(format string, params ...interface{}) {
	log.Debugf(ctx.getLogString(format, params...))
}

func (ctx *apiServerContext) infof(format string, params ...interface{}) {
	log.Infof(ctx.getLogString(format, params...))
}

func (ctx *apiServerContext) warnf(format string, params ...interface{}) {
	log.Warnf(ctx.getLogString(format, params...))
}

func (ctx *apiServerContext) errorf(format string, params ...interface{}) {
	log.Errorf(ctx.getLogString(format, params...))
}

func (ctx *apiServerContext) criticalf(format string, params ...interface{}) {
	log.Criticalf(ctx.getLogString(format, params...))
}
