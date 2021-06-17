package dbmgr

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/tutumagi/pitaya/logger"

	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

//条件查找的标识
const (
	SELECTOR = iota
	GT       //大于
	GTE      //大于等于
	LT       //小于
	LTE      //小于等于
	NE       //不等于
	IN       //在[]中
	NIN      //不在[]中
)

const (
	DESC = -1 //降序
	ASC  = 1  //升序
)

type MongoHandler struct {
	host   string        //host地址
	port   int32         //端口
	user   string        //用户
	pwd    string        //密码
	dbName string        //库名
	auth   bool          //是否授权
	url    string        //连接信息
	conn   *mongo.Client //前端连接实例
}

//初始化
func (this *MongoHandler) Init(cfg *viper.Viper) error {
	//参数初始化
	this.host = cfg.GetString("dbmgr.mongo.host")
	this.port = cfg.GetInt32("dbmgr.mongo.port")
	this.auth = cfg.GetBool("dbmgr.mongo.auth")
	this.user = cfg.GetString("dbmgr.mongo.user")
	this.pwd = cfg.GetString("dbmgr.mongo.pwd")
	this.dbName = cfg.GetString("dbmgr.mongo.dbname")
	return this.Reconn()
}

//重连
func (this *MongoHandler) Reconn() error {
	var err error = nil
	var needReconn = false
	if this.conn == nil {
		needReconn = true
	} else if err = this.conn.Ping(context.TODO(), nil); err != nil { //ping不通也需要重连
		this.Close()
		needReconn = true
	}
	if needReconn {
		this.url = fmt.Sprintf("mongodb://%s:%d", this.host, this.port)
		option := options.Client()
		option.ApplyURI(this.url)
		option.SetMaxPoolSize(1000)
		option.SetMaxConnIdleTime(24 * time.Hour)

		if this.auth {
			option.SetAuth(options.Credential{
				Username: this.user,
				Password: this.pwd,
				//AuthSource: this.dbName,
			})
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(5))
		defer cancel()
		loop := 0
		for loop < ReconnTimes { //重连三次
			this.conn, err = mongo.Connect(ctx, option)
			if err == nil {
				logger.Debugf("MongoHandler::Reconn successfully:%s", this.url)
				break
			}
			loop++
		}
	}

	return err
}

//关闭
func (this *MongoHandler) Close() error {
	if this.conn != nil {
		err := this.conn.Disconnect(context.TODO())
		this.conn = nil
		return err
	}
	return nil
}

//将参数解析成Mongodb的select语句
func (this *MongoHandler) ParseSelector(para QueryPara) bson.M {
	selector := bson.M{}
	if para.FieldName != "" && para.Field != nil {
		selector[para.FieldName] = para.Field
	}
	if para.KeyName != "" && para.Key != nil {
		selector[para.KeyName] = para.Key
	}
	for k, v := range para.Selector {
		switch v.Flag {
		case GT:
			selector[k] = bson.M{"$gt": v.Val}
		case GTE:
			selector[k] = bson.M{"$gte": v.Val}
		case LT:
			selector[k] = bson.M{"$lt": v.Val}
		case LTE:
			selector[k] = bson.M{"$lte": v.Val}
		case NE:
			selector[k] = bson.M{"$ne": v.Val}
		case NIN:
			selector[k] = bson.M{"$nin": v.Val}
		case IN:
			selector[k] = bson.M{"$in": v.Val}
		default:
			selector[k] = v.Val
		}
	}
	return selector
}

//解析查找器
func (this *MongoHandler) ParseFinder(para QueryPara) *options.FindOptions {
	findOptions := options.Find()
	for k, v := range para.Sort { //排序设置
		findOptions.SetSort(bson.M{k: v})
	}
	if para.Skip > 0 { //设置返回的条数限制
		findOptions.SetSkip(int64(para.Skip))
	}
	if para.Limit > 0 { //设置返回的数量
		findOptions.SetLimit(int64(para.Limit))
	}
	return findOptions
}

//获取
func (this *MongoHandler) Get(para QueryPara, resVal interface{}) (err error) {
	defer func() { //捕获异常
		if e := recover(); e != nil {
			err = fmt.Errorf("MongoHandler::Get panic:%+v", e)
		}
	}()
	//参数校验，返回的结果必须为地址传值
	if para.TblName == "" {
		err = fmt.Errorf("MongoHandler::Get function parameter error, not about table name[parameter:%+v]", para)
		return
	}
	//必须为址传递
	if reflect.TypeOf(resVal).Kind() != reflect.Ptr {
		err = fmt.Errorf("MongoHandler::Get function error, result parameter not pointer:%+v", resVal)
		return
	}
	if err = this.Reconn(); err == nil { //检测连接是否正常,连接正常才能使用
		collection := this.conn.Database(this.dbName).Collection((para.TblName))
		if collection == nil { //如果database跟table的collection不存在，则返回错误
			err = fmt.Errorf("MongoHandler::Get error, the database table collection not existed[dbname:%s, table:%s]", this.dbName, para.TblName)
			return
		}
		selector := this.ParseSelector(para) //解析参数，形成mongodb的select语句
		refVal := reflect.ValueOf(resVal)
		if refVal.Elem().Kind() == reflect.Array || refVal.Elem().Kind() == reflect.Slice { //查找多个
			findOptions := this.ParseFinder(para) //解析查找返回限制
			var cursor *mongo.Cursor              //游标
			if cursor, err = collection.Find(context.TODO(), selector, findOptions); err != nil {
				return
			}
			sliceElem := refVal.Elem()
			elemType := sliceElem.Type().Elem()
			index := 0 //遍历游标获取结果
			for cursor.Next(context.TODO()) {
				if sliceElem.Len() == index { //如果长度不够，则会生成一个新的
					newElem := reflect.New(elemType)
					sliceElem = reflect.Append(sliceElem, newElem.Elem())
					sliceElem = sliceElem.Slice(0, sliceElem.Cap())
				}
				elem := sliceElem.Index(index).Addr().Interface()
				err = cursor.Decode(elem)
				if err != nil {
					return
				}
				index++
			}
			refVal.Elem().Set(sliceElem.Slice(0, index))
			cursor.Close(context.TODO()) //关闭游标
			if index <= 0 {              //数据为空则返回错误
				err = ErrDataEmpty
			}
		} else { //查找一个
			singleRes := collection.FindOne(context.TODO(), selector)
			if err = singleRes.Err(); err != nil {
				if err == mongo.ErrNoDocuments { //没有数据则返回数据为空的错误
					err = ErrDataEmpty
				}
				return
			}
			err = singleRes.Decode(resVal) //解压
		}
	}
	return
}

//设置，没有则会插入
func (this *MongoHandler) Set(para QueryPara, setVal interface{}) (err error) {
	defer func() { //捕获异常
		if e := recover(); e != nil {
			err = fmt.Errorf("MongoHandler::Set panic:%+v", e)
		}
	}()
	if para.TblName == "" || (para.FieldName == "" && para.KeyName == "") { //参数缺失
		err = fmt.Errorf("MongoHandler::Set parameter missing:%+v", para)
		return
	}
	if err = this.Reconn(); err == nil {
		collection := this.conn.Database(this.dbName).Collection((para.TblName))
		if collection == nil { //如果database跟table的collection不存在，则返回错误
			err = fmt.Errorf("MongoHandler::Set error, the database table collection not existed[dbname:%s, table:%s]", this.dbName, para.TblName)
			return
		}
		selector := this.ParseSelector(para)
		dbSet := bson.M{"$set": setVal}
		//更新数据，没有则插入，批量更新，bulkWrite只是可以分批执行命令，底层还是用基本的updateMany/updateOne/deleteMany/deleteOne/insertOne/insertMany,
		_, err = collection.UpdateMany(context.TODO(), selector, dbSet, options.Update().SetUpsert(true))

	}
	return
}

//删除
func (this *MongoHandler) Del(para QueryPara) (err error) {
	defer func() { //捕获异常
		if e := recover(); e != nil {
			err = fmt.Errorf("MongoHandler::Del panic:%+v", e)
		}
	}()
	if para.TblName == "" { //参数缺失
		err = fmt.Errorf("MongoHandler::Del parameter missing:%+v", para)
		return
	}
	if err = this.Reconn(); err == nil {
		collection := this.conn.Database(this.dbName).Collection((para.TblName))
		if collection == nil { //如果database跟table的collection不存在，则返回错误
			err = fmt.Errorf("MongoHandler::Del error, the database table collection not existed[dbname:%s, table:%s]", this.dbName, para.TblName)
			return
		}
		selector := this.ParseSelector(para)
		_, err = collection.DeleteMany(context.TODO(), selector)
	}
	return
}

//索引
func (this *MongoHandler) Inx(para QueryPara) (err error) {
	defer func() { //捕获异常
		if e := recover(); e != nil {
			err = fmt.Errorf("MongoHandler::Inx panic:%+v", e)
		}
	}()
	if para.TblName == "" { //参数缺失
		err = fmt.Errorf("MongoHandler::Inx parameter missing:%+v", para)
		return
	}
	if err = this.Reconn(); err == nil {
		collection := this.conn.Database(this.dbName).Collection((para.TblName))
		if collection == nil { //如果database跟table的collection不存在，则返回错误
			err = fmt.Errorf("MongoHandler::Inx error, the database table collection not existed[dbname:%s, table:%s]", this.dbName, para.TblName)
			return
		}
		idx := mongo.IndexModel{}
		bSlice := bsonx.Doc{}
		//雙key的索引
		if para.FieldName != "" {
			bSlice = append(bSlice, bsonx.Elem{Key: para.FieldName, Value: bsonx.Int32(1)})
		}
		if para.KeyName != "" {
			bSlice = append(bSlice, bsonx.Elem{Key: para.KeyName, Value: bsonx.Int32(1)})
		}

		for k := range para.Selector {
			if k != "" {
				bSlice = append(bSlice, bsonx.Elem{Key: k, Value: bsonx.Int32(1)})
			}
		}
		if len(bSlice) > 0 { //有设置索引key
			idx.Keys = bSlice
			_, err = collection.Indexes().CreateOne(context.Background(), idx)
		}
	}
	return
}
