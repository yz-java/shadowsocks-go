package db

import (
	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
	"time"
	"log/syslog"
	"shadowsocks-go/shadowsocks/log"
)


var Engine *xorm.Engine = nil


func CreateEngine() error  {
	var err error
	Engine,err=xorm.NewEngine("sqlite3","ss_console.db")
	if err != nil {
		return err
	}
	logWriter, err := syslog.New(syslog.LOG_DEBUG, "rest-xorm-example")
	if err != nil {
		return err
	}
	logger := xorm.NewSimpleLogger(logWriter)
	Engine.SetLogger(logger)
	Engine.ShowSQL(true)
	return nil
}
//自动重连
func StartAutoConnect() {
	go func() {
		for{
			time.Sleep(10*time.Second)
			err := Engine.Ping()
			if err != nil {
				log.Logger.Error(err)
				CreateEngine()
			}
		}
	}()

}
