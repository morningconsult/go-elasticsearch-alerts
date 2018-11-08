package utils

import (
	"regexp"
	"strings"
	"strconv"
)

var re = regexp.MustCompile("\\[[0-9]*\\]$")

func Get(json map[string]interface{}, path string) interface{} {
	return get(strings.Split(path, "."), json)
}

func get(indices []string, data interface{}) interface{} {
	var err error

	if len(indices) < 1 {
		return data
	}

	m, ok := data.(map[string]interface{})
	if !ok {
		return nil
	}

	first := indices[0]
	i := -1
	if idx := re.FindString(first); idx != "" {
		i, err = strconv.Atoi(strings.Trim(strings.Trim(idx, "["), "]"))
		if err != nil {
			return nil
		}
		first = first[:strings.Index(first, "[")]
	}

	elem := m[first]
	if i == -1 {
		return get(dequeue(indices), elem)
	}

	list, ok := elem.([]interface{})
	if !ok {
		return nil
	}
	if len(list) < i + 1 {
		return nil
	}
	return get(dequeue(indices), list[i])
}

func dequeue(is []string) []string {
	if len(is) < 2 {
		return []string{}
	}
	return is[1:]
}

func GetAll(json map[string]interface{}, path string) []interface{} {
	raw := getall(0, strings.Split(path, "."), json, "")
	if v, ok := raw.([]interface{}); ok {
		return v
	}
	return []interface{}{raw}
}

func getall(i int, stack []string, elem interface{}, keychain string) interface{} {
	if i > len(stack) - 1 {
		if list, ok := elem.([]interface{}); ok {
			var mod []interface{}
			for _, e := range list {
				mod = append(mod, addkey(e, keychain))
			}
			return mod
		}
		if m, ok := elem.(map[string]interface{}); ok {
			return addkey(m, keychain)
		}
		return elem
	}

	key := stack[i]

	if m, ok := elem.(map[string]interface{}); ok {
		v, ok := m[key]
		if !ok {
			return nil
		}
		i += 1
		return getall(i, stack, v, keychain)
	}

	buckets, ok := elem.([]interface{})
	if !ok {
		return nil
	}

	var mod []interface{}
	for _, item := range buckets {
		kc := keychain
		if e, ok := item.(map[string]interface{}); ok {
			if k, ok := e["key"].(string); ok {
				if kc == "" {
					kc = k
				} else {
					kc = kc+" - "+k
				}
			}
		}

		a := getall(i, stack, item, kc)
		if v, ok := a.(map[string]interface{}); ok {
			mod = append(mod, v)
		}
		if v, ok := a.([]interface{}); ok {
			mod = append(mod, v...)
		}
	}
	return mod
}

func addkey(i interface{}, keychain string) interface{} {
	obj, ok := i.(map[string]interface{})
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
		obj["key"] = keychain+" - "+key
	}
	return obj
}
