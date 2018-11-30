// Copyright 2018 The Morning Consult, LLC or its affiliates. All Rights Reserved.
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

package query

import (
	"encoding/json"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert"
	"github.com/morningconsult/go-elasticsearch-alerts/utils"
)

func (q *QueryHandler) Transform(respData map[string]interface{}) ([]*alert.Record, int, error) {
	var records []*alert.Record
	for _, filter := range q.filters {
		elems := utils.GetAll(respData, filter)
		if elems == nil || len(elems) < 1 {
			continue
		}

		record := &alert.Record{
			Title: filter,
		}

		var fields []*alert.Field
		for _, elem := range elems {
			obj, ok := elem.(map[string]interface{})
			if !ok {
				continue
			}

			field := new(alert.Field)
			if err := mapstructure.Decode(obj, field); err != nil {
				return nil, -1, err
			}

			if field.Key == "" || field.Count < 1 {
				continue
			}

			fields = append(fields, field)
		}
		if len(fields) < 1 {
			continue
		}
		record.Fields = fields
		records = append(records, record)
	}

	// Make one record per hits.hits
	hitsRaw := utils.Get(respData, "hits.hits")
	if hitsRaw == nil {
		return records, 0, nil
	}

	hits, ok := hitsRaw.([]interface{})
	if !ok {
		return records, 0, nil
	}

	var hitsArr []string
	for _, hit := range hits {
		obj, ok := hit.(map[string]interface{})
		if !ok {
			continue
		}

		source, ok := obj["_source"].(map[string]interface{})
		if !ok {
			continue
		}

		data, err := json.MarshalIndent(source, "", "    ")
		if err != nil {
			return nil, -1, err
		}
		hitsArr = append(hitsArr, string(data))
	}

	if len(hitsArr) > 0 {
		record := &alert.Record{
			Title: "hits.hits._source",
			Text:  strings.Join(hitsArr, "\n----------------------------------------\n"),
		}
		records = append(records, record)
	}
	return records, len(hitsArr), nil
}
