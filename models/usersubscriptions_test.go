package models

import (
	"testing"

	"github.com/weirwei/ikit/iutil"
	"github.com/weirwei/rss2email/constants"
	"github.com/weirwei/rss2email/test"
)

func TestGetByEmailAndSubscriptionIDAndSubscriptionType(t *testing.T) {
	test.Init()
	ctx := test.NewCtx()
	t.Run("succ", func(t *testing.T) {
		dao := NewUserSubscriptionDao()
		res, err := dao.GetByEmailAndSubscriptionIDAndSubscriptionType(ctx, "weirwei@qq.com", constants.SubscriptionRuanyifeng, constants.SubscriptionTypeRss)
		if err != nil {
			t.Fatal(err.Error())
		} else if res == nil {
			t.Fatalf("res is nil")
		}
		t.Log(iutil.ToJson(res))
	})
}

func TestListBySubscriptionIDAndSubscriptionType(t *testing.T) {
	test.Init()
	ctx := test.NewCtx()
	t.Run("success", func(t *testing.T) {
		dao := NewUserSubscriptionDao()
		res, err := dao.ListBySubscriptionIDAndSubscriptionType(ctx, constants.SubscriptionRuanyifeng, constants.SubscriptionTypeRss)
		if err != nil {
			t.Fatal(err.Error())
		} else if len(res) == 0 {
			t.Fatalf("res is empty")
		}
		t.Log(iutil.ToJson(res))
	})
}
