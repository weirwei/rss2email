package models

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/weirwei/rss2email/constants"
	"github.com/weirwei/rss2email/helpers"
	"gorm.io/gorm"
)

// FeedSource RSS 源配置
type FeedSource struct {
	ID           uint64                     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Subscription constants.SubscriptionID   `json:"subscription_id" gorm:"column:subscription_id"`
	Name         string                     `json:"name" gorm:"column:name"`
	FeedURL      string                     `json:"feed_url" gorm:"column:feed_url"`
	Language     string                     `json:"language" gorm:"column:language"`
	ContentField constants.FeedContentField `json:"content_field" gorm:"column:content_field"`
	ScheduleType constants.ScheduleType     `json:"schedule_type" gorm:"column:schedule_type"`
	CronSpec     string                     `json:"cron_spec" gorm:"column:cron_spec"`
	CreatedAt    time.Time                  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time                  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	Deleted      int                        `json:"deleted" gorm:"column:deleted"`
}

func (f *FeedSource) TableName() string {
	return "feed_sources"
}

type feedSourceDao struct{}

func NewFeedSourceDao() *feedSourceDao {
	return &feedSourceDao{}
}

// GetBySubscriptionID 根据订阅ID获取订阅源配置
func (f *feedSourceDao) GetBySubscriptionID(ctx context.Context, subscriptionID constants.SubscriptionID) (*FeedSource, error) {
	var feedSource FeedSource
	db := helpers.RSSSQLiteHelper.WithContext(ctx)
	db = db.Where("subscription_id = ? AND deleted = 0", subscriptionID)
	err := db.Take(&feedSource).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &feedSource, nil
}

// ExistsBySubscriptionID 判断订阅源是否存在
func (f *feedSourceDao) ExistsBySubscriptionID(ctx context.Context, subscriptionID constants.SubscriptionID) (bool, error) {
	var count int64
	db := helpers.RSSSQLiteHelper.WithContext(ctx)
	err := db.Model(&FeedSource{}).
		Where("subscription_id = ? AND deleted = 0", subscriptionID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ListAll 获取所有订阅源
func (f *feedSourceDao) ListAll(ctx context.Context) ([]FeedSource, error) {
	var feedSources []FeedSource
	db := helpers.RSSSQLiteHelper.WithContext(ctx)
	err := db.Where("deleted = 0").Find(&feedSources).Error
	if err != nil {
		return nil, err
	}
	return feedSources, nil
}

// Upsert 根据订阅ID插入或更新订阅源
func (f *feedSourceDao) Upsert(ctx context.Context, feedSource *FeedSource) error {
	if feedSource == nil {
		return nil
	}
	var existing FeedSource
	db := helpers.RSSSQLiteHelper.WithContext(ctx)
	err := db.Where("subscription_id = ?", feedSource.Subscription).Take(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if f.hasLanguageColumn(ctx) {
				return db.Create(feedSource).Error
			}
			return db.Omit("language").Create(feedSource).Error
		}
		return err
	}
	updates := map[string]interface{}{
		"name":          feedSource.Name,
		"feed_url":      feedSource.FeedURL,
		"content_field": feedSource.ContentField,
		"schedule_type": feedSource.ScheduleType,
		"cron_spec":     feedSource.CronSpec,
		"deleted":       0,
		"updated_at":    time.Now(),
	}
	if f.hasLanguageColumn(ctx) {
		updates["language"] = feedSource.Language
	}
	return db.Model(&FeedSource{}).Where("id = ?", existing.ID).Updates(updates).Error
}

// UpdateByID 根据主键更新订阅源
func (f *feedSourceDao) UpdateByID(ctx context.Context, id uint64, data map[string]interface{}) error {
	if id == 0 || len(data) == 0 {
		return nil
	}
	if !f.hasLanguageColumn(ctx) {
		delete(data, "language")
	}
	return helpers.RSSSQLiteHelper.WithContext(ctx).
		Model(&FeedSource{}).
		Where("id = ?", id).
		Updates(data).Error
}

// SoftDeleteByID 软删除订阅源
func (f *feedSourceDao) SoftDeleteByID(ctx context.Context, id uint64) error {
	return helpers.RSSSQLiteHelper.WithContext(ctx).
		Model(&FeedSource{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"deleted":    1,
			"updated_at": time.Now(),
		}).Error
}

func (f *feedSourceDao) hasLanguageColumn(ctx context.Context) bool {
	rows, err := helpers.RSSSQLiteHelper.WithContext(ctx).Raw("PRAGMA table_info(feed_sources)").Rows()
	if err != nil {
		return false
	}
	defer rows.Close()
	var (
		cid       int
		name      string
		colType   string
		notnull   int
		dfltValue any
		pk        int
	)
	for rows.Next() {
		if err = rows.Scan(&cid, &name, &colType, &notnull, &dfltValue, &pk); err != nil {
			continue
		}
		if strings.EqualFold(name, "language") {
			return true
		}
	}
	return false
}
