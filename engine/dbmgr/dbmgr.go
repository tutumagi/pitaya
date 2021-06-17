package dbmgr

import (
	"errors"
	"fmt"

	"github.com/tutumagi/pitaya/logger"
	"gitlab.gamesword.com/nut/dreamcity/game/define"

	"github.com/spf13/viper"
)

//此模块连接redis及mongodb或mysql,业务本地直接操作redis,
//异步通过redis list类型指定的字段异步至mysql或mongodb进行二次数据保存
//异步的只有插入、更新、删除，查寻获取直接本地同步操作
var ReconnTimes = 3                                        //重连次数
var ErrDataEmpty = errors.New("database data is empty...") //数据为空时会返回这个错误

const (
	DB_OPT_TYPE = iota
	GET         //获取
	SET         //设置
	DEL         //删除
	INS         //插入
)

//同步的消息格式
type SyncMsg struct {
	Topic   string      //topic
	Query   QueryPara   //查找的参数
	OptType uint        //操作类型
	Val     interface{} //保存的数据
}

//查找的条件
type Selector struct {
	Val  interface{} //数据
	Flag int         //标志，等于，不等于，小于，不小于，大于，不大于，在其中，不在其中
}

type QueryPara struct {
	DbName    string      //数据库名
	TblName   string      //表格
	FieldName string      //field名字，第一主键 在调用get时为必填(如果表格有设置主键)，因为当redis没有数据时，从mongodb同步过来时需要在redis指定field
	Field     interface{} //field数据
	KeyName   string      //key名字 第二主键 在调用get时为必填(如果表格有设置主键)，因为当redis没有数据时，从mongodb同步过来时需要在redis指定key
	//key值 NOTE: 这里的 Key 的值，还影响了 redis 中的 key（如果开启了 redis 的话）
	Key      interface{}
	Selector map[string]*Selector //条件查找
	Sort     map[string]int32     //排序数据，key:字段名， value:-1倒序，1升序
	Limit    int32                //限制返回的数量
	Skip     int32                //设置skip
}

//redis key字段规则：dbname.tblName.field, key为数据查找的主键

func GetKeyName(para QueryPara, dbName string) string {
	if para.Field != nil {
		return fmt.Sprintf("%s.%s.%+v", dbName, para.TblName, para.Field)
	}
	return fmt.Sprintf("%s.%s", dbName, para.TblName)
}

//返回新的queryPara
func NewQueryPara() *QueryPara {
	return &QueryPara{
		Selector: map[string]*Selector{},
		Sort:     map[string]int32{},
	}
}

//dbmgr结构体
type DbMgrHandler struct {
	cfg           *viper.Viper
	UseRedis      bool //是否使用redis
	UseKafka      bool //是否使用kafka
	UseMongo      bool //是否使用mongo
	UseNatsStream bool //是否使用nats_streaming

	Redis      *RedisHandler      //redis实例
	Kafka      *KafkaHandler      //kafka实例
	Mongo      *MongoHandler      //mongo实例
	NatsStream *NatsStreamHandler //nats_streaming实例
}

//初始化
func (this *DbMgrHandler) Init(cfg *viper.Viper) error {
	if cfg != nil {
		this.cfg = cfg
		this.UseRedis = cfg.GetBool("dbmgr.use_redis")
		this.UseKafka = cfg.GetBool("dbmgr.use_kafka")
		this.UseMongo = cfg.GetBool("dbmgr.use_mongo")
		this.UseNatsStream = cfg.GetBool("dbmgr.use_nats_streaming")
		return this.Reconn()
	}
	return fmt.Errorf("DbMgrHandler::Init error, the configure pointer is empty....")
}

//重连
func (this *DbMgrHandler) Reconn() (err error) {
	if this.UseRedis { //kafka初始化
		this.Redis = &RedisHandler{}
		if err = this.Redis.Init(this.cfg); err != nil {
			return
		}
	}
	if this.UseKafka { //kafka初始化
		this.Kafka = &KafkaHandler{}
		if err = this.Kafka.Init(this.cfg); err != nil {
			return
		}
	}
	if this.UseMongo { //mongodb初始化
		this.Mongo = &MongoHandler{}
		if err = this.Mongo.Init(this.cfg); err != nil {
			return
		}
	}
	if this.UseNatsStream {
		this.NatsStream = &NatsStreamHandler{}
		if err = this.NatsStream.Init(this.cfg); err != nil {
			return
		}
	}
	return
}

//以下接口供外部调用=======================================================================
//全局变量
var gDbMgr = &DbMgrHandler{}

//初始化
func Init(cfg *viper.Viper) error {
	return gDbMgr.Init(cfg)
}

//从redis中获取数据
func GetFromRedis(para QueryPara, resVal interface{}) (err error) {
	if gDbMgr.UseRedis && gDbMgr.Redis != nil {
		err = gDbMgr.Redis.Get(para, resVal)
		return
	}
	err = fmt.Errorf("dbmgr::GetFromRedis error, not set use redis......")
	return
}

//把数据设置至redis中
func SetFromRedis(para QueryPara, resVal interface{}) (err error) {
	if gDbMgr.UseRedis && gDbMgr.Redis != nil {
		err = gDbMgr.Redis.Set(para, resVal)
		if err != nil { //记录错误日志
			logger.DayLogRecord(define.DB_ERROR_LOG, `SetFromRedis:{para:%+v, val:%+v, error:%+v}`, para, resVal, err)
		} else {
			logger.DayLogRecord(define.DB_SUCC_LOG, `SetFromRedis:{para:%+v, val:%+v}`, para, resVal)
		}
		return
	}
	err = fmt.Errorf("dbmgr::SetFromRedis error, not set use redis......")
	return
}

//删除redis中的数据
//把数据设置至redis中
func DelFromRedis(para QueryPara) (err error) {
	if gDbMgr.UseRedis && gDbMgr.Redis != nil {
		err = gDbMgr.Redis.Del(para)
		if err != nil { //记录错误日志
			logger.DayLogRecord(define.DB_ERROR_LOG, `DelFromRedis:{para:%+v, error:%+v}`, para, err)
		} else {
			logger.DayLogRecord(define.DB_SUCC_LOG, `DelFromRedis:{para:%+v}`, para)
		}
		return
	}
	err = fmt.Errorf("dbmgr::DelFromRedis error, not set use redis......")
	return
}

//把从mongodb中读取的数据同步至redis中
func SyncDataToRedis(para QueryPara, val interface{}) (err error) {
	if gDbMgr.UseRedis && gDbMgr.Redis != nil {
		err = gDbMgr.Redis.SyncData(para, val)
		if err != nil { //记录错误日志
			logger.DayLogRecord(define.DB_ERROR_LOG, `SyncDataToRedis:{para:%+v, val:%+v, error:%+v}`, para, val, err)
		} else {
			logger.DayLogRecord(define.DB_SUCC_LOG, `SyncDataToRedis:{para:%+v, val:%+v}`, para, val)
		}
	}
	err = fmt.Errorf("dbmgr::SyncDataToRedis error, not set use redis......")
	return
}

//从mongodb中获取数据
func GetFromMongo(para QueryPara, resVal interface{}) (err error) {
	if gDbMgr.UseMongo && gDbMgr.Mongo != nil {
		err = gDbMgr.Mongo.Get(para, resVal)
		return
	}
	err = fmt.Errorf("dbmgr::GetFromMongo error, not set use mongodb......")
	return
}

//把数据设置至mongodb
func SetFromMongo(para QueryPara, resVal interface{}) (err error) {
	if gDbMgr.UseMongo && gDbMgr.Mongo != nil {
		err = gDbMgr.Mongo.Set(para, resVal)
		if err != nil { //记录错误日志
			logger.DayLogRecord(define.DB_ERROR_LOG, `SetFromMongo:{para:%+v, val:%+v, error:%+v}`, para, resVal, err)
		} else {
			logger.DayLogRecord(define.DB_SUCC_LOG, `SetFromMongo:{para:%+v, val:%+v}`, para, resVal)
		}
		return
	}
	err = fmt.Errorf("dbmgr::SetFromMongo error, not set use mongodb......")
	return
}

//删除mongodb中的数据
func DelFromMongo(para QueryPara) (err error) {
	if gDbMgr.UseMongo && gDbMgr.Mongo != nil {
		err = gDbMgr.Mongo.Del(para)
		if err != nil { //记录错误日志
			logger.DayLogRecord(define.DB_ERROR_LOG, `DelFromMongo:{para:%+v, error:%+v}`, para, err)
		} else {
			logger.DayLogRecord(define.DB_SUCC_LOG, `DelFromMongo:{para:%+v}`, para)
		}
		return
	}
	err = fmt.Errorf("dbmgr::DelFromMongo error, not set use mongodb......")
	return
}

//异步消息的回调函数
func AsyncMsgManage(kMsg *SyncMsg) (err error) {
	logger.Debugf("AsyncMsgManage recvier msg:%+v", kMsg)
	switch kMsg.OptType {
	case SET: //设置
		err = SetFromMongo(kMsg.Query, kMsg.Val)
	case DEL: //删除
		err = DelFromMongo(kMsg.Query)
	}
	return
}

//生产消息至kafka
func ProductionMsg(opt uint, para QueryPara, val interface{}) (err error) {
	if gDbMgr.UseKafka && gDbMgr.Kafka != nil {
		kMsg := SyncMsg{
			OptType: opt,
			Val:     val,
			Query:   para,
		}
		err = gDbMgr.Kafka.ProductionMsg(kMsg)
		return
	}
	err = fmt.Errorf("dbmgr::ProductionMsg error, not set use kafka......")
	return
}

//从kafka中消费消息
func ConsumeMsg() (err error) {
	if gDbMgr.UseKafka && gDbMgr.Kafka != nil {
		//消费消息的回调函数,消费的消息只同步至mongo,不支持异步get操作
		err = gDbMgr.Kafka.ConsumeStart(AsyncMsgManage)
		return
	}
	err = fmt.Errorf("dbmgr::ConsumeMsg error, not set use kafka......")
	return
}

//发布消息 nats_streaming_server
func PublishMsg(opt uint, para QueryPara, val interface{}) (err error) {
	if gDbMgr.UseNatsStream && gDbMgr.NatsStream != nil {
		msg := SyncMsg{
			OptType: opt,
			Val:     val,
			Query:   para,
		}
		err = gDbMgr.NatsStream.PublishMsg(msg)
		if err != nil { //记录错误日志
			logger.DayLogRecord(define.DB_ASYNC_ERROR_LOG, "PublishMsg:%+v", msg)
		}
		return
	}
	err = fmt.Errorf("dbmgr::PublishMsg error, not set use nats_streaming_server......")
	return
}

//订阅消息 nats_streaming_server
func SubscribeMsg() (err error) {
	if gDbMgr.UseNatsStream && gDbMgr.NatsStream != nil {
		//消费消息的回调函数,消费的消息只同步至mongo,不支持异步get操作
		err = gDbMgr.NatsStream.Subscribe(AsyncMsgManage)
		return
	}
	err = fmt.Errorf("dbmgr::SubscribeMsg error, not set use kafka......")
	return
}

//获取
func Get(para QueryPara, resVal interface{}) (err error) {
	//优先从redis中获取数据，没有数据再从mongodb中获取
	err = GetFromRedis(para, resVal)
	if err != nil { //有错误则从mongodb获取数据
		err = GetFromMongo(para, resVal)
		if err == nil { //从mongodb获取数据则同步至redis中
			SyncDataToRedis(para, resVal)
		}
	}
	return
}

//设置
func Set(para QueryPara, setVal interface{}) (err error) {
	async := true
	if gDbMgr.UseRedis && gDbMgr.Redis != nil { //如果有用户redis则先设置数据至redis
		err = SetFromRedis(para, setVal)
	} else if gDbMgr.UseMongo && gDbMgr.Mongo != nil { //如果没有使用redis，直接使用mongo的话，则设置数据至mongo
		err = SetFromMongo(para, setVal)
		if err == nil {
			async = false
		}
	}
	if async && err == nil {
		if gDbMgr.UseKafka && gDbMgr.Kafka != nil { //使用异步入库的话,则生产消息至kafka
			err = ProductionMsg(SET, para, setVal)
		}
		if gDbMgr.UseNatsStream && gDbMgr.NatsStream != nil { //使用异步入库的话,则发布消息至nats_streaming_server
			err = PublishMsg(SET, para, setVal)
		}
	}

	return
}

//删除
func Del(para QueryPara) (err error) {
	async := true
	if gDbMgr.UseRedis && gDbMgr.Redis != nil { //如果有用户redis则先设置数据至redis
		err = DelFromRedis(para)
	} else if gDbMgr.UseMongo && gDbMgr.Mongo != nil { //如果没有使用redis，直接使用mongo的话，则设置数据至mongo
		err = DelFromMongo(para)
		if err == nil {
			async = false
		}
	}
	if async && err == nil {
		if gDbMgr.UseKafka && gDbMgr.Kafka != nil { //使用异步入库的话,则生产消息至kafka
			ProductionMsg(DEL, para, nil)
		}
		if gDbMgr.UseNatsStream && gDbMgr.NatsStream != nil { //使用异步入库的话,则发布消息至nats_streaming_server
			err = PublishMsg(DEL, para, nil)
		}
	}
	return
}
