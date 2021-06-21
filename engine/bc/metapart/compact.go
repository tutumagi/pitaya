package metapart

import (
	"fmt"

	e "github.com/tutumagi/pitaya/errors"
)

// 实体类型
const TypNamePlayer = "role"

// MasterSpaceID 主场景SpaceID
const MasterSpaceID = "MasterSpaceID"

// SpaceKind
const (
	// 默认空场景
	NilSpaceKind = 0
	// 主场景
	MasterSpaceKind = 1
)

const GateAppSvr = "gateapp"

// 玩家进入/切换场景返回的路由
const RouterAvatarEnterScene = "avatar.enterscene"

func EntityTableName(entityTypName string) string {
	return fmt.Sprintf("tbl_%s", entityTypName)
}

// entity 相关
var (
	ErrEntityNotRegister = e.NewError(fmt.Errorf("实体未注册"), "Entity_001")

	ErrLoadDB  = e.NewError(fmt.Errorf("数据库读取错误"), "DB_001")
	ErrWriteDB = e.NewError(fmt.Errorf("数据库写入错误"), "DB_002")
)

const (
	DB_SUCC_LOG        = 1006 //记录db入库成功记录
	DB_ERROR_LOG       = 1007 //记录db入库失败记录
	DB_ASYNC_ERROR_LOG = 1008 //异步db发布失败记录
)
