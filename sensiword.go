package sensiword

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

const (
	SensitiveStatusInit        int32 = iota // 初始化
	SensitiveStatusReady                    // 准备完成
	SensitiveStatusPartRebuild              // 局部重建中，增减敏感词后，仅重建ac
	SensitiveStatusRebuild                  // 整体重建中, 重建trie和ac
)

const (
	TrieRebuildDelta      = 50
	TrieRebuildEscapeTime = 10 * time.Microsecond
)

var (
	ErrSatausNotCompare = errors.New("status not compare")
)

type Sensitive struct {
	sync.RWMutex
	status    int32
	onlyLower bool // 大小写敏感，true则为不敏感，false则为敏感
	ac        *AC
	trie      *Trie
	version   string
}

func NewSensitive(onlyLower bool) *Sensitive {
	trie := NewTrie(onlyLower)
	ac := NewAC(trie)
	return &Sensitive{
		trie:      trie,
		ac:        ac,
		onlyLower: onlyLower,
	}
}

func (s *Sensitive) addWord(words ...string) {
	s.trie.Add(words...)
}

func (s *Sensitive) Init(version string, words ...string) {
	s.Lock()
	defer s.Unlock()
	s.addWord(words...)
	s.ac.BuildFailurePaths()
	s.version = version
	s.ready(s.status)
}

// Filter 过滤敏感词
func (s *Sensitive) Validate(text string) (bool, string) {
	s.RLock()
	defer s.RUnlock()
	const EMPTY = ""
	var (
		node  = s.trie.Root
		next  *Node
		runes = []rune(text)
	)

	for position := 0; position < len(runes); position++ {
		next = s.ac.next(node, runes[position])
		if next == nil {
			next = s.ac.fail(node, runes[position])
		}

		node = next
		if first := s.ac.firstOutput(node, runes, position); len(first) > 0 {
			return false, first
		}
	}

	return true, EMPTY
}

// Replace 和谐敏感词
func (s *Sensitive) Replace(text string, repl rune) string {
	s.RLock()
	defer s.RUnlock()
	var (
		node  = s.trie.Root
		next  *Node
		runes = []rune(text)
	)

	for position := 0; position < len(runes); position++ {
		next = s.ac.next(node, runes[position])
		if next == nil {
			next = s.ac.fail(node, runes[position])
		}

		node = next
		s.ac.replace(node, runes, position, repl)
	}
	return string(runes)
}

// FindIn 检测敏感词
func (s *Sensitive) FindIn(text string) (bool, string) {
	validated, first := s.Validate(text)
	return !validated, first
}

// FindAll 找到所有匹配词
func (s *Sensitive) FindAll(text string) []string {
	s.RLock()
	defer s.RUnlock()
	var (
		node  = s.trie.Root
		next  *Node
		runes = []rune(text)
	)
	var res = make([]string, 0)
	for position := 0; position < len(runes); position++ {
		next = s.ac.next(node, runes[position])
		if next == nil {
			next = s.ac.fail(node, runes[position])
		}

		node = next
		res = s.ac.output(node, runes, position, res)
	}
	return res
}

func (s *Sensitive) Version() string {
	s.RLock()
	defer s.RUnlock()
	return s.version
}

func (s *Sensitive) Status() int32 {
	s.RLock()
	defer s.RUnlock()
	return s.status
}

func (s *Sensitive) ready(old int32) bool {
	newV := SensitiveStatusReady
	return atomic.CompareAndSwapInt32(&s.status, old, newV)
}

func (s *Sensitive) partRebuild(old int32) bool {
	newV := SensitiveStatusPartRebuild
	return atomic.CompareAndSwapInt32(&s.status, old, newV)
}

func (s *Sensitive) rebuild(old int32) bool {
	newV := SensitiveStatusRebuild
	return atomic.CompareAndSwapInt32(&s.status, old, newV)
}

// 与完全重建时间只节省20%，修改trie树过程中有锁的竞争。
func (s *Sensitive) PartRebuild(version string, addWords []string, delWords []string) error {
	status := s.Status()
	if status == SensitiveStatusReady {
		s.Lock()
		if !s.partRebuild(status) {
			s.Unlock()
			return ErrSatausNotCompare
		}
		s.Unlock()
		wg := sync.WaitGroup{}
		addLen := len(addWords)
		delLen := len(delWords)
		for i := 0; i < delLen; i += TrieRebuildDelta {
			end := i + TrieRebuildDelta
			if end > delLen {
				end = delLen
			}
			s.Lock()
			s.trie.Del(delWords[i:end]...)
			s.Unlock()
			time.Sleep(TrieRebuildEscapeTime)
		}
		for i := 0; i < addLen; i += TrieRebuildDelta {
			end := i + TrieRebuildDelta
			if end > addLen {
				end = addLen
			}
			s.Lock()
			s.trie.Add(addWords[i:end]...)
			s.Unlock()
			time.Sleep(TrieRebuildEscapeTime)
		}
		// ac 重建
		ac := NewAC(s.trie)
		wg.Add(1)
		go func() {
			defer wg.Done()
			ac.BuildFailurePaths()
		}()
		wg.Wait()
		s.Lock()
		defer s.Unlock()
		s.ac = ac
		s.version = version
		s.ready(s.status)
	}
	return nil
}

func (s *Sensitive) Rebuild(version string, words ...string) error {
	status := s.Status()
	if status == SensitiveStatusReady {
		s.Lock()
		if !s.rebuild(status) {
			s.Unlock()
			return ErrSatausNotCompare
		}
		// 可以重建
		s.Unlock()
		wg := sync.WaitGroup{}
		// trie 重建
		trie := NewTrie(s.onlyLower)
		wg.Add(1)
		go func() {
			defer wg.Done()
			trie.Add(words...)
		}()
		wg.Wait()
		// ac 重建
		ac := NewAC(trie)
		wg.Add(1)
		go func() {
			defer wg.Done()
			ac.BuildFailurePaths()
		}()
		wg.Wait()
		s.Lock()
		defer s.Unlock()
		s.trie = trie
		s.ac = ac
		s.version = version
		s.ready(s.status)
	}
	return nil
}
