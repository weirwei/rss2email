package service

import (
	"context"

	"github.com/weirwei/ikit/ilog"
	"github.com/weirwei/rss2email/constants"
	"github.com/weirwei/rss2email/models"
)

func Register(ctx context.Context, email string, subscriptions []constants.SubscriptionID) error {
	var userSubscriptions []models.UserSubscription
	for _, subscription := range subscriptions {
		userSubscription := models.UserSubscription{
			Email:            email,
			SubscriptionID:   subscription,
			SubscriptionType: constants.SubscriptionTypeRss, // 目前订阅类型都是 RSS
			Process:          "",                            // 初始进度为空
			ProcessType:      constants.ProcessTypeGUID,     // 默认进度记录类型为 GUID
		}
		userSubscriptions = append(userSubscriptions, userSubscription)
	}
	err := models.NewUserSubscriptionDao().BatchInsert(ctx, userSubscriptions)
	if err != nil {
		ilog.Warnf("批量插入用户订阅记录失败，%v", err)
		return err
	}
	return nil
}
