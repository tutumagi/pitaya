package logger

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/spf13/viper"
)

type LogConfig struct {
	FilePath   string //文件路径
	FileName   string `json:"filename"` //文件名字
	Level      int    `json:"level"`    // 日志保存的时候的级别，默认是 Trace 级别
	Maxlines   int    `json:"maxlines"` // 每个文件保存的最大行数，默认值 1000000
	Maxsize    int    `json:"maxsize"`  // 每个文件保存的最大尺寸，默认值是 1 << 28, //256 MB
	Daily      bool   `json:"daily"`    // 是否按照每天 logrotate，默认是 true
	Maxdays    int    `json:"maxdays"`  // 文件最多保存多少天，默认保存 7 天
	Rotate     bool   `json:"rotate"`   // 是否开启 logrotate，默认是 true
	Perm       string `json:"perm"`     // 日志文件权限
	RotatePerm string `json:"rotateperm"`
	Color      bool   //是否支持颜色
}

//daylog
type DayLog struct {
	Cfg        *LogConfig      //配置
	Instance   *logs.BeeLogger //日志实例
	LogName    string          //配置的名字
	Dir        string          //路径
	CurLogName string          //当前使用的名字
	CurHours   int             //当前点数
}

//初始化
func (d *DayLog) init(v *viper.Viper, name string) bool {
	d.Cfg = &LogConfig{}
	d.Cfg.FilePath = v.GetString("daylog.filepath")
	d.Cfg.Level = logs.LevelInfo
	d.Cfg.Daily = false  //beego logs如果为true则，会开一个协程去处理 此模块已自己处理这些功能，所以不需要再开一个协程去处理
	d.Cfg.Rotate = false //beego logs如果为true则，会开一个协程去处理 此模块已自己处理这些功能，所以不需要再开一个协程去处理
	d.Cfg.Perm = "777"
	d.Cfg.RotatePerm = "777"
	d.Cfg.Color = true
	//原本的配置
	d.Dir = v.GetString("daylog.filepath")
	d.LogName = name
	//检测重置状态
	d.checkResetStatus()
	return true
}

//检测重置状态
func (d *DayLog) checkResetStatus() bool {
	//获取当前的点数与上一个点数比较，如果不同，则重置
	now := time.Now()
	hours := now.Hour()
	//点数不同则重置Logger
	if hours != d.CurHours {
		year := now.Year()
		month := now.Format("01")
		day := now.Day()
		d.CurHours = hours
		//文件的名字,包括文件的路径
		d.CurLogName = fmt.Sprintf("%s/%d-%s-%d/%s.%d", d.Dir, year, month, day, d.LogName, hours)
		d.Cfg.FileName = d.CurLogName
		if d.Instance != nil { //把以前的关闭
			d.Instance.Reset()
		} else {
			d.Instance = logs.NewLogger()
		}
		jsonConfig, _ := json.Marshal(d.Cfg)
		d.Instance.SetLogger(logs.AdapterFile, string(jsonConfig))
		//d.Instance.Async() //异步,一个日志实例开一个协程，资源浪费
		return true
	}
	return false
}

//info
func (d *DayLog) Info(format string, v ...interface{}) {
	if d.Instance != nil {
		d.checkResetStatus()
		d.Instance.Info(format, v...)
	}
}

//beego logs模块单协程只支持单文件读写，不支持单协程多文件读写，daylog一般都是多文件读写，为节约资源，参照beego logs模块中的log.go文件改造
type DayLogMgr struct {
	LogMap     map[int]*DayLog
	MsgChannel chan *AsyncMsg
}

type AsyncMsg struct {
	flag int
	msg  string
}

func (g *DayLogMgr) start() {
	for msg := range g.MsgChannel {
		defer func() { //捕捉异常，避免coredump
			if err := recover(); err != nil {
				fmt.Println(err)
			}
		}()
		if log, ok := gDLMgr.LogMap[msg.flag]; ok {
			log.Info(msg.msg)
		}
	}
}

var gDLMgr = new(DayLogMgr)

const ASYNC_CHANNEL_SIZE = 10000

//异步写log的channel size
func InitDayLog(v *viper.Viper) bool {
	gDLMgr.LogMap = map[int]*DayLog{}
	gDLMgr.MsgChannel = make(chan *AsyncMsg, ASYNC_CHANNEL_SIZE)
	if dlArr := v.GetStringSlice("daylog.name"); len(dlArr) > 0 {
		for _, cfg := range dlArr {
			name := strings.Split(cfg, "-")
			if len(name) == 2 {
				dayLog := &DayLog{}
				if dayLog.init(v, name[1]) {
					flag, err := strconv.Atoi(name[0])
					if err == nil {
						gDLMgr.LogMap[flag] = dayLog
					}
				}
			}
		}
		//异步
		if len(gDLMgr.LogMap) > 0 {
			go gDLMgr.start()
		}
		return true
	}
	return false
}

//异步写log
func DayLogRecord(flag int, format string, v ...interface{}) {
	async := &AsyncMsg{
		flag: flag,
		msg:  fmt.Sprintf(format, v...),
	}
	gDLMgr.MsgChannel <- async
}
