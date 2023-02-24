package loader

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/graduation-fci/phase1-demo/dependencies"
	"github.com/graduation-fci/phase1-demo/model"
	"github.com/graduation-fci/phase1-demo/repository"
	"github.com/graduation-fci/phase1-demo/service"
	"github.com/schollz/progressbar/v3"
)

type drugProgress struct {
	bar  *progressbar.ProgressBar
	idx  int
	drug model.Drug
}

type pipeline struct {
	stream         *jsonStream
	drugRepository *repository.DrugRepository
}

func NewPipeline(dp *dependencies.DP) *pipeline {
	return &pipeline{
		stream:         NewJsonStream(),
		drugRepository: repository.NewDrugRepository(dp),
	}
}

func (p pipeline) LoadNodes() {
	done := make(chan struct{}, 1)
	go func() {
		for message := range p.stream.Watch() {
			fmt.Println(message.Drug.Name)
			p.drugRepository.InsetNode(message.Drug.Name)
		}
		done <- struct{}{}
	}()
	p.stream.Start("input/drugs.json")
	<-done
}

func (p pipeline) LoadEdges() {
	done := make(chan struct{}, 1)
	go func() {
		for message := range p.stream.Watch() {
			p.drugRepository.BuildRelations(&message.Drug)
		}
		done <- struct{}{}
	}()
	p.stream.Start("input/mappings.json")
	<-done
}

func (p pipeline) ExtractInteractions() {
	done := make(chan struct{}, 1)
	extractor := service.NewExtractor("output/mappings.json")
	defer extractor.Exit()
	go func() {
		for message := range p.stream.Watch() {
			drug := message.Drug
			interactions := extractor.Interactions(drug)
			drug.Interactions = interactions
			extractor.WriteToDisk(&drug)
		}
		done <- struct{}{}
	}()

	p.stream.Start("input/drugs.json")
	<-done
}

func (p pipeline) ExtractInteractionsV2(workers int) {
	extractor := service.NewExtractor("output/mappings.json")
	defer extractor.Exit()

	var drugs []model.Drug
	jsonFile, err := os.ReadFile("input/drugs.json")
	if err != nil {
		log.Fatal("can't read file")
	}
	err = json.Unmarshal([]byte(jsonFile), &drugs)
	if err != nil {
		log.Fatal("can't Unmarshal file")
	}

	drugsChannel := make(chan drugProgress)

	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 1; i <= workers; i++ {
		go p.extractInteractionsWorker(drugsChannel, extractor, &wg)
	}
	drugsSize := len(drugs)
	bar := progressbar.Default(int64(drugsSize))

	for idx, drug := range drugs {
		drugsChannel <- drugProgress{
			bar:  bar,
			drug: drug,
			idx:  idx,
		}
	}

	close(drugsChannel)
	wg.Wait()
}

func (p pipeline) extractInteractionsWorker(drugProgress <-chan drugProgress, scrapper *service.Extractor, wg *sync.WaitGroup) {
	for progress := range drugProgress {
		log.Println("Started", progress.drug.Name)
		interactions := scrapper.Interactions(progress.drug)
		progress.drug.Interactions = interactions
		scrapper.WriteToDisk(&progress.drug)
		progress.bar.Add(1)
		log.Println("Finished", progress.drug.Name)
	}

	wg.Done()
}
