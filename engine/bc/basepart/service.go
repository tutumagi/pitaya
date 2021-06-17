package basepart

import (
	"github.com/tutumagi/pitaya/engine/bc/internal/consts"
	"github.com/tutumagi/pitaya/engine/bc/metapart"
)

// 这里的id
func CreateService(typName string, id string) *Entity {
	// createBaseEntityOnlyInit(metapart.NewUUID(), consts.ServiceTypeName(typName))
	if id == "" {
		id = metapart.NewUUID()
	}
	return CreateEntity(consts.ServiceTypeName(typName), id, nil, false)
}

func GetService(typName string, id string) *Entity {
	return baseEntManager.get(consts.ServiceTypeName(typName), id)
}
