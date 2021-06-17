package aoi

import (
	"fmt"
	"math"
	"testing"

	"github.com/tutumagi/pitaya/engine/math32"

	. "github.com/go-playground/assert/v2"
)

// `go test gitlab.gamesword.com/nut/dreamcity/engine/aoi -count 1 -run CoordSystem -v`

func Test_TriggerNode(t *testing.T) {
	t.Run("one", func(t *testing.T) {
		t.Run("normal", func(t *testing.T) {
			OneNode(t, false)
		})
		t.Run("ref", func(t *testing.T) {
			OneNode(t, true)
		})
	})
	t.Run("two", func(t *testing.T) {
		t.Run("normal", func(t *testing.T) {
			TwoNode(t, false)
		})
		t.Run("ref", func(t *testing.T) {
			TwoNode(t, true)
		})
	})
}

// 测试只有一个节点的时候排序情况
func OneNode(t *testing.T, withRef bool) {
	m := &Map{}
	m.init(30, 30)
	m.system = NewCoordSystem().(*CoordSystem)
	{
		e := newEntityMock("0", math32.Vector3{X: 0, Y: 0, Z: 0}, AvatarRadius)

		if withRef {
			// m.entityEnterWithRef(e, nil)
			m.entityEnter(e)
		} else {
			m.entityEnter(e)
		}

		t.Log(m.system.Dump())

		Equal(t, e.interestByCount(), 0)
		Equal(t, e.interestInCount(), 0)

		Equal(t, m.system.(*CoordSystem).xSweepList.Count(), 3)

		checkList(t, m,
			[]*BaseCoord{
				e.negative(),
				e.coord.BaseCoord,
				e.positive(),
			},
			[]*BaseCoord{
				e.negative(),
				e.coord.BaseCoord,
				e.positive(),
			},
		)
	}
}

func TwoNode(t *testing.T, withRef bool) {
	m := &Map{}
	m.init(30, 30)
	m.system = NewCoordSystem()

	A := newEntityMock("A", math32.Vector3{X: -4, Y: 0, Z: -6}, AvatarRadius)
	B := newEntityMock("B", math32.Vector3{X: -8, Y: 0, Z: -1}, AvatarRadius)

	if withRef {
		// A实体自己的节点目前还无法直接通过参考点B，进入插入，因为在参考点之前可能有其他实体M的aoi半径包括了A，然后M的半径非常大，M离参考点B非常远
		// 这是A插入就触发不了M的进入视野
		// m.entityEnterWithRef(A, nil)
		// m.entityEnterWithRef(B, A)
		m.entityEnter(A)
		m.entityEnter(B)
	} else {
		m.entityEnter(A)
		m.entityEnter(B)
	}

	t.Logf("after %s", m.system.Dump())
	checkInterest(t, m)

	m.entityLeave(A)
	t.Logf(m.system.Dump())
	checkInterest(t, m)

	if withRef {
		// m.entityEnterWithRef(A, B)
	} else {
		m.entityEnter(A)
	}

	t.Logf(m.system.Dump())
	checkInterest(t, m)
}

func Test_CoordSystem(t *testing.T) {
	m := &Map{}
	m.init(30, 30)
	m.system = NewCoordSystem()

	/************************************ 实体进入 ************************************/

	// A(-5) A(0) A(5)
	A := newEntityMock("A", math32.Vector3{X: 0, Y: 0, Z: 0}, AvatarRadius)
	m.entityEnter(A)

	Equal(t, A.interestByCount(), 0)
	Equal(t, A.interestInCount(), 0)

	// B(-3) B(2) B(7)
	B := newEntityMock("B", math32.Vector3{X: 2, Y: 0, Z: 2}, AvatarRadius)
	m.entityEnter(B)

	checkInterest(t, m)

	// A(-5) B(-3) A(0) B(2) A(5) B(7)
	checkList(t, m,
		[]*BaseCoord{
			A.negative(),
			B.negative(),
			A.coord.BaseCoord,
			B.coord.BaseCoord,
			A.positive(),
			B.positive(),
		},
		[]*BaseCoord{
			A.negative(),
			B.negative(),
			A.coord.BaseCoord,
			B.coord.BaseCoord,
			A.positive(),
			B.positive(),
		},
	)

	t.Log(m.system.Dump())

	// C(0) C(5) C(10)
	C := newEntityMock("C", math32.Vector3{X: 5, Y: 0, Z: 5}, AvatarRadius)
	m.entityEnter(C)

	checkInterest(t, m)

	checkList(t, m,
		[]*BaseCoord{
			A.negative(),
			B.negative(),
			C.negative(),
			A.coord.BaseCoord,
			B.coord.BaseCoord,
			C.coord.BaseCoord,
			A.positive(),
			B.positive(),
			C.positive(),
		},
		[]*BaseCoord{
			A.negative(),
			B.negative(),
			C.negative(),
			A.coord.BaseCoord,
			B.coord.BaseCoord,
			C.coord.BaseCoord,
			A.positive(),
			B.positive(),
			C.positive(),
		},
	)

	t.Log(m.system.Dump())

	/************************************ 实体离开 ************************************/

	m.entityLeave(A)
	t.Log(m.system.Dump())

	checkInterest(t, m)

	// A(-5) B(-3) A(0) B(2) A(5) B(7)
	checkList(t, m,
		[]*BaseCoord{
			B.negative(),
			C.negative(),
			B.coord.BaseCoord,
			C.coord.BaseCoord,
			B.positive(),
			C.positive(),
		},
		[]*BaseCoord{
			B.negative(),
			C.negative(),
			B.coord.BaseCoord,
			C.coord.BaseCoord,
			B.positive(),
			C.positive(),
		},
	)

	m.entityLeave(B)

	checkInterest(t, m)

	// A(-5) B(-3) A(0) B(2) A(5) B(7)
	checkList(t, m,
		[]*BaseCoord{
			C.negative(),
			C.coord.BaseCoord,
			C.positive(),
		},
		[]*BaseCoord{
			C.negative(),
			C.coord.BaseCoord,
			C.positive(),
		},
	)

	m.entityLeave(C)
	checkList(t, m,
		[]*BaseCoord{},
		[]*BaseCoord{},
	)

	/****************************** 多个实体，使用不同的radius ***************************/
	m.entityEnter(A)
	t.Log(m.system.Dump())

	m.entityEnter(B)
	t.Log(m.system.Dump())

	checkInterest(t, m)

	m.entityEnter(C)
	t.Log(m.system.Dump())

	checkInterest(t, m)

	// D(-70) D(100) D(130)
	D := newEntityMock("D", math32.Vector3{X: 30, Y: 0, Z: 30}, 100)
	m.entityEnter(D)

	checkInterest(t, m)

	checkList(t, m,
		[]*BaseCoord{
			D.negative(),
			A.negative(),
			B.negative(),
			C.negative(),
			A.coord.BaseCoord,
			B.coord.BaseCoord,
			C.coord.BaseCoord,
			A.positive(),
			B.positive(),
			C.positive(),
			D.coord.BaseCoord,
			D.positive(),
		},
		[]*BaseCoord{
			D.negative(),
			A.negative(),
			B.negative(),
			C.negative(),
			A.coord.BaseCoord,
			B.coord.BaseCoord,
			C.coord.BaseCoord,
			A.positive(),
			B.positive(),
			C.positive(),
			D.coord.BaseCoord,
			D.positive(),
		},
	)

	/************************************ modify radius ***********************************/
	// 修改 radius 测试用例

	A.setRadius(5)
	checkInterest(t, m)

	A.setRadius(0)
	checkInterest(t, m)

	A.setRadius(200)
	checkInterest(t, m)

	t.Log(m.system.Dump())

	/************************************ 实体移动 ***********************************/
	// A(0,0,0)		-A(-5,0,-5)		+A(5,0,5)
	// B(2,0,2)		-B(-3,0,-3)		+B(3,0,3)
	// C(5,0,5)		-C(0,0,0)		+C(10,0,10)
	// D(30,0,30) 	-D(-70,0-70)	+D(130,0,130)

	// 先全部离开
	m.entityLeave(A)
	checkInterest(t, m)
	m.entityLeave(B)
	checkInterest(t, m)

	m.entityEnter(A)
	t.Log(m.system.Dump())
	checkInterest(t, m)

	m.entityLeave(C)
	checkInterest(t, m)
	m.entityLeave(D)
	checkInterest(t, m)

	m.entityLeave(A)
	checkInterest(t, m)

	// 只测试一个节点移动的时候
	m.entityEnter(A)
	m.entityMove(A, math32.Vector3{X: 100, Y: 0, Z: 0})
	t.Log(m.system.Dump())

	m.entityEnter(B)
	checkInterest(t, m)

	t.Log(m.system.Dump())
	m.entityMove(A, math32.Vector3{X: 3, Y: 0, Z: 2})
	t.Log(m.system.Dump())
	checkInterest(t, m)

	m.entityMove(B, math32.Vector3{X: 3, Y: 0, Z: 100})
	checkInterest(t, m)

	m.entityEnter(D)
	checkInterest(t, m)
	m.entityEnter(C)
	checkInterest(t, m)
	t.Run("move", func(t *testing.T) {
		for _, e := range m.entities {
			t.Logf("before %s", m.system.Dump())
			m.entityMove(e, randPos(-10, 10))
			t.Logf("after %s", m.system.Dump())
			checkInterest(t, m)
		}
	})
}

// go test gitlab.gamesword.com/nut/dreamcity/engine/aoi -run ManyEntities -count 100
func Test_ManyEntities(t *testing.T) {
	// 批量插入0视距半径
	t.Run("batch", func(t *testing.T) {
		optimize := true
		// 随机坐标
		t.Run("random", func(t *testing.T) {
			randomCoord := true
			t.Run("100", func(t *testing.T) {
				testInsertManyStaticWithCount(t, 100, 5, optimize, randomCoord)
			})
			t.Run("500", func(t *testing.T) {
				testInsertManyStaticWithCount(t, 500, 50, optimize, randomCoord)
			})
		})
		// 非随机坐标
		t.Run("norandom", func(t *testing.T) {
			randomCoord := false
			t.Run("100", func(t *testing.T) {
				testInsertManyStaticWithCount(t, 100, 5, optimize, randomCoord)
			})
			t.Run("500", func(t *testing.T) {
				testInsertManyStaticWithCount(t, 500, 50, optimize, randomCoord)
			})
		})
	})

	// 单个插入
	t.Run("nobatch", func(t *testing.T) {
		optimize := false
		// 随机坐标
		t.Run("random", func(t *testing.T) {
			randomCoord := true
			t.Run("100", func(t *testing.T) {
				testInsertManyStaticWithCount(t, 100, 5, optimize, randomCoord)
			})
			t.Run("500", func(t *testing.T) {
				testInsertManyStaticWithCount(t, 500, 50, optimize, randomCoord)
			})
		})
		// 非随机坐标
		t.Run("norandom", func(t *testing.T) {
			randomCoord := false
			t.Run("100", func(t *testing.T) {
				testInsertManyStaticWithCount(t, 100, 5, optimize, randomCoord)
			})
			t.Run("500", func(t *testing.T) {
				testInsertManyStaticWithCount(t, 500, 50, optimize, randomCoord)
			})
		})
	})
}

func testInsertManyStaticWithCount(
	t *testing.T,
	staticEntityCount int,
	avatarCount int,
	batchZeroRadiusOptimize bool,
	randomeCoord bool,
) {
	side := int(math.Ceil(math.Sqrt(float64(staticEntityCount))))

	m := &Map{}
	m.init(side, side)
	m.system = NewCoordSystem()

	// 进入一些没有aoi半径的实体
	t.Run("enter-static", func(t *testing.T) {
		var staticEntities []Entityer
		if batchZeroRadiusOptimize {
			staticEntities = make([]Entityer, 0, m.height*m.width)
		}
		for i := 0; i < m.height; i++ {
			for j := 0; j < m.width; j++ {
				pos := math32.Vector3{X: float32(j), Y: 0, Z: float32(i)}
				if randomeCoord {
					pos.X = randMM(0, m.width)
					pos.Y = randMM(0, m.height)
				}
				e := newEntityMock(fmt.Sprintf("x%dy%d", j, i), pos, 0)
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
		checkInterest(t, m)
	})

	// 进入玩家
	avatars := make([]*EntityMock, 0, avatarCount)

	t.Run("enter-avatar", func(t *testing.T) {
		for i := 0; i < avatarCount; i++ {
			pos := randPos(0, MapWidth)
			// t.Logf("enter avatar pos %s", pos)
			e := newEntityMock(fmt.Sprintf("Avatar%d", i), pos, randMM(0, m.width))
			m.entityEnter(e)

			avatars = append(avatars, e)
		}
		// println(m.system.Dump())
		checkInterest(t, m)
	})

	t.Run("move", func(t *testing.T) {
		for _, e := range avatars {
			m.entityMove(e, math32.Vector3{X: e.coord.pos.X + randMM(-10, 10), Y: 0, Z: e.coord.pos.Z + randMM(-10, 10)})
		}
		checkInterest(t, m)
	})

	t.Run("leave", func(t *testing.T) {
		for _, e := range avatars {
			m.entityLeave(e)
			checkInterest(t, m)
		}
		checkInterest(t, m)
	})

	// t.Run("enter-leave-run", func(t *testing.T) {
	// 	ranIdx := rand.Intn(len(avatars))
	// 	e := avatars[ranIdx]
	// 	m.entityLeave(e)
	// })
}

func checkList(t *testing.T, m *Map, expectXCoords []*BaseCoord, expectZCoords []*BaseCoord) {
	// 检查x轴排序是否ok
	EqualSkip(t, 2, m.system.(*CoordSystem).xSweepList.Count(), len(expectXCoords))
	for idx, expectX := range expectXCoords {
		EqualSkip(t, 2, m.getXNodeAtIndex(idx), expectX)
	}

	// 检查z轴排序是否ok
	EqualSkip(t, 2, m.system.(*CoordSystem).zSweepList.Count(), len(expectZCoords))
	for idx, expectZ := range expectZCoords {
		EqualSkip(t, 2, m.getZNodeAtIndex(idx), expectZ)
	}
}

// 暴力 for 循环，检测map中所有实体的 aoi 数据是否正常，仅在测试正确性中使用
func checkInterest(t *testing.T, m *Map) {

	type ExpectEntity map[*EntityMock]*struct {
		interestIn int
		interestBy int
	}
	expectData := make(ExpectEntity, len(m.entities))
	for _, e := range m.entities {
		expectData[e] = &struct {
			interestIn int
			interestBy int
		}{0, 0}
	}

	for _, entity := range m.entities {
		expect1 := expectData[entity]

		pos := entity.coord.pos
		radius := entity.Witness().radius

		// 如果 aoi 半径为0，则肯定没有关心的
		if radius == 0 {
			continue
		}

		// 半径是矩形的半径
		xLowerbound := pos.X - radius
		xUpperbound := pos.X + radius
		zLowerbound := pos.Z - radius
		zUpperbound := pos.Z + radius

		// fmt.Printf("%s x(%.2f,%.2f) z(%.2f,%.2f) \n", entity.id, xLowerbound, xUpperbound, zLowerbound, zUpperbound)

		for _, other := range m.entities {
			if other == entity {
				continue
			}
			expect2 := expectData[other]

			otherPos := other.coord.pos

			if otherPos.X >= xLowerbound && otherPos.X <= xUpperbound &&
				otherPos.Z >= zLowerbound && otherPos.Z <= zUpperbound {
				expect1.interestIn++
				expect2.interestBy++
			}
		}
	}

	for entity, expect := range expectData {
		// t.Logf("%s in:%d by:%d \n", entity.id, expect.interestIn, expect.interestBy)

		// 如果测试结果有问题，但是又不确定是哪里的问题，则打开用例测试方法，可以看出是哪个节点出错了
		// t.Run(fmt.Sprintf("check-%s-in", entity.id), func(t *testing.T) {
		EqualSkip(t, 2, entity.interestInCount(), expect.interestIn)
		// })
		// t.Run(fmt.Sprintf("check-%s-by", entity.id), func(t *testing.T) {
		EqualSkip(t, 2, entity.interestByCount(), expect.interestBy)
		// })
	}
}
