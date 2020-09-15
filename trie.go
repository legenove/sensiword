package sensiword

import "sync/atomic"

var (
	RuneA = []rune("A")[0]
	RuneZ = []rune("Z")[0]
	Runea = []rune("a")[0]
)

type Trie struct {
	Root           *Node
	count          uint32
	lowerMatchCase bool // 大小写敏感, true 则全部都是小写匹配
}

func NewTrie(onlyLower bool) *Trie {
	return &Trie{
		Root:           NewRootNode(0),
		lowerMatchCase: onlyLower,
	}
}

func (trie *Trie) bfs() <-chan *Node {
	ch := make(chan *Node)
	go func() {
		queue := new(LinkList)
		for _, child := range trie.Root.Children {
			queue.Push(child)
		}

		for !queue.Empty() {
			n := queue.Pop().(*Node)
			ch <- n
			for _, child := range n.Children {
				queue.Push(child)
			}
		}
		close(ch)
	}()
	return ch
}

func (trie *Trie) Add(words ...string) {
	for _, word := range words {
		trie.addNode(word)
	}
}

func (trie *Trie) Del(words ...string) {
	for _, word := range words {
		trie.delNode(word)
	}
}

func (trie *Trie) Find(word string) bool {
	b, _, _ := trie.find([]rune(word))
	return b
}

func (trie *Trie) getRune(runeWord []rune, position int) rune {
	r := runeWord[position]
	if trie.lowerMatchCase == true && r >= RuneA && r <= RuneZ {
		r = r - (RuneA - Runea)
	}
	return r
}

// 根结点查找当前文本
func (trie *Trie) find(runeWord []rune) (bool, int, *Node) {
	var current = trie.Root
	for position := 0; position < len(runeWord); position++ {
		r := trie.getRune(runeWord, position)
		if next, ok := current.Children[r]; ok {
			current = next
		} else {
			return false, position, current
		}
		if position == len(runeWord)-1 {
			if current.IsPathEnd() {
				return true, position, current
			} else {
				return false, position, current
			}
		}
	}
	return false, 0, current
}

func (trie *Trie) addNode(word string) {
	var current = trie.Root
	var runeWord = []rune(word)
	for position := 0; position < len(runeWord); position++ {
		r := trie.getRune(runeWord, position)
		if next, ok := current.Children[r]; ok {
			current = next
		} else {
			count := atomic.AddUint32(&trie.count, 1)
			n := NewNode(r, count)
			n.depth = current.depth + 1
			n.Parent = current
			current.Children[r] = n
			current = n
		}
		if position == len(runeWord)-1 {
			current.isPathEnd = true
		}
	}
}

func (trie *Trie) delNode(word string) {
	var runeWord = []rune(word)
	// find the word
	has, _, current := trie.find(runeWord)
	if has {
		current.isPathEnd = false
		// 如果是页节点，则开始删除
		parent := current.Parent
		for current.IsLeafNode() || !current.isPathEnd {
			delete(parent.Children, current.Character)
			current = parent
			if current.IsRootNode() {
				break
			}
			parent = current.Parent
		}
	}
}

type Node struct {
	isRootNode bool
	isPathEnd  bool
	Character  rune
	Children   map[rune]*Node
	Parent     *Node
	depth      int
	ID         uint32
}

func (node *Node) IsLeafNode() bool {
	return len(node.Children) == 0
}

func (node *Node) IsRootNode() bool {
	return node.isRootNode
}

func (node *Node) IsPathEnd() bool {
	return node.isPathEnd
}

func NewRootNode(character rune) *Node {
	root := &Node{
		isRootNode: true,
		Character:  character,
		Children:   make(map[rune]*Node, 0),
		depth:      0,
		ID:         0,
	}

	return root
}

func NewNode(character rune, id uint32) *Node {
	return &Node{
		ID:        id,
		Character: character,
		Children:  make(map[rune]*Node, 0),
	}
}
