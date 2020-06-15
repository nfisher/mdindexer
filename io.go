package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/scanner"
)

type StopWords map[string]bool

func WordFrequency(filename string, r io.Reader, stopWords StopWords) *Document {
	wordCount := make(map[string]int)
	var s scanner.Scanner
	s.Init(r)
	s.Filename = filename
	s.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.ScanFloats
	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		if tok == scanner.Ident {
			txt := strings.ToLower(s.TokenText())
			if stopWords[txt] {
				continue
			}
			c, ok := wordCount[txt]
			if !ok {
				c = 0
			}
			wordCount[txt] = c + 1
		}
	}
	return &Document{WordCount: wordCount}
}

func DocumentList(start string, expr string) ([]string, error) {
	var docs []string
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	err = filepath.Walk(start, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() && info.Name() == "target" {
			return filepath.SkipDir
		}
		if !info.IsDir() && re.MatchString(info.Name()) {
			info.Sys()
			docs = append(docs, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return docs, err
}