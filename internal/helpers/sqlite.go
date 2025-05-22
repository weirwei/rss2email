package helpers

import (
	"log"

	"github.com/weirwei/ikit/iutil"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	RSSSQLiteHelper *gorm.DB
	err             error
)

func InitSQLite() {
	RSSSQLiteHelper, err = gorm.Open(sqlite.Open(iutil.GetRootPath()+"/db/rss.sqlite"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
}
