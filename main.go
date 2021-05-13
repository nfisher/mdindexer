package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

func filePattern(language string) string {
	switch language {
	case "go":
		return "\\.go$"
	case "java":
		return "\\.java$"
	case "js":
		return "\\.js$"
	}
	return ""
}

func main() {
	var language string
	var start string

	flag.StringVar(&start, "start", ".", "search start")
	flag.StringVar(&language, "lang", "", "language (e.g. java, go, english)")
	flag.Parse()

	pattern := filePattern(language)
	if pattern == "" {
		log.Fatal("invalid language specified")
	}

	paths := strings.Split(start, ",")
	var stopWords = make(StopWords)
	for _, w := range stopWordMap[language] {
		stopWords[w] = true
	}

	ts := time.Now()
	filenames, err := DocumentList(paths, pattern)
	if err != nil {
		log.Fatalf("glob=failed start=%s pattern=%s error='%v'", start, pattern, err)
	}
	log.Printf("documentList=success start=`%s` pattern=`%s` count=%d\n", start, pattern, len(filenames))

	index := New(len(filenames))

	fnch := make(chan string, runtime.NumCPU()*4)
	doch := make(chan *Document, runtime.NumCPU()*4)

	var docClose sync.Once
	var wg sync.WaitGroup
	for i := 0; i < runtime.NumCPU()*2; i++ {
		go readDoc(fnch, doch, &wg, &docClose, stopWords)
	}

	var wgig sync.WaitGroup
	go updateIndex(doch, index, &wgig)

	for _, filename := range filenames {
		fnch <- filename
	}
	close(fnch)
	wg.Wait()
	wgig.Wait()

	log.Printf("documents=%d words=%d latency=%v\n", index.Capacity(), index.WordCount(), time.Since(ts))

	mux := BuildRoutes(paths, index)
	log.Println("addr=127.0.0.1:8000")
	err = http.ListenAndServe("127.0.0.1:8000", mux)
	if err != nil {
		log.Fatalf("listen=failed error='%v'\n", err)
	}
}

func readDoc(fnch chan string, doch chan *Document, wg *sync.WaitGroup, docClose *sync.Once, stopWords StopWords) {
	wg.Add(1)
	for filename := range fnch {
		doc, err := readFile(filename, stopWords)
		if err != nil {
			log.Printf("readFile=failed filename=%s error='%v'\n", filename, err)
			continue
		}
		doch <- doc
	}
	wg.Done()
	wg.Wait()
	docClose.Do(func() {
		close(doch)
	})
}

func updateIndex(ch chan *Document, index *Index, wgig *sync.WaitGroup) {
	wgig.Add(1)
	for doc := range ch {
		index.Update(doc)
	}
	wgig.Done()
}

var jsStopWords = []string{
	"abstract",
	"arguments",
	"await",
	"boolean",
	"break",
	"byte",
	"case",
	"catch",
	"char",
	"class",
	"const",
	"continue",
	"debugger",
	"default",
	"delete",
	"do",
	"double",
	"else",
	"enum",
	"eval",
	"export",
	"extends",
	"false",
	"final",
	"finally",
	"float",
	"for",
	"function",
	"goto",
	"if",
	"implements",
	"import",
	"in",
	"instanceof",
	"int",
	"interface",
	"let",
	"long",
	"native",
	"new",
	"null",
	"package",
	"private",
	"protected",
	"public",
	"return",
	"short",
	"static",
	"super",
	"switch",
	"synchronized",
	"this",
	"throw",
	"throws",
	"transient",
	"true",
	"try",
	"typeof",
	"var",
	"void",
	"volatile",
	"while",
	"with",
	"yield",
}

var goStopWords = []string{
	"break",
	"case",
	"chan",
	"const",
	"continue",
	"default",
	"defer",
	"else",
	"fallthrough",
	"for",
	"func",
	"go",
	"goto",
	"if",
	"import",
	"interface",
	"map",
	"package",
	"range",
	"return",
	"select",
	"struct",
	"switch",
	"type",
	"var",
}

var javaStopWords = []string{
	"abstract",
	"assert",
	"boolean",
	"break",
	"byte",
	"case",
	"catch",
	"char",
	"class",
	"const",
	"continue",
	"default",
	"double",
	"do",
	"else",
	"enum",
	"extends",
	"false",
	"final",
	"finally",
	"float",
	"for",
	"goto",
	"if",
	"implements",
	"import",
	"instanceof",
	"int",
	"interface",
	"long",
	"native",
	"new",
	"null",
	"package",
	"private",
	"protected",
	"public",
	"return",
	"short",
	"static",
	"strictfp",
	"super",
	"switch",
	"synchronized",
	"this",
	"throw",
	"throws",
	"transient",
	"true",
	"try",
	"void",
	"volatile",
	"while",
}
var englishStopWords = []string{"a", "about", "above", "above", "across", "after", "afterwards", "again", "against", "all", "almost", "alone", "along", "already", "also", "although", "always", "am", "among", "amongst", "amoungst", "amount", "an", "and", "another", "any", "anyhow", "anyone", "anything", "anyway", "anywhere", "are", "around", "as", "at", "back", "be", "became", "because", "become", "becomes", "becoming", "been", "before", "beforehand", "behind", "being", "below", "beside", "besides", "between", "beyond", "bill", "both", "bottom", "but", "by", "call", "can", "cannot", "cant", "co", "con", "could", "couldnt", "cry", "de", "describe", "detail", "do", "done", "down", "due", "during", "each", "eg", "eight", "either", "eleven", "else", "elsewhere", "empty", "enough", "etc", "even", "ever", "every", "everyone", "everything", "everywhere", "except", "few", "fifteen", "fify", "fill", "find", "fire", "first", "five", "for", "former", "formerly", "forty", "found", "four", "from", "front", "full", "further", "get", "give", "had", "has", "hasnt", "have", "he", "hence", "her", "here", "hereafter", "hereby", "herein", "hereupon", "hers", "herself", "him", "himself", "his", "how", "however", "hundred", "ie", "if", "in", "inc", "indeed", "interest", "into", "is", "it", "its", "itself", "keep", "last", "latter", "latterly", "least", "less", "ltd", "made", "many", "may", "me", "meanwhile", "might", "mill", "mine", "more", "moreover", "most", "mostly", "move", "much", "must", "my", "myself", "name", "namely", "neither", "never", "nevertheless", "next", "nine", "no", "nobody", "none", "noone", "nor", "not", "nothing", "now", "nowhere", "of", "off", "often", "on", "once", "one", "only", "onto", "or", "other", "others", "otherwise", "our", "ours", "ourselves", "out", "over", "own", "part", "per", "perhaps", "please", "put", "rather", "re", "same", "see", "seem", "seemed", "seeming", "seems", "serious", "several", "she", "should", "show", "side", "since", "sincere", "six", "sixty", "so", "some", "somehow", "someone", "something", "sometime", "sometimes", "somewhere", "still", "such", "system", "take", "ten", "than", "that", "the", "their", "them", "themselves", "then", "thence", "there", "thereafter", "thereby", "therefore", "therein", "thereupon", "these", "they", "thickv", "thin", "third", "this", "those", "though", "three", "through", "throughout", "thru", "thus", "to", "together", "too", "top", "toward", "towards", "twelve", "twenty", "two", "un", "under", "until", "up", "upon", "us", "very", "via", "was", "we", "well", "were", "what", "whatever", "when", "whence", "whenever", "where", "whereafter", "whereas", "whereby", "wherein", "whereupon", "wherever", "whether", "which", "while", "whither", "who", "whoever", "whole", "whom", "whose", "why", "will", "with", "within", "without", "would", "yet", "you", "your", "yours", "yourself", "yourselves"}

var stopWordMap = map[string][]string{
	"english": englishStopWords,
	"go":      goStopWords,
	"java":    javaStopWords,
	"js":      jsStopWords,
}

func readFile(filename string, stopWords StopWords) (*Document, error) {
	r, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	doc := WordFrequency(filename, r, stopWords)
	doc.Name = filename
	return doc, nil
}
