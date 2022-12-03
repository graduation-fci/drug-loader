package dependencies

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type DP struct {
	neo4jDriver neo4j.DriverWithContext
}

func NewDependencyInjection() *DP {
	neo4j := flag.String("neo4j", "neo4j://localhost:7687", "Used to define neo4j uri")
	neo4jUserName := flag.String("neo4jUsername", "neo4j", "username for neo4j")
	neo4jPassword := flag.String("neo4jPassword", "neo4j", "password for neo4j")
	flag.Parse()
	os.Setenv("neo4j", *neo4j)
	os.Setenv("neo4jusername", *neo4jUserName)
	os.Setenv("neo4jpassword", *neo4jPassword)

	return &DP{}
}

func (d *DP) WithNeo4j() *DP {
	uri := os.Getenv("neo4j")
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(os.Getenv("neo4jusername"), os.Getenv("neo4jpassword"), ""))
	if err != nil {
		log.Fatalf("error while connecting to neo4j with error %s", err.Error())
	}

	d.neo4jDriver = driver
	return d
}

func (d *DP) GetNeo4j() neo4j.DriverWithContext {
	if d.neo4jDriver == nil {
		d.WithNeo4j()
		return d.neo4jDriver
	}

	return d.neo4jDriver
}

func (d *DP) Shutdown() {
	ctx := context.Background()
	if d.neo4jDriver != nil {
		d.neo4jDriver.Close(ctx)
	}
}
