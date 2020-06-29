package main

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"

	"github.com/nfisher/mdindexer/edit"
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
func (z *Index) Capacity() int {
	z.RLock()
	defer z.RUnlock()
	return cap(z.Names)
}

// WordCount provides the number of Words in the index.
func (z *Index) WordCount() int {
	z.RLock()
	defer z.RUnlock()
	return len(z.Words)
}

// Update incorporates the documents word count frequency into the index.
func (z *Index) Update(doc *Document) {
	z.Lock()
	defer z.Unlock()
	var isNew bool
	pos := z.byName(doc.Name)
	if pos == nameNotFound {
		pos = len(z.Names)
		z.Names = append(z.Names, doc.Name)
		isNew = true
	}

	cur := make(map[string]bool)
	for word, count := range doc.WordCount {
		cur[word] = true
		col, ok := z.Words[word]
		if !ok {
			col = NewColumn(word)
		}
		col.Upsert(pos, count)
		z.Words[word] = col
	}

	if !isNew {
		z.clean(pos, cur)
	}
}

func (z *Index) clean(pos int, cur map[string]bool) {
	for word, col := range z.Words {
		if cur[word] {
			continue
		}
		col.Remove(pos)
		if col.Empty() {
			delete(z.Words, word)
		}
	}
}

// Search returns the list of documents that contain needle.
func (z *Index) Search(needle string) (DocList, error) {
	z.RLock()
	defer z.RUnlock()
	if 1 > len(z.Words) {
		return nil, ErrWordNotIndexed
	}
	var words = Words{{needle, 0}}
	_, ok := z.Words[needle]
	if !ok {
		words = Words{}
		for k := range z.Words {
			d := edit.Distance2(needle, k)
			words = append(words, WordDist{k, d})
		}
		sort.Sort(words)
		end := len(words)
		min := words[0].Distance
		for i := 1; i < len(words); i++ {
			if words[i].Distance > min {
				break
			}
			end = i
		}
		words = words[0:end]
	}

	pos := make(map[string]int)
	var docs = make(DocList, 0, len(words))
	for _, word := range words {
		col := z.Words[word.Word]
		col.Apply(func(id int, count int) {
			if count > 0 {
				doc := z.byId(id)
				relevance := DocRelevance{Document: doc}
				relevance.Count = count
				relevance.Distance = word.Distance
				p, ok := pos[doc]
				if !ok {
					p = len(docs)
					docs = append(docs, relevance)
					pos[doc] = p
					return
				}
				if docs[p].Distance < relevance.Distance {
					return
				}

				docs[p].Distance = relevance.Distance
				docs[p].Count = relevance.Count
			}
		})
	}

	sort.Slice(docs, func(i, j int) bool {
		return docs[i].Count < docs[j].Count
	})

	return docs, nil
}

const nameNotFound = -1

func (z *Index) byName(name string) int {
	var id int
	for ; id < len(z.Names); id++ {
		if name == z.Names[id] {
			return id
		}
	}
	return nameNotFound
}

func (z *Index) byId(id int) string {
	return z.Names[id]
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

func (z *WordColumn) Upsert(pos int, count int) {
	// lazily build the index so a read from file does not break
	if z.idx == nil {
		z.lazyIndex()
	}

	i, ok := z.idx[pos]

	// make a sparse matrices, don't store count < 1
	if count < 1 {
		return
	}

	tup := [2]int{pos, count}
	if !ok {
		i := len(z.Docs)
		z.idx[pos] = i
		z.Docs = append(z.Docs, tup)
		return
	}
	z.Docs[i] = tup
}

func (z *WordColumn) lazyIndex() {
	z.idx = make(map[int]int)
	for i, tup := range z.Docs {
		pos := tup[0]
		z.idx[pos] = i
	}
}

func (z *WordColumn) Apply(each func(int, int)) {
	for _, tup := range z.Docs {
		if tup[1] > 0 {
			each(tup[0], tup[1])
		}
	}
}

func (z *WordColumn) Empty() bool {
	var count int
	z.Apply(func(i int, i2 int) {
		count++
	})
	return count == 0
}

func (z *WordColumn) Remove(pos int) {
	i, ok := z.idx[pos]
	if !ok {
		return
	}
	delete(z.idx, pos)
	// TODO: should consider compacting the array instead of using 0
	z.Docs[i] = [2]int{pos, 0}
}

// Search executes the query against the index returning a document list.
func Search(query string, index *Index) ScoreList {
	var result = make(Scores)
	if query == "" {
		return ScoreList{}
	}

	query = strings.ToLower(query)
	terms := strings.Split(query, " ")
	union := make(StrSet)
	for i, term := range terms {
		docs, err := index.Search(term)
		if err != nil {
			log.Printf("search=failed query=%s error='%v'\n", term, err)
			continue
		}

		s := make(StrSet)
		for _, d := range docs {
			n := d.Document
			if i == 0 {
				result[n] = d.Distance + d.Rank
				s[n] = true
				continue
			}
			if !union[n] {
				continue
			}
			result[n] += d.Distance + d.Rank
			s[n] = true
		}
		union = s
	}

	list := make(ScoreList, 0, len(union))
	for n := range result {
		if !union[n] {
			continue
		}
		length := len(n) - len(query)
		dist := edit.Distance2(query, n) - length
		list = append(list, Score{Document: n, Rank: result[n], NameDistance: dist})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].NameDistance < list[j].NameDistance
	})
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].Rank < list[j].Rank
	})

	return list
}

type Score struct {
	Document     string
	Rank         int
	NameDistance int
}
type ScoreList []Score

type Scores map[string]int

type StrSet map[string]bool

func (s StrSet) Union(o StrSet) StrSet {
	r := make(StrSet)
	for k := range s {
		if o[k] {
			r[k] = true
		}
	}
	return r
}

type DocList []DocRelevance

type DocRelevance struct {
	Document string
	Count    int
	Distance int
	Rank     int
}

type WordDist struct {
	Word     string
	Distance int
}

type Words []WordDist

func (z Words) Len() int           { return len(z) }
func (z Words) Swap(i, j int)      { z[i], z[j] = z[j], z[i] }
func (z Words) Less(i, j int) bool { return z[i].Distance < z[j].Distance }
