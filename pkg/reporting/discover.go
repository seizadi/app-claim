package reporting

import (
	"fmt"
	"strings"
	
	"github.com/seizadi/app-claim/pkg/graphdb"
)

func Discover(searchResults map[string][]string) error {
	// Initialize Driver and Services
	driver, err := graphdb.NewDriver()
	if err != nil {
		return err
	}
	
	appSvc := graphdb.NewAppService(driver)
	resourceSvc := graphdb.NewResourceService(driver)
	accessSvc := graphdb.NewAccessService(driver)
	
	for resource, apps := range searchResults {
		// For resource and apps create the nodes and access relationships
		fmt.Printf("%s %v\n", resource, apps)
		// Extract Resource name
		v := strings.Split(resource, ":")
		// Ignore this resource if we can not find a name
		if len(v) == 2 {
			resourceName := v[1]
			// Check if resource is already created?
			r, err := resourceSvc.FindOneByName(resourceName)
			// Check for record not being found
			if domainError, ok := err.(*graphdb.DomainError); ok && domainError.StatusCode() == graphdb.NotFound {
				// Create Resource
				r, err = resourceSvc.Save(resourceName)
			}
			
			if err != nil {
				return err
			}
			
			for _, app := range apps {
				// Check if app is already created?
				a, err := appSvc.FindOneByName(app)
				// Check for record not being found
				if domainError, ok := err.(*graphdb.DomainError); ok && domainError.StatusCode() == graphdb.NotFound {
					// Create App
					a, err = appSvc.Save(app)
				}
				
				if err != nil {
					return err
				}
				
				_, err = accessSvc.Save(graphdb.Access{Read: true}, a["appId"].(string), r["resourceId"].(string))
				if err != nil {
					//return err
				}
			}
		}

	}
	return nil
}
