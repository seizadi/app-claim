package graphdb

import (
	"fmt"
	
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

type Access struct {
	Read bool
	Write bool
}

type AccessService interface {
	Save(access Access, appId, resourceId string) (Resource, error)
	FindAllByAppId(appId string, page *Paging) ([]Resource, error)
	Delete(appId, resourceId string) (Resource, error)
}

type neo4jAccessService struct {
	driver neo4j.Driver
}

func NewAccessService(driver neo4j.Driver) AccessService {
	return &neo4jAccessService{driver: driver}
}

// Save should create a `:ACCESS` relationship between
// the App and Resource ID nodes provided.
//
// If either the app or resource cannot be found, a `NotFoundError` should be thrown.

//func (fs *neo4jAccessService) Save(appId, resourceId string) (_ Resource, err error) {
//	// Open a new Session
//	session := fs.driver.NewSession(neo4j.SessionConfig{})
//	defer func() {
//		err = DeferredClose(session, err)
//	}()
//
//	// Create ACCESS relationship within a write Transaction
//	resource, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
//		result, err := tx.Run(`
//				MATCH (u:App {appId: $appId})
//				MATCH (m:Resource {resourceId: $resourceId})
//
//				MERGE (u)-[r:ACCESS]->(m)
//						ON CREATE SET u.createdAt = datetime()
//
//				RETURN m {
//					.*,
//					favorite: true
//				} AS resource
//		`, map[string]interface{}{
//			"appId":  appId,
//			"resourceId": resourceId,
//		})
//		if err != nil {
//			return nil, err
//		}
//
//		record, err := result.Single()
//		if err != nil {
//			return nil, err
//		}
//		resource, _ := record.Get("resource")
//		return resource.(map[string]interface{}), nil
//	})
//
//	// Throw an error if the app or resource could not be found
//	if err != nil {
//		return nil, err
//	}
//
//	// Return resource details and `favorite` property
//	return resource.(Resource), nil
//}

// Save adds a relationship between an App and Resource with Access properties.
// The Access properties will be converted to a Neo4j types.
//
// If the App or Resource cannot be found, a NotFoundError should be thrown
func (fs *neo4jAccessService) Save(access Access, appId, resourceId string) (_ Resource, err error) {
	// Open a new session
	session := fs.driver.NewSession(neo4j.SessionConfig{})
	defer func() {
		err = DeferredClose(session, err)
	}()
	
	// Save the rating in the database
	result, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run(`
				MATCH (u:App {appId: $appId})
				MATCH (m:Resource {resourceId: $resourceId})

				MERGE (u)-[r:ACCESS]->(m)
				SET r.write = $write, r.read = $read, r.timestamp = timestamp()

				RETURN m { .*, write: r.write, read: r.read } AS resource
		`, map[string]interface{}{
			"appId":  appId,
			"resourceId": resourceId,
			"write":  access.Write,
			"read":  access.Read,
		})
		
		// Handle error from driver
		if err != nil {
			return nil, err
		}
		
		// Get the one and only record
		record, err := result.Single()
		if err != nil {
			return nil, err
		}
		
		// Extract resource properties
		resource, _ := record.Get("resource")
		return resource.(map[string]interface{}), nil
	})
	
	// Handle Errors from the Unit of Work
	if err != nil {
		return nil, err
	}
	
	// Return resource details and access properties
	return result.(Resource), nil
}

// FindAllByAppId should retrieve a list of resources that have an incoming :ACCESS
// relationship from a App node with the supplied `appId`.
//
// Results should be ordered by the `sort` parameter, and in the direction specified
// in the `order` parameter.
// Results should be limited to the number passed as `limit`.
// The `skip` variable should be used to skip a certain number of rows.
func (fs *neo4jAccessService) FindAllByAppId(appId string, page *Paging) (_ []Resource, err error) {
	// Open a new Session
	session := fs.driver.NewSession(neo4j.SessionConfig{})
	defer func() {
		err = DeferredClose(session, err)
	}()
	
	// Execute the query
	results, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run(fmt.Sprintf(`
			MATCH (u:App {appId: $appId})-[r:ACCESS]->(m:Resource)
			RETURN m {
				.*,
				favorite: true
			} AS resource
			ORDER BY m.`+"`%s`"+` %s
			SKIP $skip
			LIMIT $limit`, page.Sort(), page.Order()),
			map[string]interface{}{
				"appId": appId,
				"skip":   page.Skip(),
				"limit":  page.Limit(),
			})
		
		if err != nil {
			return nil, err
		}
		
		// Consume the results
		records, err := result.Collect()
		if err != nil {
			return nil, err
		}
		
		var resources []map[string]interface{}
		for _, record := range records {
			resource, _ := record.Get("resource")
			resources = append(resources, resource.(map[string]interface{}))
		}
		return resources, nil
	})
	
	if err != nil {
		return nil, err
	}
	return results.([]Resource), nil
}

// Delete should remove the `:ACCESS` relationship between
// the App and Resource ID nodes provided.
// If either the app, resource or the relationship between them cannot be found,
// a `NotFoundError` should be thrown.
func (fs *neo4jAccessService) Delete(appId, resourceId string) (_ Resource, err error) {
	// Open a new session
	session := fs.driver.NewSession(neo4j.SessionConfig{})
	defer func() {
		err = DeferredClose(session, err)
	}()
	
	// Delete ACCESS relationship within a write Transaction
	resource, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run(`
				MATCH (u:App {appId: $appId})-[r:ACCESS]->(m:Resource {resourceId: $resourceId})
				DELETE r

				RETURN m {
					.*,
					favorite: false
				} AS resource
		`, map[string]interface{}{
			"appId":  appId,
			"resourceId": resourceId,
		})
		
		if err != nil {
			return nil, err
		}
		
		record, err := result.Single()
		if err != nil {
			return nil, err
		}
		resource, _ := record.Get("resource")
		return resource.(map[string]interface{}), nil
	})
	
	// Throw an error if the app or resource could not be found
	if err != nil {
		return nil, err
	}
	
	// Return resource details and `favorite` property
	return resource.(Resource), nil
}
