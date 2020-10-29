package aoi

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

const (
	MINX = -500
	MAXX = 500
	MINY = -500
	MAXY = 500

	NumObjs             = 10000
	VerifyNeighborCount = false
)

func TestAOIManager(t *testing.T) {
	const AOIRadius = 10
	const ItemRow = 20
	const ItemColumn = 20

	mgr := NewXZListAOIManager(AOIRadius)
	objs := []*_TestCallback{}
	for i := 0; i < ItemRow; i++ {
		for j := 0; j < ItemColumn; j++ {

			obj := &_TestCallback{
				id: i*ItemColumn + j,
			}
			// obj.aoi = *NewInterestedItem(obj)
			obj.aoi = *NewItem(AOIRadius, obj, obj)

			mgr.Enter(&obj.aoi, Coord(j), Coord(i))
			mgr.Ajust(&obj.aoi)

			objs = append(objs, obj)
		}
	}

	for i := 0; i < ItemRow*ItemColumn; i++ {
		obj := objs[i]
		neighborsCount := len(obj.aoi.impData.neighbors)
		t.Logf("objs[%v].neighbors count %v", i, neighborsCount)
		if neighborsCount > (AOIRadius * 2 * AOIRadius * 2) {
			t.Errorf("neighbors should less than 100, but %v", neighborsCount)
		}
	}
}

func TestXZListAOIManager(t *testing.T) {
	testAOI(t, NumObjs)
	// testCAOI(t, NumObjs)
}

func BenchmarkXZListAOIManager(b *testing.B) {
	testAOI(b, NumObjs)
}

func randCoord(min, max int) Coord {
	return Coord(min) + Coord(rand.Intn(int(max)-int(min)))
}

func testAOI(tb testing.TB, numAOI int) {

	mgr := NewXZListAOIManager(100)
	objs := []*_TestCallback{}
	for i := 0; i < numAOI; i++ {
		obj := &_TestCallback{
			id: i + 1,
		}
		// obj.aoi = *NewInterestedItem(obj)
		obj.aoi = *NewItem(100, obj, obj)
		objs = append(objs, obj)
		mgr.Enter(&obj.aoi, randCoord(MINX, MAXX), randCoord(MINY, MAXY))
		// mgr.Enter(&obj.aoi, Coord(numAOI-i), Coord(numAOI-i))

		mgr.Ajust(&obj.aoi)
	}

	for i := 0; i < 10; i++ {
		t0 := time.Now()
		for _, obj := range objs[0:1000] {
			mgr.Moved(&obj.aoi, obj.aoi.x+randCoord(-10, 10), obj.aoi.z+randCoord(-10, 10))
			mgr.Leave(&obj.aoi)

			mgr.Enter(&obj.aoi, obj.aoi.x+randCoord(-10, 10), obj.aoi.z+randCoord(-10, 10))
			mgr.Ajust(&obj.aoi)
		}
		dt := time.Now().Sub(t0)
		tb.Logf("tick %d objects takes %s", numAOI, dt)
	}

	// for _, obj := range objs {
	// 	mgr.Leave(&obj.aoi)
	// 	mgr.Ajust(&obj.aoi)
	// }

	// if VerifyNeighborCount {
	// 	totalCalc := int64(0)
	// 	for _, obj := range objs {
	// 		totalCalc += obj.nCalc
	// 	}
	// 	println("Average calculate count: ", totalCalc/int64(len(objs)))
	// }

}

func testCAOI(t *testing.T, numAOI int) {
	// t.Run("CAOI", func(t *testing.T) {
	// 	bm := _NewBindManager(0, 0, 4800, 4800, 4)

	// 	objs := make([]*Unit, 0, numAOI)
	// 	for i := int32(0); i < int32(numAOI); i++ {
	// 		obj := NewUnitNative(i, float32(randCoord(0, 4800)), float32(randCoord(0, 4800)))
	// 		objs = append(objs, obj)
	// 		bm.Enter(obj)
	// 	}

	// 	for i := 0; i < 10; i++ {
	// 		t0 := time.Now()
	// 		for _, obj := range objs[0:1000] {
	// 			bm.Moved(obj, randCoord(-10, 10), randCoord(-10, 10))
	// 			bm.Leave(obj)

	// 			bm.Enter(obj)
	// 			bm.Ajust(obj)
	// 		}

	// 		dt := time.Now().Sub(t0)
	// 		t.Logf("tick %d objects takes %s", numAOI, dt)
	// 	}

	// 	// for _, obj := range objs {
	// 	// 	bm.Leave(obj)
	// 	// }
	// })

}

type _TestCallback struct {
	aoi            Item
	id             int
	totalNeighbors int64
	nCalc          int64
}

func (tc *_TestCallback) OnEnterAOI(self *Item, other *Item) {
	if VerifyNeighborCount {

	}
}

func (tc *_TestCallback) OnLeaveAOI(self *Item, other *Item) {
	if VerifyNeighborCount {
	}
}

func (tc *_TestCallback) String() string {
	return fmt.Sprintf("TestCallback <%d>", tc.id)
}

func (tc *_TestCallback) getObj(aoi *Item) *_TestCallback {
	return aoi.Data.(*_TestCallback)
}
