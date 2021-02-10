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

package config

import (
	"encoding/json"
	"errors"
	"github.com/morningconsult/go-elasticsearch-alerts/utils"
	"math"
	"sort"
	"strconv"
	"sync"

	hclog "github.com/hashicorp/go-hclog"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/shopspring/decimal"
	"golang.org/x/xerrors"
)

const (
	keyCommonField  = "comonfield"
	keyFiltersField = "filtersfield"
	keyQuantifier   = "quantifier"

	quantifierAny  = "any"
	quantifierAll  = "all"
	quantifierNone = "none"

	operatorEqual                = "eq"
	operatorNotEqual             = "ne"
	operatorLessThan             = "lt"
	operatorLessThanOrEqualTo    = "le"
	operatorGreaterThan          = "gt"
	operatorGreaterThanOrEqualTo = "ge"

	keyType               = "type"
	typeStandardDeviation = "standardDeviation"

	volumeBuffer = 5
)

var (
	lastValue map[string][]int // для хранения последних значений (нужно для StandardDeviation)
	one       sync.Once
	mx        *sync.RWMutex
)

// Condition is an optional parameter that can be used to limit
// when alerts are triggered
type Condition map[string]interface{}

func (c Condition) field() string {
	return c[keyCommonField].(string)
}

func (c Condition) Fieldfier() string {
	if v, ok := c[keyFiltersField]; ok {
		return v.(string)
	} else {
		return ""
	}
}

func (c Condition) quantifier() string {
	return c[keyQuantifier].(string)
}

func (c Condition) getType() string {
	if v, ok := c[keyType]; ok {
		return v.(string)
	} else {
		return ""
	}
}

func (c Condition) validate() error {
	var allErrors *multierror.Error

	if err := c.validateField(); err != nil {
		allErrors = multierror.Append(allErrors, err)
	}

	if err := c.validateQuantifier(); err != nil {
		allErrors = multierror.Append(allErrors, err)
	}

	if errs := c.validateNumOperators(); len(errs) != 0 {
		allErrors = multierror.Append(allErrors, errs...)
	}

	if errs := c.validateMultiOperators(); len(errs) != 0 {
		allErrors = multierror.Append(allErrors, errs...)
	}

	return allErrors.ErrorOrNil()
}

func (c Condition) validateField() error {
	fieldRaw, fieldOK := c[keyCommonField]
	if !fieldOK {
		return errors.New("condition must have the field 'comonfield'")
	}

	v, ok := fieldRaw.(string)
	if !ok || v == "" {
		return errors.New("field 'field' of condition must not be empty")
	}

	return nil
}

func (c Condition) validateQuantifier() error {
	raw, ok := c[keyQuantifier]
	if !ok {
		c[keyQuantifier] = quantifierAny
		return nil
	}

	v, ok := raw.(string)
	if !ok {
		return errors.New("field 'quantifier' of condition must be a string")
	}

	if v != quantifierAny && v != quantifierAll && v != quantifierNone {
		return errors.New("field 'quantifier' of condition must either be 'any', 'all', or 'none'")
	}

	return nil
}

func (c Condition) validateNumOperators() []error {
	numOperators := []string{
		operatorLessThanOrEqualTo,
		operatorLessThan,
		operatorGreaterThan,
		operatorGreaterThanOrEqualTo,
	}

	errors := make([]error, 0)
	for _, operator := range numOperators {
		if raw, ok := c[operator]; ok {
			if v, ok := raw.(json.Number); !ok {
				errors = append(errors, xerrors.Errorf("value of operator '%s' should be a number", operator))
			} else if v.String() == "" {
				errors = append(errors, xerrors.Errorf("value of operator '%s' should not be empty", operator))
			}
		}
	}

	return errors
}

func (c Condition) validateMultiOperators() []error {
	strOrNumOperators := []string{
		operatorEqual,
		operatorNotEqual,
	}

	errors := make([]error, 0)
	for _, operator := range strOrNumOperators {
		if raw, ok := c[operator]; ok {
			switch v := raw.(type) {
			case json.Number:
				if string(v) == "" {
					errors = append(errors, xerrors.Errorf("value of operator '%s' should not be empty", operator))
				}
			case string:
				if v == "" {
					errors = append(errors, xerrors.Errorf("value of operator '%s' should not be empty", operator))
				}
			default:
				errors = append(errors, xerrors.Errorf("value of operator '%s' should either be a number or a string", operator))
			}
		}
	}

	return errors
}

// ConditionsMet returns true if the response JSON meets the given conditions.
func ConditionsMet(logger hclog.Logger, resp map[string]interface{}, conditions []Condition) bool {
	for _, condition := range conditions {
		if !ConditionMet(logger, resp, condition, condition.field()) {
			logger.Info("Conditions false")
			return false
		}
	}

	logger.Info("Conditions true")
	return true
}

func ConditionMet(logger hclog.Logger, resp map[string]interface{}, condition Condition, fieldPath string) (res bool) {
	matches := []interface{}{ resp }
	if fieldPath != "" {
		matches = utils.GetAll(resp, fieldPath)
	}

	switch condition.quantifier() {
	case quantifierAll:
		res = allSatisfied(logger, matches, condition)
	case quantifierAny:
		res = anySatisfied(logger, matches, condition)
	case quantifierNone:
		res = noneSatisfied(logger, matches, condition)
	default:
		res = false
	}

	return
}

func allSatisfied(logger hclog.Logger, matches []interface{}, condition Condition) bool {
	for _, match := range matches {
		if match == nil {
			continue
		}

		sat := satisfied(logger, match, condition)
		if !sat {
			return false
		}
	}

	return true
}

func anySatisfied(logger hclog.Logger, matches []interface{}, condition Condition) bool {
	for _, match := range matches {
		if match == nil {
			continue
		}

		sat := satisfied(logger, match, condition)
		if sat {
			return true
		}
	}

	return false
}

func noneSatisfied(logger hclog.Logger, matches []interface{}, condition Condition) bool {
	for _, match := range matches {
		if match == nil {
			continue
		}

		sat := satisfied(logger, match, condition)
		if sat {
			return false
		}
	}

	return true
}

func satisfied(logger hclog.Logger, match interface{}, condition Condition) bool {
	switch v := match.(type) {
	case string:
		return stringSatisfied(v, condition)
	case json.Number:
		return numberSatisfied(v, condition)
	case bool:
		return boolSatisfied(v, condition)
	default:
		switch condition.getType() {
		case typeStandardDeviation:
			return standardDeviation(logger, v, condition)
		default:
			fields := make([]interface{}, 0, 4)
			if f, ok := condition[keyCommonField].(string); ok {
				fields = append(fields, "field", f)
			}

			if d, err := json.Marshal(match); err == nil {
				fields = append(fields, "value", string(d))
			} else {
				fields = append(fields, "value", match)
			}

			logger.Error("Value of field in Elasticsearch response is not a string, number, or boolean. Ignoring condition for this value", fields...) // nolint: lll
			return true
		}
	}
}

func numberSatisfied(k json.Number, condition Condition) bool { // nolint: gocyclo, gocognit
	d := decimal.RequireFromString(k.String())

	dec := decimal.RequireFromString

	sat := true

	if v, ok := condition[operatorEqual].(json.Number); ok {
		sat = sat && d.Equal(dec(string(v)))
	}

	if v, ok := condition[operatorNotEqual].(json.Number); ok {
		sat = sat && !d.Equal(dec(string(v)))
	}

	if v, ok := condition[operatorLessThan].(json.Number); ok {
		sat = sat && d.LessThan(dec(string(v)))
	}

	if v, ok := condition[operatorLessThanOrEqualTo].(json.Number); ok {
		sat = sat && d.LessThanOrEqual(dec(string(v)))
	}

	if v, ok := condition[operatorGreaterThan].(json.Number); ok {
		sat = sat && d.GreaterThan(dec(string(v)))
	}

	if v, ok := condition[operatorGreaterThanOrEqualTo].(json.Number); ok {
		sat = sat && d.GreaterThanOrEqual(dec(string(v)))
	}

	return sat
}

func stringSatisfied(s string, condition Condition) bool {
	sat := true

	if v, ok := condition[operatorEqual].(string); ok && v != "" {
		sat = sat && s == v
	}

	if v, ok := condition[operatorNotEqual].(string); ok && v != "" {
		sat = sat && s != v
	}

	return sat
}

func boolSatisfied(b bool, condition Condition) bool {
	sat := true

	if v, ok := condition[operatorEqual].(bool); ok {
		sat = sat && b == v
	}

	if v, ok := condition[operatorNotEqual].(bool); ok {
		sat = sat && b == v
	}

	return sat
}

func standardDeviation(logger hclog.Logger, i interface{}, condition Condition) bool {
	if data, ok := i.(map[string]interface{}); !ok {
		return false
	} else {
		if doc_count, err := strconv.Atoi(string(data["doc_count"].(json.Number))); err == nil {
			key := data["key"].(string)
			lv := setlastValue(key, doc_count)

			// Если текущее значение меньше чем предыдущее, значит произошло падение, на такое мы не реагируем. Потому что на резкие падения нам не нужно реагировать
			downturn := len(lv[key]) > 1 && lv[key][len(lv[key])-2] > doc_count

			dev := stDeviation(lv[key])
			m := mediana(lv[key])
			logger.With("key", key, "deviation", dev, "mediana", m, "data", lv[key], "downturn", downturn, "doc_count", doc_count).Info("standardDeviation")
			return !downturn && dev > m
		}
	}

	return false
}

func getlastValue() map[string][]int {
	one.Do(func() {
		lastValue = map[string][]int{}
		mx = new(sync.RWMutex)
	})
	return lastValue
}

func setlastValue(k string, v int) map[string][]int {
	lv := getlastValue()

	// если standardDeviation вызовится 2 раза подряд, то в массив lv[k] добавится еще раз тоже самое значение и это повлияет на расчет стандартного отклонения
	// по этому исключаем такие случаи вот так
	//if len(lv[k]) > 0 && lv[k][len(lv[k])-1] == v {
	//	return lv
	//}

	mx.Lock()
	defer mx.Unlock()

	lv[k] = append(lv[k], v)
	if len(lv[k]) > volumeBuffer {
		lv[k] = lv[k][:volumeBuffer]
	}

	return lv
}

func stDeviation(selection []int) float64 {
	var av, result float64
	for _, v := range selection {
		av += float64(v) / float64(len(selection))
	}
	for _, v := range selection {
		result += math.Pow(float64(v) - av, 2) / float64(len(selection)-1)
	}

	if math.IsNaN(result) {
		result = 0
	}

	return math.Sqrt(result)
}

func mediana(selection []int) float64 {
	sort.Ints(selection)

	if len(selection)%2 != 0 {
		return float64(selection[((len(selection)-1)/2)])
	} else {
		return float64(selection[(len(selection)/2)-1]+selection[(len(selection)/2)]) / 2
	}
}