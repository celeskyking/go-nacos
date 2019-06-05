package v1

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/celeskyking/go-nacos/client/loadbalancer"
	"github.com/celeskyking/go-nacos/pkg/util"
	"github.com/celeskyking/go-nacos/types"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"math/rand"
	"net"
	"strconv"
	"time"
)

var (
	UDPPort = 0
)

var (
	GzipMagicCode = []byte("\x1F\x8B")
)

func NewPushReceiver(lb loadbalancer.LB) *PushReceiver {
	return &PushReceiver{
		QuitC:   make(chan struct{}, 0),
		NotifyC: make(chan *PushMessage, 100),
		LB:      lb,
	}
}

type PushReceiver struct {
	Port int

	QuitC chan struct{}

	NotifyC chan *PushMessage

	LB loadbalancer.LB
}

type PushData struct {
	Type string `json:"type"`

	Data string `json:"data"`

	LastRefTime int64 `json:"lastRefTime"`
}

func (u *PushReceiver) Start() {
	var conn *net.UDPConn
	retries := 0
	max := 60
	for {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		port := r.Intn(1000) + 45000
		u.Port = port
		c, ok := u.Listen()
		if !ok {
			retries = retries + 1
			time.Sleep(time.Duration(util.Min(retries, max)) * time.Second)
			continue
		} else {
			conn = c
			UDPPort = port
			logrus.Info("connect to nacos success")
			break
		}
	}
	defer func() {
		_ = conn.Close()
	}()
	for {
		select {
		case <-u.QuitC:
			return
		default:
			u.consume(conn)
		}
	}
}

func (u *PushReceiver) GetNotifyChannel() <-chan *PushMessage {
	return u.NotifyC
}

func (u *PushReceiver) Listen() (*net.UDPConn, bool) {
	s := u.selectServer(u.LB)
	if s == nil {
		logrus.Errorf("no nacos server worked")
		return nil, false
	}
	addr, err := net.ResolveUDPAddr("udp", s.GetHost()+":"+strconv.Itoa(u.Port))
	if err != nil {
		logrus.Errorf("listen to nacos push service failed:%+v", err)
		return nil, false
	}
	connection, err := net.ListenUDP("udp", addr)
	if err != nil {
		logrus.Error("listen error", err)
		return nil, false
	}
	return connection, true
}

func (u *PushReceiver) consume(conn *net.UDPConn) {
	data := make([]byte, 4096)
	n, remoteAddr, err := conn.ReadFromUDP(data)
	if err != nil {
		logrus.Error("failed to read UDP msg because of ", err)
		return
	}
	s := DecompressData(data[:n])
	logrus.Info("receive push: "+s+" from: ", remoteAddr)
	var pushData PushData
	er := json.Unmarshal([]byte(s), &pushData)
	if er != nil {
		logrus.Error("failed to process push data", er)
		return
	}
	var pushMessage PushMessage
	er = json.Unmarshal([]byte(pushData.Data), &pushMessage)
	if er != nil {
		logrus.Errorf("failed to unmarshal json string:%+v", er)
	}
	if len(pushMessage.Hosts) == 0 {
		logrus.Errorf("get empty ip list, ignore it, dom:%s", pushMessage.Dom)
		return
	}
	u.NotifyC <- &pushMessage
	logrus.Infof("receive message:%+v", pushMessage)

}

func (u *PushReceiver) selectServer(lb loadbalancer.LB) *loadbalancer.Server {
	l := len(lb.GetServers())
	if l == 0 {
		return nil
	}
	return lb.GetServers()[rand.Intn(l-1)]
}

func GetUDPPort() int {
	return UDPPort
}

func IsGzipFile(data []byte) bool {
	if len(data) < 2 {
		return false
	}

	return bytes.HasPrefix(data, GzipMagicCode)
}

func DecompressData(data []byte) string {
	if IsGzipFile(data) {
		return string(data)
	}
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		logrus.Error("failed to decompress gzip data", err)
		return ""
	}

	defer func() {
		_ = reader.Close()
	}()
	bs, er := ioutil.ReadAll(reader)
	if er != nil {
		logrus.Warn("failed to decompress gzip data", er)
		return ""
	}
	return string(bs)
}

type Ack struct {
	Type string `json:"type"`

	LastRefTime string `json:"lastRefTime"`

	Data string
}

type PushMessage struct {
	Name string `json:"name"`
	//serviceName
	Dom string `json:"dom"`

	Clusters string `json:"clusters"`

	CacheMillis int64 `json:"cacheMillis"`

	LastRefTime int64 `json:"lastRefTime"`

	Checksum string `json:"checksum"`

	Hosts []*types.ServiceInstance `json:"hosts"`
}
