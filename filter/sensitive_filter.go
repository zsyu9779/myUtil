package util

import (
	"sync"
	"sync/atomic"
)

const InvalidWords = " ,~,!,@,#,$,%,^,&,*,(,),_,-,+,=,?,<,>,.,—,，,。,/,\\,|,《,》,？,;,:,：,',‘,；,“,"

var invalidMap = make(map[rune]struct{})

func init() {
	for _, v := range InvalidWords {
		invalidMap[v] = struct{}{}
	}
}

// Trie树节点
type TrieNode struct {
	// 当前节点的值
	value rune
	// 当前节点是否为敏感词结尾
	isEnd bool
	// 子节点
	children map[rune]*TrieNode
	// 失败指针
	fail *TrieNode
}

// 构建Trie树
func BuildTrieTree(words []string) *TrieNode {
	root := &TrieNode{
		value:    '/',
		isEnd:    false,
		children: make(map[rune]*TrieNode),
	}
	for _, word := range words {
		cur := root
		for _, w := range word {
			if _, ok := cur.children[w]; !ok {
				cur.children[w] = &TrieNode{
					value:    w,
					isEnd:    false,
					children: make(map[rune]*TrieNode),
				}
			}
			cur = cur.children[w]
		}
		cur.isEnd = true
	}
	return root
}

// 构建失败指针
func buildFailPoint(root *TrieNode) {
	queue := make([]*TrieNode, 0)
	queue = append(queue, root)
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		// 失败指针指向的节点锁代表的单词 是当前节点所代表单词的最大后缀
		for _, child := range cur.children {
			if cur == root {
				// 根节点的子节点失败指针指向根节点
				child.fail = root
			} else {
				// 某节点的失败指针一定是其父节点的失败指针指向的节点的子节点 且值相同
				fail := cur.fail
				for fail != nil {
					//找到某节点的父节点的失败指针指向的节点的与当前节点值相同的子节点
					if _, ok := fail.children[child.value]; ok {
						//当前节点的失败指针指向该节点
						child.fail = fail.children[child.value]
						break
					}
					//如果找不到 则继续向上找
					fail = fail.fail
				}
				//代表找到头了 也没找到
				if fail == nil {
					child.fail = root
				}
			}
			queue = append(queue, child)
		}
	}
}

type SensitiveFilter interface {
	build(words []string)
	// Refresh 刷新词库
	Refresh(words []string)
	// Filter 过滤敏感词
	Filter(text string) string
	// FindIn 过滤敏感词
	FindIn(text string) []string
	// Contains 判断是否包含敏感词
	Contains(text string) bool
}

// ACFilter AC算法敏感词过滤器
type ACFilter struct {
	// 敏感词库 根节点
	root *TrieNode
	// 用于刷新的敏感词库 防止构建敏感词库时影响敏感词过滤
	refreshRoot *TrieNode
	// 读写锁 用于更新敏感词库
	rwLock *sync.RWMutex
	// 敏感词库状态 0:正常 1:刷新中 用于防止多个协程同时刷新
	status atomic.Int32
}

// NewACFilter 新建AC算法敏感词过滤器
func NewACFilter() *ACFilter {
	return &ACFilter{
		rwLock: &sync.RWMutex{},
		status: atomic.Int32{},
	}
}

func (A *ACFilter) Build(words []string) *ACFilter {
	A.build(words)
	return A
}

func (A *ACFilter) BuildWithFunc(f func() ([]string, error)) (*ACFilter, error) {
	words, err := f()
	if err != nil || len(words) == 0 {
		words = []string{""}
	}
	A.build(words)
	return A, err
}

// build 构建敏感词库
func (A *ACFilter) build(words []string) {
	A.rwLock.Lock()
	defer A.rwLock.Unlock()
	A.root = BuildTrieTree(words)
	buildFailPoint(A.root)
}

// Refresh 刷新词库 保证同一时间只有一个协程在刷新
func (A *ACFilter) Refresh(words []string) {
	// 如果当前有协程在刷新 则直接返回
	if !A.status.CompareAndSwap(0, 1) {
		return
	}
	// 构建新的敏感词库
	refreshRoot := BuildTrieTree(words)
	buildFailPoint(refreshRoot)
	// 用新的敏感词库替换旧的敏感词库
	A.rwLock.Lock()
	A.refreshRoot = refreshRoot
	A.rwLock.Unlock()
	A.refreshRoot = nil
	// 替换成功后 将状态置为正常
	A.status.Store(0)
}
func (A *ACFilter) traverseText(text string, matchHandler func(start, end int, cur *TrieNode)) {
	root := A.root
	runes := []rune(text)
	cur := root
	start := 0

	for i, r := range runes {
		if _, ok := invalidMap[r]; ok {
			continue
		}

		for cur.children[r] == nil && cur != root {
			cur = cur.fail
		}

		cur, ok := cur.children[r]
		if !ok {
			cur = root // 当前字符在Trie树中不存在，重置为root
			start = i + 1
		} else if cur.isEnd {
			matchHandler(start, i, cur)
			cur = root
			start = i + 1 // 从下一个字符重新开始匹配
		}
	}
}

func (A *ACFilter) Filter(text string) string {
	A.rwLock.RLock()
	defer A.rwLock.RUnlock()
	root := A.root
	runes := []rune(text)
	cur := root
	start := 0
	// 遍历文本 忽略无效字符
	for i, r := range runes {
		if _, ok := invalidMap[r]; ok {
			continue
		}
		// 如果当前节点的子节点中没有当前字符 则找失败指针指向的节点的子节点中是否有当前字符
		for cur.children[r] == nil && cur != root {
			cur = cur.fail
		}
		// 上面的for循环结束后，cur一定是root或者cur.children[r] != nil
		cur = cur.children[r]
		// 如果还没有 则证明之前的for循环中cur已经是root了 重置cur为root 从下一个字符开始匹配
		if cur == nil {
			cur = root
			start = i + 1
		}
		if cur.isEnd {
			//敏感词替换
			runes = replace(runes, start, i)
			cur = root
		}
	}
	return string(runes)
}

// 替换敏感词
func replace(runes []rune, start, end int) []rune {
	for i := start; i <= end; i++ {
		runes[i] = '*'
	}
	return runes
}

func (A *ACFilter) FindIn(text string) []string {
	A.rwLock.RLock()
	defer A.rwLock.RUnlock()
	result := make([]string, 0)
	root := A.root
	runes := []rune(text)
	cur := root
	start := 0
	// 遍历文本 忽略无效字符
	for i, r := range runes {
		if _, ok := invalidMap[r]; ok {
			continue
		}
		// 如果当前节点的子节点中没有当前字符 则找失败指针指向的节点的子节点中是否有当前字符
		for cur.children[r] == nil && cur != root {
			cur = cur.fail
		}
		// 上面的for循环结束后，cur一定是root或者cur.children[r] != nil
		cur = cur.children[r]
		// 如果还没有 则证明之前的for循环中cur已经是root了 重置cur为root 从下一个字符开始匹配
		if cur == nil {
			cur = root
			start = i + 1
		}
		if cur.isEnd {
			//敏感词替换
			result = append(result, getSensitiveWord(runes, start, i))
			cur = root
		}
	}
	return result
}

func getSensitiveWord(runes []rune, start, end int) string {
	res := ""
	for i := start; i <= end; i++ {
		if _, ok := invalidMap[runes[i]]; ok {
			continue
		}
		res += string(runes[i])
	}
	return res
}

func (A *ACFilter) Contains(text string) bool {
	A.rwLock.RLock()
	defer A.rwLock.RUnlock()
	root := A.root
	runes := []rune(text)
	cur := root
	// 遍历文本 忽略无效字符
	for _, r := range runes {
		if _, ok := invalidMap[r]; ok {
			continue
		}
		// 如果当前节点的子节点中没有当前字符 则找失败指针指向的节点的子节点中是否有当前字符
		for cur.children[r] == nil && cur != root {
			cur = cur.fail
		}
		// 上面的for循环结束后，cur一定是root或者cur.children[r] != nil
		cur = cur.children[r]
		// 如果还没有 则证明之前的for循环中cur已经是root了 重置cur为root 从下一个字符开始匹配
		if cur == nil {
			cur = root
		}
		if cur.isEnd {
			return true
		}
	}
	return false
}
