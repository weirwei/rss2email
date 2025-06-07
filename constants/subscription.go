package constants

type SubscriptionID string
type SubscriptionType string

const (
	SubscriptionTypeRss SubscriptionType = "rss"

	SubscriptionDecoHack   SubscriptionID = "decohack"
	SubscriptionRuanyifeng SubscriptionID = "ruanyifeng"
	SubscriptionV2ex       SubscriptionID = "v2ex"
	SubscriptionSspai      SubscriptionID = "sspai"
)

var AllSubscription = []SubscriptionID{
	SubscriptionDecoHack,
	SubscriptionRuanyifeng,
	SubscriptionV2ex,
	SubscriptionSspai,
}

type ProcessType string

const (
	ProcessTypeGUID ProcessType = "guid"
	ProcessTypeTime ProcessType = "time"
)
