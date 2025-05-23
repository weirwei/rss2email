// Package test
// 单元测试工具包
package test

import (
	"context"
	"path"
	"runtime"

	"github.com/weirwei/ikit/iutil"
	"github.com/weirwei/rss2email/conf"
	"github.com/weirwei/rss2email/helpers"
)

func Init() {
	_, file, _, _ := runtime.Caller(0)
	rootDir := path.Dir(path.Dir(file))
	iutil.SetRootPath(rootDir)
	conf.InitConfig()

	helpers.Init()
}

func NewCtx() context.Context {
	return context.Background()
}
