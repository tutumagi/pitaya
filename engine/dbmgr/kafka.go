package dbmgr

import (
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	KafkaProducer = 0 //生产者
	KafkaConsumer = 1 //消费者
)

//kafka实例
type KafkaHandler struct {
	ID            uint                 //身份，0生产者 1消费者
	HostCluster   []string             //集群连接信息
	Topic         []string             //topic
	Sync          bool                 //是否同步
	GroupID       string               //groud id,消费者消费使用
	SyncProducer  sarama.SyncProducer  //同步生产者
	AsyncProducer sarama.AsyncProducer //异步生产者
	Consumer      *cluster.Consumer    //消费者
}

//初始化
func (this *KafkaHandler) Init(cfg *viper.Viper) error {
	if cfg != nil {
		this.ID = cfg.GetUint("dbmgr.kafka.type")
		this.HostCluster = cfg.GetStringSlice("dbmgr.kafka.hostCluster")
		this.Sync = cfg.GetBool("dbmgr.kafka.sync")
		this.Topic = cfg.GetStringSlice("dbmgr.kafka.topic")
		this.GroupID = cfg.GetString("dbmgr.kafka.groud_id")
		return this.Reconn()
	}
	return fmt.Errorf("config this is empty.......")
}

func (this *KafkaHandler) Close() error {
	var err error = nil
	if this.SyncProducer != nil {
		err = this.SyncProducer.Close()
	}
	if this.AsyncProducer != nil {
		err = this.AsyncProducer.Close()
	}
	return err
}

//重连
func (this *KafkaHandler) Reconn() error {
	if this.ID == KafkaProducer { //生产者
		return this.ProducerReconn()
	} else { //消费者
		return this.ConsumerReconn()
	}
}

//生产者重连
func (this *KafkaHandler) ProducerReconn() error {
	needConn := false
	var err error = nil
	if this.Sync && this.SyncProducer == nil {
		needConn = true
	} else if this.AsyncProducer == nil {
		needConn = true
	}
	defer func() { //捕获异常
		if e := recover(); e != nil {
			this.Close()
			this.Reconn()
		}
	}()
	if needConn {
		config := sarama.NewConfig()
		config.Producer.Timeout = 5 * time.Second
		config.Producer.Partitioner = sarama.NewRandomPartitioner //随机的分区类型
		loop := 0
		for loop < ReconnTimes {
			if this.Sync { //同步
				config.Producer.RequiredAcks = sarama.WaitForAll //等待服务器所有副本都保存成功后的响应,确保消息的保存
				//是否等待成功和失败后的响应,只有上面的RequireAcks设置不是NoReponse这里才有用.
				config.Producer.Return.Successes = true
				config.Producer.Return.Errors = true
				this.SyncProducer, err = sarama.NewSyncProducer(this.HostCluster, config)
				if err == nil {
					break
				}
			} else { //异步
				this.AsyncProducer, err = sarama.NewAsyncProducer(this.HostCluster, config)
				if err == nil {
					break
				}
			}
			loop++
		}
	}
	return err
}

//消费者重连
func (this *KafkaHandler) ConsumerReconn() error {
	var err error = nil
	if this.Consumer == nil {
		loop := 0
		config := cluster.NewConfig()
		config.Group.Return.Notifications = false
		config.Consumer.Offsets.CommitInterval = 1 * time.Second
		for loop < ReconnTimes {
			//不能使用sarama.NewConsumer,进程重启后只能从offset为0的或重启后再生产的消息，
			//不会从上一次重启的offset读取消息，sarama.cluster则会从上一次的offset再次获取
			this.Consumer, err = cluster.NewConsumer(this.HostCluster, this.GroupID, this.Topic, config)
			if err == nil {
				break
			}
			loop++
		}
	}
	return err
}

//生产者生产消息
func (this *KafkaHandler) ProductionMsg(msg SyncMsg) error {
	var err error = nil
	if err = this.Reconn(); err == nil {
		if this.ID == KafkaProducer {
			defer func() { //捕获异步
				if e := recover(); e != nil {
					this.Close()
					this.Reconn()
				}
			}()

			kMsg := &sarama.ProducerMessage{}
			if msg.Topic != "" {
				kMsg.Topic = msg.Topic
			} else if len(this.Topic) > 0 { //没有topic则默认配置
				kMsg.Topic = this.Topic[0]
			}
			jByte := []byte{}
			jByte, err = bson.Marshal(msg)
			if err == nil {
				kMsg.Value = sarama.ByteEncoder(jByte) //将interface转成json string
				var partition int32
				var offset int64
				loop := 0
				for loop < ReconnTimes {
					if this.Sync { //同步
						partition, offset, err = this.SyncProducer.SendMessage(kMsg)
						fmt.Printf("KafkaHandler::ProductionMsg result[partition:%d, offset:%d, err:%+v\n]", partition, offset, err)
					} else { //异步，交给sarama的异步,不需要等待结果
						this.AsyncProducer.Input() <- kMsg
					}
					if err == nil {
						break
					}
					loop++
				}
			}

		} else {
			return fmt.Errorf("KafkaHandler::ProductionMsg error, not the producer type.....")
		}
	}
	return err
}

//消费者消费消息
func (this *KafkaHandler) ConsumeStart(consumeFunc func(kMsg *SyncMsg) error) error {
	var err error = nil
	//必须设置是消费者身份
	if this.ID == KafkaConsumer {
		if this.Reconn() == nil {
			//捕获异常
			defer func() {
				if e := recover(); e != nil {
					this.Close()
					this.Reconn()
				}
			}()
			for msg := range this.Consumer.Messages() {
				val := &SyncMsg{}
				if err = bson.Unmarshal(msg.Value, val); err == nil {
					fmt.Printf("KafkaHandler::ConsumeMsg json convert result:%+v\n", val)
					consumeFunc(val)
				}
				fmt.Printf("KafkaHandler::ConsumeMsg result:%+v\n", msg)
				this.Consumer.MarkOffset(msg, "") //并不是实时写入kafka，有可能在程序crash时丢掉未提交的offset，帮需要实时改变offset
			}
		}
	} else {
		return fmt.Errorf("KafkaHandler::ConsumeMsg error, not the consumer type.....")
	}
	return err
}
