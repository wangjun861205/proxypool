package sensitivewords

import (
	"sync"
)

//=========first vision=========
// type Letter struct {
// 	Value string
// 	Word     string
// 	Parent *Letter
// 	Children []*Letter
// }

type Letter struct {
	Value string
	Words map[string]bool
	// Parents  map[*Letter]bool
	Children map[string]*Letter
}

func NewLetter(s string) *Letter {
	return &Letter{
		Value: s,
		Words: make(map[string]bool),
		// Parents:  make(map[*Letter]bool),
		Children: make(map[string]*Letter),
	}
}

//=========first vision=========
// func (l *Letter) next(s string) *Letter {
// 	for _, nl := range l.Children {
// 		if nl.Value == s {
// 			return nl
// 		}
// 	}
// 	return nil
// }

func (l *Letter) next(s string) *Letter {
	if nl, ok := l.Children[s]; ok {
		return nl
	}
	return nil
}

// func (l *Letter) addParent(pl *Letter) {
// 	l.Parents[pl] = true
// }

func (l *Letter) addChild(cl *Letter) {
	l.Children[cl.Value] = cl
}

func (l *Letter) addWord(w string) {
	l.Words[w] = true
}

// func (l *Letter) getOrAddChildren(s string) *Letter {
// 	if nl := l.next(s); nl == nil {
// 		newLetter := NewLetter(s)
// 		newLetter.addParent(l)
// 		l.addChild(newLetter)
// 		return newLetter
// 	} else {
// 		return nl
// 	}
// }

// func (l *Letter) Iterate(f func(le *Letter)) {
// 	for _, child := range l.Children {
// 		f(child)
// 		child.Iterate(f)
// 	}
// }

//=========first vision=========
// func (l *Letter) String() string {
// 	return fmt.Sprintf("Value: %s, Word: %s, Parent: %s, Children: %v", l.Value, l.Word, l.Parent, l.Children)
// }

type Tree struct {
	Root       *Letter
	LetterPool map[string]*Letter
}

func (tree *Tree) getOrNewLetter(s string) *Letter {
	if l, ok := tree.LetterPool[s]; ok {
		return l
	}
	newLetter := NewLetter(s)
	tree.LetterPool[s] = newLetter
	return newLetter
}

func (tree *Tree) addWord(w string) {
	currentNode := tree.Root
	for _, l := range w {
		newLetter := tree.getOrNewLetter(string(l))
		// newLetter.addParent(currentNode)
		currentNode.addChild(newLetter)
		currentNode = newLetter
	}
	currentNode.addWord(w)
}

func NewTree(wordSlice []string) *Tree {
	tree := &Tree{
		Root:       NewLetter(""),
		LetterPool: make(map[string]*Letter),
	}
	for _, w := range wordSlice {
		tree.addWord(w)
	}
	return tree
}

func (tree *Tree) match(s string) (string, bool) {
	// matchedLetters := make([]string, 0, 64)
	lastIndex := len(s)
	currentNode := tree.Root
	for i, l := range s {
		if nextNode := currentNode.next(string(l)); nextNode != nil {
			// matchedLetters = append(matchedLetters, string(l))
			currentNode = nextNode
			continue
		}
		lastIndex = i
		break
	}
	// if len(matchedLetters) == 0 {
	// 	return "", false
	// }
	if lastIndex == 0 {
		return "", false
	}
	// matchedWord := strings.Join(matchedLetters, "")
	matchedWord := s[:lastIndex]
	if _, ok := currentNode.Words[matchedWord]; ok {
		return matchedWord, true
	}
	return "", false
}

func (tree *Tree) search(s string) []string {
	results := make([]string, 0, 64)
	doneChan := make(chan interface{})
	resultChan := make(chan string)
	var wg sync.WaitGroup
	runeSlice := []rune(s)
	for i, _ := range runeSlice {
		wg.Add(1)
		go func(startIndex int) {
			if sw, ok := tree.match(string(runeSlice[startIndex:])); ok {
				resultChan <- sw
				wg.Done()
				return
			}
			wg.Done()
			return
		}(i)
	}
	go func() {
		wg.Wait()
		close(doneChan)
	}()
OUTER:
	for {
		select {
		case sw := <-resultChan:
			results = append(results, sw)
		case <-doneChan:
			close(resultChan)
			break OUTER
		}
	}
	return results
}
