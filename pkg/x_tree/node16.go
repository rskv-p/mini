// file:mini/pkg/x_tree/node16.go
package x_tree

// Node16 (up to 16 children)
//---------------------

type node16 struct {
	child [16]node
	meta
	key [16]byte
}

// newNode16 creates a new node16 with the given prefix.
func newNode16(prefix []byte) *node16 {
	nn := &node16{}
	nn.setPrefix(prefix)
	return nn
}

//---------------------
// Node Interface Impl
//---------------------

func (n *node16) isFull() bool { return n.size >= 16 }

func (n *node16) grow() node {
	nn := newNode48(n.prefix)
	for i := 0; i < 16; i++ {
		nn.addChild(n.key[i], n.child[i])
	}
	return nn
}

func (n *node16) shrink() node {
	if n.size > 10 {
		return nil
	}
	nn := newNode10(nil)
	for i := uint16(0); i < n.size; i++ {
		nn.addChild(n.key[i], n.child[i])
	}
	return nn
}

func (n *node16) addChild(c byte, nn node) {
	if n.size >= 16 {
		panic("node16 full")
	}
	n.key[n.size] = c
	n.child[n.size] = nn
	n.size++
}

func (n *node16) findChild(c byte) *node {
	for i := uint16(0); i < n.size; i++ {
		if n.key[i] == c {
			return &n.child[i]
		}
	}
	return nil
}

func (n *node16) deleteChild(c byte) {
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

func (n *node16) iter(f func(node) bool) {
	for i := uint16(0); i < n.size; i++ {
		if !f(n.child[i]) {
			return
		}
	}
}

func (n *node16) children() []node {
	return n.child[:n.size]
}
