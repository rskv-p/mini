// file:mini/pkg/x_tree/leaf.go
package bus_tree

import (
	"bytes"
)

//---------------------
// Leaf Node
//---------------------

// leaf represents a terminal node in the subject tree.
type leaf[T any] struct {
	value  T
	suffix []byte // remaining suffix of subject
}

// newLeaf creates a new leaf node.
func newLeaf[T any](suffix []byte, value T) *leaf[T] {
	return &leaf[T]{value, copyBytes(suffix)}
}

//---------------------
// Interface Implementation
//---------------------

func (n *leaf[T]) isLeaf() bool                               { return true }
func (n *leaf[T]) base() *meta                                { return nil }
func (n *leaf[T]) isFull() bool                               { return true }
func (n *leaf[T]) children() []node                           { return nil }
func (n *leaf[T]) numChildren() uint16                        { return 0 }
func (n *leaf[T]) iter(f func(node) bool)                     {}
func (n *leaf[T]) path() []byte                               { return n.suffix }
func (n *leaf[T]) match(subject []byte) bool                  { return bytes.Equal(subject, n.suffix) }
func (n *leaf[T]) setSuffix(suffix []byte)                    { n.suffix = copyBytes(suffix) }
func (n *leaf[T]) matchParts(parts [][]byte) ([][]byte, bool) { return matchParts(parts, n.suffix) }

//---------------------
// Unsupported Operations
//---------------------

func (n *leaf[T]) setPrefix(pre []byte)    { panic("setPrefix called on leaf") }
func (n *leaf[T]) addChild(_ byte, _ node) { panic("addChild called on leaf") }
func (n *leaf[T]) findChild(_ byte) *node  { panic("findChild called on leaf") }
func (n *leaf[T]) grow() node              { panic("grow called on leaf") }
func (n *leaf[T]) deleteChild(_ byte)      { panic("deleteChild called on leaf") }
func (n *leaf[T]) shrink() node            { panic("shrink called on leaf") }
