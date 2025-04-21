// file:mini/pkg/x_tree/parts.go
package bus_tree

import (
	"bytes"
)

//---------------------
// Filter Split & Match
//---------------------

// genParts splits a filter into wildcard-aware segments.
func genParts(filter []byte, parts [][]byte) [][]byte {
	var start int
	for i, e := 0, len(filter)-1; i < len(filter); i++ {
		if filter[i] == tsep {
			// Handle '*'
			if i < e && filter[i+1] == pwc && (i+2 <= e && filter[i+2] == tsep || i+1 == e) {
				if i > start {
					parts = append(parts, filter[start:i+1])
				}
				parts = append(parts, filter[i+1:i+2])
				i++
				if i+2 <= e {
					i++
				}
				start = i + 1
			} else if i < e && filter[i+1] == fwc && i+1 == e {
				// Handle '>'
				if i > start {
					parts = append(parts, filter[start:i+1])
				}
				parts = append(parts, filter[i+1:i+2])
				i++
				start = i + 1
			}
		} else if filter[i] == pwc || filter[i] == fwc {
			// Wildcard must be at the beginning of token
			if prev := i - 1; prev >= 0 && filter[prev] != tsep {
				continue
			}
			// Wildcard must be at the end of token
			if next := i + 1; next == e || next < e && filter[next] != tsep {
				continue
			}
			parts = append(parts, filter[i:i+1])
			if i+1 <= e {
				i++
			}
			start = i + 1
		}
	}
	if start < len(filter) {
		if filter[start] == tsep {
			start++
		}
		parts = append(parts, filter[start:])
	}
	return parts
}

// matchParts checks if subject fragment matches the parts.
func matchParts(parts [][]byte, frag []byte) ([][]byte, bool) {
	lf := len(frag)
	if lf == 0 {
		return parts, true
	}

	var si int
	lpi := len(parts) - 1

	for i, part := range parts {
		if si >= lf {
			return parts[i:], true
		}
		lp := len(part)

		// Handle wildcards
		if lp == 1 {
			switch part[0] {
			case pwc:
				index := bytes.IndexByte(frag[si:], tsep)
				if index < 0 {
					if i == lpi {
						return nil, true
					}
					return parts[i:], true
				}
				si += index + 1
				continue
			case fwc:
				return nil, true
			}
		}

		end := min(si+lp, lf)
		if si+lp > end {
			part = part[:end-si]
		}
		if !bytes.Equal(part, frag[si:end]) {
			return parts, false
		}

		if end < lf {
			si = end
			continue
		}

		if end < si+lp {
			if end >= lf {
				parts = append([][]byte{}, parts...)
				parts[i] = parts[i][lf-si:]
			} else {
				i++
			}
			return parts[i:], true
		}

		if i == lpi {
			return nil, true
		}

		si += len(part)
	}
	return parts, false
}
