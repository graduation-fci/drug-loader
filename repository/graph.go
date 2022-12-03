package repository

import (
	"context"
	"fmt"

	"github.com/graduation-fci/phase1-demo/dependencies"
	"github.com/graduation-fci/phase1-demo/model"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type DrugRepository struct {
	graph neo4j.DriverWithContext
}

func NewDrugRepository(dp *dependencies.DP) *DrugRepository {
	return &DrugRepository{
		graph: dp.GetNeo4j(),
	}
}

func (d DrugRepository) InsetNode(name string) {
	ctx := context.TODO()
	session := d.graph.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)
	result, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		records, err := tx.Run(ctx, "CREATE (n:Drug {name: $name }) RETURN n.name", map[string]any{
			"name": name,
		})
		if err != nil {
			return nil, err
		}
		record, err := records.Single(ctx)
		if err != nil {
			return nil, err
		}
		return &model.Drug{
			Name: record.Values[0].(string),
		}, nil
	})

	fmt.Println(result, err)
}

func (d DrugRepository) BuildRelations(drug *model.Drug) {
}
