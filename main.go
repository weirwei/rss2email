package main

import (
	"fmt"
	"os"

	"github.com/weirwei/ikit/iutil"
	"github.com/weirwei/rss2email/cmd"
	"github.com/weirwei/rss2email/conf"
	"github.com/weirwei/rss2email/helpers"
)

func main() {
	iutil.SetRootPath(helpers.ResolveProjectRoot())
	conf.InitConfig()
	helpers.Init()
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
