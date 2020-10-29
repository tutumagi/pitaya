package entity

import (
	"fmt"

	"github.com/tutumagi/pitaya/logger"
)

var spaceManager = newSpaceManager()

const _SpaceEntityType = "__space__"
const _SpaceKindType = "__space_kind__"

type _SpaceManager struct {
	spaces map[string]*Space
}

func newSpaceManager() *_SpaceManager {
	return &_SpaceManager{
		spaces: map[string]*Space{},
	}
}

func (sm *_SpaceManager) putSpace(space *Space) {
	sm.spaces[space.ID] = space
}

func (sm *_SpaceManager) delSpace(id string) {
	delete(sm.spaces, id)
}

func (sm *_SpaceManager) getSpace(id string) *Space {
	return sm.spaces[id]
}

// RegisterSpace register custom space
func RegisterSpace(kind int32, spacePtr ISpace) {
	// spaceVal := reflect.Indirect(reflect.ValueOf(spacePtr))
	// spaceType = spaceVal.Type()

	RegisterEntity(spaceTypeName(kind), spacePtr, nil, false)
}

// GetSpace get space
func GetSpace(id string) *Space {
	return spaceManager.getSpace(id)
}

// CreateSpace 创建space
func CreateSpace(kind int32, id string) *Space {
	logger.Log.Debugf("create space !!!!!!!!!!!!!!!!!!!!!!! %d %s", kind, id)
	s := createEntityOnlyInit(spaceTypeName(kind), id, nil, nil, true)

	ss := s.AsSpace()

	ss.kind = kind

	s.I.OnCreated()

	return ss
}

func spaceTypeName(kind int32) string {
	return fmt.Sprintf("%s_%d", _SpaceEntityType, kind)
}
