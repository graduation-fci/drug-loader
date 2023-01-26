package service

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/graduation-fci/phase1-demo/model"
)

type Extractor struct {
	fd *os.File
}

func NewExtractor(fileName string) *Extractor {
	extractor := Extractor{}
	extractor.createFd(fileName)

	return &extractor
}

func (e *Extractor) createFd(fileName string) {
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("Cannot create file %q: %s\n", fileName, err)
		return
	}
	e.fd = file
}

func (e *Extractor) Exit() {
	e.fd.Close()
}

func (ex Extractor) Interactions(drug model.Drug) {
	fmt.Println("Starting", drug.Name)

	mainWorker := colly.NewCollector(
		colly.AllowedDomains("www.drugs.com"),
		colly.CacheDir("./cache"),
	)

	mainWorker.OnRequest(func(r *colly.Request) {
		log.Println("Visited home url", r.URL)
	})

	interactionsWorker := mainWorker.Clone()
	interactionsWorker.OnRequest(func(r *colly.Request) {
		log.Println("Visited details interactions", r.URL)
	})

	mainWorker.OnHTML("a[href]", func(e *colly.HTMLElement) {
		if e.Text == "Check interactions" {
			prefixs := strings.Split(e.Attr("href"), "/")
			file := strings.Split(prefixs[len(prefixs)-1], ".html")
			interactionsUrl := "https://www.drugs.com/drug-interactions/" + file[0] + "-index.html"
			log.Println("Visiting", interactionsUrl)
			interactionsWorker.Request("GET", interactionsUrl, nil, e.Request.Ctx, nil)
			time.Sleep(time.Duration(time.Second * 2))
		}
	})

	interactionsWorker.OnHTML(`#container #contentWrap #content .contentBox`, func(e1 *colly.HTMLElement) {
		alphabeticsCount := false
		e1.ForEachWithBreak(".col-list-az", func(_ int, h *colly.HTMLElement) bool {
			alphabeticsCount = true
			return true
		})
		time.Sleep(time.Duration(time.Second * 2))
	})

	// interactionsWorker.OnHTML(`#container #contentWrap #content .contentBox .ddc-list-unstyled`, func(e1 *colly.HTMLElement) {
	// 	e1.ForEach("li", func(_ int, e2 *colly.HTMLElement) {
	// 		log.Println(e2.Text)
	// 	})
	// })

	err := mainWorker.Visit(drug.Url)
	if err != nil {
		log.Fatalln("error occured", err)
	}
}
