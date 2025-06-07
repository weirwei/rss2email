package constants

const (
	SubscriptionTypeRss = "rss"

	SubscriptionDecoHack   = "decohack"
	SubscriptionRuanyifeng = "ruanyifeng"
	SubscriptionV2ex       = "v2ex"
)

var AllSubscription = []string{
	SubscriptionDecoHack,
	SubscriptionRuanyifeng,
	SubscriptionV2ex,
}

const (
	ProcessTypeGUID = "guid"
	ProcessTypeTime = "time"
)
