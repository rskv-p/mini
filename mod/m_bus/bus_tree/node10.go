// file:mini/pkg/x_tree/node10.go
package bus_tree

//---------------------
// Node10 (0–9 children)
//---------------------

// node10 is optimized for numeric pivots (0–9).
type node10 struct {
	child [10]node
	meta
	key [10]byte
}

// newNode10 creates a new node10 with given prefix.
func newNode10(prefix []byte) *node10 {
	nn := &node10{}
	nn.setPrefix(prefix)
	return nn
}

//---------------------
// Node Interface Impl
//---------------------

func (n *node10) isFull() bool { return n.size >= 10 }

func (n *node10) grow() node {
	nn := newNode16(n.prefix)
	for i := 0; i < 10; i++ {
		nn.addChild(n.key[i], n.child[i])
	}
	return nn
}

func (n *node10) shrink() node {
	if n.size > 4 {
		return nil
	}
	nn := newNode4(nil)
	for i := uint16(0); i < n.size; i++ {
		nn.addChild(n.key[i], n.child[i])
	}
	return nn
}

func (n *node10) addChild(c byte, nn node) {
	if n.size >= 10 {
		panic("node10 full")
	}
	n.key[n.size] = c
	n.child[n.size] = nn
	n.size++
}

func (n *node10) findChild(c byte) *node {
	for i := uint16(0); i < n.size; i++ {
		if n.key[i] == c {
			return &n.child[i]
		}
	}
	return nil
}

func (n *node10) deleteChild(c byte) {
	for i, last := uint16(0), n.size-1; i < n.size; i++ {
		if n.key[i] == c {
			if i < last {
				n.key[i] = n.key[last]
				n.child[i] = n.child[last]
			}
			n.key[last] = 0
			n.child[last] = nil
			n.size--
			return
		}
	}
}

func (n *node10) iter(f func(node) bool) {
	for i := uint16(0); i < n.size; i++ {
		if !f(n.child[i]) {
			return
		}
	}
}

func (n *node10) children() []node {
	return n.child[:n.size]
}
