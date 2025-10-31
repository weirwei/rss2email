package constants

type SubscriptionID string
type SubscriptionType string

const (
	SubscriptionTypeRss SubscriptionType = "rss"

	SubscriptionDecoHack       SubscriptionID = "decohack"
	SubscriptionRuanyifeng     SubscriptionID = "ruanyifeng"
	SubscriptionV2ex           SubscriptionID = "v2ex"
	SubscriptionSspai          SubscriptionID = "sspai"
	SubscriptionZhihu          SubscriptionID = "zhihu"
	SubscriptionKitekagi       SubscriptionID = "kitekagi"
	SubscriptionKitekagiAI     SubscriptionID = "kitekagi-ai"
	SubscriptionAIInsightDaily SubscriptionID = "ai-insight-daily"
)

var AllSubscription = []SubscriptionID{
	SubscriptionDecoHack,
	SubscriptionRuanyifeng,
	SubscriptionV2ex,
	SubscriptionSspai,
	SubscriptionZhihu,
	SubscriptionKitekagi,
	SubscriptionKitekagiAI,
	SubscriptionAIInsightDaily,
}

type ProcessType string

const (
	ProcessTypeGUID ProcessType = "guid"
	ProcessTypeTime ProcessType = "time"
)
