package service

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/gocolly/colly"
	"github.com/graduation-fci/phase1-demo/model"
)

const (
	PageTypeProfessional = "professional"
	PageTypeConsumer     = "consumer"
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
	file.WriteString("[")
	e.fd = file
}

func (e *Extractor) WriteToDisk(drug *model.Drug) {
	bytes, _ := json.Marshal(drug)
	e.fd.Write(bytes)
	e.fd.WriteString(", \n")
}

func (e *Extractor) Exit() {
	e.fd.WriteString("]")
	e.fd.Close()
}

func (ext Extractor) Interactions(drug model.Drug) []model.Interaction {
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
		}
	})

	interactionsWorker.OnHTML(`.contentBox`, func(contentBox *colly.HTMLElement) {
		var drugs []model.Drug
		if contentBox.DOM.Find(".col-list-az").Length() > 0 {
			drugs = ext.interactionList(".ddc-list-column-2", contentBox)
		} else {
			drugs = ext.interactionList(".ddc-list-unstyled", contentBox)
		}
		for _, drugInstance := range drugs {
			drugName := strings.Join(strings.Fields(drugInstance.Name), "-")
			detialsWorker.Request("GET", drugInstance.Url+"?drugName="+drugName, nil, contentBox.Request.Ctx, nil)
			detialsWorker.Request("GET", drugInstance.Url+"?professional=1&drugName="+drugName, nil, contentBox.Request.Ctx, nil)
		}
	})
	var publicInteractions []model.Interaction
	var professionalInteractions []model.Interaction
	detialsWorker.OnHTML(".contentBox .interactions-reference-wrapper", func(interactionsWrapper *colly.HTMLElement) {
		drugNameQuery := interactionsWrapper.Request.URL.Query()["drugName"]
		drugName := strings.Join(strings.Split(drugNameQuery[0], "-"), " ")
		interactionsWrapper.ForEach(".interactions-reference", func(i int, interaction *colly.HTMLElement) {
			if _, ok := interactionsWrapper.Request.URL.Query()[PageTypeProfessional]; ok {
				professionalInteractions = append(professionalInteractions, model.Interaction{
					ProfessionalEffect: EffectDescription(interaction),
				})
			} else {
				publicInteractions = append(publicInteractions, model.Interaction{
					Name:           drugName,
					HashedName:     InteractionName(interaction),
					Level:          interaction.ChildText(".ddc-status-label"),
					ConsumerEffect: EffectDescription(interaction),
				})
			}
		})
	})

	err := mainWorker.Visit(drug.Url)
	if err != nil {
		log.Fatalln("error occured", err)
	}

	for idx := range publicInteractions {
		if len(professionalInteractions) <= idx {
			break
		}

		publicInteractions[idx].ProfessionalEffect = professionalInteractions[idx].ProfessionalEffect
	}

	log.Println("finished")
	return publicInteractions
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
