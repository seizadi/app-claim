package commands

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/seizadi/app-claim/pkg/reporting"
	"io/ioutil"
	"log"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
)

// addSearch implements the search command
var addSearch = &cobra.Command{
	Use:   "search",
	Short: "search kubernetes manifest yaml",
	Long: `search kubernetes manifest yaml
It assumes that the directory supplied has the manifests
in it in YAML format.`,
	Run: func(cmd *cobra.Command, args []string) {

		dir, err := cmd.Flags().GetString("dir")
		if err != nil {
			fmt.Println(err)
			return
		}

		// If stage is specified search only selected
		stage, err := cmd.Flags().GetString("stage")
		if err != nil {
			fmt.Println(err)
			return
		}

		// If environment is specified search only selected
		env, err := cmd.Flags().GetString("env")
		if err != nil {
			fmt.Println(err)
			return
		}
		// If application is specified search only selected
		app, err := cmd.Flags().GetString("app")
		if err != nil {
			fmt.Println(err)
			return
		}

		// Check if we should save results to graphdb
		graphOptions, err := cmd.Flags().GetString("graphdb")
		if err != nil {
			fmt.Println(err)
			return
		}

		// Search
		r := NewSearchRunner()
		searchResults, err := r.SearchForTokens(args, dir, stage, env, app)
		if err != nil {
			fmt.Println(err)
			return
		}

		for k, v := range searchResults {
			fmt.Printf("%s %v\n", k, v)
		}

		if len(graphOptions) > 0 {
			err = reporting.Discover(graphOptions, searchResults)
			if err != nil {
				log.Print(err)
				return
			}
		}
	},
}

// DatabaseClaimReconciler reconciles a DatabaseClaim object
type SearchRunner struct {
	SearchResults map[string][]string
}

func NewSearchRunner() *SearchRunner {
	return &SearchRunner{
		SearchResults: map[string][]string{},
	}
}

func (r *SearchRunner) SearchForTokens(tokens []string, dir, stage, env, app string) (map[string][]string, error) {
	stageFound := false

	stages, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	for _, s := range stages {
		if s.IsDir() && !IsHiddenFile(s.Name()) {
			if len(stage) > 0 {
				if stage == s.Name() {
					stageFound = true
					r.SearchEnv(dir, stage, env, app, tokens)
				}
			} else {
				r.SearchEnv(dir, s.Name(), env, app, tokens)
			}
		}
	}

	if len(stage) > 0 && !stageFound {
		return nil, errors.New(fmt.Sprintf("stage %s not found\n", stage))
	}

	return r.SearchResults, nil
}

func init() {
	rootCmd.AddCommand(addSearch)
	addSearch.Flags().StringP("dir", "d", "", "search directory")
	addSearch.Flags().StringP("stage", "s", "", "search stage")
	addSearch.Flags().StringP("env", "e", "", "search environment")
	addSearch.Flags().StringP("app", "a", "", "search application")
	addSearch.Flags().StringP("graphdb", "g", "", "use graph database")
	//addSearch.Flags().StringP("s3", "s3", "", "search for s3 buckets")
}

func (r *SearchRunner) SearchEnv(dir string, stage string, env string, app string, args []string) {
	envFound := false
	path := dir + "/" + stage
	envs, err := ioutil.ReadDir(path)
	if err != nil {
		log.Print(err)
		return
	}

	for _, e := range envs {
		if e.IsDir() && !IsHiddenFile(e.Name()) {
			if len(env) > 0 {
				if env == e.Name() {
					envFound = true
					r.SearchManifest(dir, stage, env, app, args)
				}
			} else {
				r.SearchManifest(dir, stage, e.Name(), app, args)
			}
		}
	}

	if len(env) > 0 && !envFound {
		log.Printf("environment %s not found\n", env)
	}
}

func GetManifest(filename string) (*[]map[interface{}]interface{}, error) {
	var out []map[interface{}]interface{}
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Print(err)
		return &out, err
	}

	dec := yaml.NewDecoder(bytes.NewReader(source))

	for {
		m := make(map[interface{}]interface{})
		if dec.Decode(&m) != nil {
			break
		}
		out = append(out, m)
	}

	return &out, err
}

func (r *SearchRunner) SearchManifest(dir string, stage string, env string, app string, args []string) ([]string, error) {
	var searchOut []string
	apps, err := ioutil.ReadDir(dir + "/" + stage + "/" + env)
	if err != nil {
		log.Print(err)
		return searchOut, err
	}

	rootPath := dir + "/" + stage + "/" + env + "/"
	for _, a := range apps {
		if a.IsDir() && !IsHiddenFile(a.Name()) {
			out, err := GetManifest(rootPath + a.Name() + "/manifest.yaml")
			if err != nil {
				fmt.Printf("Got error %s\n", err.Error())
				return searchOut, err
			}

			for _, m := range *out {
				r.recurseSearch(m, stage, env, a.Name(), args)
			}
		}
	}
	return searchOut, nil
}

func (r *SearchRunner) recurseSearch(m interface{}, stage string, env string, app string, args []string) {
	reflectM := reflect.ValueOf(m)

	switch reflectM.Kind() {
	case reflect.String:
		v := reflectM.String()
		r.SearchMatch(v, stage, env, app, args)

	case reflect.Slice:
		for i := 0; i < reflectM.Len(); i++ {
			r.recurseSearch(reflectM.Index(i), stage, env, app, args)
		}

	case reflect.Map:
		for _, key := range reflectM.MapKeys() {
			strct := reflectM.MapIndex(key)
			value := reflect.ValueOf(strct.Interface())
			if value.Kind() == reflect.String {
				r.recurseSearch(fmt.Sprintf("%v:%s", key.Interface(), value.String()), stage, env, app, args)
			} else {
				// Search Key for match
				r.SearchMatch(fmt.Sprintf("%v", key.Interface()), stage, env, app, args)
				r.recurseSearch(strct.Interface(), stage, env, app, args)
			}
		}
	}
}

func (r *SearchRunner) SearchMatch(s string, stage string, env string, app string, args []string) {
	for _, a := range args {
		if strings.Contains(strings.ToLower(s), a) {
			if len(s) < 128 {
				// fmt.Printf("%s/%s/%s %s %s\n", stage, env, app, a, s)
				tag := fmt.Sprintf("%s/%s/%s", stage, env, app)
				r.SearchResults[s] = append(r.SearchResults[s], tag)
			}
		}
	}
}

func IsHiddenFile(filename string) bool {
	return filename[0:1] == "."
}
