package conf

import "github.com/weirwei/ikit/iutil"

var FeedSourceConf FeedSourceConfig

type FeedSourceConfig struct {
	DecoHack   string `yaml:"decohack"`
	Ruanyifeng string `yaml:"ruanyifeng"`
	Sspai      string `yaml:"sspai"`
	V2ex       string `yaml:"v2ex"`
	Zhihu      string `yaml:"zhihu"`
	Kitekagi   string `yaml:"kitekagi"`
	KitekagiAI string `yaml:"kitekagi-ai"`
}

func FeedSourceInit() {
	iutil.LoadYaml("yaml/feedsource.yaml", "conf", &FeedSourceConf)
}
