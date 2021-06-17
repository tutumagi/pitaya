package aoi

// 绑定 cgo 目录下 c语言实现的 aoi
// // #cgo LDFLAGS: -L ./cgo/build/ -laoi
// // #include "./cgo/aoi.h"
// import "C"
// import (
// 	"fmt"
// 	"math/rand"
// 	"runtime"
// 	"unsafe"
// )

// var (
// 	// UnitRadius 单位半径
// 	UnitRadius = 10
// 	// SearchRadius 搜索半径
// 	SearchRadius float64 = 200.0
// )

// // ------------------ aoi unit ------------------

// // Unit aoi单位
// type Unit struct {
// 	U *C.struct_iunit
// }

// // NewUnitNative 新建aoi单位
// func NewUnitNative(id int32, x, y float32) *Unit {
// 	return NewUnit(C.iid(id), C.ireal(x), C.ireal(y))
// }

// // NewUnit 新建aoi单位
// func NewUnit(uniqueID C.iid, x, y C.ireal) *Unit {
// 	u := &Unit{
// 		U: C.imakeunit(uniqueID, x, y),
// 	}
// 	u.U.radius = C.ireal(int(rand.Int31n(int32(UnitRadius/2))) + UnitRadius/2)

// 	runtime.SetFinalizer(u, (*Unit).Free)
// 	return u
// }

// // Free 释放 aoi 单位
// func (u *Unit) Free() {
// 	if u.U == nil {
// 		return
// 	}
// 	C.ifreeunit(u.U)
// 	u.U = nil
// }

// func (u *Unit) String() string {
// 	return fmt.Sprintf("<Unit>(uid:%v pos:{%v,%v} radius:%v)", u.U.id, u.U.pos.x, u.U.pos.y, u.U.radius)
// }

// // ------------------ Aoi map ------------------

// // Map struct
// type Map struct {
// 	M *C.struct_imap

// 	Units map[int64]*Unit
// }

// // NewAoiMap new map divide 几级分割
// func NewAoiMap(p *C.struct_ipos, s *C.struct_isize, divide int) *Map {
// 	m := &Map{
// 		M:     C.imapmake(p, s, C.int(divide)),
// 		Units: make(map[int64]*Unit),
// 	}
// 	runtime.SetFinalizer(m, (*Map).Free)
// 	return m
// }

// // Print 打印整个地图
// func (m *Map) Print(require int) {
// 	C._aoi_print(m.M, C.int(require))
// }

// // AddUnit 添加单位
// func (m *Map) AddUnit(u *Unit) bool {
// 	id := int64(u.U.id)
// 	if _, ok := m.Units[id]; ok {
// 		return false
// 	}
// 	C.imapaddunit(m.M, u.U)
// 	m.Units[id] = u
// 	return true
// }

// // RemoveUnit 移除单位
// func (m *Map) RemoveUnit(u *Unit) bool {
// 	id := int64(u.U.id)
// 	return m.RemoveUnitByID(id)
// }

// // RemoveUnitByID 根据Id移除单位
// func (m *Map) RemoveUnitByID(id int64) bool {
// 	if u, ok := m.Units[id]; ok {
// 		C.imapremoveunit(m.M, u.U)
// 		delete(m.Units, id)
// 		return true
// 	}
// 	return false
// }

// // Search 搜索x,y位置上，半径为radius的所有元素
// func (m *Map) Search(result *SearchResult, x, y, radius float64) int {
// 	// result.MarkUnits(ColorUnit)

// 	pos := C.struct_ipos{C.ireal(x), C.ireal(y)}
// 	C.imapsearchfrompos(m.M, &pos, result.S, C.ireal(radius))
// 	result.M = m
// 	result.X = x
// 	result.Y = y
// 	result.Radius = radius
// 	// result.Mark

// 	return result.Len()
// }

// // UpdateUnit unit位置有变化后，进行更新
// func (m *Map) UpdateUnit(u *Unit) {
// 	C.imapupdateunit(m.M, u.U)
// }

// // Free 释放
// func (m *Map) Free() {
// 	if m.M == nil {
// 		return
// 	}
// 	C.imapfree(m.M)
// 	m.M = nil
// 	m.Units = nil
// }

// // ------------------ Aoi Search Result ------------------

// // SearchResult 搜索结果
// type SearchResult struct {
// 	S *C.struct_isearchresult

// 	X      float64
// 	Y      float64
// 	Radius float64

// 	M *Map
// }

// // NewSearchResult 构造搜索结果
// func NewSearchResult() *SearchResult {
// 	s := &SearchResult{S: C.isearchresultmake()}
// 	runtime.SetFinalizer(s, (*SearchResult).Free)
// 	return s
// }

// // Free 释放搜索结果
// func (s *SearchResult) Free() {
// 	if s.S == nil {
// 		return
// 	}

// 	C.isearchresultfree(s.S)
// 	s.S = nil
// 	s.M = nil
// }

// // Clean 清楚掉搜索结果里的内容
// func (s *SearchResult) Clean() {
// 	C.isearchresultclean(s.S)
// }

// // Units 返回搜索结果中的各个单位
// func (s *SearchResult) Units() []*Unit {
// 	units := []*Unit{}
// 	for first := C.ireflistfirst(s.S.units); first != nil; first = first.next {
// 		u := (*C.struct_iunit)(unsafe.Pointer(first.value))
// 		units = append(units, s.M.Units[(int64)(u.id)])
// 	}
// 	return units
// }

// // Len 结果数量
// func (s *SearchResult) Len() int {
// 	return int(C.ireflistlen(s.S.units))
// }

// type _BindManager struct {
// 	xmap   *Map
// 	search *SearchResult

// 	idx uint32
// }

// // _NewBindManager 新建aoi地图
// func _NewBindManager(posX int32, posZ int32, width float32, height float32, divide int) *_BindManager {
// 	bm := &_BindManager{}

// 	pos := C.struct_ipos{
// 		x: C.ireal(posX),
// 		y: C.ireal(posZ),
// 	}
// 	size := C.struct_isize{
// 		w: C.ireal(width),
// 		h: C.ireal(height),
// 	}

// 	bm.xmap = NewAoiMap(&pos, &size, divide)
// 	bm.search = NewSearchResult()

// 	return bm
// }

// // Enter 实体进入地图
// func (bm *_BindManager) Enter(aoi *Unit) {
// 	bm.xmap.AddUnit(aoi)
// }

// // Leave 实体离开地图
// func (bm *_BindManager) Leave(aoi *Unit) {
// 	bm.xmap.RemoveUnit(aoi)
// }

// // Moved 实体移动
// func (bm *_BindManager) Moved(aoi *Unit, x, y Coord) {

// 	aoi.U.pos.x += C.ireal(x)
// 	aoi.U.pos.y += C.ireal(y)

// 	bm.xmap.UpdateUnit(aoi)

// }

// func (bm *_BindManager) Ajust(aoi *Unit) {
// 	bm.xmap.Search(bm.search, float64(aoi.U.pos.x), float64(aoi.U.pos.y), 100.0)
// }
