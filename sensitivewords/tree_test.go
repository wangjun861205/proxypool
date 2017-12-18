package sensitivewords

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestTree(t *testing.T) {
	f, err := os.Open("./originwords/sw1.txt")
	if err != nil {
		fmt.Println("open file error:", err)
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Println("read file error:", err)
	}
	s := strings.Replace(string(b[:]), "\r", "", -1)
	wordSlice := strings.Split(s, "\n")
	tree := NewTree(wordSlice)
	text := `先问一句话，你希望你家孩子以后做什么工作？跟技术打交道的技术类工作？还是跟人打交道的管理类工作？
如果你希望你家孩子做技术类工作，那您可以点右上角的红叉了。
如果你希望你家孩子做管理类工作，那终究要跟人打交道吧？
跟人打交道，那么必须要先了解人。不仅要了解人，而且要了解社会里各色人等。因为，社会是各种人组成的，而历史是人性的展开。不了解历史，如何了解人性？又如何了解人性之复杂？？
李敖曾向钱穆请教治学方法，他回答说并没有协警具体方法，要多读书、多求解，而且当以原文为主，免受他人成见的约束。
钱穆说，选书最好选已经有两三百年以上历史的书，这种书经两三百年犹未被淘汰，必有价值。新书则不然，新书有否价值，犹待考验也。
就问各位一句，就在你们刷的最多朋友圈里，有多少是三百年以上文？每天那么多心灵鸡汤成功学，都经过时间检验淘汰了吗？所以，大部分的时间，都是在做无用功也。。。
这就是为什么只能读纸质书，而且只读百年以上的书。。。打倒江泽民`
	w := tree.search(text)
	fmt.Println(w)
}
