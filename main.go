package main

import (
	"context"
	"log"
	"time"

	"github.com/weirwei/ikit/ilog"
	"github.com/weirwei/rss2email/internal/service"
)

func main() {
	for {
		log.Println("开始执行 DecoHackService...")
		err := service.DecoHackService(context.Background())
		if err != nil {
			ilog.Error(err)
		}
		time.Sleep(6 * time.Hour)
	}
}
