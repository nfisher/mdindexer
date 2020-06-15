package main

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const minimalDoc = `---
title: Hello World
---
Hello world how are you today?`

var minimalDocFreq = map[string]int{
	"hello": 2,
	"world": 2,
	"how":   1,
	"are":   1,
	"today": 1,
	"title": 1,
	"you":   1,
}

func Test_extract_word_count_from_document_excluding_stop_words(t *testing.T) {
	t.Parallel()
	r := strings.NewReader(minimalDoc)
	sw := StopWords{
		"title": true,
	}
	expected := make(map[string]int)
	for k, v := range minimalDocFreq {
		if k == "title" {
			continue
		}
		expected[k] = v
	}
	doc := WordFrequency("test", r, sw)
	if !cmp.Equal(doc.WordCount, expected) {
		t.Errorf("WordFrequency(minimalDoc) mismatch (-want +got)\n%s", cmp.Diff(doc.WordCount, expected))
	}
}

func Test_extract_word_count_from_document(t *testing.T) {
	t.Parallel()
	r := strings.NewReader(minimalDoc)
	doc := WordFrequency("test", r, make(StopWords))
	if !cmp.Equal(doc.WordCount, minimalDocFreq) {
		t.Errorf("WordFrequency(minimalDoc) mismatch (-want +got)\n%s", cmp.Diff(doc.WordCount, minimalDocFreq))
	}
}

func Test_document_list(t *testing.T) {
	t.Parallel()
	list, _ := DocumentList("./testdata", ".*.md")
	expected := []string{"testdata/2018-04-06-Docker-for-Development.md", "testdata/2019-06-20-Maven-to-bazel-prep.md"}
	if !cmp.Equal(list, expected) {
		t.Errorf("DocumentList(\"./testdata\", \".*.md\") mismatch (-want +got)\n%s",
			cmp.Diff(list, expected))
	}
}


