package repository

import (
	"context"
	"log"
	"strings"

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

		//"CREATE (n:Drug {name: $name }) RETURN n.name"


		records, err := tx.Run(ctx, "MERGE (n:Drug {name: $name }) RETURN n.name", map[string]any{
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
	log.Println("INSERTED " + name + " AS AN NODE")
}

func (d DrugRepository) BuildRelations(drug *model.Drug) {
	ctx := context.TODO()
	session := d.graph.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)
	for _, interaction := range drug.Interactions {
		_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {

			// "MATCH (a:Drug {name: $src}) MATCH(b:Drug {name: $dest}) CREATE (a)-[rel:INTERACTS {consumerEffect: $consumerEffect, professionalEffect: $professionalEffect, hashedName: $hashedName, level: $level}]->(b) RETURN a.name"
			

			// "MATCH (a:Drug {name: $src})
			// MATCH (b:Drug {name: $dest})
			// MERGE (a)-[rel:INTERACTS {hashedName: $hashedName}]->(b)
			// ON CREATE SET rel.consumerEffect = $consumerEffect, rel.professionalEffect = $professionalEffect, rel.level = $level
			// ON MATCH SET rel.consumerEffect = $consumerEffect, rel.professionalEffect = $professionalEffect, rel.level = $level
			// RETURN a.name"


			records, err := tx.Run(ctx, "MATCH (a:Drug {name: $src}) MATCH (b:Drug {name: $dest}) MERGE (a)-[rel:INTERACTS {hashedName: $hashedName}]->(b) ON CREATE SET rel.consumerEffect = $consumerEffect, rel.professionalEffect = $professionalEffect, rel.level = $level RETURN a.name", map[string]any{
				"src":                strings.ToLower(drug.Name),
				"dest":               strings.ToLower(interaction.Name),
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
			log.Println("$src", strings.ToLower(drug.Name), "$dest", strings.ToLower(interaction.Name))
			log.Println("Inserting Node and retrying")
			d.InsetNode(strings.ToLower(interaction.Name))
			retryDrug := &model.Drug{
				Name: drug.Name,
				Interactions: []model.Interaction{
					interaction,
				},
			}
			d.BuildRelations(retryDrug)
		}
	}
}
