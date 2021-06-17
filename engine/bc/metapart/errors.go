package metapart

import (
	"fmt"

	e "github.com/tutumagi/pitaya/errors"
)

// 场景切换相关
var (
	// 这个比较特殊，因为进场景是异步消息，code直接放在返回进入场景的协议字段中，所以这里定义了一个成功的code
	SuccessEnterSpaceCode = "Space_000"
	ErrFindSpace          = func(spaceID string) *e.Error {
		return e.NewError(fmt.Errorf("场景进入失败，找不到该场景 %s", spaceID), "Space_001")
	}
	ErrMasterSpaceCreating = e.NewError(fmt.Errorf("主场景创建中，请稍候再试"), "Space_002")
	ErrEnteringSpace       = e.NewError(fmt.Errorf("当前实体已正在进入场景"), "Space_003")
	ErrAlreadyInSpace      = e.NewError(fmt.Errorf("实体已场景在当前场景中"), "Space_004")
	ErrSpaceDestroyed      = e.NewError(fmt.Errorf("场景已被释放，请稍候再试"), "Space_005")
	ErrChangeViewport      = e.NewError(fmt.Errorf("切换场景视图失败，请稍候再试"), "Space_006")
	ErrSpaceRequestFailed  = e.NewError(fmt.Errorf("请求进入场景失败"), "Space_007")
	ErrSpaceNotInAnySpace  = e.NewError(fmt.Errorf("玩家不在任何场景中"), "Space_008")

	ErrSpaceCannotFindEntity = func(id string, label string) *e.Error {
		return e.NewError(fmt.Errorf("找不到该实体%s:%s", label, id), "Space_009")
	}
	ErrSpaceUnknown = func(err error) *e.Error {
		return e.NewError(fmt.Errorf("未知的场景错误 %s", err), "Space_010")
	}
	ErrGenEnterSceneToken = func(err error) *e.Error {
		return e.NewError(fmt.Errorf("生成token失败 %v", err), "Space_011")
	}
	ErrParseEnterSceneToken = func(err error) *e.Error {
		return e.NewError(fmt.Errorf("解析token失败 %v", err), "Space_012")
	}
	ErrValidToken = func(str string) *e.Error {
		return e.NewError(fmt.Errorf("token 校验失败 %s", str), "Space_013")
	}
)
