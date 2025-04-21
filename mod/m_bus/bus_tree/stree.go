// file:mini/pkg/x_tree/stree.go
package bus_tree

import (
	"bytes"
	"slices"
)

// SubjectTree
//---------------------

// SubjectTree is an adaptive radix tree for storing subjects.
type SubjectTree[T any] struct {
	root node
	size int
}

// NewSubjectTree creates a new SubjectTree.
func NewSubjectTree[T any]() *SubjectTree[T] {
	return &SubjectTree[T]{}
}

// Size returns number of entries.
func (t *SubjectTree[T]) Size() int {
	if t == nil {
		return 0
	}
	return t.size
}

// Empty clears the tree.
func (t *SubjectTree[T]) Empty() *SubjectTree[T] {
	if t == nil {
		return NewSubjectTree[T]()
	}
	t.root, t.size = nil, 0
	return t
}

// Insert adds or replaces an entry.
func (t *SubjectTree[T]) Insert(subject []byte, value T) (*T, bool) {
	if t == nil || bytes.IndexByte(subject, noPivot) >= 0 {
		return nil, false
	}
	old, updated := t.insert(&t.root, subject, value, 0)
	if !updated {
		t.size++
	}
	return old, updated
}

// Find returns value for exact subject match.
func (t *SubjectTree[T]) Find(subject []byte) (*T, bool) {
	if t == nil {
		return nil, false
	}
	var si int
	for n := t.root; n != nil; {
		if n.isLeaf() {
			if ln := n.(*leaf[T]); ln.match(subject[si:]) {
				return &ln.value, true
			}
			return nil, false
		}
		if bn := n.base(); len(bn.prefix) > 0 {
			end := min(si+len(bn.prefix), len(subject))
			if !bytes.Equal(subject[si:end], bn.prefix) {
				return nil, false
			}
			si += len(bn.prefix)
		}
		if an := n.findChild(pivot(subject, si)); an != nil {
			n = *an
		} else {
			return nil, false
		}
	}
	return nil, false
}

// Delete removes subject and returns its value.
func (t *SubjectTree[T]) Delete(subject []byte) (*T, bool) {
	if t == nil {
		return nil, false
	}
	val, deleted := t.delete(&t.root, subject, 0)
	if deleted {
		t.size--
	}
	return val, deleted
}

// Match finds all values matching a wildcard pattern.
func (t *SubjectTree[T]) Match(filter []byte, cb func(subject []byte, val *T)) {
	if t == nil || t.root == nil || len(filter) == 0 || cb == nil {
		return
	}
	var raw [16][]byte
	parts := genParts(filter, raw[:0])
	var pre [256]byte
	t.match(t.root, parts, pre[:0], cb)
}

// IterOrdered traverses the tree in lexicographic order.
func (t *SubjectTree[T]) IterOrdered(cb func(subject []byte, val *T) bool) {
	if t == nil || t.root == nil {
		return
	}
	var pre [256]byte
	t.iter(t.root, pre[:0], true, cb)
}

// IterFast traverses the tree in internal order.
func (t *SubjectTree[T]) IterFast(cb func(subject []byte, val *T) bool) {
	if t == nil || t.root == nil {
		return
	}
	var pre [256]byte
	t.iter(t.root, pre[:0], false, cb)
}

//---------------------
// Internal
//---------------------

func (t *SubjectTree[T]) insert(np *node, subject []byte, value T, si int) (*T, bool) {
	n := *np
	if n == nil {
		*np = newLeaf(subject, value)
		return nil, false
	}
	if n.isLeaf() {
		ln := n.(*leaf[T])
		if ln.match(subject[si:]) {
			old := ln.value
			ln.value = value
			return &old, true
		}
		cpi := commonPrefixLen(ln.suffix, subject[si:])
		nn := newNode4(subject[si : si+cpi])
		ln.setSuffix(ln.suffix[cpi:])
		si += cpi
		if p := pivot(ln.suffix, 0); cpi > 0 && si < len(subject) && p == subject[si] {
			t.insert(np, subject, value, si)
			nn.addChild(p, *np)
		} else {
			nl := newLeaf(subject[si:], value)
			nn.addChild(pivot(nl.suffix, 0), nl)
			nn.addChild(pivot(ln.suffix, 0), ln)
		}
		*np = nn
		return nil, false
	}

	bn := n.base()
	if len(bn.prefix) > 0 {
		cpi := commonPrefixLen(bn.prefix, subject[si:])
		if pli := len(bn.prefix); cpi >= pli {
			si += pli
			if nn := n.findChild(pivot(subject, si)); nn != nil {
				return t.insert(nn, subject, value, si)
			}
			if n.isFull() {
				n = n.grow()
				*np = n
			}
			n.addChild(pivot(subject, si), newLeaf(subject[si:], value))
			return nil, false
		}
		prefix := subject[si : si+cpi]
		si += len(prefix)
		nn := newNode4(prefix)
		n.setPrefix(bn.prefix[cpi:])
		nn.addChild(pivot(bn.prefix[:], 0), n)
		nn.addChild(pivot(subject[si:], 0), newLeaf(subject[si:], value))
		*np = nn
	} else {
		if nn := n.findChild(pivot(subject, si)); nn != nil {
			return t.insert(nn, subject, value, si)
		}
		if n.isFull() {
			n = n.grow()
			*np = n
		}
		n.addChild(pivot(subject, si), newLeaf(subject[si:], value))
	}
	return nil, false
}

func (t *SubjectTree[T]) delete(np *node, subject []byte, si int) (*T, bool) {
	if t == nil || np == nil || *np == nil || len(subject) == 0 {
		return nil, false
	}
	n := *np
	if n.isLeaf() {
		ln := n.(*leaf[T])
		if ln.match(subject[si:]) {
			*np = nil
			return &ln.value, true
		}
		return nil, false
	}
	if bn := n.base(); len(bn.prefix) > 0 {
		if !bytes.Equal(subject[si:si+len(bn.prefix)], bn.prefix) {
			return nil, false
		}
		si += len(bn.prefix)
	}
	p := pivot(subject, si)
	nna := n.findChild(p)
	if nna == nil {
		return nil, false
	}
	nn := *nna
	if nn.isLeaf() {
		ln := nn.(*leaf[T])
		if ln.match(subject[si:]) {
			n.deleteChild(p)
			if sn := n.shrink(); sn != nil {
				bn := n.base()
				pre := bn.prefix[:len(bn.prefix):len(bn.prefix)]
				if sn.isLeaf() {
					ln := sn.(*leaf[T])
					ln.suffix = append(pre, ln.suffix...)
				} else {
					if len(pre) > 0 {
						bsn := sn.base()
						sn.setPrefix(append(pre, bsn.prefix...))
					}
				}
				*np = sn
			}
			return &ln.value, true
		}
		return nil, false
	}
	return t.delete(nna, subject, si)
}

func (t *SubjectTree[T]) match(n node, parts [][]byte, pre []byte, cb func(subject []byte, val *T)) {
	var hasFWC bool
	if lp := len(parts); lp > 0 && len(parts[lp-1]) > 0 && parts[lp-1][0] == fwc {
		hasFWC = true
	}
	for n != nil {
		nparts, matched := n.matchParts(parts)
		if !matched {
			return
		}
		if n.isLeaf() {
			if len(nparts) == 0 || (hasFWC && len(nparts) == 1) {
				ln := n.(*leaf[T])
				cb(append(pre, ln.suffix...), &ln.value)
			}
			return
		}
		bn := n.base()
		if len(bn.prefix) > 0 {
			pre = append(pre, bn.prefix...)
		}
		if len(nparts) == 0 && !hasFWC {
			var hasTermPWC bool
			if lp := len(parts); lp > 0 && len(parts[lp-1]) == 1 && parts[lp-1][0] == pwc {
				nparts = parts[len(parts)-1:]
				hasTermPWC = true
			}
			for _, cn := range n.children() {
				if cn == nil {
					continue
				}
				if cn.isLeaf() {
					ln := cn.(*leaf[T])
					if len(ln.suffix) == 0 {
						cb(append(pre, ln.suffix...), &ln.value)
					} else if hasTermPWC && bytes.IndexByte(ln.suffix, tsep) < 0 {
						cb(append(pre, ln.suffix...), &ln.value)
					}
				} else if hasTermPWC {
					t.match(cn, nparts, pre, cb)
				}
			}
			return
		}
		if hasFWC && len(nparts) == 0 {
			nparts = parts[len(parts)-1:]
		}
		fp := nparts[0]
		p := pivot(fp, 0)
		if len(fp) == 1 && (p == pwc || p == fwc) {
			for _, cn := range n.children() {
				if cn != nil {
					t.match(cn, nparts, pre, cb)
				}
			}
			return
		}
		nn := n.findChild(p)
		if nn == nil {
			return
		}
		n, parts = *nn, nparts
	}
}

func (t *SubjectTree[T]) iter(n node, pre []byte, ordered bool, cb func(subject []byte, val *T) bool) bool {
	if n.isLeaf() {
		ln := n.(*leaf[T])
		return cb(append(pre, ln.suffix...), &ln.value)
	}
	bn := n.base()
	pre = append(pre, bn.prefix...)
	if !ordered {
		for _, cn := range n.children() {
			if cn != nil && !t.iter(cn, pre, false, cb) {
				return false
			}
		}
		return true
	}
	var _nodes [256]node
	nodes := _nodes[:0]
	for _, cn := range n.children() {
		if cn != nil {
			nodes = append(nodes, cn)
		}
	}
	slices.SortStableFunc(nodes, func(a, b node) int {
		return bytes.Compare(a.path(), b.path())
	})
	for _, cn := range nodes {
		if !t.iter(cn, pre, true, cb) {
			return false
		}
	}
	return true
}
