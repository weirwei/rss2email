package conf

import "github.com/weirwei/ikit/iutil"

var EmailConf EmailConfig

type EmailConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
}

func EmailInit() {
	iutil.LoadYaml("yaml/email.yaml", "conf", &EmailConf)
}
