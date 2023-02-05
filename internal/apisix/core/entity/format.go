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
package entity

import (
	"errors"
	"strconv"
	"strings"
)

func mapKV2Node(key string, val float64) (*Node, error) {
	hp := strings.Split(key, ":")
	host := hp[0]
	//  according to APISIX upstream nodes policy, port is optional
	port := "0"

	if len(hp) > 2 {
		return nil, errors.New("invalid upstream node")
	} else if len(hp) == 2 {
		port = hp[1]
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		log().Error("parse port to int fail", "err", err)
		return nil, err
	}

	node := &Node{
		Host:   host,
		Port:   portInt,
		Weight: int(val),
	}

	return node, nil
}

func NodesFormat(obj interface{}) interface{} {
	nodes := make([]*Node, 0)
	switch objType := obj.(type) {
	case map[string]float64:
		log().Info("nodes type map float64", "type", objType)
		value := obj.(map[string]float64)
		for key, val := range value {
			node, err := mapKV2Node(key, val)
			if err != nil {
				return obj
			}
			nodes = append(nodes, node)
		}
		return nodes
	case map[string]interface{}:
		log().Info("nodes type map interface", "type", objType)
		value := obj.(map[string]interface{})
		for key, val := range value {
			node, err := mapKV2Node(key, val.(float64))
			if err != nil {
				return obj
			}
			nodes = append(nodes, node)
		}
		return nodes
	case []*Node:
		log().Info("nodes type array node", "type", objType)
		return obj
	case []interface{}:
		log().Info("nodes type []interface{}", "type", objType)
		list := obj.([]interface{})
		for _, v := range list {
			val := v.(map[string]interface{})
			node := &Node{
				Host:   val["host"].(string),
				Port:   int(val["port"].(float64)),
				Weight: int(val["weight"].(float64)),
			}
			if _, ok := val["priority"]; ok {
				node.Priority = int(val["priority"].(float64))
			}
			nodes = append(nodes, node)
		}
		return nodes
	}

	return obj
}
