package entity

import "bytes"

// Map is the data structure for maintaining entity IDs to entities
type Map map[string]*Entity

// Add adds a new entity to EntityMap
func (em Map) Add(entity *Entity) {
	em[entity.ID] = entity
}

// Del deletes an entity from EntityMap
func (em Map) Del(id string) {
	delete(em, id)
}

// Get returns the Entity of specified entity ID in EntityMap
func (em Map) Get(id string) *Entity {
	return em[id]
}

// Keys return keys of the EntityMap in a slice
func (em Map) Keys() (keys []string) {
	for eid := range em {
		keys = append(keys, eid)
	}
	return
}

// Values return values of the EntityMap in a slice
func (em Map) Values() (vals []*Entity) {
	for _, e := range em {
		vals = append(vals, e)
	}
	return
}

// Filter filter map
func (em Map) Filter(filter func(*Entity) bool) Map {
	r := Map{}
	for _, e := range em {
		if filter(e) {
			r.Add(e)
		}
	}
	return r
}

// Set is the data structure for a set of entities
type Set map[*Entity]struct{}

// Add adds an entity to the EntitySet
func (es Set) Add(entity *Entity) {
	es[entity] = struct{}{}
}

// Del deletes an entity from the EntitySet
func (es Set) Del(entity *Entity) {
	delete(es, entity)
}

// Contains returns if the entity is in the EntitySet
func (es Set) Contains(entity *Entity) bool {
	_, ok := es[entity]
	return ok
}

// ForEach do function
func (es Set) ForEach(f func(e *Entity)) {
	for e := range es {
		f(e)
	}
}

func (es Set) String() string {
	b := bytes.Buffer{}
	b.WriteString("{")
	first := true
	for entity := range es {
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
