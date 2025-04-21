// file: mini/pkg/x_sub/sublist.go
package bus_sub

import (
	"bytes"
	"errors"
	"strings"

	"github.com/rskv-p/mini/mod/m_bus/bus_tree"
)

//---------------------
// Sublist
//---------------------

// Sublist manages subscriptions, including exact and wildcard matches.
type Sublist struct {
	tree          *bus_tree.SubjectTree[*Subscription] // Tree structure for subject subscriptions
	exactCache    map[string][]*Subscription           // Cache for exact matches
	wildcardCache map[string][]*Subscription           // Cache for wildcard matches
	cacheSize     int                                  // Cache size limit
}

// NewSublist creates a new Sublist with a specified cache size.
func NewSublist(cacheSize int) *Sublist {
	return &Sublist{
		tree:          bus_tree.NewSubjectTree[*Subscription](),
		exactCache:    make(map[string][]*Subscription),
		wildcardCache: make(map[string][]*Subscription),
		cacheSize:     cacheSize,
	}
}

// Insert adds a Subscription to the tree and updates the cache.
func (s *Sublist) Insert(sub *Subscription) {
	s.tree.Insert(sub.Subject, sub)
	key := string(sub.Subject)

	// Check if it's a wildcard and update the appropriate cache
	if isWildcard(key) {
		s.wildcardCache[key] = append(s.wildcardCache[key], sub)
		//	x_log.Debug("inserted wildcard Subscription", "subject", key)
	} else {
		s.exactCache[key] = append(s.exactCache[key], sub)

		// Evict cache entry if it exceeds the cache size limit
		if len(s.exactCache) > s.cacheSize {
			for k := range s.exactCache {
				delete(s.exactCache, k)
				//		x_log.Warn("evicted exact Subscription from cache", "subject", k)
				break
			}
		}

		//	x_log.Debug("inserted exact Subscription", "subject", key)
	}
}

// Remove deletes a Subscription from the tree.
func (s *Sublist) Remove(sub *Subscription) {
	key := string(sub.Subject)
	s.tree.Delete(sub.Subject)

	// Remove from the appropriate cache (exact or wildcard)
	if isWildcard(key) {
		delete(s.wildcardCache, key)
		//	x_log.Debug("removed wildcard Subscription", "subject", key)
	} else {
		delete(s.exactCache, key)
		//	x_log.Debug("removed exact Subscription", "subject", key)
	}
}

// SublistResult holds the subscriptions that match a subject.
type SublistResult struct {
	Psubs []*Subscription // List of matching subscriptions
}

// Match returns Subscriptions matching the given subject.
func (s *Sublist) Match(subject []byte) *SublistResult {
	key := string(subject)
	res := &SublistResult{}

	// Handle regular subjects
	if subs, ok := s.exactCache[key]; ok {
		//	x_log.Debug("exact cache hit", "subject", key)
		res.Psubs = append(res.Psubs, subs...)
	}

	// Use the tree to match subscriptions (e.g., with wildcards)
	s.tree.Match(subject, func(_ []byte, sub **Subscription) {
		if sub != nil && *sub != nil && !contains(res.Psubs, *sub) {
			res.Psubs = append(res.Psubs, *sub)
		}
	})

	//	x_log.Debug("tree matched Subscriptions", "count", len(res.Psubs), "subject", key)
	return res
}

// HasInterest checks if any subscription matches the subject.
func (s *Sublist) HasInterest(subject []byte) bool {
	found := false
	s.tree.Match(subject, func(_ []byte, _ **Subscription) {
		found = true
	})
	return found
}

//---------------------
// Helpers
//---------------------

// isWildcard checks if the subject contains wildcard characters ('*' or '>').
func isWildcard(subj string) bool {
	return bytes.ContainsAny([]byte(subj), ">*/")
}

// contains checks if a subscription is already in the list.
func contains(list []*Subscription, target *Subscription) bool {
	for _, s := range list {
		if s == target {
			return true
		}
	}
	return false
}

//---------------------
// Subject Transformation
//---------------------

// SubjectTransform handles the transformation of subjects based on specific rules.
type SubjectTransform struct {
	srcTokens  []string
	destTokens []string
}

// NewSubjectTransform creates a new transformer for converting src → dest.
func NewSubjectTransform(src, dest string) (*SubjectTransform, error) {
	srcTokens := strings.Split(src, ".")
	destTokens := strings.Split(dest, ".")

	// Check if the number of wildcards matches between source and destination
	if countWildcards(srcTokens) != countWildcards(destTokens) {
		//x_log.Warn("wildcard count mismatch in transform", "src", src, "dest", dest)
		return nil, errors.New("wildcard count mismatch between src and dest")
	}

	//	x_log.Debug("created subject transform", "src", src, "dest", dest)
	return &SubjectTransform{
		srcTokens:  srcTokens,
		destTokens: destTokens,
	}, nil
}

// TransformSubject applies the transformation to the subject based on the src → dest rule.
func (st *SubjectTransform) TransformSubject(subject string) (string, error) {
	inputTokens := strings.Split(subject, ".")
	mapping := make([]string, 0)
	i := 0

	// Apply the transformation rules to the source subject
	for _, token := range st.srcTokens {
		switch token {
		case "*":
			// Handle wildcard "*" by ensuring enough tokens remain
			if i >= len(inputTokens) {
				//		x_log.Error("subject too short for *", "subject", subject)
				return "", errors.New("subject too short for *")
			}
			mapping = append(mapping, inputTokens[i])
			i++
		case ">":
			// Handle wildcard ">" by adding all remaining tokens
			if i >= len(inputTokens) {
				//		x_log.Error("no tokens available for >", "subject", subject)
				return "", errors.New("no tokens available for >")
			}
			mapping = append(mapping, strings.Join(inputTokens[i:], "."))
			i = len(inputTokens) // Move the index to the end
		default:
			// Ensure the current token matches the expected token
			if i >= len(inputTokens) || token != inputTokens[i] {
				//	x_log.Error("subject does not match pattern", "expected", token, "got", inputTokens[i])
				return "", errors.New("subject does not match source pattern")
			}
			mapping = append(mapping, inputTokens[i])
			i++
		}
	}

	// Build the transformed result based on destTokens
	var result []string
	wcIndex := 0
	for _, token := range st.destTokens {
		if token == "*" || token == ">" {
			// If we encounter a wildcard, fill it with values from the mapping
			if wcIndex >= len(mapping) {
				//		x_log.Error("not enough wildcards for destination", "subject", subject)
				return "", errors.New("not enough wildcard values to fill destination")
			}
			result = append(result, mapping[wcIndex])
			wcIndex++
		} else {
			result = append(result, token)
		}
	}

	// Join the final transformed subject
	transformed := strings.Join(result, ".")
	//x_log.Debug("transformed subject", "input", subject, "output", transformed)
	return transformed, nil
}

//---------------------
// Helpers
//---------------------

// countWildcards counts the number of wildcard characters ('*' or '>') in the tokens.
func countWildcards(tokens []string) int {
	count := 0
	for _, t := range tokens {
		if t == "*" || t == ">" {
			count++
		}
	}
	return count
}
