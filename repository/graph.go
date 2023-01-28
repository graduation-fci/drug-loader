package repository

import (
	"context"
	"log"

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
	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
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
	if err != nil {
		log.Println("Error while insertion; err: ", err)
		return
	}
	log.Println("INSERTED " + name + " AS AN EDGE")
}

func (d DrugRepository) BuildRelations(drug *model.Drug) {
	ctx := context.TODO()
	session := d.graph.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)
	for i, interaction := range drug.Interactions {
		_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
			records, err := tx.Run(ctx, "MATCH (a:Drug {name: $src}) MATCH(b:Drug {name: $dest}) CREATE (a)-[rel:INTERACTS {effect: $effect, level: $level}]->(b) RETURN a.name", map[string]any{
				"src":                drug.Name,
				"dest":               interaction.Name,
				"hashedName":         interaction.HashedName,
				"consumerEffect":     interaction.ConsumerEffect,
				"professionalEffect": interaction.ProfessionalEffect,
				"level":              interaction.Level,
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
		if err != nil {
			log.Printf("Error while building relation %d ; err: %s \n", i, err)
		}
	}
}
