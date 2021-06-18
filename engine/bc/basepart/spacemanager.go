package basepart

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/tutumagi/pitaya/engine/bc/metapart"
	"github.com/tutumagi/pitaya/engine/common"
	"github.com/tutumagi/pitaya/logger"
)

var spaceManager = newSpaceManager()

// const _SpaceKindType = "__space_kind__"

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

// GetSpace get space
func GetSpace(id string) *Space {
	return spaceManager.getSpace(id)
}

// CreateSpace 创建space
// initCellServerID 初始创建场景时，在哪个 cellapp server 上
func CreateSpace(kind int32, id string, initCellServerID string) *Space {
	logger.Infof("create space %d %s", kind, id)
	if id == "" {
		id = metapart.NewUUID()
	}
	alreadySpace := spaceManager.getSpace(id)
	if alreadySpace != nil {
		logger.Warnf("space id duplicated kind:%d id%s", kind, id)
		return alreadySpace
	}

	// TODO 这里是否需要存储到db里面
	s := createBaseEntityOnlyInit(id, common.SpaceTypeName(kind))

	ss := s.AsSpace()
	// 这里 150000， 大概有11万的资源，3万的土地
	cap := 150000
	ss.entities = newBaseSet(cap)
	ss.I = ss.Entity.I.(ISpace)
	ss.kind = kind
	ss.initCellServerID = initCellServerID

	err := ss.I.OnSpaceInit()
	if err != nil {
		logger.Warnf("create space failed. id:%s kind:%s initCellServerID:%s err:%s",
			id, kind, err,
		)
		return nil
	}

	props := actor.PropsFromProducer(func() actor.Actor {
		return ss
	})
	ss.pid, _ = baseEntManager.system.Root.SpawnNamed(props, "space_"+ss.ID)

	s.I.OnCreated()

	return ss
}
