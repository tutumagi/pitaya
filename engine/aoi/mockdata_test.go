package aoi

import (
	"fmt"
	"math/rand"

	"github.com/tutumagi/pitaya/engine/math32"
	"github.com/tutumagi/pitaya/logger"
)

type Map struct {
	system Systemer

	width  int
	height int

	entities map[string]*EntityMock
}

func (m *Map) init(width int, height int) {
	m.width = width
	m.height = height
	m.entities = make(map[string]*EntityMock)

}

func (m *Map) entityEnter(e *EntityMock) {
	if _, ok := m.entities[e.id]; ok {
		logger.Infof("fuck duplicate enter %s", e.id)
		return
	}
	// 这里 实体节点 移除坐标系统后，再进入坐标系统时，需要重置一下flag
	e.Coord().ResetFlags()
	e.setPos(*e.pos)
	e.witness = NewWitness()
	e.witness.Attach(e)
	e.witness.SetViewRadius(e.radius, 0)

	m.system.Insert(e.Coord().BaseCoord)

	e.onEnterSpace()

	m.entities[e.AoiID()] = e
}

func (m *Map) zeroRaidusEntitiesEnter(entities []Entityer) {
	// 实体的aoi半径必须是0
	for _, e := range entities {

		e.Coord().ResetFlags()

		e.(*EntityMock).witness = NewWitness()
		e.Witness().Attach(e)
		e.Witness().SetViewRadius(0, 0)
	}

	m.system.(*CoordSystem).InsertZeroRadiusEntities(entities)

	for _, e := range entities {
		e.(*EntityMock).onEnterSpace()

		m.entities[e.AoiID()] = e.(*EntityMock)
	}
}

func (m *Map) entityLeave(e *EntityMock) {
	if _, ok := m.entities[e.id]; !ok {
		return
	}
	// 时序很重要
	m.system.Remove(e.Coord().BaseCoord)
	e.Witness().Detach(e)
	e.witness = nil

	delete(m.entities, e.id)
}

func (m *Map) entityMove(e Entityer, pos math32.Vector3) {
	e.(*EntityMock).setPos(pos)
	e.Coord().Update()
}

func (m *Map) getXNodeAtIndex(idx int) *BaseCoord {
	return m.system.(*CoordSystem).xSweepList.GetDataByIndex(idx).(*node).coord
}

func (m *Map) getZNodeAtIndex(idx int) *BaseCoord {
	return m.system.(*CoordSystem).zSweepList.GetDataByIndex(idx).(*node).coord
}

type EntityMock struct {
	id string

	nCalc int

	pos     *math32.Vector3
	coord   *EntityCoord
	witness *Witness

	radius float32

	sights map[string]struct{}
}

func newEntityMock(id string, pos math32.Vector3, radius float32) *EntityMock {
	e := &EntityMock{}
	// e.id = fmt.Sprintf("%d", id)
	e.id = id

	e.coord = NewEntityNode(e)
	e.setPos(pos)

	e.radius = radius

	e.sights = make(map[string]struct{})
	return e
}

func (a *EntityMock) AoiID() string {
	return a.id
}

func (a *EntityMock) setPos(pos math32.Vector3) {
	a.pos = &pos
	a.coord.SetVec3(&pos)
}

func (a *EntityMock) String() string {
	return fmt.Sprintf("<EntityMock>(id:%s coord:%s)", a.id, a.coord)
}

func (a *EntityMock) setRadius(radius float32) {
	a.witness.SetViewRadius(radius, 0)
}

func (a *EntityMock) Coord() *EntityCoord {
	return a.coord
}

func (a *EntityMock) OnEnterAOI(other Entityer) {
	if _, ok := a.sights[other.AoiID()]; ok {
		panic(fmt.Sprintf("enter sight multi sight %s", other.AoiID()))
	}
	a.sights[other.AoiID()] = struct{}{}
}

func (a *EntityMock) OnLeaveAOI(other Entityer) {
	if _, ok := a.sights[other.AoiID()]; !ok {
		panic(fmt.Sprintf("leave sight multi sight %s", other.AoiID()))
	}
	delete(a.sights, other.AoiID())
}

func (a *EntityMock) interestByCount() int {
	if a.witness == nil {
		return 0
	}
	return len(a.witness.InterestedBy)
}

func (a *EntityMock) interestInCount() int {
	if a.witness == nil {
		return 0
	}
	return len(a.witness.InterestIn)
}

func (a *EntityMock) Witness() *Witness {
	return a.witness
}

func (a *EntityMock) onEnterSpace() {
	a.witness.InstallViewTrigger()
}

func (a *EntityMock) negative() *BaseCoord {
	return a.witness.trigger.negativeBoundary.BaseCoord
}

func (a *EntityMock) positive() *BaseCoord {
	return a.witness.trigger.positiveBoundary.BaseCoord
}

func randMM(min, max int) float32 {
	return float32(min) + float32(rand.Intn(int(max)-int(min)))
}

func randPos(min, max int) math32.Vector3 {
	return math32.Vector3{
		X: randMM(min, max),
		Y: 0,
		Z: randMM(min, max),
	}
}
