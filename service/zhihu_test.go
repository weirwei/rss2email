package service

import (
	"testing"

	"github.com/weirwei/rss2email/test"
)

func TestZhihuService(t *testing.T) {
	test.Init()
	ctx := test.NewCtx()
	err := ZhihuService(ctx)
	if err != nil {
		t.Fatal(err)
	}
}
