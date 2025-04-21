// file:mini/pkg/x_tree/node.go
package bus_tree

//---------------------
// Node Interface
//---------------------

// node represents a single tree node (leaf or internal).
type node interface {
	isLeaf() bool
	base() *meta
	setPrefix(pre []byte)
	addChild(c byte, n node)
	findChild(c byte) *node
	deleteChild(c byte)
	isFull() bool
	grow() node
	shrink() node
	matchParts(parts [][]byte) ([][]byte, bool)
	kind() string
	iter(f func(node) bool)
	children() []node
	numChildren() uint16
	path() []byte
}

//---------------------
// Node Metadata (Shared)
//---------------------

type meta struct {
	prefix []byte
	size   uint16
}

func (n *meta) isLeaf() bool         { return false }
func (n *meta) base() *meta          { return n }
func (n *meta) setPrefix(pre []byte) { n.prefix = append([]byte(nil), pre...) }
func (n *meta) numChildren() uint16  { return n.size }
func (n *meta) path() []byte         { return n.prefix }
func (n *meta) matchParts(parts [][]byte) ([][]byte, bool) {
	return matchParts(parts, n.prefix)
}
