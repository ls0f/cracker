package cracker

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	g "github.com/golang/glog"
	"gopkg.in/bufio.v1"
)

const (
	PerHostNum = 10
)

var tr = &http.Transport{

	DisableKeepAlives:   false,
	MaxIdleConnsPerHost: PerHostNum,
	Proxy:               http.ProxyFromEnvironment,
	TLSHandshakeTimeout: time.Second * timeout / 2,
	IdleConnTimeout:     time.Second * 90,
}

func Init(cert string) {
	if f, err := os.Stat(cert); err == nil && !f.IsDir() {
		var CAPOOL *x509.CertPool
		CAPOOL, err := x509.SystemCertPool()
		if err != nil {
			g.Warning(err)
			CAPOOL = x509.NewCertPool()
		}
		serverCert, err := ioutil.ReadFile(cert)
		if err != nil {
			g.Errorf("read cert.pem err:%s ", err)
			return
		}
		CAPOOL.AppendCertsFromPEM(serverCert)
		config := &tls.Config{RootCAs: CAPOOL}
		tr.TLSClientConfig = config
		g.Infof("load %s success ... ", cert)
	} else if err != nil {
		g.Error(err)
	} else {
		g.Errorf("%s is a dir", cert)
	}
}

type localProxyConn struct {
	uuid     string
	server   string
	secret   string
	source   io.ReadCloser
	close    chan bool
	interval time.Duration
}

func (c *localProxyConn) gen_sign(req *http.Request) {

	ts := fmt.Sprintf("%d", time.Now().Unix())
	req.Header.Set("UUID", c.uuid)
	req.Header.Set("timestamp", ts)
	req.Header.Set("sign", GenHMACSHA1(c.secret, ts))
}

func (c *localProxyConn) push(data []byte, typ string) error {
	hc := &http.Client{Transport: tr, Timeout: time.Second * timeout}
	buf := bufio.NewBuffer(data)
	req, _ := http.NewRequest("POST", c.server+PUSH, buf)
	c.gen_sign(req)
	req.Header.Set("TYP", typ)
	req.ContentLength = int64(len(data))
	req.Header.Set("Content-Type", "image/jpeg")
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
}

func (c *localProxyConn) connect(dstHost, dstPort string) (uuid string, err error) {
	hc := &http.Client{Transport: tr, Timeout: time.Second * timeout}
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

func (c *localProxyConn) pull() error {

	var (
		hc *http.Client
	)
	if c.interval > 0 {
		hc = &http.Client{Transport: tr, Timeout: time.Second * timeout}
	} else {
		hc = &http.Client{Transport: tr}
	}

	req, _ := http.NewRequest("GET", c.server+PULL, nil)
	req.Header.Set("Interval", fmt.Sprintf("%d", c.interval))
	c.gen_sign(req)
	res, err := hc.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != HeadOK {
		body, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()
		return errors.New(fmt.Sprintf("status code is %d, body is %s", res.StatusCode, string(body)))
	}
	c.source = res.Body
	return nil
}

func (c *localProxyConn) Read(b []byte) (n int, err error) {

	if c.source == nil {
		if c.interval > 0 {
			if err = c.pull(); err != nil {
				return
			}
		} else {
			return 0, errors.New("pull http connection is not ready")
		}
	}
	n, err = c.source.Read(b)
	if err != nil {
		c.source.Close()
		c.source = nil
	}
	if err == io.EOF && c.interval > 0 {
		err = nil
	}
	return
}

func (c *localProxyConn) Write(b []byte) (int, error) {

	err := c.push(b, DATA_TYP)
	if err != nil {
		g.V(LDEBUG).Infof("push: %v", err)
		return 0, err
	}

	return len(b), nil
}

func (c *localProxyConn) alive() {
	for {
		select {
		case <-c.close:
			return
		case <-time.After(time.Second * heartTTL / 2):
			if err := c.push([]byte("alive"), HEART_TYP); err != nil {
				return
			}
		}
	}
}

func (c *localProxyConn) quit() error {
	return c.push([]byte("quit"), QUIT_TYP)
}

func (c *localProxyConn) Close() error {
	close(c.close)
	return c.quit()
}
