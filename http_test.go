package cracker

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	lock            sync.Mutex
	serverHaveStart bool
	testP           *httpProxy
)

const (
	testAddr   = ":12245"
	testSecret = "12345"
)

func startProxyServer() {
	lock.Lock()
	defer lock.Unlock()
	if serverHaveStart {
		return
	}
	testP = NewHttpProxy(testAddr, testSecret, false)
	go testP.Listen()
	time.Sleep(time.Millisecond * 100)
	serverHaveStart = true
}

func TestHandler_Connect(t *testing.T) {
	startProxyServer()

	res, err := http.Get(fmt.Sprintf("http://127.0.0.1%s%s", testAddr, CONNECT))
	if err != nil {
		assert.NoError(t, err)
	}
	assert.Equal(t, res.StatusCode, 404)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	assert.Equal(t, "404", string(body))
}
