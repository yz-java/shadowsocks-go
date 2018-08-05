/**
 * Created with IntelliJ IDEA.
 * Description: 
 * User: yangzhao
 * Date: 2018-08-05
 * Time: 14:46
 */
package model

import "time"

type ServerConfig struct {

	Id int

	ServerIp string

	EncryptWay string

	ServerPort int

	ServerTimeout int

	CreateTime time.Time

}

