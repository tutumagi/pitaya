package basepart

import (
	"github.com/tutumagi/pitaya/engine/common"
)

// 这里的id
func CreateService(serviceName string) *Entity {
	// createBaseEntityOnlyInit(metapart.NewUUID(), consts.ServiceTypeName(typName))

	return CreateEntity(common.ServiceTypeName(serviceName), common.ServiceID(serviceName), nil, false)
}

func GetService(typName string, id string) *Entity {
	return baseEntManager.get(common.ServiceTypeName(typName), id)
}
