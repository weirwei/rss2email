package helpers

import (
	"github.com/weirwei/rss2email/conf"
	"github.com/weirwei/rss2email/email"
)

var EmailHelper *email.Email

func InitEmailHelper() {
	EmailHelper = email.NewEmail(&email.Config{
		Host: conf.EmailConf.Host,
		Port: conf.EmailConf.Port,
		User: conf.EmailConf.User,
		Pass: conf.EmailConf.Pass,
	})
}
