/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/chain5j/logger"
	"github.com/tidwall/gjson"
	"github.com/xeipuuv/gojsonschema"
	entity2 "github.com/xwc1125/apisix-go/internal/apisix/core/entity"
	"go.uber.org/zap/buffer"
)

type Validator interface {
	Validate(obj interface{}) error
}
type JsonSchemaValidator struct {
	schema *gojsonschema.Schema
}

func NewJsonSchemaValidator(jsonPath string) (Validator, error) {
	bs, err := ioutil.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("get abs path failed: %s", err)
	}
	s, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(string(bs)))
	if err != nil {
		return nil, fmt.Errorf("new schema failed: %s", err)
	}
	return &JsonSchemaValidator{
		schema: s,
	}, nil
}

func NewSchemaValidator(schemaJson string) (Validator, error) {
	s, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(schemaJson))
	if err != nil {
		return nil, fmt.Errorf("new schema failed: %s", err)
	}
	return &JsonSchemaValidator{
		schema: s,
	}, nil
}

func (v *JsonSchemaValidator) Validate(obj interface{}) error {
	ret, err := v.schema.Validate(gojsonschema.NewGoLoader(obj))
	if err != nil {
		return fmt.Errorf("validate failed: %s", err)
	}

	if !ret.Valid() {
		errString := buffer.Buffer{}
		for i, vErr := range ret.Errors() {
			if i != 0 {
				errString.AppendString("\n")
			}
			errString.AppendString(vErr.String())
		}
		return errors.New(errString.String())
	}
	return nil
}

type APISIXJsonSchemaValidator struct {
	schema    *gojsonschema.Schema
	schemaDef string

	gschema gjson.Result
}

func NewAPISIXJsonSchemaValidator(schema gjson.Result, jsonPath string) (Validator, error) {
	schemaDef := schema.Get(jsonPath).String()
	if schemaDef == "" {
		logger.Error("schema validate failed: schema not found", "path", jsonPath)
		return nil, fmt.Errorf("schema validate failed: schema not found, path: %s", jsonPath)
	}

	s, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(schemaDef))
	if err != nil {
		logger.Error("new schema failed", "err", err)
		return nil, fmt.Errorf("new schema failed: %s", err)
	}
	return &APISIXJsonSchemaValidator{
		schema:    s,
		schemaDef: schemaDef,
		gschema:   schema,
	}, nil
}

func getPlugins(reqBody interface{}) (map[string]interface{}, string) {
	switch bodyType := reqBody.(type) {
	case *entity2.Route:
		logger.Info("type of reqBody", "bodyType", bodyType)
		route := reqBody.(*entity2.Route)
		return route.Plugins, "schema"
	case *entity2.Service:
		logger.Info("type of reqBody", "bodyType", bodyType)
		service := reqBody.(*entity2.Service)
		return service.Plugins, "schema"
	case *entity2.Consumer:
		logger.Info("type of reqBody", "bodyType", bodyType)
		consumer := reqBody.(*entity2.Consumer)
		return consumer.Plugins, "consumer_schema"
	}
	return nil, ""
}

func cHashKeySchemaCheck(schema gjson.Result, upstream *entity2.UpstreamDef) error {
	if upstream.HashOn == "consumer" {
		return nil
	}
	if upstream.HashOn != "vars" &&
		upstream.HashOn != "header" &&
		upstream.HashOn != "cookie" {
		return fmt.Errorf("invalid hash_on type: %s", upstream.HashOn)
	}

	var schemaDef string
	if upstream.HashOn == "vars" {
		schemaDef = schema.Get("main.upstream_hash_vars_schema").String()
		if schemaDef == "" {
			return fmt.Errorf("schema validate failed: schema not found, path: main.upstream_hash_vars_schema")
		}
	}

	if upstream.HashOn == "header" || upstream.HashOn == "cookie" {
		schemaDef = schema.Get("main.upstream_hash_header_schema").String()
		if schemaDef == "" {
			return fmt.Errorf("schema validate failed: schema not found, path: main.upstream_hash_header_schema")
		}
	}

	s, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(schemaDef))
	if err != nil {
		return fmt.Errorf("schema validate failed: %s", err)
	}

	ret, err := s.Validate(gojsonschema.NewGoLoader(upstream.Key))
	if err != nil {
		return fmt.Errorf("schema validate failed: %s", err)
	}

	if !ret.Valid() {
		errString := buffer.Buffer{}
		for i, vErr := range ret.Errors() {
			if i != 0 {
				errString.AppendString("\n")
			}
			errString.AppendString(vErr.String())
		}
		return fmt.Errorf("schema validate failed: %s", errString.String())
	}

	return nil
}

func checkUpstream(schema gjson.Result, upstream *entity2.UpstreamDef) error {
	if upstream == nil {
		return nil
	}

	if upstream.PassHost == "node" && upstream.Nodes != nil {
		nodes, ok := entity2.NodesFormat(upstream.Nodes).([]*entity2.Node)
		if !ok {
			return fmt.Errorf("upstrams nodes not support value %v when `pass_host` is `node`", nodes)
		} else if len(nodes) != 1 {
			return fmt.Errorf("only support single node for `node` mode currentlywhen `pass_host` is `node`")
		}
	}

	if upstream.PassHost == "rewrite" && upstream.UpstreamHost == "" {
		return fmt.Errorf("`upstream_host` can't be empty when `pass_host` is `rewrite`")
	}

	if upstream.Type != "chash" {
		return nil
	}

	// to confirm
	if upstream.HashOn == "" {
		upstream.HashOn = "vars"
	}

	if upstream.HashOn != "consumer" && upstream.Key == "" {
		return fmt.Errorf("missing key")
	}

	if err := cHashKeySchemaCheck(schema, upstream); err != nil {
		return err
	}

	return nil
}

func checkRemoteAddr(remoteAddrs []string) error {
	for _, remoteAddr := range remoteAddrs {
		if remoteAddr == "" {
			return fmt.Errorf("schema validate failed: invalid field remote_addrs")
		}
	}
	return nil
}

func checkConf(schema gjson.Result, reqBody interface{}) error {
	switch bodyType := reqBody.(type) {
	case *entity2.Route:
		route := reqBody.(*entity2.Route)
		logger.Info("type of reqBody", "bodyType", bodyType)
		if err := checkUpstream(schema, route.Upstream); err != nil {
			return err
		}
		// todo: this is a temporary method, we'll drop it later
		if err := checkRemoteAddr(route.RemoteAddrs); err != nil {
			return err
		}
	case *entity2.Service:
		service := reqBody.(*entity2.Service)
		if err := checkUpstream(schema, service.Upstream); err != nil {
			return err
		}
	case *entity2.Upstream:
		upstream := reqBody.(*entity2.Upstream)
		if err := checkUpstream(schema, &upstream.UpstreamDef); err != nil {
			return err
		}
	}
	return nil
}

func (v *APISIXJsonSchemaValidator) Validate(obj interface{}) error {
	ret, err := v.schema.Validate(gojsonschema.NewGoLoader(obj))
	if err != nil {
		logger.Error("schema validate failed", "err", err, "schema", v.schema, "obj", obj)
		return fmt.Errorf("schema validate failed: %s", err)
	}

	if !ret.Valid() {
		errString := buffer.Buffer{}
		for i, vErr := range ret.Errors() {
			if i != 0 {
				errString.AppendString("\n")
			}
			errString.AppendString(vErr.String())
		}
		logger.Error("schema validate failed", "schemaDef", v.schemaDef, "obj", obj)
		return fmt.Errorf("schema validate failed: %s", errString.String())
	}

	// custom check
	if err := checkConf(v.gschema, obj); err != nil {
		return err
	}

	plugins, schemaType := getPlugins(obj)
	for pluginName, pluginConf := range plugins {
		schemaValue := v.gschema.Get("plugins." + pluginName + "." + schemaType).Value()
		if schemaValue == nil && schemaType == "consumer_schema" {
			schemaValue = v.gschema.Get("plugins." + pluginName + ".schema").Value()
		}

		if schemaValue == nil {
			logger.Error("schema validate failed: schema not found", "pluginName", "plugins."+pluginName, "schemaType", schemaType)
			return fmt.Errorf("schema validate failed: schema not found, path: %s", "plugins."+pluginName)
		}
		schemaMap := schemaValue.(map[string]interface{})
		schemaByte, err := json.Marshal(schemaMap)
		if err != nil {
			logger.Warn("schema validate failed: schema json encode failed", "path", "plugins."+pluginName, "err", err)
			return fmt.Errorf("schema validate failed: schema json encode failed, path: %s, %w", "plugins."+pluginName, err)
		}

		s, err := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(schemaByte))
		if err != nil {
			logger.Error("init schema validate failed", "err", err)
			return fmt.Errorf("schema validate failed: %s", err)
		}

		// check property disable, if is bool, remove from json schema checking
		conf := pluginConf.(map[string]interface{})
		var exchange bool
		disable, ok := conf["disable"]
		if ok {
			if fmt.Sprintf("%T", disable) == "bool" {
				delete(conf, "disable")
				exchange = true
			}
		}

		// check schema
		ret, err := s.Validate(gojsonschema.NewGoLoader(conf))
		if err != nil {
			logger.Error("schema validate failed", "err", err)
			return fmt.Errorf("schema validate failed: %s", err)
		}

		// put the value back to the property disable
		if exchange {
			conf["disable"] = disable
		}

		if !ret.Valid() {
			errString := buffer.Buffer{}
			for i, vErr := range ret.Errors() {
				if i != 0 {
					errString.AppendString("\n")
				}
				errString.AppendString(vErr.String())
			}
			return fmt.Errorf("schema validate failed: %s", errString.String())
		}
	}

	return nil
}

type APISIXSchemaValidator struct {
	schema *gojsonschema.Schema
}

func NewAPISIXSchemaValidator(schema gjson.Result, jsonPath string) (Validator, error) {
	schemaDef := schema.Get(jsonPath).String()
	if schemaDef == "" {
		logger.Warn("schema validate failed: schema not found", "path", jsonPath)
		return nil, fmt.Errorf("schema validate failed: schema not found, path: %s", jsonPath)
	}

	s, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(schemaDef))
	if err != nil {
		logger.Warn("new schema failed", "err", err)
		return nil, fmt.Errorf("new schema failed: %w", err)
	}
	return &APISIXSchemaValidator{
		schema: s,
	}, nil
}

func (v *APISIXSchemaValidator) Validate(obj interface{}) error {
	ret, err := v.schema.Validate(gojsonschema.NewBytesLoader(obj.([]byte)))
	if err != nil {
		logger.Warn("schema validate failed", "err", err)
		return fmt.Errorf("schema validate failed: %w", err)
	}

	if !ret.Valid() {
		errString := buffer.Buffer{}
		for i, vErr := range ret.Errors() {
			if i != 0 {
				errString.AppendString("\n")
			}
			errString.AppendString(vErr.String())
		}
		return fmt.Errorf("schema validate failed: %s", errString.String())
	}

	return nil
}
