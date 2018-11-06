package utils

import (
	"regexp"
	"strings"
	"strconv"
)

var re = regexp.MustCompile("\\[[0-9]*\\]$")

func Get(path string, data map[string]interface{}) interface{} {
	return get(strings.Split(path, "."), data)
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