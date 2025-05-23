package models

import (
	"context"
	"errors"
	"time"

	"github.com/weirwei/rss2email/helpers"
	"gorm.io/gorm"
)

// 用户订阅记录，记录用户和订阅的关系以及订阅进度
type UserSubscription struct {
	ID               uint64    `json:"id"`
	Email            string    `json:"email"`
	SubscriptionID   string    `json:"subscription_id"`
	SubscriptionType string    `json:"subscription_type"`
	Process          string    `json:"process"`
	ProcessType      string    `json:"process_type"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Deleted          int       `json:"deleted"`
}

func (u *UserSubscription) TableName() string {
	return "user_subscriptions"
}

type userSubscriptionDao struct{}

func NewUserSubscriptionDao() *userSubscriptionDao {
	return &userSubscriptionDao{}
}

func (u *userSubscriptionDao) Create(userSubscription *UserSubscription) error {
	return helpers.RSSSQLiteHelper.Create(userSubscription).Error
}

func (u *userSubscriptionDao) GetByEmailAndSubscriptionIDAndSubscriptionType(ctx context.Context, email string, subscriptionID string, subscriptionType string) (*UserSubscription, error) {
	var userSubscription UserSubscription
	db := helpers.RSSSQLiteHelper.WithContext(ctx)
	db.Where("email = ? AND subscription_id = ? AND subscription_type = ?", email, subscriptionID, subscriptionType)
	err := db.Take(&userSubscription).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &userSubscription, nil
}

// ListBySubscriptionIDAndSubscriptionType 根据订阅ID和订阅类型获取订阅列表
func (u *userSubscriptionDao) ListBySubscriptionIDAndSubscriptionType(ctx context.Context, subscriptionID string, subscriptionType string) ([]UserSubscription, error) {
	var userSubscriptions []UserSubscription
	db := helpers.RSSSQLiteHelper.WithContext(ctx)
	db.Where("subscription_id = ? AND subscription_type = ?", subscriptionID, subscriptionType)
	err := db.Find(&userSubscriptions).Error
	if err != nil {
		return nil, err
	}
	return userSubscriptions, nil
}

func (u *userSubscriptionDao) Update(ctx context.Context, id uint64, data map[string]interface{}) error {
	return helpers.RSSSQLiteHelper.WithContext(ctx).Model(&UserSubscription{}).Where("id = ?", id).Updates(data).Error
}
