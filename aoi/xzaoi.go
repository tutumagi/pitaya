package aoi

import (
	"fmt"
)

// xzaoi 四叉树结构
type xzaoi struct {
	aoi       *Item
	neighbors map[*xzaoi]struct{}
	xl        *xList
	zl        *zList

	xnode   *Node
	znode   *Node
	markVal int
}

func (xz xzaoi) String() string {
	return fmt.Sprintf("<xzaoi> aoi item data %v", xz.aoi.Data)
}

func (xz *xzaoi) xPrev() *xzaoi {
	if xz.xnode != nil && xz.xnode.Prev != nil {
		return xz.xnode.Prev.Data.(*xzaoi)
	}
	return nil
}

func (xz *xzaoi) xNext() *xzaoi {
	if xz.xnode != nil && xz.xnode.Next != nil {
		return xz.xnode.Next.Data.(*xzaoi)
	}
	return nil
}

func (xz *xzaoi) zPrev() *xzaoi {
	if xz.znode != nil && xz.znode.Prev != nil {
		return xz.znode.Prev.Data.(*xzaoi)
	}
	return nil
}
func (xz *xzaoi) zNext() *xzaoi {
	if xz.znode != nil && xz.znode.Next != nil {
		return xz.znode.Next.Data.(*xzaoi)
	}
	return nil
}
