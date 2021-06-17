package aoi

import "testing"

// cgo 目录下 的测试用例
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
