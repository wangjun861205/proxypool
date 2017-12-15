package sensitivewords

import (
	"fmt"
)

type Letter struct {
	Value    string
	Word     string
	Parent   *Letter
	Children []*Letter
}

func (l *Letter) next(s string) *Letter {
	for _, nl := range l.Children {
		if nl.Value == s {
			return nl
		}
	}
	return nil
}

func (l *Letter) Iterate(f func(le *Letter)) {
	for _, child := range l.Children {
		f(child)
		child.Iterate(f)
	}
}

func (l *Letter) String() string {
	return fmt.Sprintf("Value: %s, Word: %s, Parent: %s, Children: %v", l.Value, l.Word, l.Parent, l.Children)
}
