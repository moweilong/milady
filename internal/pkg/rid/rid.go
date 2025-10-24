package rid

import (
	"github.com/moweilong/milady/pkg/id"
)

const defaultABC = "abcdefghijklmnopqrstuvwxyz1234567890"

type ResourceID string

const (
	// UserID defines the resource identifier for a user.
	UserID ResourceID = "user"
)

// String converts the resource identifier to a string.
func (rid ResourceID) String() string {
	return string(rid)
}

// New creates a unique identifier with a prefix.
func (rid ResourceID) New(counter uint64) string {
	// Generate a unique identifier using custom options.
	uniqueStr := id.NewCode(
		counter,
		id.WithCodeChars([]rune(defaultABC)),
		id.WithCodeL(6),
		id.WithCodeSalt(Salt()),
	)
	return rid.String() + "-" + uniqueStr
}

// NewResourceID 创建一个新的资源标识符
func NewResourceID(prefix string) ResourceID {
	return ResourceID(prefix)
}
