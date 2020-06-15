package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func fooMD() *Document {
	return &Document{
		Name:      "foo.md",
		WordCount: map[string]int{"hello": 1, "world": 1},
	}
}

func bazMD(words ...string) *Document {
	var m = make(map[string]int)
	for _, word := range words {
		m[word] = 1
	}

	return &Document{
		Name:      "baz.md",
		WordCount: m,
	}
}

func barMD() *Document {
	return &Document{
		Name:      "bar.md",
		WordCount: map[string]int{"ciao": 1, "world": 1},
	}
}

func Test_new_has_capacity_matching_document_size(t *testing.T) {
	t.Parallel()
	// Given a list of 20 documents
	// When an index is created
	// Then it should have a capacity of 20 documents.
	index := New(20)
	if index.Capacity() != 20 {
		t.Errorf("index.Capacity()=%v, want 20", index.Capacity())
	}
}

func Test_new_has_zero_word_count_after_initialisation(t *testing.T) {
	t.Parallel()
	index := New(20)
	if index.WordCount() != 0 {
		t.Errorf("index.WordCount()=%v, want 0", index.WordCount())
	}
}

func Test_update_with_single_document_updates_word_count_correctly(t *testing.T) {
	t.Parallel()
	index := New(20)
	doc := fooMD()
	index.Update(doc)

	if index.WordCount() != 2 {
		t.Errorf("index.WordCount()=%v, want 2", index.WordCount())
	}
}

func Test_search_should_error_if_term_not_found_in_index(t *testing.T) {
	t.Parallel()
	index := New(20)
	_, err := index.Search("world")
	if err != ErrWordNotIndexed {
		t.Errorf("index.Search(`world`) error=%v, want ErrWordNotIndexed", err)
	}
}

func Test_searching_multiple_documents_should_return_document_matching_term(t *testing.T) {
	t.Parallel()
	index := New(20)
	doc := fooMD()
	index.Update(doc)
	doc = barMD()
	index.Update(doc)

	docs, _ := index.Search("hello")
	if !cmp.Equal(docs, DocList{"foo.md"}) {
		t.Errorf("index.Search(`hello`) mismatch (-want +got)\n%s", cmp.Diff(docs, DocList{"foo.md"}))
	}
}

func Test_searching_multiple_documents_should_return_documents_matching_common_term(t *testing.T) {
	t.Parallel()
	index := New(2)
	doc := fooMD()
	index.Update(doc)
	doc = barMD()
	index.Update(doc)

	docs, _ := index.Search("world")
	if !cmp.Equal(docs, DocList{"foo.md", "bar.md"}) {
		t.Errorf("index.Search(`hello`) mismatch (-want +got)\n%s", cmp.Diff(docs, DocList{"foo.md", "bar.md"}))
	}
}

func Test_update_to_existing_document_does_not_append_another_record(t *testing.T) {
	t.Parallel()
	index := New(20)
	doc := fooMD()
	index.Update(doc)
	index.Update(doc)

	docs, _ := index.Search("hello")
	if !cmp.Equal(docs, DocList{"foo.md"}) {
		t.Errorf("index.Search(`hello`) mismatch (-want +got)\n%s", cmp.Diff(DocList{"foo.md"}, docs))
	}
}

func Test_update_removing_word_should_remove_column_from_index_when_empty(t *testing.T) {
	t.Parallel()
	index := New(20)
	doc := bazMD("hello", "world")
	index.Update(doc)
	doc = bazMD("hello")
	index.Update(doc)

	_, err := index.Search("world")
	if err != ErrWordNotIndexed {
		t.Errorf("index.Search(`world`) error=%v, want ErrWordNotIndexed", err)
	}
}

func Test_update_removing_word_should_update_search_terms(t *testing.T) {
	t.Parallel()
	index := New(20)
	doc := bazMD("ciao", "world")
	index.Update(doc)
	doc = barMD()
	index.Update(doc)
	doc = bazMD("world")
	index.Update(doc)

	expected := DocList{"bar.md"}
	docs, _ := index.Search("ciao")
	if !cmp.Equal(docs, expected) {
		t.Errorf("index.Search(`ciao`) mismatch (-want +got)\n%s", cmp.Diff(expected, docs))
	}

	expected = DocList{"baz.md", "bar.md"}
	docs, _ = index.Search("world")
	if !cmp.Equal(docs, expected) {
		t.Errorf("index.Search(`world`) mismatch (-want +got)\n%s", cmp.Diff(expected, docs))
	}
}
