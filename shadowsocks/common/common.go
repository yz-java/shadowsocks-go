/**
 * Created with IntelliJ IDEA.
 * Description: 
 * User: yangzhao
 * Date: 2018-08-05
 * Time: 21:46
 */
package common

import (
	"fmt"
	"os"
	"strconv"
	"shadowsocks-go/shadowsocks/config"
	"errors"
)

const AddrMask        byte = 0xf

func UnifyPortPassword(config *config.Config) (err error) {
	if len(config.PortPassword) == 0 { // this handles both nil PortPassword and empty one
		if !enoughOptions(config) {
			fmt.Fprintln(os.Stderr, "must specify both port and password")
			return errors.New("not enough options")
		}
		port := strconv.Itoa(config.ServerPort)
		config.PortPassword = map[string]string{port: config.Password}
	} else {
		if config.Password != "" || config.ServerPort != 0 {
			fmt.Fprintln(os.Stderr, "given port_password, ignore server_port and password option")
		}
	}
	return
}

func enoughOptions(config *config.Config) bool {
	return config.ServerPort != 0 && config.Password != ""
}
