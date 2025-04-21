// file:mini/pkg/x_tree/node4.go
package bus_tree

//---------------------
// Node4 (up to 4 children)
//---------------------

type node4 struct {
	child [4]node // child pointers
	meta          // prefix + size
	key   [4]byte // corresponding child keys
}

// newNode4 creates a new node4 with the given prefix.
func newNode4(prefix []byte) *node4 {
	nn := &node4{}
	nn.setPrefix(prefix)
	return nn
}

//---------------------
// Node Interface Impl
//---------------------

func (n *node4) isFull() bool { return n.size >= 4 }

func (n *node4) grow() node {
	nn := newNode10(n.prefix)
	for i := 0; i < 4; i++ {
		nn.addChild(n.key[i], n.child[i])
	}
	return nn
}

func (n *node4) shrink() node {
	if n.size == 1 {
		return n.child[0]
	}
	return nil
}

func (n *node4) addChild(c byte, nn node) {
	if n.size >= 4 {
		panic("node4 full")
	}
	n.key[n.size] = c
	n.child[n.size] = nn
	n.size++
}

func (n *node4) findChild(c byte) *node {
	for i := uint16(0); i < n.size; i++ {
		if n.key[i] == c {
			return &n.child[i]
		}
	}
	return nil
}

func (n *node4) deleteChild(c byte) {
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

func (n *node4) iter(f func(node) bool) {
	for i := uint16(0); i < n.size; i++ {
		if !f(n.child[i]) {
			return
		}
	}
}

func (n *node4) children() []node {
	return n.child[:n.size]
}
