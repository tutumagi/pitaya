package bc

// import (
// 	"strings"
// 	"testing"

// 	. "github.com/go-playground/assert/v2"
// )

// func newEntity(id string) *Entity {
// 	return &Entity{ID: id}
// }

// func checkForeach(t *testing.T, set *Set, elist []*Entity) {
// 	var expect strings.Builder

// 	for _, e := range elist {
// 		expect.WriteString(e.ID)
// 	}

// 	var str strings.Builder
// 	set.ForEach(func(e *Entity) {
// 		str.WriteString(e.ID)
// 	})

// 	EqualSkip(t, 2, str.String(), expect.String())
// }

// func TestEntitySet(t *testing.T) {
// 	set := newSet(100)

// 	Equal(t, cap(set.list), 100)
// 	Equal(t, set.Count(), 0)

// 	e1 := newEntity("e1")
// 	set.Add(e1)
// 	Equal(t, set.Count(), 1)
// 	Equal(t, set.Contains(e1), true)
// 	checkForeach(t, set, []*Entity{e1})

// 	e2 := newEntity("e2")
// 	set.Add(e2)
// 	Equal(t, set.Count(), 2)
// 	Equal(t, set.Contains(e1), true)
// 	Equal(t, set.Contains(e2), true)
// 	checkForeach(t, set, []*Entity{e1, e2})

// 	set.Del(e1)
// 	Equal(t, set.Count(), 1)
// 	Equal(t, set.Contains(e1), false)
// 	Equal(t, set.Contains(e2), true)
// 	checkForeach(t, set, []*Entity{e2})

// 	set.Del(e1)
// 	Equal(t, set.Count(), 1)
// 	Equal(t, set.Contains(e1), false)
// 	Equal(t, set.Contains(e2), true)
// 	checkForeach(t, set, []*Entity{e2})

// 	set.Del(e2)
// 	Equal(t, set.Count(), 0)
// 	Equal(t, set.Contains(e1), false)
// 	Equal(t, set.Contains(e2), false)
// 	checkForeach(t, set, []*Entity{})

// 	e3 := newEntity("e3")
// 	set.Add(e3)
// 	Equal(t, set.Count(), 1)
// 	Equal(t, set.Contains(e1), false)
// 	Equal(t, set.Contains(e2), false)
// 	Equal(t, set.Contains(e3), true)
// 	checkForeach(t, set, []*Entity{e3})

// 	set.Clear()
// 	Equal(t, len(set.src), 0)
// 	Equal(t, len(set.list), 0)
// 	Equal(t, set.Contains(e1), false)
// 	Equal(t, set.Contains(e2), false)
// 	Equal(t, set.Contains(e3), false)
// 	checkForeach(t, set, []*Entity{})

// }
