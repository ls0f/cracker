package proxy

import (
	"net/http"
	"testing"
	"time"

	"bytes"

	"github.com/stretchr/testify/assert"
	"gopkg.in/bufio.v1"
	"strings"
)

func TestHttpProxy_serve(t *testing.T) {
	hp := NewHttpProxy("localhost:12345")
	go hp.Listen()
	time.Sleep(1e8)
	client := &http.Client{}

	buf := bytes.Buffer{}
	buf.WriteString("GET / HTTP/1.1\r\n")
	buf.WriteString("Host: baidu.com\r\n\r\n")
	r := bufio.NewBuffer(buf.Bytes())
	req, err := http.NewRequest("POST", "http://localhost:12345/proxy", r)
	assert.NoError(t, err)
	req.ContentLength = -1
	req.Close = false
	req.TransferEncoding = []string{"chunked"}
	req.Header.Set("Content-Type", "image/jpeg")
	req.Header.Set("Transfer-Encoding", "chunked")
	req.Header.Set("proxy_addr", "baidu.com:80")
	res, err := client.Do(req)
	assert.NoError(t, err)
	defer res.Body.Close()
	body := make([]byte, 1024)
	n, err := res.Body.Read(body)
	assert.NoError(t, err)
	println("res:", string(body[:n]))
	assert.True(t, true, strings.Contains(string(body[:n]), "HTTP/1.1 200 OK"))

}
