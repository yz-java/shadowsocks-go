/**
 * Created with IntelliJ IDEA.
 * Description: 
 * User: yangzhao
 * Date: 2018-08-05
 * Time: 14:58
 */
package dao

import (
	"github.com/go-xorm/xorm"
	"shadowsocks-go/cmd/shadowsocks-console/db"
	"shadowsocks-go/cmd/shadowsocks-console/model"
)

var configDao *ServerConfigDao = nil

type ServerConfigDao struct {

	*xorm.Engine

	tableName string

}

func GetServerConfigDao() *ServerConfigDao {
	if configDao == nil {
		configDao = &ServerConfigDao{db.Engine,"ss_server_config"}
	}
	return configDao
}

func (this *ServerConfigDao)GetList() []model.ServerConfig  {
	configs := make([]model.ServerConfig,0)
	this.Table(this.tableName).Find(&configs)
	return configs
}

