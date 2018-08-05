/**
 * Created with IntelliJ IDEA.
 * Description: 
 * User: yangzhao
 * Date: 2018-08-01
 * Time: 16:16
 */
package main

import (
	"net/http"
	"shadowsocks-go/cmd/shadowsocks-console/filter"
	"errors"
	"shadowsocks-go/cmd/shadowsocks-console/db"
	"shadowsocks-go/cmd/shadowsocks-console/config"
	"os"
	"fmt"
	"shadowsocks-go/shadowsocks/encrypt"
	"shadowsocks-go/shadowsocks/common"
	"shadowsocks-go/shadowsocks/tcp"
	"shadowsocks-go/shadowsocks/log"
	"shadowsocks-go/shadowsocks"
)

type HttpServer struct {
	http.Server
}

func (server *HttpServer) StartServer() {
	log.Logger.Info("web server start " + server.Addr)
	err := server.ListenAndServe()
	if err != nil {
		log.Logger.Panic(err)
	}
}

func main() {
	web_filter.Register("/safe/**", func(rw http.ResponseWriter, r *http.Request) error {
		//return errors.New("解密失败")
		return nil
	})
	web_filter.Register("/safe/user/**", func(rw http.ResponseWriter, r *http.Request) error {
		return errors.New("请登录")
		//return nil
	})
	http.HandleFunc("/safe", web_filter.Handle(func(wr http.ResponseWriter, req *http.Request) error {
		wr.Write([]byte(req.RequestURI))
		return nil
	}))

	http.HandleFunc("/safe/user/test", web_filter.Handle(func(wr http.ResponseWriter, req *http.Request) error {
		wr.Write([]byte(req.RequestURI))
		return nil
	}))

	http.HandleFunc("/safe/user", web_filter.Handle(func(wr http.ResponseWriter, req *http.Request) error {
		wr.Write([]byte(req.RequestURI))
		return nil
	}))

	log.CreateLog("shadowsocks.log")
	shadowsocks.PrintVersion()

	err := db.CreateEngine()
	if err != nil {
		panic(err)
		return
	}
	db.StartAutoConnect()

	config := console_config.ParseConfig()
	if config != nil {
		if config.Method == "" {
			config.Method = "aes-256-cfb"
		}
		if err = encrypt.CheckCipherMethod(config.Method); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err = common.UnifyPortPassword(config); err != nil {
			os.Exit(1)
		}

		for port, password := range config.PortPassword {
			go tcp.GetTcpProxy().Run(port, password)
			//if ss.Udp {
			//	go ss.RunUDP(port, password)
			//} TODO UDP待处理
		}
	}

	server := &HttpServer{}
	server.Addr = ":8080"
	server.StartServer()
}
