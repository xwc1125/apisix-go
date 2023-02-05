// Package serve
//
// @author: xwc1125
package serve

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/tidwall/gjson"
)

func (p *ProxyServe) initSchema(dir string) {
	var (
		apisixSchemaPath       = dir + "/conf/schema.json"
		apisixSchemaContent    []byte
		customizeSchemaContent []byte
		err                    error
	)

	if apisixSchemaContent, err = ioutil.ReadFile(apisixSchemaPath); err != nil {
		panic(fmt.Errorf("fail to read configuration: %s, error: %s", apisixSchemaPath, err.Error()))
	}

	content, err := mergeSchema(apisixSchemaContent, customizeSchemaContent)
	if err != nil {
		panic(err)
	}

	p.schema = gjson.ParseBytes(content)
}

func mergeSchema(apisixSchema, customizeSchema []byte) ([]byte, error) {
	var (
		apisixSchemaMap    map[string]map[string]interface{}
		customizeSchemaMap map[string]map[string]interface{}
	)
	if len(customizeSchema) == 0 {
		return apisixSchema, nil
	}

	if err := json.Unmarshal(apisixSchema, &apisixSchemaMap); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(customizeSchema, &customizeSchemaMap); err != nil {
		return nil, err
	}

	for key := range apisixSchemaMap["main"] {
		if _, ok := customizeSchemaMap["main"][key]; ok {
			return nil, fmt.Errorf("duplicates key: main.%s between schema.json and customize_schema.json", key)
		}
	}

	for k, v := range customizeSchemaMap["main"] {
		apisixSchemaMap["main"][k] = v
	}

	return json.Marshal(apisixSchemaMap)
}
