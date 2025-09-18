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

package query

import (
	"encoding/json"
	"strings"

	"github.com/mitchellh/mapstructure"

	"github.com/morningconsult/go-elasticsearch-alerts/command/alert"
	"github.com/morningconsult/go-elasticsearch-alerts/config"
	"github.com/morningconsult/go-elasticsearch-alerts/internal/jsonpath"
)

const hitsDelimiter = "\n----------------------------------------\n"

// process converts the raw response returned from Elasticsearch into a
// []*github.com/morningconsult/go-elasticsearch-alerts/command/alert.Record
// array and returns that array, the response fields grouped by
// *QueryHandler.bodyField (if any), and an error if there was an error.
// If process returns a non-nil error, the other returned values will
// be nil.
//
//nolint:gocognit
func (q *QueryHandler) process(
	respData map[string]any,
) ([]*alert.Record, []map[string]any, error) {
	if len(q.conditions) != 0 && !config.ConditionsMet(q.logger.Named("conditions"), respData, q.conditions) {
		return nil, nil, nil
	}

	records := make([]*alert.Record, 0)
	for _, filter := range q.filters {
		elems := jsonpath.GetAll(respData, filter)
		if len(elems) < 1 {
			continue
		}

		fields, err := q.gatherFields(elems)
		if err != nil {
			return nil, nil, err
		}

		if len(fields) < 1 {
			continue
		}

		record := &alert.Record{
			Filter: filter,
			Fields: fields,
		}

		records = append(records, record)
	}

	// Get the body field
	body := jsonpath.GetAll(respData, q.bodyField)
	if body == nil {
		return records, nil, nil
	}

	stringifiedHits, hits, err := q.gatherHits(body)
	if err != nil {
		return nil, nil, err
	}

	if len(stringifiedHits) > 0 {
		record := &alert.Record{
			Filter:    q.bodyField,
			Text:      strings.Join(stringifiedHits, hitsDelimiter),
			BodyField: true,
		}
		records = append(records, record)
	}

	return records, hits, nil
}

func (q *QueryHandler) gatherHits(body []any) ([]string, []map[string]any, error) {
	stringifiedHits := make([]string, 0, len(body))
	hits := make([]map[string]any, 0, len(body))
	for _, elem := range body {
		hit, ok := elem.(map[string]any)
		if !ok {
			continue
		}

		hits = append(hits, hit)

		data, err := json.MarshalIndent(hit, "", "    ")
		if err != nil {
			return nil, nil, err
		}
		stringifiedHits = append(stringifiedHits, string(data))
	}
	return stringifiedHits, hits, nil
}

func (q *QueryHandler) gatherFields(elems []any) ([]*alert.Field, error) {
	fields := make([]*alert.Field, 0, len(elems))
	for _, elem := range elems {
		obj, ok := elem.(map[string]any)
		if !ok {
			continue
		}

		field := new(alert.Field)
		if err := mapstructure.Decode(obj, field); err != nil {
			return nil, err
		}

		if field.Key == "" || field.Count < 1 {
			continue
		}

		fields = append(fields, field)
	}
	return fields, nil
}
