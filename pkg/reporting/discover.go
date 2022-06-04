package reporting

import (
	"strings"

	"github.com/seizadi/app-claim/pkg/graphdb"
)

func Discover(graphOptions string, searchResults map[string][]string, kind string) error {
	// Initialize Driver and Services
	driver, err := graphdb.NewDriver(graphOptions)
	if err != nil {
		return err
	}

	appSvc := graphdb.NewAppService(driver)
	resourceSvc := graphdb.NewResourceService(driver)
	accessSvc := graphdb.NewAccessService(driver)

	for resource, apps := range searchResults {
		// For resource and apps create the nodes and access relationships
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
				r, err = resourceSvc.Save(resourceName, kind)
			}

			if err != nil {
				return err
			}

			for _, app := range apps {
				// Check if app is already created?
				a, err := appSvc.FindOneByName(app)
				// Check for record not being found
				if domainError, ok := err.(*graphdb.DomainError); ok && domainError.StatusCode() == graphdb.NotFound {
					// Create
					s := strings.Split(app, "/")
					appRecord := graphdb.App{
						Name:        app,
						Stage:       s[0],
						Environment: s[1],
						ShortName:   s[2],
					}
					a, err = appSvc.Save(appRecord)
				}

				if err != nil {
					return err
				}

				_, err = accessSvc.Save(graphdb.Access{Read: true}, a.AppId, r["resourceId"].(string))
				if err != nil {
					//return err
				}
			}
		}

	}
	return nil
}
