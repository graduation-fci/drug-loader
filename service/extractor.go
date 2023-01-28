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

func (ext Extractor) Interactions(drug model.Drug) {
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
		log.Println("Visiting list of interactions", r.URL)
	})

	detialsWorker := mainWorker.Clone()
	detialsWorker.OnRequest(func(r *colly.Request) {
		log.Println("Visiting interaction details", r.URL)
	})

	mainWorker.OnHTML("a[href]", func(e *colly.HTMLElement) {
		if e.Text == "Check interactions" {
			prefixs := strings.Split(e.Attr("href"), "/")
			file := strings.Split(prefixs[len(prefixs)-1], ".html")
			interactionsUrl := "https://www.drugs.com/drug-interactions/" + file[0] + "-index.html"
			interactionsWorker.Request("GET", interactionsUrl, nil, e.Request.Ctx, nil)
			time.Sleep(time.Duration(time.Second * 2))
		}
	})

	interactionsWorker.OnHTML(`.contentBox`, func(contentBox *colly.HTMLElement) {
		var drugs []model.Drug
		if contentBox.DOM.Find(".col-list-az").Length() > 0 {
			drugs = ext.interactionList(".ddc-list-column-2", contentBox)
		} else {
			drugs = ext.interactionList(".ddc-list-unstyled", contentBox)
		}
		for _, drug := range drugs {
			detialsWorker.Request("GET", drug.Url, nil, contentBox.Request.Ctx, nil)
			detialsWorker.Request("GET", drug.Url+"?professional=1", nil, contentBox.Request.Ctx, nil)
		}
		time.Sleep(time.Duration(time.Second * 2))
	})

	// detialsWorker.OnHTML(".contentBox ")

	err := mainWorker.Visit(drug.Url)
	if err != nil {
		log.Fatalln("error occured", err)
	}
}

func (ex Extractor) interactionList(selector string, boxElement *colly.HTMLElement) []model.Drug {
	var drugs []model.Drug
	boxElement.ForEach(selector, func(i int, drugParentDiv *colly.HTMLElement) {
		if drugParentDiv.DOM.HasClass("interactions-label") {
			return
		}
		drugParentDiv.ForEach("li", func(j int, drugInteraction *colly.HTMLElement) {
			drug := model.Drug{
				Name: drugInteraction.ChildText("a"),
				Url:  "https://www.drugs.com" + drugInteraction.ChildAttr("a", "href"),
			}
			drugs = append(drugs, drug)
			log.Println("found drug ", drug.Name, drug.Url)
		})
	})

	return drugs
}