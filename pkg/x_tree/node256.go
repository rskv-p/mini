// file:mini/pkg/x_tree/node256.go
package x_tree

//---------------------
// Node256 (up to 256 children)
//---------------------

type node256 struct {
	child [256]node
	meta
}

// newNode256 creates a new node256 with given prefix.
func newNode256(prefix []byte) *node256 {
	nn := &node256{}
	nn.setPrefix(prefix)
	return nn
}

//---------------------
// Node Interface Impl
//---------------------

func (n *node256) isFull() bool { return false }

func (n *node256) grow() node {
	panic("grow cannot be called on node256")
}

func (n *node256) shrink() node {
	if n.size > 48 {
		return nil
	}
	nn := newNode48(nil)
	for c, child := range n.child {
		if child != nil {
			nn.addChild(byte(c), child)
		}
	}
	return nn
}

func (n *node256) addChild(c byte, nn node) {
	n.child[c] = nn
	n.size++
}

func (n *node256) findChild(c byte) *node {
	if n.child[c] != nil {
		return &n.child[c]
	}
	return nil
}

func (n *node256) deleteChild(c byte) {
	if n.child[c] != nil {
		n.child[c] = nil
		n.size--
	}
}

func (n *node256) iter(f func(node) bool) {
	for i := 0; i < 256; i++ {
		if n.child[i] != nil {
			if !f(n.child[i]) {
				return
			}
		}
	}
}

func (n *node256) children() []node {
	return n.child[:256]
}
