package sensiword

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var sensiWordTestData = []string{
	"17大",
	"naive",
	"敏感词",
}

func TestNewTrie(t *testing.T) {
	tree := NewTrie(false)
	tree.Add(sensiWordTestData...)
	assert.Equal(t, tree.Find("17大"), true)
	assert.Equal(t, tree.Find("naive"), true)
	assert.Equal(t, tree.Find("NaIve"), false)
	assert.Equal(t, tree.Find("额外"), false)
	tree.Add("额外")
	assert.Equal(t, tree.Find("额外"), true)
}

func TestDelTrieNode(t *testing.T) {
	tree := NewTrie(false)
	tree.Add(sensiWordTestData...)
	assert.Equal(t, tree.Find("17大"), true)
	tree.Del("17")
	assert.Equal(t, tree.Find("17大"), true)
	tree.Del("17大大")
	assert.Equal(t, tree.Find("17大"), true)
	tree.Del("17大")
	assert.Equal(t, tree.Find("17大"), false)
}

func TestOnlyLowerTrieNode(t *testing.T) {
	tree := NewTrie(true)
	tree.Add(sensiWordTestData...)
	assert.Equal(t, tree.Find("naive"), true)
	assert.Equal(t, tree.Find("nAiVe"), true)
	tree.Del("nAiVe")
	assert.Equal(t, tree.Find("naive"), false)
	assert.Equal(t, tree.Find("nAiVe"), false)
}