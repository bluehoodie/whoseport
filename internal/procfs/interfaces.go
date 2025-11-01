// Package procfs provides /proc filesystem parsing and process enhancement.
package procfs

import "github.com/bluehoodie/whoseport/internal/model"

// Enhancer enhances ProcessInfo with data from /proc filesystem.
type Enhancer interface {
	Enhance(info *model.ProcessInfo) error
}
