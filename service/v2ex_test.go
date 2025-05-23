package service

import (
	"testing"

	"github.com/weirwei/rss2email/test"
)

func TestV2exService(t *testing.T) {
	test.Init()
	ctx := test.NewCtx()
	err := V2exService(ctx)
	if err != nil {
		t.Fatal(err)
	}
}
