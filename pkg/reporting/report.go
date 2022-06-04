package reporting

import (
	"errors"
	"fmt"

	"github.com/seizadi/app-claim/pkg/graphdb"
)

func Report(graphOptions string, stage string, env string, apps []string) error {
	// Initialize Driver and Services
	driver, err := graphdb.NewDriver(graphOptions)
	if err != nil {
		return err
	}

	appSvc := graphdb.NewAppService(driver)

	for _, app := range apps {
		// For app in the list query database and return result
		r, err := appSvc.QueryAppResources(stage, env, app)
		// Check for record not being found
		if domainError, ok := err.(*graphdb.DomainError); ok && domainError.StatusCode() == graphdb.NotFound {
			// application not found
			msg := fmt.Sprintf("application %s not found.", app)
			return errors.New(msg)
		}

		if err != nil {
			return err
		}

		fmt.Printf("%v", r)
	}
	return nil
}
