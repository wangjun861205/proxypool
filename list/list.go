package list

import (
	"fmt"
	"reflect"
)

type List []interface{}

func FromSlice(s interface{}) (*List, error) {
	var v reflect.Value
	t := reflect.TypeOf(s).Kind()
	switch t {
	case reflect.Ptr:
		if ev := reflect.ValueOf(s).Elem(); ev.Type().Kind() == reflect.Slice {
			v = ev
		} else {
			return nil, fmt.Errorf("%v is not a valid slice pointer", s)
		}
	case reflect.Slice:
		v = reflect.ValueOf(s)
	default:
		return nil, fmt.Errorf("%v is not a valid slice pointer or a valid slice", s)
	}
	l := make(List, 0, 64)
	for i := 0; i < v.Len(); i++ {
		l = append(l, v.Index(i).Interface())
	}
	return &l, nil
}

func (l *List) Len() int {
	return len(*l)
}

func (l *List) Pop() (interface{}, bool) {
	if len(*l) == 0 {
		return nil, false
	}
	var element interface{}
	*l, element = []interface{}(*l)[:len(*l)-1], []interface{}(*l)[len(*l)-1]
	return element, true
}

func (l *List) LeftPop() (interface{}, bool) {
	if len(*l) == 0 {
		return nil, false
	}
	var element interface{}
	element, *l = []interface{}(*l)[0], []interface{}(*l)[1:]
	return element, true
}

func (l *List) Append(ele interface{}) {
	*l = append(*l, ele)
}

func (l *List) Remove(ele interface{}, num int) {
	indexList = make([]int, 0, 16)
	for i, element := range []interface{}(*l) {
		if ele == element {
			if len(indexList) < num || num == 0 {
				indexList = append(indexList, i)
			} else {
				break
			}
		} else {
			continue
		}
	}

}

func (l *List) Iterate(f func(e interface{})) {
	for _, ele := range []interface{}(*l) {
		f(ele)
	}
}
