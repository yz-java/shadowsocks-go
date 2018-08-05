package udp

import (
	"fmt"
	"net"
	"time"
	"shadowsocks-go/shadowsocks/encrypt"
)

const (
	maxPacketSize = 4096 // increase it if error occurs
)

var (
	errPacketTooSmall  = fmt.Errorf("[udp]read error: cannot decrypt, received packet is smaller than ivLen")
	errPacketTooLarge  = fmt.Errorf("[udp]read error: received packet is latger than maxPacketSize(%d)", maxPacketSize)
	errBufferTooSmall  = fmt.Errorf("[udp]read error: given buffer is too small to hold data")
)

type SecurePacketConn struct {
	net.PacketConn
	*encrypt.Cipher
}

func NewSecurePacketConn(c net.PacketConn, cipher *encrypt.Cipher) *SecurePacketConn {
	return &SecurePacketConn{
		PacketConn: c,
		Cipher:     cipher,
	}
}

func (c *SecurePacketConn) Close() error {
	return c.PacketConn.Close()
}

func (c *SecurePacketConn) ReadFrom(b []byte) (n int, src net.Addr, err error) {
	cipher := c.Copy()
	buf := make([]byte, 4096)
	n, src, err = c.PacketConn.ReadFrom(buf)
	if err != nil {
		return
	}

	if n < c.Info.IvLen {
		return 0, nil, errPacketTooSmall
	}

	if len(b) < n-c.Info.IvLen {
		err = errBufferTooSmall // just a warning
	}

	iv := make([]byte, c.Info.IvLen)
	copy(iv, buf[:c.Info.IvLen])

	if err = cipher.InitDecrypt(iv); err != nil {
		return
	}

	cipher.Decrypt(b[0:], buf[c.Info.IvLen:n])
	n -= c.Info.IvLen

	return
}

func (c *SecurePacketConn) WriteTo(b []byte, dst net.Addr) (n int, err error) {
	cipher := c.Copy()
	iv, err := cipher.InitEncrypt()
	if err != nil {
		return
	}
	packetLen := len(b) + len(iv)

	cipherData := make([]byte, packetLen)
	copy(cipherData, iv)

	cipher.Encrypt(cipherData[len(iv):], b)
	n, err = c.PacketConn.WriteTo(cipherData, dst)
	return
}

func (c *SecurePacketConn) LocalAddr() net.Addr {
	return c.PacketConn.LocalAddr()
}

func (c *SecurePacketConn) SetDeadline(t time.Time) error {
	return c.PacketConn.SetDeadline(t)
}

func (c *SecurePacketConn) SetReadDeadline(t time.Time) error {
	return c.PacketConn.SetReadDeadline(t)
}

func (c *SecurePacketConn) SetWriteDeadline(t time.Time) error {
	return c.PacketConn.SetWriteDeadline(t)
}
