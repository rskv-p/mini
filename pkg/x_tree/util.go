// file:mini/pkg/x_tree/util.go
package x_tree

// Subject Match Constants
//---------------------

const (
	pwc  = '*' // partial wildcard
	fwc  = '>' // full wildcard
	tsep = '.' // token separator
)

//---------------------
// Utilities
//---------------------

// commonPrefixLen returns length of common prefix.
func commonPrefixLen(s1, s2 []byte) int {
	limit := min(len(s1), len(s2))
	var i int
	for ; i < limit; i++ {
		if s1[i] != s2[i] {
			break
		}
	}
	return i
}

// copyBytes returns a new copy of the byte slice.
func copyBytes(src []byte) []byte {
	if len(src) == 0 {
		return nil
	}
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}

//---------------------
// Pivot Helpers
//---------------------

type position interface {
	int | uint16
}

const noPivot = byte(127) // special value for no pivot

// pivot returns subject[pos] or noPivot if out of bounds.
func pivot[N position](subject []byte, pos N) byte {
	if int(pos) >= len(subject) {
		return noPivot
	}
	return subject[pos]
}
