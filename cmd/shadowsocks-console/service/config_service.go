/**
 * Created with IntelliJ IDEA.
 * Description: 
 * User: yangzhao
 * Date: 2018-08-05
 * Time: 14:57
 */
package service

import (
	"shadowsocks-go/cmd/shadowsocks-console/model"
	"shadowsocks-go/cmd/shadowsocks-console/dao"
)

func GetServerConfig() *model.ServerConfig {
	serverConfigs := dao.GetServerConfigDao().GetList()
	if len(serverConfigs) == 0 {
		return nil
	}
	return &serverConfigs[0]
}

func GetClientConfig() []model.ClientConfig  {
	return dao.GetClientConfigDao().GetList()
}
