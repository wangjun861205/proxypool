package list

import (
	"fmt"
	"testing"
)

func TestList(t *testing.T) {
	l, err := FromSlice([]int{1, 2, 3})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(len(*l))
}
