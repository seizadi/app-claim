package graphdb

import (
	"errors"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"strings"
)

func NewDriver(graphOptions string) (neo4j.Driver, error) {
	// TODO - These should come from config
	graphValues := strings.Split(graphOptions, "@")
	if len(graphValues) < 2 {
		return nil, errors.New("graphdb value should be of form username:password@neo4j://<host>:<port>")
	}
	uri := graphValues[1]
	graphCreds := strings.Split(graphValues[0], ":")
	if len(graphCreds) < 2 {
		return nil, errors.New("graphdb value should be of form username:password@neo4j://<host>:<port>")
	}
	username := graphCreds[0]
	password := graphCreds[1]

	driver, err := neo4j.NewDriver(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		return driver, err
	}

	return driver, nil
}
