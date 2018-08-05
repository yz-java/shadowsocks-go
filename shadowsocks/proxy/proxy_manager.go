/**
 * Created with IntelliJ IDEA.
 * Description: 
 * User: yangzhao
 * Date: 2018-08-05
 * Time: 20:41
 */
package proxy

import (
	"sync"
	"net"
	"shadowsocks-go/shadowsocks/config"


)

type PortListener struct {
	password string
	listener net.Listener
}

type UDPListener struct {
	password string
	listener *net.UDPConn
}

var manager *Manager = nil

type Manager struct {
	sync.Mutex
	portListener map[string]*PortListener
	udpListener  map[string]*UDPListener
	trafficStats map[string]int64
}

func GetManager() *Manager {
	if manager == nil {
		manager = &Manager{
			portListener: map[string]*PortListener{},
			udpListener:  map[string]*UDPListener{},
			trafficStats: map[string]int64{},
		}
	}
	return manager
}

func (pm *Manager) Add(port, password string, listener net.Listener) {
	pm.Lock()
	pm.portListener[port] = &PortListener{password, listener}
	pm.trafficStats[port] = 0
	pm.Unlock()
}

func (pm *Manager) AddUDP(port, password string, listener *net.UDPConn) {
	pm.Lock()
	pm.udpListener[port] = &UDPListener{password, listener}
	pm.Unlock()
}

func (pm *Manager) Get(port string) (pl *PortListener, ok bool) {
	pm.Lock()
	pl, ok = pm.portListener[port]
	pm.Unlock()
	return
}

func (pm *Manager) GetUDP(port string) (pl *UDPListener, ok bool) {
	pm.Lock()
	pl, ok = pm.udpListener[port]
	pm.Unlock()
	return
}

func (pm *Manager) Del(port string) {
	pl, ok := pm.Get(port)
	if !ok {
		return
	}
	if config.Udp {
		upl, ok := pm.GetUDP(port)
		if !ok {
			return
		}
		upl.listener.Close()
	}
	pl.listener.Close()
	pm.Lock()
	delete(pm.portListener, port)
	delete(pm.trafficStats, port)
	if config.Udp {
		delete(pm.udpListener, port)
	}
	pm.Unlock()
}

func (pm *Manager) AddTraffic(port string, n int) {
	pm.Lock()
	pm.trafficStats[port] = pm.trafficStats[port] + int64(n)
	pm.Unlock()
	return
}

func (pm *Manager) GetTrafficStats() map[string]int64 {
	pm.Lock()
	copy := make(map[string]int64)
	for k, v := range pm.trafficStats {
		copy[k] = v
	}
	pm.Unlock()
	return copy
}
