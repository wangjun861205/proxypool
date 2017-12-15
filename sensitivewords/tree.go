package sensitivewords

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

var Root = &Letter{Value: "", Word: "", Parent: nil, Children: make([]*Letter, 0, 1024)}

func generateTree() {
	fs, err := ioutil.ReadDir("./originwords")
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, f := range fs {
		rf, err := os.Open("./originwords/" + f.Name())
		if err != nil {
			fmt.Println(err)
			return
		}
		defer rf.Close()
		b, err := ioutil.ReadAll(rf)
		if err != nil {
			fmt.Println(err)
			return
		}
		s := string(b[:])
		s = strings.Replace(s, "\r", "", -1)
		words := strings.Split(s, "\n")
		for _, word := range words {
			currentNode := Root
			for _, l := range word {
				if nl := currentNode.next(string(l)); nl != nil {
					currentNode = nl
				} else {
					newLetter := &Letter{Value: string(l), Word: "", Parent: currentNode, Children: make([]*Letter, 0, 1024)}
					currentNode.Children = append(currentNode.Children, newLetter)
					currentNode = newLetter
				}
			}
			currentNode.Word = word
		}
	}
}

func Match(source string) (string, bool) {
	currentNode := Root
	for _, l := range source {
		if nl := currentNode.next(string(l)); nl != nil {
			if nl.Word == "" {
				currentNode = nl
			} else {
				return nl.Word, true
			}
		} else {
			return "", false
		}
	}
	return "", false
}

func Search(source string) []string {
	var wg sync.WaitGroup
	source = processText(source)
	sourceLen := len(source)
	output := make(chan string)
	done := make(chan interface{})
	sensitiveWords := make([]string, 0, 64)
	for i, _ := range source {
		slice := string([]byte(source[i:sourceLen]))
		wg.Add(1)
		go func(s string) {
			if ss, ok := Match(s); ok {
				output <- ss
				wg.Done()
				return
			}
			wg.Done()
		}(slice)
	}
	go func() {
		wg.Wait()
		close(done)
	}()
	for {
		select {
		case <-done:
			return sensitiveWords
		case ss := <-output:
			sensitiveWords = append(sensitiveWords, ss)
		}
	}
}

func init() {
	generateTree()
	fmt.Println("init success")
}
