package sensiword

// ac
type AC struct {
	trie        *Trie
	FailurePath map[uint8]map[uint32]*Node
}

func NewAC(trie *Trie) *AC {
	return &AC{
		trie: trie,
		FailurePath: make(map[uint8]map[uint32]*Node),
	}
}

func (ac *AC) fail(node *Node, c rune) *Node {
	var next *Node
	for {
		next = ac.next(ac.getFailurePath(node), c)
		if next == nil {
			if node.IsRootNode() {
				return node
			}
			node = ac.getFailurePath(node)
			continue
		}
		return next
	}
}

func (ac *AC) getBucket(id uint32) map[uint32]*Node {
	if ac.FailurePath == nil {
		ac.FailurePath = make(map[uint8]map[uint32]*Node)
	}
	if _, ok := ac.FailurePath[uint8(id)]; !ok {
		ac.FailurePath[uint8(id)] = make(map[uint32]*Node, 0)
	}
	return ac.FailurePath[uint8(id)]
}

// BuildFailurePaths 更新Aho-Corasick的失败表
func (ac *AC) BuildFailurePaths() {
	for node := range ac.trie.bfs() {
		pointer := node.Parent
		var link *Node
		for link == nil {
			if pointer.IsRootNode() {
				link = pointer
				break
			}
			link = ac.getFailurePath(pointer).Children[node.Character]
			pointer = ac.getFailurePath(pointer)
		}
		if !link.IsRootNode() && link.ID != node.ID {
			ac.getBucket(node.ID)[node.ID] = link
		}
	}
}

func (ac *AC) getFailurePath(node *Node) *Node {
	if n, ok := ac.getBucket(node.ID)[node.ID]; ok {
		return n
	}
	return ac.trie.Root
}

func (ac *AC) next(node *Node, c rune) *Node {
	next, ok := node.Children[c]
	if ok {
		return next
	}
	return nil
}

func (ac *AC) output(node *Node, runes []rune, position int, results []string) []string {
	if node.IsRootNode() {
		return results
	}

	if node.IsPathEnd() {
		results = append(results, string(runes[position+1-node.depth:position+1]))
	}

	return ac.output(ac.getFailurePath(node), runes, position, results)
}

func (ac *AC) firstOutput(node *Node, runes []rune, position int) string {
	if node.IsRootNode() {
		return ""
	}

	if node.IsPathEnd() {
		return string(runes[position+1-node.depth : position+1])
	}

	return ac.firstOutput(ac.getFailurePath(node), runes, position)
}

func (ac *AC) replace(node *Node, runes []rune, position int, replace rune) {
	if node.IsRootNode() {
		return
	}

	if node.IsPathEnd() {
		for i := position + 1 - node.depth; i < position+1; i++ {
			runes[i] = replace
		}
	}
	ac.replace(ac.getFailurePath(node), runes, position, replace)
}
