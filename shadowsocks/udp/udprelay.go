package udp

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"shadowsocks-go/shadowsocks/common"
	"shadowsocks-go/shadowsocks/log"
)

const (
	idType  = 0 // address type index
	idIP0   = 1 // ip address start index
	idDmLen = 1 // domain address length index
	idDm0   = 2 // domain address start index

	typeIPv4 = 1 // type is ipv4 address
	typeDm   = 3 // type is domain address
	typeIPv6 = 4 // type is ipv6 address

	lenIPv4     = 1 + net.IPv4len + 2 // 1addrType + ipv4 + 2port
	lenIPv6     = 1 + net.IPv6len + 2 // 1addrType + ipv6 + 2port
	lenDmBase   = 1 + 1 + 2           // 1addrType + 1addrLen + 2port, plus addrLen
	// lenHmacSha1 = 10
)

var (
	reqList            = newReqList()
	natlist            = newNatTable()
	udpTimeout         = 30 * time.Second
	reqListRefreshTime = 5 * time.Minute
)

type natTable struct {
	sync.Mutex
	conns map[string]net.PacketConn
}

func newNatTable() *natTable {
	return &natTable{conns: map[string]net.PacketConn{}}
}

func (table *natTable) Delete(index string) net.PacketConn {
	table.Lock()
	defer table.Unlock()
	c, ok := table.conns[index]
	if ok {
		delete(table.conns, index)
		return c
	}
	return nil
}

func (table *natTable) Get(index string) (c net.PacketConn, ok bool, err error) {
	table.Lock()
	defer table.Unlock()
	c, ok = table.conns[index]
	if !ok {
		c, err = net.ListenPacket("udp", "")
		if err != nil {
			return nil, false, err
		}
		table.conns[index] = c
	}
	return
}

//noinspection GoRedundantParens
type requestHeaderList struct {
	sync.Mutex
	List map[string]([]byte)
}

//noinspection GoRedundantParens
func newReqList() *requestHeaderList {
	ret := &requestHeaderList{List: map[string]([]byte){}}
	go func() {
		for {
			time.Sleep(reqListRefreshTime)
			ret.Refresh()
		}
	}()
	return ret
}

func (r *requestHeaderList) Refresh() {
	r.Lock()
	defer r.Unlock()
	for k := range r.List {
		delete(r.List, k)
	}
}

func (r *requestHeaderList) Get(dstaddr string) (req []byte, ok bool) {
	r.Lock()
	defer r.Unlock()
	req, ok = r.List[dstaddr]
	return
}

func (r *requestHeaderList) Put(dstaddr string, req []byte) {
	r.Lock()
	defer r.Unlock()
	r.List[dstaddr] = req
	return
}

func parseHeaderFromAddr(addr net.Addr) ([]byte, int) {
	// if the request address type is domain, it cannot be reverselookuped
	ip, port, err := net.SplitHostPort(addr.String())
	if err != nil {
		return nil, 0
	}
	buf := make([]byte, 20)
	IP := net.ParseIP(ip)
	b1 := IP.To4()
	iplen := 0
	if b1 == nil { //ipv6
		b1 = IP.To16()
		buf[0] = typeIPv6
		iplen = net.IPv6len
	} else { //ipv4
		buf[0] = typeIPv4
		iplen = net.IPv4len
	}
	copy(buf[1:], b1)
	port_i, _ := strconv.Atoi(port)
	binary.BigEndian.PutUint16(buf[1+iplen:], uint16(port_i))
	return buf[:1+iplen+2], 1 + iplen + 2
}

func Pipeloop(write net.PacketConn, writeAddr net.Addr, readClose net.PacketConn, addTraffic func(int)) {
	buf := common.LB.Get()
	defer common.LB.Put(buf)
	defer readClose.Close()
	for {
		readClose.SetDeadline(time.Now().Add(udpTimeout))
		n, raddr, err := readClose.ReadFrom(buf)
		if err != nil {
			if ne, ok := err.(*net.OpError); ok {
				if ne.Err == syscall.EMFILE || ne.Err == syscall.ENFILE {
					// log too many open file error
					// EMFILE is process reaches open file limits, ENFILE is system limit
					log.Logger.Info("[udp]read error:", err)
				}
			}
			log.Logger.Debug("[udp]closed pipe %s<-%s\n", writeAddr, readClose.LocalAddr())
			return
		}
		// need improvement here
		if req, ok := reqList.Get(raddr.String()); ok {
			n, _ := write.WriteTo(append(req, buf[:n]...), writeAddr)
			addTraffic(n)
		} else {
			header, hlen := parseHeaderFromAddr(raddr)
			n, _ := write.WriteTo(append(header[:hlen], buf[:n]...), writeAddr)
			addTraffic(n)
		}
	}
}

func handleUDPConnection(handle *SecurePacketConn, n int, src net.Addr, receive []byte, addTraffic func(int)) {
	var dstIP net.IP
	var reqLen int
	addrType := receive[idType]
	defer common.LB.Put(receive)

	switch addrType & common.AddrMask {
	case typeIPv4:
		reqLen = lenIPv4
		if len(receive) < reqLen {
			log.Logger.Debug("[udp]invalid received message.")
		}
		dstIP = net.IP(receive[idIP0 : idIP0+net.IPv4len])
	case typeIPv6:
		reqLen = lenIPv6
		if len(receive) < reqLen {
			log.Logger.Debug("[udp]invalid received message.")
		}
		dstIP = net.IP(receive[idIP0 : idIP0+net.IPv6len])
	case typeDm:
		reqLen = int(receive[idDmLen]) + lenDmBase
		if len(receive) < reqLen {
			log.Logger.Debug("[udp]invalid received message.")
		}
		name := string(receive[idDm0 : idDm0+int(receive[idDmLen])])
		// avoid panic: syscall: string with NUL passed to StringToUTF16 on windows.
		if strings.ContainsRune(name, 0x00) {
			fmt.Println("[udp]invalid domain name.")
			return
		}
		dIP, err := net.ResolveIPAddr("ip", name) // carefully with const type
		if err != nil {
			log.Logger.Debug("[udp]failed to resolve domain name: %s\n", string(receive[idDm0:idDm0+receive[idDmLen]]))
			return
		}
		dstIP = dIP.IP
	default:
		log.Logger.Debug("[udp]addrType %d not supported", addrType)
		return
	}
	dst := &net.UDPAddr{
		IP:   dstIP,
		Port: int(binary.BigEndian.Uint16(receive[reqLen-2 : reqLen])),
	}
	if _, ok := reqList.Get(dst.String()); !ok {
		req := make([]byte, reqLen)
		copy(req, receive)
		reqList.Put(dst.String(), req)
	}

	remote, exist, err := natlist.Get(src.String())
	if err != nil {
		return
	}
	if !exist {
		log.Logger.Debug("[udp]new client %s->%s via %s\n", src, dst, remote.LocalAddr())
		go func() {
			Pipeloop(handle, src, remote, addTraffic)
			natlist.Delete(src.String())
		}()
	} else {
		log.Logger.Info("[udp]using cached client %s->%s via %s\n", src, dst, remote.LocalAddr())
	}
	if remote == nil {
		fmt.Println("WTF")
	}
	remote.SetDeadline(time.Now().Add(udpTimeout))
	n, err = remote.WriteTo(receive[reqLen:n], dst)
	addTraffic(n)
	if err != nil {
		if ne, ok := err.(*net.OpError); ok && (ne.Err == syscall.EMFILE || ne.Err == syscall.ENFILE) {
			// log too many open file error
			// EMFILE is process reaches open file limits, ENFILE is system limit
			log.Logger.Debug("[udp]write error:", err)
		} else {
			log.Logger.Debug("[udp]error connecting to:", dst, err)
		}
		if conn := natlist.Delete(src.String()); conn != nil {
			conn.Close()
		}
	}
	// Pipeloop
	return
}

func ReadAndHandleUDPReq(c *SecurePacketConn, addTraffic func(int)) error {
	buf := common.LB.Get()
	n, src, err := c.ReadFrom(buf[0:])
	if err != nil {
		return err
	}
	go handleUDPConnection(c, n, src, buf, addTraffic)
	return nil
}
