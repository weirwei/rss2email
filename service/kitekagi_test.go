package service

import (
	"testing"

	"github.com/weirwei/rss2email/test"
)

func TestKitekagiService(t *testing.T) {
	test.Init()
	ctx := test.NewCtx()
	err := KitekagiService(ctx)
	if err != nil {
		t.Fatal(err)
	}
}

func TestKitekagiAIService(t *testing.T) {
	test.Init()
	ctx := test.NewCtx()
	err := KitekagiAIService(ctx)
	if err != nil {
		t.Fatal(err)
	}
}
