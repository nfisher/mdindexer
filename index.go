package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
)

//go:generate msgp

var (
	ErrWordNotIndexed = fmt.Errorf("index does not contain word")
)

type Document struct {
	Name      string
	WordCount map[string]int
}

// New creates an index that can accommodate the number of documents specified by size.
func New(size int) *Index {
	return &Index{
		Words: make(map[string]*WordColumn),
		Names: make([]string, 0, size),
	}
}

// Index is a matrix that counts word occurrences in documents.
type Index struct {
	Words        map[string]*WordColumn
	Names        []string
	sync.RWMutex `msg:"-"`
}

// Capacity returns the number of documents in the index.
func (i *Index) Capacity() int {
	i.RLock()
	defer i.RUnlock()
	return cap(i.Names)
}

// WordCount provides the number of Words in the index.
func (i *Index) WordCount() int {
	i.RLock()
	defer i.RUnlock()
	return len(i.Words)
}

// Update incorporates the documents word count frequency into the index.
func (i *Index) Update(doc *Document) {
	i.Lock()
	defer i.Unlock()
	var isNew bool
	pos := i.byName(doc.Name)
	if pos == nameNotFound {
		pos = len(i.Names)
		i.Names = append(i.Names, doc.Name)
		isNew = true
	}

	cur := make(map[string]bool)
	for word, count := range doc.WordCount {
		cur[word] = true
		col, ok := i.Words[word]
		if !ok {
			col = NewColumn(word)
		}
		col.Upsert(pos, count)
		i.Words[word] = col
	}

	if !isNew {
		i.clean(pos, cur)
	}
}

func (i *Index) clean(pos int, cur map[string]bool) {
	for word, col := range i.Words {
		if cur[word] {
			continue
		}
		col.Remove(pos)
		if col.Empty() {
			delete(i.Words, word)
		}
	}
}

// DocList provides a list of documents found in the index.
type DocList []string

// Search returns the list of documents that contain needle.
func (i *Index) Search(needle string) (DocList, error) {
	i.RLock()
	defer i.RUnlock()
	col, ok := i.Words[needle]
	if !ok {
		return nil, ErrWordNotIndexed
	}

	var docs DocList
	col.Apply(func(id int, count int) {
		if count > 0 {
			doc := i.byId(id)
			docs = append(docs, doc)
		}
	})

	return docs, nil
}

const nameNotFound = -1

func (i *Index) byName(name string) int {
	var id int
	for ; id < len(i.Names); id++ {
		if name == i.Names[id] {
			return id
		}
	}
	return nameNotFound
}

func (i *Index) byId(id int) string {
	return i.Names[id]
}

// NewColumn returns a newly initialised WordColumn.
func NewColumn(name string) *WordColumn {
	return &WordColumn{
		Name: name,
		Docs: make([][2]int, 0, 64),
	}
}

// WordColumn maintains the frequency a word occurs in the named document.
type WordColumn struct {
	Name string
	Docs [][2]int
	// msgpack/json can't serialise map with int keys
	idx map[int]int
}

func (wc *WordColumn) Upsert(pos int, count int) {
	// lazily build the index so a read from file does not break
	if wc.idx == nil {
		wc.lazyIndex()
	}

	i, ok := wc.idx[pos]

	// make a sparse matrices, don't store count < 1
	if count < 1 {
		return
	}

	tup := [2]int{pos, count}
	if !ok {
		i := len(wc.Docs)
		wc.idx[pos] = i
		wc.Docs = append(wc.Docs, tup)
		return
	}
	wc.Docs[i] = tup
}

func (wc *WordColumn) lazyIndex() {
	wc.idx = make(map[int]int)
	for i, tup := range wc.Docs {
		pos := tup[0]
		wc.idx[pos] = i
	}
}

func (wc *WordColumn) Apply(each func(int, int)) {
	for _, tup := range wc.Docs {
		if tup[1] > 0 {
			each(tup[0], tup[1])
		}
	}
}

func (wc *WordColumn) Empty() bool {
	var count int
	wc.Apply(func(i int, i2 int) {
		count++
	})
	return count == 0
}

func (wc *WordColumn) Remove(pos int) {
	i, ok := wc.idx[pos]
	if !ok {
		return
	}
	delete(wc.idx, pos)
	// TODO: should consider compacting the array instead of using 0
	wc.Docs[i] = [2]int{pos, 0}
}

func ExactMatch(needle string, index *Index) map[string]bool {
	needle = strings.ToLower(needle)
	exact := make(map[string]bool)
	if needle != "" {
		needles := strings.Split(needle, " ")
		for _, s := range needles {
			match := make(map[string]bool)
			docs, err := index.Search(s)
			if err != nil {
				log.Printf("search=failed needle=%s error='%v'\n", s, err)
				continue
			}

			for i := range docs {
				match[docs[i]] = true
			}
			if len(exact) == 0 {
				exact = match
			}
			for k := range exact {
				if !match[k] {
					delete(exact, k)
				}
			}
		}
	}
	return exact
}
