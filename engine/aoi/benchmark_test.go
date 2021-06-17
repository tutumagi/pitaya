package aoi

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/tutumagi/pitaya/engine/math32"
	// . "github.com/go-playground/assert/v2"
)

const (
	// AvatarObjs   = 1000
	AvatarRadius = 5

	MapWidth  = 142
	MapHeight = 142

	MinX = 0
	MaxX = MapWidth
	MinZ = 0
	MaxZ = MapHeight
)

func Benchmark_Empty(b *testing.B) {
	benchCaseWithCount(b, 0, 1000, false)
}

func Benchmark_10000(b *testing.B) {
	benchCaseWithCount(b, 10000, 1000, false)
}
func Benchmark_20000(b *testing.B) {
	benchCaseWithCount(b, 20000, 1000, false)
}
func Benchmark_30000(b *testing.B) {
	benchCaseWithCount(b, 30000, 1000, false)
}
func Benchmark_50000(b *testing.B) {
	benchCaseWithCount(b, 50000, 1000, false)
}

func Benchmark_80000(b *testing.B) {
	benchCaseWithCount(b, 80000, 1000, false)
}

func Benchmark_Opt10000(b *testing.B) {
	benchCaseWithCount(b, 10000, 1000, true)
}
func Benchmark_Opt20000(b *testing.B) {
	benchCaseWithCount(b, 20000, 1000, true)
}
func Benchmark_Opt30000(b *testing.B) {
	benchCaseWithCount(b, 30000, 1000, true)
}
func Benchmark_Opt50000(b *testing.B) {
	benchCaseWithCount(b, 50000, 1000, true)
}

func Benchmark_Opt80000(b *testing.B) {
	benchCaseWithCount(b, 80000, 1000, true)
}

func Benchmark_Opt200000(b *testing.B) {
	benchCaseWithCount(b, 200000, 1000, true)
}

func benchCaseWithCount(tb *testing.B, staticEntityCount int, avatarCount int, batchZeroRadiusOptimize bool) {
	side := int(math.Ceil(math.Sqrt(float64(staticEntityCount))))

	m := &Map{}
	m.init(side, side)
	m.system = NewCoordSystem()

	now := time.Now()

	// 进入一些静态实体
	var staticEntities []Entityer
	if batchZeroRadiusOptimize {
		staticEntities = make([]Entityer, 0, m.height*m.width)
	}

	for j := 0; j < m.height; j++ {
		for i := 0; i < m.width; i++ {
			e := newEntityMock(fmt.Sprintf("x%dy%d", i, j), math32.Vector3{X: float32(i), Y: 0, Z: float32(j)}, 0)
			if batchZeroRadiusOptimize {
				staticEntities = append(staticEntities, e)
			} else {
				m.entityEnter(e)
			}
		}
	}
	if batchZeroRadiusOptimize {
		m.zeroRaidusEntitiesEnter(staticEntities)
	}
	tb.Logf("enter %d entities, time:%s, coord count %d \n",
		len(m.entities), time.Now().Sub(now), m.system.(*CoordSystem).xSweepList.Count())

	// 进入玩家
	avatars := make([]*EntityMock, 0, avatarCount)

	now = time.Now()
	for i := 0; i < avatarCount; i++ {
		e := newEntityMock(fmt.Sprintf("Avatar%d", i), randPos(0, MapWidth), AvatarRadius)
		m.entityEnter(e)

		avatars = append(avatars, e)
	}
	tb.Logf("enter %d avatars, time:%s, coord count %d \n",
		len(avatars), time.Now().Sub(now), m.system.(*CoordSystem).xSweepList.Count())

	// 玩家移动
	tb.Run("move", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			now = time.Now()
			for _, e := range avatars {
				m.entityMove(e, math32.Vector3{X: e.coord.pos.X + randMM(-10, 10), Y: 0, Z: e.coord.pos.Z + randMM(-10, 10)})
			}
			b.Logf("%d avatars, move time:%s,\n",
				len(avatars), time.Now().Sub(now))
		}
	})

	// 玩家离开
	tb.Run("leave", func(b *testing.B) {
		now = time.Now()
		for _, e := range avatars {
			m.entityLeave(e)
		}
		tb.Logf("%d avatars, leave time:%s,\n",
			len(avatars), time.Now().Sub(now))
	})
	tb.Logf("\n")
}

// func Benchmark_Insert(tb *testing.B, staticEntityCount int) {
// 	side := int(math.Ceil(math.Sqrt(float64(staticEntityCount))))

// 	m := &Map{}
// 	m.init(side, side)

// 	// 进入一些静态实体
// 	now := time.Now()

// 	staticEntities := make([]*EntityMock, 0, m.width*m.height)
// 	for i := 0; i < m.height; i++ {
// 		for j := 0; j < m.width; j++ {
// 			e := newEntityMock(fmt.Sprintf("%d", i), math32.Vector3{X: float32(j), Y: 0, Z: float32(i)})
// 			staticEntities = append(staticEntities, e)
// 		}
// 	}
// 	m.batchEntityEnter(staticEntities)
// 	tb.Logf("enter %d entities, time:%s, coord count %d \n",
// 		len(m.entities), time.Now().Sub(now), m.system.xSweepList.Count())
// }
