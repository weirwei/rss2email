package service

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/weirwei/ikit/ilog"
	"github.com/weirwei/rss2email/constants"
	"github.com/weirwei/rss2email/models"
)

const defaultFeedSourceReloadInterval = time.Minute

type scheduledFeed struct {
	entryID   cron.EntryID
	signature string
}

type scheduleState struct {
	entries map[uint64]scheduledFeed
}

func newScheduleState() *scheduleState {
	return &scheduleState{
		entries: make(map[uint64]scheduledFeed),
	}
}

// RunAllByDB executes all feed sources from database once.
func RunAllByDB(ctx context.Context) {
	feedSources, err := models.NewFeedSourceDao().ListAll(ctx)
	if err != nil {
		logScheduleError("list feed_sources failed", err)
		return
	}
	for _, feedSource := range feedSources {
		config, err := BuildConfigFromFeedSource(&feedSource)
		if err != nil {
			logScheduleError(fmt.Sprintf("build config failed [%s]", feedSource.Subscription), err)
			continue
		}
		if err := exponentialBackoffRetry(ctx, func(ctx context.Context) error {
			return CommonService(ctx, config)
		}); err != nil {
			logScheduleError(fmt.Sprintf("run service failed [%s]", feedSource.Subscription), err)
		}
	}
}

// ScheduleFromDB registers schedule tasks based on feed_sources table.
func ScheduleFromDB(ctx context.Context, c *cron.Cron) {
	if c == nil {
		return
	}
	reconcileSchedules(ctx, c, nil, newScheduleState())
}

// ScheduleFromDBDynamic polls feed_sources and applies incremental schedule changes without restarting the process.
func ScheduleFromDBDynamic(ctx context.Context, c *cron.Cron, reloadInterval time.Duration) {
	if c == nil {
		return
	}
	if reloadInterval <= 0 {
		reloadInterval = defaultFeedSourceReloadInterval
	}
	state := newScheduleState()
	reconcileSchedules(ctx, c, nil, state)
	spec := "@every " + reloadInterval.String()
	if _, err := c.AddFunc(spec, func() {
		reconcileSchedules(ctx, c, nil, state)
	}); err != nil {
		logScheduleError(fmt.Sprintf("register feed_sources reloader failed [%s]", spec), err)
	}
}

func reconcileSchedules(ctx context.Context, c *cron.Cron, feedSources []models.FeedSource, state *scheduleState) {
	if c == nil || state == nil {
		return
	}
	if feedSources == nil {
		var err error
		feedSources, err = models.NewFeedSourceDao().ListAll(ctx)
		if err != nil {
			logScheduleError("list feed_sources failed", err)
			return
		}
	}
	next := make(map[uint64]scheduledFeed, len(feedSources))
	// Keep a deterministic order for clearer logs and stable tests.
	sort.Slice(feedSources, func(i, j int) bool {
		return feedSources[i].ID < feedSources[j].ID
	})
	for _, feedSource := range feedSources {
		if feedSource.ID == 0 {
			logScheduleError(fmt.Sprintf("skip feed_source with invalid id [%s]", feedSource.Subscription), nil)
			continue
		}
		scheduleType := normalizeScheduleType(feedSource.ScheduleType)
		config, err := BuildConfigFromFeedSource(&feedSource)
		if err != nil {
			logScheduleError(fmt.Sprintf("build config failed [%s]", feedSource.Subscription), err)
			continue
		}
		signature := scheduleSignature(feedSource)
		if current, ok := state.entries[feedSource.ID]; ok && current.signature == signature {
			next[feedSource.ID] = current
			continue
		}
		if current, ok := state.entries[feedSource.ID]; ok {
			c.Remove(current.entryID)
		}
		entryID, ok := addSchedule(ctx, c, config, scheduleType, feedSource.CronSpec)
		if ok {
			next[feedSource.ID] = scheduledFeed{
				entryID:   entryID,
				signature: signature,
			}
		}
	}
	for id, current := range state.entries {
		if _, ok := next[id]; !ok {
			c.Remove(current.entryID)
		}
	}
	state.entries = next
}

func scheduleSignature(feedSource models.FeedSource) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s|%s",
		feedSource.Subscription,
		feedSource.FeedURL,
		strings.TrimSpace(feedSource.Language),
		feedSource.ContentField,
		normalizeScheduleType(feedSource.ScheduleType),
		strings.TrimSpace(feedSource.CronSpec),
	)
}

func normalizeScheduleType(scheduleType constants.ScheduleType) constants.ScheduleType {
	if scheduleType == "" {
		return constants.ScheduleTypeLive
	}
	return scheduleType
}

// exponentialBackoffRetry 使用指数退避策略重试函数调用
func exponentialBackoffRetry(ctx context.Context, fn func(ctx context.Context) error) error {
	const (
		initialDelay = 1 * time.Second // 初始延迟1秒
		maxDelay     = 5 * time.Minute // 最大延迟5分钟
		maxRetries   = 5               // 最大重试次数
		multiplier   = 2.0             // 指数增长因子
		jitter       = 0.1             // 抖动因子
	)

	delay := initialDelay

	for i := 0; i <= maxRetries; i++ {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 执行函数
		err := fn(ctx)
		if err == nil {
			// 成功执行，返回nil
			return nil
		}

		// 如果是最后一次重试，返回错误
		if i == maxRetries {
			return err
		}

		// 记录重试信息
		logScheduleError("函数执行失败，将进行重试", err)

		// 等待退避时间或上下文取消
		select {
		case <-time.After(delay):
			// 计算下一次的延迟时间（指数增长 + 随机抖动）
			delay = time.Duration(float64(delay) * multiplier)
			// 添加随机抖动避免惊群效应
			jitterAmount := time.Duration(float64(delay) * jitter * (0.5 - rand.Float64()) * 2)
			delay += jitterAmount
			// 确保不超过最大延迟时间
			if delay > maxDelay {
				delay = maxDelay
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

func addSchedule(ctx context.Context, c *cron.Cron, config Config, scheduleType constants.ScheduleType, cronSpec string) (cron.EntryID, bool) {
	switch scheduleType {
	case constants.ScheduleTypeStartup:
		return 0, false
	case constants.ScheduleTypeCron:
		spec := strings.TrimSpace(cronSpec)
		if spec == "" {
			logScheduleError(fmt.Sprintf("cron spec empty [%s]", config.Subscription), nil)
			return 0, false
		}
		if _, err := cron.ParseStandard(spec); err != nil {
			logScheduleError(fmt.Sprintf("cron spec invalid [%s]", config.Subscription), err)
			return 0, false
		}
		entryID, err := c.AddFunc(spec, func() {
			ilog.Infof("开始自定义订阅...[%s]", spec)
			if err := CommonService(ctx, config); err != nil {
				logScheduleError(fmt.Sprintf("cron run failed [%s]", config.Subscription), err)
			}
			time.Sleep(1 * time.Minute)
		})
		if err != nil {
			logScheduleError(fmt.Sprintf("register cron failed [%s]", config.Subscription), err)
			return 0, false
		}
		return entryID, true
	default:
		entryID, err := c.AddFunc("0 */1 * * *", func() {
			ilog.Info("开始实时订阅...")
			if err := CommonService(ctx, config); err != nil {
				logScheduleError(fmt.Sprintf("live run failed [%s]", config.Subscription), err)
			}
		})
		if err != nil {
			logScheduleError(fmt.Sprintf("register live failed [%s]", config.Subscription), err)
			return 0, false
		}
		return entryID, true
	}
}

func logScheduleError(message string, err error) {
	if err == nil {
		ilog.Warnf("%s", message)
		return
	}
	ilog.Warnf("%s，%v", message, err)
}
