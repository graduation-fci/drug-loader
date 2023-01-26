package loader

import (
	"fmt"
	"time"

	"github.com/graduation-fci/phase1-demo/dependencies"
	"github.com/graduation-fci/phase1-demo/repository"
	"github.com/graduation-fci/phase1-demo/service"
)

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
			extractor.Interactions(message.Drug)
			time.Sleep(time.Duration(time.Second * 3))
		}
		done <- struct{}{}
	}()

	p.stream.Start("input/drugs.json")
	<-done
}
