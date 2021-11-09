package clog

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"
)

const (
	ctxMapKey = "ctxMap"
)

var mut sync.Mutex

func WithCtx(ctx context.Context, key string, value interface{}) context.Context {
	mut.Lock()
	defer mut.Unlock()
	var ctxMap map[string]interface{}
	if mp, ok := ctx.Value(ctxMapKey).(map[string]interface{}); ok {
		ctxMap = mp
	} else {
		ctxMap = make(map[string]interface{})
	}
	ctxMap[key] = value
	return context.WithValue(ctx, ctxMapKey, ctxMap)
}

func Logger(ctxOpt ...context.Context) *logrus.Entry {

	if len(ctxOpt) != 1 {
		return logrus.WithContext(context.TODO())
	}
	ctx := ctxOpt[0]
	if mp, ok := ctx.Value(ctxMapKey).(map[string]interface{}); ok {
		return logrus.WithContext(ctx).WithFields(mp)
	}
	return logrus.WithContext(ctx)

}
