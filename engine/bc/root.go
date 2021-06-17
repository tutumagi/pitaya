package bc

import (
	"github.com/tutumagi/pitaya/engine/bc/basepart"
	"gitlab.gamesword.com/nut/entitygen/attr"
)

const rootEntityLabel = "__root__entity___"

// _RootEntity 根实体，用来初始化主场景的
type _RootEntity struct {
	basepart.Entity
}

func init() {
	typeDesc := RegisterEntity(rootEntityLabel, &_RootEntity{}, nil, false)
	typeDesc.SetUseAOI(false, 0)
	typeDesc.SetMeta(&attr.Meta{})
}

func (r *_RootEntity) CellAttrChanged(keys map[string]struct{}) {
}

var _rootEntity *_RootEntity

func RootEntity() *basepart.Entity {
	if _rootEntity == nil {
		_rootEntity = basepart.CreateEntity(rootEntityLabel, "", nil, false).Val().(*_RootEntity)
	}
	return &_rootEntity.Entity
}
