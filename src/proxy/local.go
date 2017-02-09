package proxy

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
	"gopkg.in/bufio.v1"
)

const (
	PerHostNum = 30
)

var tr = &http.Transport{

	DisableKeepAlives: false,

	MaxIdleConnsPerHost: PerHostNum,
}

type localProxyConn struct {
	uuid        string
	server      string
	secret      string
	read_buffer []byte
	read_mutex  sync.Mutex
	Close       chan bool
}

func (c *localProxyConn) gen_sign(req *http.Request) {

	ts := fmt.Sprintf("%d", time.Now().Unix())
	req.Header.Set("UUID", c.uuid)
	req.Header.Set("timestamp", ts)
	req.Header.Set("sign", GenHMACSHA1(c.secret, ts))
}

func (c *localProxyConn) push(data []byte, typ string) error {
	hc := &http.Client{Transport: tr, Timeout: time.Duration(time.Second * heartTTL)}
	buf := bufio.NewBufferString(base64.StdEncoding.EncodeToString(data))
	req, _ := http.NewRequest("POST", c.server+PUSH, buf)
	c.gen_sign(req)
	req.Header.Set("TYP", typ)
	res, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	switch res.StatusCode {
	case HeadOK:
		return nil
	default:
		return errors.New(fmt.Sprintf("status code is %d, body is: %s", res.StatusCode, string(body)))
	}
	return nil
}

func (c *localProxyConn) connect(dstHost, dstPort string) (uuid string, err error) {
	hc := &http.Client{Transport: tr, Timeout: time.Duration(time.Second * heartTTL)}
	req, _ := http.NewRequest("GET", c.server+CONNECT, nil)
	c.gen_sign(req)
	req.Header.Set("DSTHOST", dstHost)
	req.Header.Set("DSTPORT", dstPort)
	res, err := hc.Do(req)
	if err != nil {
		return "", err
	}
	body, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != HeadOK {
		return "", errors.New(fmt.Sprintf("status code is %d, body is:%s", res.StatusCode, string(body)))
	}
	return string(body), err

}

func (c *localProxyConn) pull() ([]byte, error) {

	hc := &http.Client{Transport: tr, Timeout: time.Duration(time.Second * heartTTL)}
	for {

		req, _ := http.NewRequest("GET", c.server+PULL, nil)
		c.gen_sign(req)
		res, err := hc.Do(req)
		if err != nil {
			return nil, err
		}
		body, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()

		switch res.StatusCode {

		case HeadData:
			return body, nil
		case HeadHeart:
			continue
		case HeadQuit:
			return nil, errors.New("should quit")
		default:
			return nil, errors.New(fmt.Sprintf("status code is %d, body is: %s", res.StatusCode, string(body)))

		}
	}

}

func (c *localProxyConn) Write(b []byte) (int, error) {

	err := c.push(b, DATA_TYP)
	if err != nil {
		log.Printf("push err %v ... \n", err)
		return 0, err
	}

	return len(b), nil
}

func (c *localProxyConn) fill() error {

	data_bytes, err := c.pull()
	if err != nil {
		return err
	}
	data := string(data_bytes)

	decodeLen := base64.StdEncoding.DecodedLen(len(data))
	bData := make([]byte, len(c.read_buffer)+decodeLen)
	n, err := base64.StdEncoding.Decode(bData[len(c.read_buffer):], data_bytes)
	if err != nil {
		return err
	}
	bData = bData[:len(c.read_buffer)+n]
	c.read_buffer = bData
	return nil
}

func (c *localProxyConn) Read(b []byte) (n int, err error) {
	c.read_mutex.Lock()
	// If local buffer is empty, get new data
	if len(c.read_buffer) == 0 {
		err := c.fill()
		if err != nil {
			log.Printf("fill err:%v \n", err)
			c.read_mutex.Unlock()
			return 0, err
		}
	}
	// Return local buffer
	count := len(b)
	if count > len(c.read_buffer) {
		count = len(c.read_buffer)
	}
	copy(b, c.read_buffer[:count])
	c.read_buffer = c.read_buffer[count:]

	c.read_mutex.Unlock()
	return count, nil
}

func (c *localProxyConn) alive() {
	for {
		select {
		case <-c.Close:
			return
		case <-time.After(time.Duration(time.Second * timeout)):
			if err := c.push([]byte("alive"), HEART_TYP); err != nil {
				return
			}
		}
	}
}

func (c *localProxyConn) Quit() {
	c.push([]byte("quit"), QUIT_TYP)
}

func Connect(server, remote, secret string) (*localProxyConn, error) {
	if strings.HasSuffix(server, "/") {
		server = server[:len(server)-1]
	}
	conn := localProxyConn{server: server, secret: secret}
	host := strings.Split(remote, ":")[0]
	port := strings.Split(remote, ":")[1]

	uuid, err := conn.connect(host, port)
	if err != nil {
		return nil, err
	}
	conn.uuid = uuid
	conn.Close = make(chan bool)
	go conn.alive()
	return &conn, nil
}
