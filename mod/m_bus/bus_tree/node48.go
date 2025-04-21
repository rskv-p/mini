// file:mini/pkg/x_tree/node48.go
package bus_tree

//---------------------
// Node48 (up to 48 children)
//---------------------

// node48 uses 1-indexed key lookup to save memory vs node256.
type node48 struct {
	child [48]node
	meta
	key [256]byte // 1-indexed: 0 = no entry
}

// newNode48 creates a new node48 with given prefix.
func newNode48(prefix []byte) *node48 {
	nn := &node48{}
	nn.setPrefix(prefix)
	return nn
}

//---------------------
// Node Interface Impl
//---------------------

func (n *node48) isFull() bool { return n.size >= 48 }

func (n *node48) grow() node {
	nn := newNode256(n.prefix)
	for c := 0; c < len(n.key); c++ {
		if i := n.key[byte(c)]; i > 0 {
			nn.addChild(byte(c), n.child[i-1])
		}
	}
	return nn
}

func (n *node48) shrink() node {
	if n.size > 16 {
		return nil
	}
	nn := newNode16(nil)
	for c := 0; c < len(n.key); c++ {
		if i := n.key[byte(c)]; i > 0 {
			nn.addChild(byte(c), n.child[i-1])
		}
	}
	return nn
}

func (n *node48) addChild(c byte, nn node) {
	if n.size >= 48 {
		panic("node48 full")
	}
	n.child[n.size] = nn
	n.key[c] = byte(n.size + 1) // 1-indexed
	n.size++
}

func (n *node48) findChild(c byte) *node {
	i := n.key[c]
	if i == 0 {
		return nil
	}
	return &n.child[i-1]
}

func (n *node48) deleteChild(c byte) {
	i := n.key[c]
	if i == 0 {
		return
	}
	i-- // to 0-based
	last := byte(n.size - 1)
	if i < last {
		n.child[i] = n.child[last]
		for ic := 0; ic < len(n.key); ic++ {
			if n.key[byte(ic)] == last+1 {
				n.key[byte(ic)] = i + 1
				break
			}
		}
	}
	n.child[last] = nil
	n.key[c] = 0
	n.size--
}

func (n *node48) iter(f func(node) bool) {
	for _, c := range n.child {
		if c != nil && !f(c) {
			return
		}
	}
}

func (n *node48) children() []node {
	return n.child[:n.size]
}
