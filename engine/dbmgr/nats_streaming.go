package dbmgr

import (
	"fmt"
	"runtime"
	"time"

	"github.com/tutumagi/pitaya/logger"

	"github.com/nats-io/stan.go"
	"github.com/spf13/viper"
	bson "go.mongodb.org/mongo-driver/bson"
)

//管理nats_streaming的发布与订阅
type NatsStreamHandler struct {
	Url       string                   //nats服务器连接数据
	ClusterID string                   //集群ID
	ClientID  string                   //客户端ID
	Durable   string                   //持久性数据标识名字
	Subject   string                   //用于订阅发布的topic标识
	Conn      stan.Conn                //连接
	MsgFunc   func(msg *SyncMsg) error //消息回调函数
}

//初始化
func (this *NatsStreamHandler) Init(cfg *viper.Viper) error {
	if cfg != nil {
		this.Url = cfg.GetString("dbmgr.nats_streaming.url")
		this.ClusterID = cfg.GetString("dbmgr.nats_streaming.cluster_id")
		this.ClientID = cfg.GetString("dbmgr.nats_streaming.client_id")
		this.Durable = cfg.GetString("dbmgr.nats_streaming.durable")
		this.Subject = cfg.GetString("dbmgr.nats_streaming.subject")
		return this.Reconn()
	}
	return fmt.Errorf("NatsStreamHandler::Init error, the configure is empty pointer")
}

//重连
func (this *NatsStreamHandler) Reconn() (err error) {
	if this.Conn == nil {
		loop := 0
		//连接失去后的处理函数
		connLostFunc := func(_ stan.Conn, reason error) {
			logger.Errorf("Connection lost, reason: %+v\n", reason)
			this.Close()
			if err = this.Reconn(); err == nil {
				//如果订阅函数没有不为空，则重新启动订阅函数去订阅消息
				if this.MsgFunc != nil {
					this.Subscribe(this.MsgFunc)
				}
			}
		}
		for loop < ReconnTimes {
			this.Conn, err = stan.Connect(this.ClusterID, this.ClientID, stan.NatsURL(this.Url), stan.SetConnectionLostHandler(connLostFunc))
			if err == nil {
				logger.Debugf("NatsStreamHandler::Reconn successfully:%s", this.Url)
				break
			}
			this.Conn = nil
			loop++
		}
		//程序一启动的时候就没办法连接成功的话，如果是订阅者，则隔一定的时间去重连,直到连接成功为止
		if err != nil && this.MsgFunc != nil {
			time.Sleep(3 * time.Second)
			logger.Debugf("subcribe reconnection begining...........")
			return this.Reconn()
		}
	}
	return
}

//关闭
func (this *NatsStreamHandler) Close() (err error) {
	if this.Conn != nil {
		err = this.Conn.Close()
		if err == nil {
			this.Conn = nil
		}
	}
	return
}

//发布消息
func (this *NatsStreamHandler) PublishMsg(msg SyncMsg) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("NatsStreamHandler::PublishMsg panic:%+v", e)
			this.Close()
			this.Reconn()
		}
	}()

	if err = this.Reconn(); err == nil {
		var jByte []byte
		if jByte, err = bson.Marshal(msg); err == nil {
			err = this.Conn.Publish(this.Subject, jByte)
			if err != nil { //失败后关闭，待下次调用再重连
				this.Close()
			}
		}
	}
	return
}

//订阅消息
func (this *NatsStreamHandler) Subscribe(msgFunc func(msg *SyncMsg) error) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("NatsStreamHandler::PublishMsg panic:%+v", e)
			this.Close()
			this.Reconn()
		}
	}()
	if this.MsgFunc == nil {
		this.MsgFunc = msgFunc
	}

	if err = this.Reconn(); err == nil {
		subfunc := func(msg *stan.Msg) { //回调函数
			sMsg := &SyncMsg{}
			if err := bson.Unmarshal(msg.Data, sMsg); err == nil {
				this.MsgFunc(sMsg)
			}
		}
		_, err = this.Conn.Subscribe(this.Subject, subfunc, stan.DurableName(this.Durable))
		if err != nil {
			panic(err)
		}
		runtime.Goexit()
	}
	return
}
