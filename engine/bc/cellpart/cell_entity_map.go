package cellpart

import "bytes"

// CellMap is the data structure for maintaining entity IDs to entities
type CellMap map[string]*Entity

// Add adds a new entity to EntityMap
func (em CellMap) Add(entity *Entity) {
	em[entity.ID] = entity
}

// Del deletes an entity from EntityMap
func (em CellMap) Del(id string) {
	delete(em, id)
}

// Get returns the Entity of specified entity ID in EntityMap
func (em CellMap) Get(id string) *Entity {
	return em[id]
}

// Keys return keys of the EntityMap in a slice
func (em CellMap) Keys() (keys []string) {
	for eid := range em {
		keys = append(keys, eid)
	}
	return
}

// Values return values of the EntityMap in a slice
func (em CellMap) Values() (vals []*Entity) {
	for _, e := range em {
		vals = append(vals, e)
	}
	return
}

// Filter filter map
func (em CellMap) Filter(filter func(*Entity) bool) CellMap {
	r := CellMap{}
	for _, e := range em {
		if filter(e) {
			r.Add(e)
		}
	}
	return r
}

// CellSet is the data structure for a set of entities
type CellSet struct {
	src  map[*Entity]struct{}
	list []*Entity // 为了快速迭代
}

func newCellSet(cap int) *CellSet {
	return &CellSet{
		src:  make(map[*Entity]struct{}, cap),
		list: make([]*Entity, 0, cap),
	}
}

// Count return count
func (es *CellSet) Count() int {
	return len(es.list)
}

// Add adds an entity to the EntitySet
func (es *CellSet) Add(entity *Entity) {
	es.src[entity] = struct{}{}
	es.list = append(es.list, entity)
}

// Clear the set
func (es *CellSet) Clear() {
	for e := range es.src {
		delete(es.src, e)
	}
	es.list = es.list[0:0]
}

// Del deletes an entity from the EntitySet
func (es *CellSet) Del(entity *Entity) {
	delete(es.src, entity)

	find := -1
	for idx, e := range es.list {
		if e == entity {
			find = idx
			break
		}
	}
	if find != -1 {
		es.list[find] = es.list[len(es.list)-1]
		es.list = es.list[:len(es.list)-1]
	}
}

// Contains returns if the entity is in the EntitySet
func (es *CellSet) Contains(entity *Entity) bool {
	_, ok := es.src[entity]
	return ok
}

// ForEach do function
func (es *CellSet) ForEach(f func(e *Entity)) {
	for _, e := range es.list {
		f(e)
	}
}

func (es *CellSet) String() string {
	b := bytes.Buffer{}
	b.WriteString("{")
	first := true
	for _, entity := range es.list {
		if !first {
			b.WriteString(", ")
		} else {
			first = false
		}
		b.WriteString(entity.String())
	}
	b.WriteString("}")
	return b.String()
}
