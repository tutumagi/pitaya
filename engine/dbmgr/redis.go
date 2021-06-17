package dbmgr

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/tutumagi/pitaya/logger"

	"github.com/go-redis/redis"
	"github.com/spf13/viper"
)

//管理redis模块的操作

var poolSize = 20

const (
	Single  = 0
	Cluster = 1
)

//支持单机跟集群
type RedisHandler struct {
	tye         int                  //类型 0单机 1集群
	hostCluster []string             //集群服务器地址
	host        string               //单机服务器地址
	auth        bool                 //是否授权
	pwd         string               //密码
	dbName      string               //数据库的名字
	conn        *redis.Client        //单机的前端连接信息
	connCluster *redis.ClusterClient //集群连接地址
}

//初始化
func (this *RedisHandler) Init(cfg *viper.Viper) error {
	if cfg != nil {
		this.tye = cfg.GetInt("dbmgr.redis.type")
		this.hostCluster = cfg.GetStringSlice("dbmgr.redis.hostCluster")
		this.host = cfg.GetString("dbmgr.redis.host")
		this.auth = cfg.GetBool("dbmgr.redis.auth")
		this.pwd = cfg.GetString("dbmgr.redis.pwd")
		this.dbName = cfg.GetString("dbmgr.redis.dbname")
		return this.Reconn()
	}
	return fmt.Errorf("RedisHandler::Init error, configure is null pointer")
}

func (this *RedisHandler) Reconn() error {
	if this.tye == Single {
		return this.singleReconn()
	} else {
		return this.clusterReconn()
	}
	return nil
}

//单机重连
func (this *RedisHandler) singleReconn() (err error) {
	needReconn := false
	//重连三次
	if this.conn == nil {
		needReconn = true
	} else if _, err = this.conn.Ping().Result(); err != nil { //ping不了也需要重连
		this.close()
		needReconn = true
	}
	if needReconn {
		loop := 0
		connPara := &redis.Options{
			Addr:     this.host,
			Password: this.pwd,
			DB:       0,
			PoolSize: poolSize,
		}
		this.conn = redis.NewClient(connPara)
		for loop < 3 {
			if _, err = this.conn.Ping().Result(); err != nil {
				logger.Errorf("RedisHandler::singleReconn single error, cann't ping the redis server[url:%s, error:%+v]", this.host, err)
				loop++
			} else {
				logger.Debugf("RedisHandler::singleReconn single successfully[url:%s]", this.host)
				return
			}
		}
	}
	return
}

//集群reconnection
func (this *RedisHandler) clusterReconn() (err error) {
	//重连三次
	needReconn := false
	if this.connCluster == nil {
		needReconn = true
	} else if _, err = this.connCluster.Ping().Result(); err != nil { //ping不了也需要重连
		this.close()
		needReconn = true
	}
	if needReconn {
		loop := 0
		connPara := &redis.ClusterOptions{
			Addrs:    this.hostCluster,
			Password: this.pwd,
			PoolSize: poolSize,
		}
		this.connCluster = redis.NewClusterClient(connPara)
		for loop < 3 { //重连三次，这是调用端
			if _, err = this.connCluster.Ping().Result(); err != nil {
				logger.Errorf("RedisHandler::clusterReconn cluster error, cann't ping the redis server[url:%+v, error:%+v]", this.hostCluster, err)
				loop++
			} else {
				logger.Debugf("RedisHandler::clusterReconn cluster successfully[url:%+v]", this.hostCluster)
				return
			}
		}
	}
	return
}

//关闭
func (this *RedisHandler) close() (err error) {
	if this.conn != nil {
		if err = this.conn.Close(); err == nil {
			this.conn = nil
		}
	}
	if this.connCluster != nil {
		if err = this.connCluster.Close(); err == nil {
			this.connCluster = nil
		}
	}
	return
}

//set
func (this *RedisHandler) Set(para QueryPara, setVal interface{}) (err error) {
	if err = this.Reconn(); err == nil {
		keyName := GetKeyName(para, this.dbName)
		field := ""
		if para.Key != nil {
			field = fmt.Sprintf("%+v", para.Key)
		}
		//转成json字符串存储
		var valByte []byte
		valByte, err = json.Marshal(setVal)
		if string(valByte) == "{}" {
			logger.Errorf("!!!!!!!!!!! save to redis empty object %+v", para)
		}
		if err == nil {
			valStr := string(valByte)
			if this.tye == Single {
				_, err = this.conn.HSet(keyName, field, valStr).Result()
			} else {
				_, err = this.connCluster.HSet(keyName, field, valStr).Result()
			}
		}
	}
	return
}

//get
func (this *RedisHandler) Get(para QueryPara, resVal interface{}) (err error) {
	if err = this.Reconn(); err == nil {
		//需要指针来接收结果
		if reflect.TypeOf(resVal).Kind() != reflect.Ptr {
			err = fmt.Errorf("parameter error, the result parameter not pointer")
			return
		}
		refVal := reflect.ValueOf(resVal)
		keyName := GetKeyName(para, this.dbName)
		field := ""
		if para.Key != nil {
			field = fmt.Sprintf("%+v", para.Key)
		} else if refVal.Elem().Kind() != reflect.Array && refVal.Elem().Kind() != reflect.Slice {
			err = fmt.Errorf("parameter error, the result parameter not array or slice")
			return
		}
		//没有指定的field,获取所有
		if field == "" {
			//获取所有必须传slice
			var ret = map[string]string{}
			if this.tye == Single { //单机
				ret, err = this.conn.HGetAll(keyName).Result()
			} else { //集群
				ret, err = this.connCluster.HGetAll(keyName).Result()
			}
			if err == nil {
				sliceElem := refVal.Elem()
				elemType := sliceElem.Type().Elem()
				index := 0
				for _, v := range ret {
					if sliceElem.Len() == index { //如果长度不够，则会生成一个新的
						newElem := reflect.New(elemType)
						sliceElem = reflect.Append(sliceElem, newElem.Elem())
						sliceElem = sliceElem.Slice(0, sliceElem.Cap())
					}
					elem := sliceElem.Index(index).Addr().Interface()
					if err = json.Unmarshal([]byte(v), &elem); err == nil {
						index++
					} else {
						continue
					}
				}
				refVal.Elem().Set(sliceElem.Slice(0, index))
				if index <= 0 { //数据为空
					err = ErrDataEmpty
				}
			}
		} else { //获取指定的field,参数必须为指针
			var ret string
			if this.tye == Single {
				ret, err = this.conn.HGet(keyName, field).Result()
			} else {
				ret, err = this.connCluster.HGet(keyName, field).Result()
			}
			if err == nil {
				if ret != "" {
					json.Unmarshal([]byte(ret), resVal)
				} else { //数据为空
					err = ErrDataEmpty
				}
			}
		}
	}
	return
}

//exists
func (this *RedisHandler) Exists(para QueryPara) bool {
	if this.Reconn() == nil {
		keyName := GetKeyName(para, this.dbName)
		//查询某一段数据是否存在,主键查询
		if para.Key != nil {
			field := fmt.Sprintf("%+v", para.Key)
			var ret bool = false
			var err error = nil
			if this.tye == Single {
				ret, err = this.conn.HExists(keyName, field).Result()
			} else {
				ret, err = this.connCluster.HExists(keyName, field).Result()
			}
			if err != nil {
				return false
			}
			return ret
		} else { //字段名是否存在
			var ret int64 = 0
			var err error = nil
			if this.tye == Single {
				ret, err = this.conn.Exists(keyName).Result()
			} else {
				ret, err = this.connCluster.Exists(keyName).Result()
			}

			if err != nil {
				return false
			}
			if ret == 0 { //0则是不存在
				return false
			}
			return true
		}
	}
	return false
}

//expire
func (this *RedisHandler) Expire(para QueryPara, sec time.Duration) bool {
	if this.Reconn() == nil {
		keyName := GetKeyName(para, this.dbName)
		var ret bool = false
		var err error = nil
		if this.tye == Single {
			ret, err = this.conn.Expire(keyName, sec).Result()
		} else {
			ret, err = this.connCluster.Expire(keyName, sec).Result()
		}

		if err != nil {
			return false
		}
		return ret
	}
	return false
}

//del
func (this *RedisHandler) Del(para QueryPara) (err error) {
	if err = this.Reconn(); err == nil {
		keyName := GetKeyName(para, this.dbName)
		field := ""
		if para.Key != nil {
			field = fmt.Sprintf("%+v", para.Key)
		}
		//没有指定的field,则删除所有
		if field == "" {
			if this.tye == Single {
				_, err = this.conn.Del(keyName).Result()
			} else {
				_, err = this.connCluster.Del(keyName).Result()
			}

		} else { //指定的field，删除指定的数据
			if this.tye == Single {
				_, err = this.conn.HDel(keyName, field).Result()
			} else {
				_, err = this.connCluster.HDel(keyName, field).Result()
			}
		}
	}
	return
}

//list push
func (this *RedisHandler) Lpush(para QueryPara, val interface{}) (err error) {
	if err = this.Reconn(); err == nil {
		keyName := GetKeyName(para, this.dbName)
		//转成json字符串存储
		var valByte []byte
		valByte, err = json.Marshal(val)
		if err == nil {
			valStr := string(valByte)
			//从列表的头部进行添加
			if this.tye == Single {
				_, err = this.conn.LPush(keyName, valStr).Result()
			} else {
				_, err = this.connCluster.LPush(keyName, valStr).Result()
			}
		}
	}
	return
}

//list pop
func (this *RedisHandler) Rpop(para QueryPara, val interface{}) (err error) {
	if err = this.Reconn(); err == nil {
		//需要指针来接收结果
		if reflect.TypeOf(val).Kind() != reflect.Ptr {
			err = fmt.Errorf("parameter error, the result parameter not pointer")
			return
		}
		keyName := GetKeyName(para, this.dbName)
		var ret string = ""
		var err error = nil
		if this.tye == Single {
			ret, err = this.conn.RPop(keyName).Result()
		} else {
			ret, err = this.connCluster.RPop(keyName).Result()
		}
		if err == nil {
			json.Unmarshal([]byte(ret), val)
		}
	}
	return
}

//同步数据
func (this *RedisHandler) SyncData(para QueryPara, val interface{}) (err error) {
	if err = this.Reconn(); err == nil {
		refElem := reflect.ValueOf(val)
		if refElem.Kind() == reflect.Ptr { //如果是指针
			refElem = refElem.Elem() //这个才是缓存数据
		}
		//解析函数
		var parseFieldKey func(rType reflect.Type, rVal reflect.Value)
		parseFieldKey = func(rType reflect.Type, rVal reflect.Value) {
			rfType := rType
			if rfType.Kind() == reflect.Ptr { //指针
				rfType = rfType.Elem()
			}
			for j := 0; j < rVal.NumField(); j++ {
				fieldVal := rfType.Field(j).Tag.Get("json")
				if fieldVal != "" {
					rrVal := rVal.Field(j)
					if fieldVal == para.FieldName {
						para.Field = rrVal.Interface()
					} else if fieldVal == para.KeyName {
						para.Key = rrVal.Interface()
					} else if rrVal.Kind() == reflect.Struct || rrVal.Kind() == reflect.Ptr { //如果是结构体或指针,则递归
						if rrVal.Kind() == reflect.Ptr { //指针需要把实际的缓存数据取出来
							rrVal = rrVal.Elem()
						}
						parseFieldKey(rrVal.Type(), rrVal)
					}
				}
			}
		}
		//数组或切片
		if refElem.Kind() == reflect.Array || refElem.Kind() == reflect.Slice {
			refType := refElem.Type().Elem()
			for i := 0; i < refElem.Len(); i++ {
				ele := refElem.Index(i).Elem()
				if para.FieldName != "" || para.KeyName != "" {
					parseFieldKey(refType, ele)
				}
				err = this.Set(para, ele.Interface())
			}
		} else {
			refType := refElem.Type()
			parseFieldKey(refType, refElem)
			err = this.Set(para, refElem.Interface())
		}
	}
	return
}
