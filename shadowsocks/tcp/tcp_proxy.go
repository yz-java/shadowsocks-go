/**
 * Created with IntelliJ IDEA.
 * Description: 
 * User: yangzhao
 * Date: 2018-08-05
 * Time: 20:47
 */
package tcp

import (
	"os"
	"net"
	"shadowsocks-go/shadowsocks/encrypt"
	"shadowsocks-go/shadowsocks/config"
	"shadowsocks-go/shadowsocks/log"
	"strings"
	"syscall"
	"io"
	"fmt"
	"encoding/binary"
	"strconv"
	"time"
	"shadowsocks-go/shadowsocks/proxy"
)

const (
	idType  = 0 // address type index
	idIP0   = 1 // ip address start index
	idDmLen = 1 // domain address length index
	idDm0   = 2 // domain address start index

	typeIPv4 = 1 // type is ipv4 address
	typeDm   = 3 // type is domain address
	typeIPv6 = 4 // type is ipv6 address

	lenIPv4   = net.IPv4len + 2 // ipv4 + 2port
	lenIPv6   = net.IPv6len + 2 // ipv6 + 2port
	lenDmBase = 2               // 1addrLen + 2port, plus addrLen
	// lenHmacSha1 = 10

	logCntDelta = 100

	AddrMask        byte = 0xf
)

var connCnt int
var nextLogConnCnt = logCntDelta
var readTimeout time.Duration

var tcpProxy *TcpProxy

type TcpProxy struct {
}

func GetTcpProxy() *TcpProxy {

	if tcpProxy == nil {
		tcpProxy = &TcpProxy{}
	}
	return tcpProxy
}

func (this *TcpProxy) Run(port, password string) {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Logger.Info("error listening port %v: %v\n", port, err)
		os.Exit(1)
	}
	proxy.GetManager().Add(port, password, ln)
	var cipher *encrypt.Cipher
	log.Logger.Info("server listening port %v ...\n", port)
	for {
		conn, err := ln.Accept()
		if err != nil {
			// listener maybe closed to update password
			log.Logger.Info("accept error: %v\n", err)
			return
		}
		// Creating cipher upon first connection.
		if cipher == nil {
			log.Logger.Info("creating cipher for port:", port)
			cipher, err = encrypt.NewCipher(config.SysConfig.Method, password)
			if err != nil {
				log.Logger.Info("Error generating cipher for port: %s %v\n", port, err)
				conn.Close()
				continue
			}
		}
		go this.handleConnection(proxy.NewConn(conn, cipher.Copy()), port)
	}
}

func (this *TcpProxy) handleConnection(conn *proxy.Conn, port string) {
	var host string

	connCnt++ // this maybe not accurate, but should be enough
	if connCnt-nextLogConnCnt >= 0 {
		// XXX There's no xadd in the atomic package, so it's difficult to log
		// the message only once with low cost. Also note nextLogConnCnt maybe
		// added twice for current peak connection number level.
		log.Logger.Info("Number of client connections reaches %d\n", nextLogConnCnt)
		nextLogConnCnt += logCntDelta
	}

	log.Logger.Info("new client %s->%s\n", conn.RemoteAddr().String(), conn.LocalAddr())
	closed := false
	defer func() {
		log.Logger.Info("closed pipe %s<->%s\n", conn.RemoteAddr().String(), host)
		connCnt--
		if !closed {
			conn.Close()
		}
	}()

	host, err := getRequest(conn)
	if err != nil {
		log.Logger.Info("error getting request", conn.RemoteAddr().String(), conn.LocalAddr(), err)
		closed = true
		return
	}
	// ensure the host does not contain some illegal characters, NUL may panic on Win32
	if strings.ContainsRune(host, 0x00) {
		log.Logger.Info("invalid domain name.")
		closed = true
		return
	}
	log.Logger.Info("connecting", host)
	remote, err := net.Dial("tcp", host)
	if err != nil {
		if ne, ok := err.(*net.OpError); ok && (ne.Err == syscall.EMFILE || ne.Err == syscall.ENFILE) {
			// log too many open file error
			// EMFILE is process reaches open file limits, ENFILE is system limit
			log.Logger.Info("dial error:", err)
		} else {
			log.Logger.Info("error connecting to:", host, err)
		}
		return
	}
	defer func() {
		if !closed {
			remote.Close()
		}
	}()

	log.Logger.Info("piping %s<->%s", conn.RemoteAddr().String(), host)

	go func() {
		proxy.PipeThenClose(conn, remote, func(Traffic int) {
			proxy.GetManager().AddTraffic(port, Traffic)
		})
	}()

	proxy.PipeThenClose(remote, conn, func(Traffic int) {
		proxy.GetManager().AddTraffic(port, Traffic)
	})

	closed = true
	return
}

func (this *TcpProxy) Stop(port string) {

}

func getRequest(conn *proxy.Conn) (host string, err error) {

	if readTimeout != 0 {
		conn.SetReadDeadline(time.Now().Add(readTimeout))
	}

	// buf size should at least have the same size with the largest possible
	// request size (when addrType is 3, domain name has at most 256 bytes)
	// 1(addrType) + 1(lenByte) + 255(max length address) + 2(port) + 10(hmac-sha1)
	buf := make([]byte, 269)
	// read till we get possible domain length field
	if _, err = io.ReadFull(conn, buf[:idType+1]); err != nil {
		return
	}

	var reqStart, reqEnd int
	addrType := buf[idType]
	switch addrType & AddrMask {
	case typeIPv4:
		reqStart, reqEnd = idIP0, idIP0+lenIPv4
	case typeIPv6:
		reqStart, reqEnd = idIP0, idIP0+lenIPv6
	case typeDm:
		if _, err = io.ReadFull(conn, buf[idType+1:idDmLen+1]); err != nil {
			return
		}
		reqStart, reqEnd = idDm0, idDm0+int(buf[idDmLen])+lenDmBase
	default:
		err = fmt.Errorf("addr type %d not supported", addrType&AddrMask)
		return
	}

	if _, err = io.ReadFull(conn, buf[reqStart:reqEnd]); err != nil {
		return
	}

	// Return string for typeIP is not most efficient, but browsers (Chrome,
	// Safari, Firefox) all seems using typeDm exclusively. So this is not a
	// big problem.
	switch addrType & AddrMask {
	case typeIPv4:
		host = net.IP(buf[idIP0:idIP0+net.IPv4len]).String()
	case typeIPv6:
		host = net.IP(buf[idIP0:idIP0+net.IPv6len]).String()
	case typeDm:
		host = string(buf[idDm0 : idDm0+int(buf[idDmLen])])
	}
	// parse port
	port := binary.BigEndian.Uint16(buf[reqEnd-2 : reqEnd])
	host = net.JoinHostPort(host, strconv.Itoa(int(port)))
	return
}
