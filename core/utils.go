package core

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	DefaultQueueGroup = "q"    // DefaultQueueGroup when no queue group is set.
	APIPrefix         = "$SRV" // APIPrefix for control verb subjects.
)

// Validation errors.
var (
	ErrConfigValidation    = errors.New("validation")               // Error for config validation failure
	ErrVerbNotSupported    = errors.New("unsupported verb")         // Error for unsupported verb
	ErrServiceNameRequired = errors.New("service name is required") // Error if service name is missing
)

// Regex patterns for validation.
var (
	semVerRegexp  = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`) // SemVer regex
	nameRegexp    = regexp.MustCompile(`^[A-Za-z0-9\-_]+$`)                                                                                                                                                                   // Service name regex
	subjectRegexp = regexp.MustCompile(`^[^ >]*[>]?$`)                                                                                                                                                                        // Subject format regex
)

// resolveQueueGroup determines the final queue group name.
func resolveQueueGroup(customQG, parentQG string, disabled, parentDisabled bool) (string, bool) {
	if disabled {
		return "", true // Return empty if disabled
	}
	if customQG != "" {
		return customQG, false // Use custom queue group if available
	}
	if parentDisabled {
		return "", true // Return empty if parent queue group is disabled
	}
	if parentQG != "" {
		return parentQG, false // Use parent queue group if available
	}
	return DefaultQueueGroup, false // Use default queue group if none is set
}

// ControlSubject returns a subject for the given verb, name, and id.
func ControlSubject(verb Verb, name, id string) (string, error) {
	verbStr := verb.String()
	if verbStr == "" {
		return "", fmt.Errorf("%w: %q", ErrVerbNotSupported, verbStr) // Return error if verb is invalid
	}
	if name == "" && id != "" {
		return "", ErrServiceNameRequired // Return error if service name is missing but id is provided
	}
	if name == "" && id == "" {
		return fmt.Sprintf("%s.%s", APIPrefix, verbStr), nil // Return subject for control verb
	}
	if id == "" {
		return fmt.Sprintf("%s.%s.%s", APIPrefix, verbStr, name), nil // Return subject with service name
	}
	return fmt.Sprintf("%s.%s.%s.%s", APIPrefix, verbStr, name, id), nil // Return subject with service name and id
}

// joinParts joins parts into a subject using '.'.
func joinParts(parts []string) string {
	return strings.Join(parts, ".") // Join subject parts
}

// valid validates required Config fields.
func (c *Config) valid() error {
	if !nameRegexp.MatchString(c.Name) {
		if c.Logger != nil {
			c.Logger.Errorw("invalid service name", "name", c.Name) // Log invalid service name
		}
		return fmt.Errorf("%w: invalid service name", ErrConfigValidation) // Return error for invalid service name
	}
	if !semVerRegexp.MatchString(c.Version) {
		if c.Logger != nil {
			c.Logger.Errorw("invalid version format", "version", c.Version) // Log invalid version format
		}
		return fmt.Errorf("%w: invalid version (expected SemVer)", ErrConfigValidation) // Return error for invalid version
	}
	if c.QueueGroup != "" && !subjectRegexp.MatchString(c.QueueGroup) {
		if c.Logger != nil {
			c.Logger.Errorw("invalid queue group", "queue_group", c.QueueGroup) // Log invalid queue group
		}
		return fmt.Errorf("%w: invalid queue group", ErrConfigValidation) // Return error for invalid queue group
	}
	return nil // Return nil if all validations pass
}
