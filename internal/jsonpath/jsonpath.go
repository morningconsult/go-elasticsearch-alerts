// Copyright 2019 The Morning Consult, LLC or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//         https://www.apache.org/licenses/LICENSE-2.0
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package jsonpath

import (
	"strings"
)

// GetAll recursively traverses the JSON via the provided path
// and returns all elements matching the path. If no elements
// are found, it will return [<nil>].
func GetAll(json map[string]any, path string) []any {
	raw := getall(0, strings.Split(path, "."), json, "")
	if v, ok := raw.([]any); ok {
		return v
	}
	return []any{raw}
}

func getall(i int, stack []string, elem any, keychain string) any { //nolint:gocyclo,gocognit
	if i > len(stack)-1 {
		if list, ok := elem.([]any); ok {
			var mod []any
			for _, e := range list {
				mod = append(mod, addkey(e, keychain))
			}
			return mod
		}
		if m, ok := elem.(map[string]any); ok {
			return addkey(m, keychain)
		}
		return elem
	}

	key := stack[i]

	if m, ok := elem.(map[string]any); ok {
		v, ok := m[key]
		if !ok {
			return nil
		}
		i++
		return getall(i, stack, v, keychain)
	}

	buckets, ok := elem.([]any)
	if !ok {
		return nil
	}

	var mod []any
	for _, item := range buckets {
		kc := keychain
		if e, ok := item.(map[string]any); ok {
			if k, ok := e["key"].(string); ok {
				if kc == "" {
					kc = k
				} else {
					kc = kc + " - " + k
				}
			}
		}

		a := getall(i, stack, item, kc)
		switch v := a.(type) {
		case map[string]any:
			mod = append(mod, v)
		case []any:
			mod = append(mod, v...)
		case nil:
		default:
			mod = append(mod, a)
		}
	}
	return mod
}

func addkey(i any, keychain string) any {
	obj, ok := i.(map[string]any)
	if !ok {
		return i
	}
	key, ok := obj["key"].(string)
	if !ok {
		return obj
	}
	if key == "" {
		return obj
	}
	if keychain != "" {
		obj["key"] = keychain + " - " + key
	}
	return obj
}
