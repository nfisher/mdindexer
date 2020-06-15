package main

import (
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"sync"
	"time"

	"github.com/fsnotify/fsevents"
	"github.com/tinylib/msgp/msgp"
)

func main() {
	var start string
	var pattern string
	flag.StringVar(&start, "start", "./", "search start")
	flag.StringVar(&pattern, "glob", "", "file path pattern (e.g. .*.md)")
	flag.Parse()

	if pattern == "" {
		log.Fatal("invalid glob pattern")
	}

	var stopWords = make(StopWords)
	for _, w := range javaStopWords {
		stopWords[w] = true
	}

	ts := time.Now()
	filenames, err := DocumentList(start, pattern)
	if err != nil {
		log.Fatalf("glob=failed start=%s pattern=%s error='%v'", start, pattern, err)
	}
	log.Printf("documentList=success start=`%s` pattern=`%s`\n", start, pattern)

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

	log.Println("parsed", index.Capacity(), "files with", index.WordCount(), "distinct words in", time.Now().Sub(ts))

	outFilename := "index.msgp"
	err = writeFile(outFilename, index)
	if err != nil {
		log.Fatalf("writeFile=failed filename=%s error='%v'\n", outFilename, err)
	}

	dev, err := fsevents.DeviceForPath(start)
	if err != nil {
		log.Fatalf("Failed to retrieve device for path: %v", err)
	}

	es := &fsevents.EventStream{
		Paths:   []string{start},
		Latency: 500 * time.Millisecond,
		Device:  dev,
		Flags:   fsevents.FileEvents | fsevents.WatchRoot}
	es.Start()
	ec := es.Events

	go func() {
		re := regexp.MustCompile(pattern)
		for msg := range ec {
			for _, event := range msg {
				logEvent(event, re)
			}
		}
	}()

	mux := BuildRoutes(start, index)
	log.Println("binding to 127.0.0.1:8000")
	err = http.ListenAndServe("127.0.0.1:8000", mux)
	if err != nil {
		log.Fatalf("listen=failed error='%v'\n", err)
	}
	//consoleLoop(index)
}

var noteDescription = map[fsevents.EventFlags]string{
	fsevents.MustScanSubDirs: "MustScanSubdirs",
	fsevents.UserDropped:     "UserDropped",
	fsevents.KernelDropped:   "KernelDropped",
	fsevents.EventIDsWrapped: "EventIDsWrapped",
	fsevents.HistoryDone:     "HistoryDone",
	fsevents.RootChanged:     "RootChanged",
	fsevents.Mount:           "Mount",
	fsevents.Unmount:         "Unmount",

	fsevents.ItemCreated:       "Created",
	fsevents.ItemRemoved:       "Removed",
	fsevents.ItemInodeMetaMod:  "InodeMetaMod",
	fsevents.ItemRenamed:       "Renamed",
	fsevents.ItemModified:      "Modified",
	fsevents.ItemFinderInfoMod: "FinderInfoMod",
	fsevents.ItemChangeOwner:   "ChangeOwner",
	fsevents.ItemXattrMod:      "XAttrMod",
	fsevents.ItemIsFile:        "IsFile",
	fsevents.ItemIsDir:         "IsDir",
	fsevents.ItemIsSymlink:     "IsSymLink",
}

func logEvent(event fsevents.Event, re *regexp.Regexp) {
	if !re.MatchString(event.Path) {
		return
	}
	note := ""
	for bit, description := range noteDescription {
		if event.Flags&bit == bit {
			note += description + " "
		}
	}
	log.Printf("EventID: %d Path: %s Flags: %s", event.ID, event.Path, note)
}

func consoleLoop(index *Index) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("\nEnter a query:")
		if !scanner.Scan() {
			break
		}
		search := scanner.Text()
		fmt.Println()

		ts := time.Now()
		exact := ExactMatch(search, index)
		for k := range exact {
			fmt.Println(k)
		}
		fmt.Println("found", len(exact), "entries in", time.Now().Sub(ts))
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
var stopWordList = []string{"a", "about", "above", "above", "across", "after", "afterwards", "again", "against", "all", "almost", "alone", "along", "already", "also", "although", "always", "am", "among", "amongst", "amoungst", "amount", "an", "and", "another", "any", "anyhow", "anyone", "anything", "anyway", "anywhere", "are", "around", "as", "at", "back", "be", "became", "because", "become", "becomes", "becoming", "been", "before", "beforehand", "behind", "being", "below", "beside", "besides", "between", "beyond", "bill", "both", "bottom", "but", "by", "call", "can", "cannot", "cant", "co", "con", "could", "couldnt", "cry", "de", "describe", "detail", "do", "done", "down", "due", "during", "each", "eg", "eight", "either", "eleven", "else", "elsewhere", "empty", "enough", "etc", "even", "ever", "every", "everyone", "everything", "everywhere", "except", "few", "fifteen", "fify", "fill", "find", "fire", "first", "five", "for", "former", "formerly", "forty", "found", "four", "from", "front", "full", "further", "get", "give", "had", "has", "hasnt", "have", "he", "hence", "her", "here", "hereafter", "hereby", "herein", "hereupon", "hers", "herself", "him", "himself", "his", "how", "however", "hundred", "ie", "if", "in", "inc", "indeed", "interest", "into", "is", "it", "its", "itself", "keep", "last", "latter", "latterly", "least", "less", "ltd", "made", "many", "may", "me", "meanwhile", "might", "mill", "mine", "more", "moreover", "most", "mostly", "move", "much", "must", "my", "myself", "name", "namely", "neither", "never", "nevertheless", "next", "nine", "no", "nobody", "none", "noone", "nor", "not", "nothing", "now", "nowhere", "of", "off", "often", "on", "once", "one", "only", "onto", "or", "other", "others", "otherwise", "our", "ours", "ourselves", "out", "over", "own", "part", "per", "perhaps", "please", "put", "rather", "re", "same", "see", "seem", "seemed", "seeming", "seems", "serious", "several", "she", "should", "show", "side", "since", "sincere", "six", "sixty", "so", "some", "somehow", "someone", "something", "sometime", "sometimes", "somewhere", "still", "such", "system", "take", "ten", "than", "that", "the", "their", "them", "themselves", "then", "thence", "there", "thereafter", "thereby", "therefore", "therein", "thereupon", "these", "they", "thickv", "thin", "third", "this", "those", "though", "three", "through", "throughout", "thru", "thus", "to", "together", "too", "top", "toward", "towards", "twelve", "twenty", "two", "un", "under", "until", "up", "upon", "us", "very", "via", "was", "we", "well", "were", "what", "whatever", "when", "whence", "whenever", "where", "whereafter", "whereas", "whereby", "wherein", "whereupon", "wherever", "whether", "which", "while", "whither", "who", "whoever", "whole", "whom", "whose", "why", "will", "with", "within", "without", "would", "yet", "you", "your", "yours", "yourself", "yourselves"}

func writeFile(filename string, index *Index) error {
	w, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer w.Close()
	gzw := gzip.NewWriter(w)
	defer gzw.Close()

	mw := msgp.NewWriter(gzw)
	err = index.EncodeMsg(mw)
	if err != nil {
		return err
	}
	return nil
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
