package utils

import (
	"context"
	"fmt"
)

type contextKey string

const (
	contextKeyRequestID contextKey = "REQUEST_ID"
)

type ApiServerContext struct {
	context.Context
}

func NewAPIServerContext(ctx context.Context) *ApiServerContext {
	if asCtx, ok := ctx.(*ApiServerContext); ok {
		return asCtx
	}
	return &ApiServerContext{Context: ctx}
}

func (ctx *ApiServerContext) SetRequestID(requestID uint64) context.Context {
	context.WithValue(ctx, contextKeyRequestID, requestID)
	return ctx
}

func (ctx *ApiServerContext) requestID() uint64 {
	id := ctx.Value(contextKeyRequestID)
	uint64ID, _ := id.(uint64)
	return uint64ID
}

func (ctx *ApiServerContext) getLogString(format string, params ...interface{}) string {
	return fmt.Sprintf("RID %d: ", ctx.requestID()) + fmt.Sprintf(format, params...)
}

func (ctx *ApiServerContext) Tracef(format string, params ...interface{}) {
	log.Tracef(ctx.getLogString(format, params...))
}

func (ctx *ApiServerContext) Debugf(format string, params ...interface{}) {
	log.Debugf(ctx.getLogString(format, params...))
}

func (ctx *ApiServerContext) Infof(format string, params ...interface{}) {
	log.Infof(ctx.getLogString(format, params...))
}

func (ctx *ApiServerContext) Warnf(format string, params ...interface{}) {
	log.Warnf(ctx.getLogString(format, params...))
}

func (ctx *ApiServerContext) Errorf(format string, params ...interface{}) {
	log.Errorf(ctx.getLogString(format, params...))
}

func (ctx *ApiServerContext) Criticalf(format string, params ...interface{}) {
	log.Criticalf(ctx.getLogString(format, params...))
}
