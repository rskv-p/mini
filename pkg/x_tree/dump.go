// file:mini/pkg/x_tree/dump.go
package x_tree

import (
	"fmt"
	"io"
	"strings"
)

//---------------------
// Tree Dump (Debug)
//---------------------

// Dump writes a visual tree representation to writer.
func (t *SubjectTree[T]) Dump(w io.Writer) {
	t.dump(w, t.root, 0)
	fmt.Fprintln(w)
}

// dump writes a single node (recursive).
func (t *SubjectTree[T]) dump(w io.Writer, n node, depth int) {
	if n == nil {
		fmt.Fprintln(w, "EMPTY")
		return
	}
	if n.isLeaf() {
		leaf := n.(*leaf[T])
		fmt.Fprintf(w, "%s LEAF: Suffix: %q Value: %+v\n", dumpPre(depth), leaf.suffix, leaf.value)
		return
	}

	bn := n.base()
	fmt.Fprintf(w, "%s %s Prefix: %q\n", dumpPre(depth), n.kind(), bn.prefix)
	depth++
	n.iter(func(cn node) bool {
		t.dump(w, cn, depth)
		return true
	})
}

//---------------------
// Node Kind Labels
//---------------------

func (n *leaf[T]) kind() string { return "LEAF" }
func (n *node4) kind() string   { return "NODE4" }
func (n *node10) kind() string  { return "NODE10" }
func (n *node16) kind() string  { return "NODE16" }
func (n *node48) kind() string  { return "NODE48" }
func (n *node256) kind() string { return "NODE256" }

//---------------------
// Indentation Helper
//---------------------

func dumpPre(depth int) string {
	if depth == 0 {
		return "-- "
	}
	var b strings.Builder
	for i := 0; i < depth; i++ {
		b.WriteString("  ")
	}
	b.WriteString("|__ ")
	return b.String()
}
