package graphdb

import "github.com/neo4j/neo4j-go-driver/v4/neo4j"

func NewDriver()(neo4j.Driver, error){
	// TODO - These should come from config
	uri := "neo4j://localhost:7687"
	username := "neo4j"
	password := "s3cr3t"
	
	driver, err := neo4j.NewDriver(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		return driver, err
	}
	
	return driver, nil
}

