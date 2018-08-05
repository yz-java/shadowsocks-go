package main

import (
	"flag"
	"net"
	"os"
	"runtime"
	ss "shadowsocks-go/shadowsocks"
	"fmt"
	config2 "shadowsocks-go/shadowsocks/config"
	"shadowsocks-go/shadowsocks/encrypt"
	"shadowsocks-go/shadowsocks/common"
	"shadowsocks-go/shadowsocks/log"
	"shadowsocks-go/shadowsocks/tcp"
)

var configFile string
var config *config2.Config


func main() {

	var cmdConfig config2.Config
	var core int

	flag.StringVar(&configFile, "c", "config.json", "specify config file")
	flag.StringVar(&cmdConfig.Password, "k", "", "password")
	flag.IntVar(&cmdConfig.ServerPort, "p", 0, "server port")
	flag.IntVar(&cmdConfig.Timeout, "t", 300, "timeout in seconds")
	flag.StringVar(&cmdConfig.Method, "m", "", "encryption method, default: aes-256-cfb")
	flag.IntVar(&core, "core", 0, "maximum number of CPU cores to use, default is determinied by Go runtime")
	flag.BoolVar(&config2.Udp, "u", false, "UDP Relay")
	flag.StringVar(&config2.ManagerAddr, "manager-address", "", "shadowsocks manager listening address")
	flag.Parse()

	ss.PrintVersion()

	var err error
	config, err = config2.ParseConfig(configFile)
	config2.SysConfig=config
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "error reading %s: %v\n", configFile, err)
			os.Exit(1)
		}
		config = &cmdConfig
		config2.UpdateConfig(config, config)
	} else {
		config2.UpdateConfig(config, &cmdConfig)
	}
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
	if core > 0 {
		runtime.GOMAXPROCS(core)
	}
	for port, password := range config.PortPassword {
		go tcp.GetTcpProxy().Run(port, password)
		//if config2.Udp {
		//	go udp.RunUDP(port, password)
		//}
	}

	if config2.ManagerAddr != "" {
		addr, err := net.ResolveUDPAddr("udp", config2.ManagerAddr)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Can't resolve address: ", err)
			os.Exit(1)
		}
		conn, err := net.ListenUDP("udp", addr)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error listening:", err)
			os.Exit(1)
		}
		log.Logger.Info("manager listening udp addr %v ...\n", config2.ManagerAddr)
		defer conn.Close()
		//go ss.ManagerDaemon(conn)
	}

	i := make(chan int)
	<-i
}
