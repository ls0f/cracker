package cracker

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProxyConn(t *testing.T) {
	startProxyServer()

	h := &Handler{Server: "http://localhost" + testAddr, Secret: testSecret}
	conn, err := h.Connect("localhost" + testAddr)
	assert.NoError(t, err)
	conn.Write([]byte("GET /ping HTTP/1.1\r\nHost: localhost\r\n\r\n"))
	go func() {
		time.Sleep(time.Millisecond * 100)
		conn.Close()
	}()
	body, _ := ioutil.ReadAll(conn)
	assert.Contains(t, string(body), "pong")
}

func TestProxyConn2(t *testing.T) {
	startProxyServer()

	h := &Handler{Server: "http://localhost" + testAddr, Secret: testSecret, Interval: time.Millisecond * 20}
	conn, err := h.Connect("localhost" + testAddr)
	assert.NoError(t, err)
	conn.Write([]byte("GET /connect HTTP/1.1\r\nHost: localhost\r\n\r\n"))
	go func() {
		time.Sleep(time.Millisecond * 100)
		conn.Close()
	}()
	body, _ := ioutil.ReadAll(conn)
	assert.Contains(t, string(body), "404")
}

func TestProxyConn3(t *testing.T) {
	startProxyServer()

	h := &Handler{Server: "http://localhost" + testAddr, Secret: testSecret}
	conn, err := h.Connect("localhost" + testAddr)
	assert.NoError(t, err)
	p, ok := conn.(*localProxyConn)
	assert.True(t, ok)
	// wrong uuid
	p.uuid = ""
	_, err = conn.Write([]byte("GET /connect HTTP/1.1\r\nHost: localhost\r\n\r\n"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "uuid don't exist")
	conn.Close()
}

func TestProxyConn4(t *testing.T) {
	startProxyServer()

	h := &Handler{Server: "http://localhost" + testAddr, Secret: testSecret, Interval: time.Millisecond * 20}
	conn, err := h.Connect("localhost" + testAddr)
	assert.NoError(t, err)
	p, ok := conn.(*localProxyConn)
	assert.True(t, ok)
	// wrong uuid
	p.uuid = ""
	body, err := ioutil.ReadAll(conn)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "uuid don't exist")
	conn.Close()
	assert.Empty(t, body)
}

func TestProxyConn5(t *testing.T) {
	startProxyServer()

	h := &Handler{Server: "http://localhost" + testAddr, Secret: testSecret, Interval: time.Millisecond * 20}
	conn, err := h.Connect("localhost" + testAddr)
	assert.NoError(t, err)
	p, ok := conn.(*localProxyConn)
	assert.True(t, ok)
	remote, ok := testP.proxyMap[p.uuid]
	assert.True(t, ok)
	assert.False(t, remote.IsClosed())
	conn.Close()
}

func TestProxyConn6(t *testing.T) {
	startProxyServer()

	h := &Handler{Server: "http://localhost" + testAddr, Secret: testSecret, Interval: time.Millisecond * 20}
	conn, err := h.Connect("localhost" + testAddr)
	assert.NoError(t, err)
	p, ok := conn.(*localProxyConn)
	assert.True(t, ok)
	remote, ok := testP.proxyMap[p.uuid]
	assert.True(t, ok)
	conn.Close()
	assert.True(t, remote.IsClosed())
}
