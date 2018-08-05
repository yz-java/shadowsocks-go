/**
 * Created with IntelliJ IDEA.
 * Description: 
 * User: yangzhao
 * Date: 2018-08-05
 * Time: 15:22
 */
package dao

import (
	"github.com/go-xorm/xorm"
	"shadowsocks-go/cmd/shadowsocks-console/db"
	"shadowsocks-go/cmd/shadowsocks-console/model"
)

var clientConfigDao *ClientConfigDao = nil

type ClientConfigDao struct {
	*xorm.Engine
	tableName string
}

func GetClientConfigDao() *ClientConfigDao  {
	if clientConfigDao == nil {
		clientConfigDao = &ClientConfigDao{db.Engine,"ss_client_config"}
	}
	return clientConfigDao
}

func (this *ClientConfigDao)GetList() []model.ClientConfig  {
	configs := make([]model.ClientConfig, 0)
	this.Table(this.tableName).Where("c_status=?",1).Find(&configs)
	return configs
}
