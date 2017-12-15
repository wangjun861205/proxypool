package sensitivewords

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestMain(t *testing.T) {
	text := `先问一句话，你希望你家孩子以后做什么工作？江泽民跟技术打交道的技术类工作？还是跟人打交道的管理类工作？
如果你希望你家孩子做技术类工作，那您可以点右上角的红叉了。
如果你希望你家孩子做管理类工作，那终究要跟人打交道吧？
跟人打交道，那么必须要先了解人。不仅要了解人，而且要了解社会里各色人等。因为，社会是各种人组成的，而历史是人性的展开。不了解历史，如何了解人性？又如何了解人性之复杂？？
李敖曾向钱穆请教治学方法，他回答说并没有具体方法，要多读书、多求解，而且当以原文为主，免受他人成见的约束。
钱穆说，选书最好选已经有两三百年以上历史的书，这种书经两三百年犹未被淘汰，必有价值。新书则不然，新书有否价值，犹待考验也。
就问各位一句，就在你们刷的最多朋友圈里，有多少是三百年以上文？每天那么多心灵鸡汤成功学，都经过时间检验淘汰了吗？所以，大部分的时间，都是在做无用功也。。。
这就是为什么只能读纸质书，而且只读百年以上的书。。。`
	request := Request{Text: text}
	postJson, err := json.Marshal(request)
	if err != nil {
		fmt.Println("json marshal error:", err)
	}
	fmt.Println(string(postJson[:]))
	server := NewServer()
	go server.ListenAndServe()
	var wg sync.WaitGroup
	startTime := time.Now()
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			// resp, err := http.Post("http://127.0.0.1:8000/sensitive_filter", "application/json", bytes.NewBuffer(postJson))
			req, _ := http.NewRequest("POST", "http://127.0.0.1:8000/sensitive_filter", bytes.NewBuffer(postJson))
			// req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				fmt.Println("request error:", err)
				wg.Done()
				return
			}
			defer resp.Body.Close()
			var response Response
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("read body error:", err)
				wg.Done()
				return
			}
			err = json.Unmarshal(body, &response)
			if err != nil {
				fmt.Println("json unmarshal error:", err)
				wg.Done()
				return
			}
			fmt.Println(response.Status, response.ErrorDetail, response.Data)
			wg.Done()
		}()
	}
	wg.Wait()
	fmt.Printf("time: %f", time.Since(startTime).Seconds())
	server.Shutdown(context.Background())
}
