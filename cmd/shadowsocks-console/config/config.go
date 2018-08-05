/**
 * Created with IntelliJ IDEA.
 * Description: 
 * User: yangzhao
 * Date: 2018-08-05
 * Time: 15:19
 */
package console_config

import (
	"shadowsocks-go/cmd/shadowsocks-console/service"
	"strconv"
	"shadowsocks-go/shadowsocks/config"
)

func ParseConfig() *config.Config {
	ssConfig := &config.Config{}

	serverConfig := service.GetServerConfig()

	if serverConfig == nil {
		return nil
	}

	ssConfig.Server=serverConfig.ServerIp
	ssConfig.Method=serverConfig.EncryptWay
	ssConfig.ServerPort=serverConfig.ServerPort
	ssConfig.Timeout=serverConfig.ServerTimeout

	clientConfigs := service.GetClientConfig()
	portPassword := make(map[string]string)
	for _,v :=range clientConfigs {
		key := strconv.Itoa(v.CPort)
		portPassword[key]=v.CPassword
	}
	ssConfig.PortPassword = portPassword

	config.SysConfig = ssConfig
	return ssConfig
}
