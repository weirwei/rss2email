package constants

type SubscriptionID string
type SubscriptionType string
type FeedContentField string
type ScheduleType string

const (
	SubscriptionTypeRss SubscriptionType = "rss"
)

const (
	FeedContentFieldDescription FeedContentField = "description"
	FeedContentFieldContent     FeedContentField = "content"
)

const (
	ScheduleTypeStartup ScheduleType = "startup"
	ScheduleTypeLive    ScheduleType = "live"
	ScheduleTypeCron    ScheduleType = "cron"
)

type ProcessType string

const (
	ProcessTypeGUID ProcessType = "guid"
	ProcessTypeTime ProcessType = "time"
)
