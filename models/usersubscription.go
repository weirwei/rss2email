package models

import (
	"context"
	"errors"
	"time"

	"github.com/weirwei/rss2email/constants"
	"github.com/weirwei/rss2email/helpers"
	"gorm.io/gorm"
)

// 用户订阅记录，记录用户和订阅的关系以及订阅进度
type UserSubscription struct {
	ID               uint64                     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Email            string                     `json:"email" gorm:"column:email"`
	SubscriptionID   constants.SubscriptionID   `json:"subscription_id" gorm:"column:subscription_id"`
	SubscriptionType constants.SubscriptionType `json:"subscription_type" gorm:"column:subscription_type"`
	Process          string                     `json:"process" gorm:"column:process"`
	ProcessType      constants.ProcessType      `json:"process_type" gorm:"column:process_type"`
	CreatedAt        time.Time                  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time                  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	Deleted          int                        `json:"deleted" gorm:"column:deleted"`
}

func (u *UserSubscription) TableName() string {
	return "user_subscriptions"
}

type userSubscriptionDao struct{}

func NewUserSubscriptionDao() *userSubscriptionDao {
	return &userSubscriptionDao{}
}

func (u *userSubscriptionDao) BatchInsert(ctx context.Context, userSubscriptions []UserSubscription) error {
	if len(userSubscriptions) == 0 {
		return nil
	}
	db := helpers.RSSSQLiteHelper.WithContext(ctx)
	return db.Create(&userSubscriptions).Error
}

// GetByEmailAndSubscriptionIDAndSubscriptionType 根据邮箱、订阅ID和订阅类型获取用户订阅记录

func (u *userSubscriptionDao) GetByEmailAndSubscriptionIDAndSubscriptionType(ctx context.Context, email string, subscriptionID constants.SubscriptionID, subscriptionType constants.SubscriptionType) (*UserSubscription, error) {
	var userSubscription UserSubscription
	db := helpers.RSSSQLiteHelper.WithContext(ctx)
	db = db.Where("email = ? AND subscription_id = ? AND subscription_type = ? AND deleted = 0", email, subscriptionID, subscriptionType)
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
func (u *userSubscriptionDao) ListBySubscriptionIDAndSubscriptionType(ctx context.Context, subscriptionID constants.SubscriptionID, subscriptionType constants.SubscriptionType) ([]UserSubscription, error) {
	var userSubscriptions []UserSubscription
	db := helpers.RSSSQLiteHelper.WithContext(ctx)
	db = db.Where("subscription_id = ? AND subscription_type = ? AND deleted = 0", subscriptionID, subscriptionType)
	err := db.Find(&userSubscriptions).Error
	if err != nil {
		return nil, err
	}
	return userSubscriptions, nil
}

func (u *userSubscriptionDao) Update(ctx context.Context, id uint64, data map[string]interface{}) error {
	return helpers.RSSSQLiteHelper.WithContext(ctx).Model(&UserSubscription{}).Where("id = ?", id).Updates(data).Error
}

// ListAll 获取所有未删除用户订阅
func (u *userSubscriptionDao) ListAll(ctx context.Context) ([]UserSubscription, error) {
	var userSubscriptions []UserSubscription
	db := helpers.RSSSQLiteHelper.WithContext(ctx)
	err := db.Where("deleted = 0").Order("id DESC").Find(&userSubscriptions).Error
	if err != nil {
		return nil, err
	}
	return userSubscriptions, nil
}

// UpsertByUnique 根据 (email, subscription_id, subscription_type) 进行插入或恢复
func (u *userSubscriptionDao) UpsertByUnique(ctx context.Context, userSubscription *UserSubscription) error {
	if userSubscription == nil {
		return nil
	}
	db := helpers.RSSSQLiteHelper.WithContext(ctx)
	var existing UserSubscription
	err := db.Where("email = ? AND subscription_id = ? AND subscription_type = ?",
		userSubscription.Email, userSubscription.SubscriptionID, userSubscription.SubscriptionType).
		Take(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return db.Create(userSubscription).Error
		}
		return err
	}
	updates := map[string]interface{}{
		"deleted":    0,
		"updated_at": time.Now(),
	}
	return db.Model(&UserSubscription{}).Where("id = ?", existing.ID).Updates(updates).Error
}

// SoftDeleteByID 软删除用户订阅
func (u *userSubscriptionDao) SoftDeleteByID(ctx context.Context, id uint64) error {
	return helpers.RSSSQLiteHelper.WithContext(ctx).
		Model(&UserSubscription{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"deleted":    1,
			"updated_at": time.Now(),
		}).Error
}

// SQLExec 使用 GORM 执行任意 SQL 语句
// query: SQL 语句
// args: 可选参数
// 返回值 any: 查询返回 []map[string]interface{}，非查询返回影响行数
func (u *userSubscriptionDao) SQLExec(ctx context.Context, query string, args ...interface{}) (any, error) {
	db := helpers.RSSSQLiteHelper
	if isSelect(query) {
		var results []map[string]interface{}
		rows, err := db.Raw(query, args...).Rows()
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		cols, err := rows.Columns()
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			columns := make([]interface{}, len(cols))
			columnPointers := make([]interface{}, len(cols))
			for i := range columns {
				columnPointers[i] = &columns[i]
			}
			if err := rows.Scan(columnPointers...); err != nil {
				return nil, err
			}
			rowMap := make(map[string]interface{})
			for i, colName := range cols {
				val := columnPointers[i].(*interface{})
				rowMap[colName] = *val
			}
			results = append(results, rowMap)
		}
		return results, nil
	} else {
		res := db.Exec(query, args...)
		return res.RowsAffected, res.Error
	}
}

// isSelect 判断是否为 SELECT 查询
func isSelect(query string) bool {
	for i := 0; i < len(query) && (query[i] == ' ' || query[i] == '\t' || query[i] == '\n'); i++ {
		query = query[1:]
	}
	return len(query) >= 6 && (query[:6] == "SELECT" || query[:6] == "select")
}
