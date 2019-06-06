package v1

import (
	"encoding/json"
	"github.com/celeskyking/go-nacos/pkg/util"
	"github.com/celeskyking/go-nacos/types"
	"github.com/sirupsen/logrus"
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

func NewPushReceiver() *PushReceiver {
	return &PushReceiver{
		QuitC:   make(chan struct{}, 0),
		NotifyC: make(chan *PushMessage, 100),
	}
}

type PushReceiver struct {
	Port int

	QuitC chan struct{}

	NotifyC chan *PushMessage
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
	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:"+strconv.Itoa(u.Port))
	if err != nil {
		logrus.Errorf("listen to nacos push service failed:%+v", err)
		return nil, false
	}
	logrus.Infof("udp server started, addr:%s", addr.String())
	connection, err := net.ListenUDP("udp", addr)
	if err != nil {
		logrus.Error("listen error", err)
		return nil, false
	}
	return connection, true
}

func (u *PushReceiver) consume(conn *net.UDPConn) {
	data := make([]byte, 64*1024)
	n, remoteAddr, err := conn.ReadFromUDP(data)
	if err != nil {
		logrus.Error("failed to read UDP msg because of ", err)
		return
	}
	logrus.Infof("receive push message from %s\n", remoteAddr.String())
	var pushData PushData
	j := data[:n]
	logrus.Infof("push message:%s", string(j))
	er := json.Unmarshal([]byte(j), &pushData)
	if er != nil {
		logrus.Error("failed to process push data", er)
		return
	}

	ack := make(map[string]string, 4)
	if pushData.Type == "dom" || pushData.Type == "service" {
		var pushMessage PushMessage
		er = json.Unmarshal([]byte(pushData.Data), &pushMessage)
		if er != nil {
			logrus.Errorf("failed to unmarshal json string:%+v", er)
			return
		}
		if len(pushMessage.Hosts) == 0 {
			logrus.Errorf("get empty ip list, ignore it, dom:%s", pushMessage.Dom)
			return
		}
		u.NotifyC <- &pushMessage
		ack["type"] = "push-ack"
		ack["data"] = ""
		//todo
	} else if pushData.Type == "dump" {
		ack["type"] = "dump-ack"
		ack["data"] = ""
	} else {
		ack["type"] = "unknown-ack"
		ack["data"] = ""
	}
	ack["lastRefTime"] = strconv.FormatInt(pushData.LastRefTime, 10)
	ackData, er := json.Marshal(ack)
	if er != nil {
		logrus.Error("push message encode ack failed", er)
		return
	}
	_, er = conn.WriteToUDP(ackData, remoteAddr)
	if er != nil {
		logrus.Error("push message ack failed", er)
	}
}

func GetUDPPort() int {
	return UDPPort
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

	GroupName string `json:"groupName"`

	CacheMillis int64 `json:"cacheMillis"`

	LastRefTime int64 `json:"lastRefTime"`

	Checksum string `json:"checksum"`

	Hosts []*types.Host `json:"hosts"`
}
