package query

import (
	"encoding/json"
	"strings"

	"github.com/mitchellh/mapstructure"
	"gitlab.morningconsult.com/mci/go-elasticsearch-alerts/utils"
	"gitlab.morningconsult.com/mci/go-elasticsearch-alerts/command/alert"
)

func (q *QueryHandler) Transform(respData map[string]interface{}) ([]*alert.Record, error) {
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
				return nil, err
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
		return records, nil
	}

	hits, ok := hitsRaw.([]interface{})
	if !ok {
		return records, nil
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
			return nil, err
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
	return records, nil
}