package main

import (
	"github.com/graduation-fci/phase1-demo/dependencies"
	"github.com/graduation-fci/phase1-demo/loader"
)

func main() {
	dp := dependencies.NewDependencyInjection().WithNeo4j()
	defer dp.Shutdown()

	pipeline := loader.NewPipeline(dp)

	// pipeline.LoadNodes()
	// pipeline.LoadEdges()
	// pipeline.ExtractInteractions()
	pipeline.ExtractInteractionsV2(2)
}
