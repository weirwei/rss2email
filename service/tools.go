package service

import (
	"context"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/weirwei/ikit/iutil"
	"github.com/weirwei/rss2email/models"
)

// SQLExec 使用 GORM 执行任意 SQL 语句
// query: SQL 语句
// args: 可选参数
// 返回值 any: 查询返回 []map[string]interface{}，非查询返回影响行数
func SQLExec(ctx context.Context, query string, args ...interface{}) error {
	result, err := models.NewUserSubscriptionDao().SQLExec(ctx, query, args...)
	if err != nil {
		return err
	}
	fmt.Println(iutil.ToJson(result))
	return nil
}
