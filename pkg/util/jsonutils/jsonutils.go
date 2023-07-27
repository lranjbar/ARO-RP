package jsonutils

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"encoding/json"
	"fmt"
)

// UpdateJsonString Updates key value pairs in a json string
func UpdateJsonString(jsonString string, jsonKeyValuePair map[string]string) (string, bool, error) {
	changed := false
	var jsonMap map[string]string

	err := json.Unmarshal([]byte(jsonString), &jsonMap)
	if err != nil {
		return "", changed, err
	}

	for k, v := range jsonKeyValuePair {
		// check for key in map
		if val, ok := jsonMap[k]; ok {
			// update existing key
			if val != v {
				jsonMap[k] = v
				changed = true
			}
		} else {
			// key does not exist
			return jsonString, changed, fmt.Errorf("key %s does not exist in json string", k)
		}
	}

	jsonStringByte, err := json.Marshal(jsonMap)

	return string(jsonStringByte), changed, err
}
