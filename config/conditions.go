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

	hclog "github.com/hashicorp/go-hclog"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/shopspring/decimal"
	"golang.org/x/xerrors"

	"github.com/morningconsult/go-elasticsearch-alerts/utils"
)

const (
	keyField      = "field"
	keyQuantifier = "quantifier"

	quantifierAny  = "any"
	quantifierAll  = "all"
	quantifierNone = "none"

	operatorEqual                = "eq"
	operatorNotEqual             = "ne"
	operatorLessThan             = "lt"
	operatorLessThanOrEqualTo    = "le"
	operatorGreaterThan          = "gt"
	operatorGreaterThanOrEqualTo = "ge"
)

// Condition is an optional parameter that can be used to limit
// when alerts are triggered.
type Condition map[string]interface{}

func (c Condition) field() string {
	return c[keyField].(string)
}

func (c Condition) quantifier() string {
	return c[keyQuantifier].(string)
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
	raw, ok := c[keyField]
	if !ok {
		return errors.New("condition must have the field 'field'")
	}

	v, ok := raw.(string)
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
		matches := utils.GetAll(resp, condition.field())

		res := false

		switch condition.quantifier() {
		case quantifierAll:
			res = allSatisfied(logger, matches, condition)
		case quantifierAny:
			res = anySatisfied(logger, matches, condition)
		case quantifierNone:
			res = noneSatisfied(logger, matches, condition)
		}

		if !res {
			return false
		}
	}

	return true
}

func allSatisfied(logger hclog.Logger, matches []interface{}, condition Condition) bool {
	for _, match := range matches {
		sat := satisfied(logger, match, condition)
		if !sat {
			return false
		}
	}

	return true
}

func anySatisfied(logger hclog.Logger, matches []interface{}, condition Condition) bool {
	for _, match := range matches {
		sat := satisfied(logger, match, condition)
		if sat {
			return true
		}
	}

	return false
}

func noneSatisfied(logger hclog.Logger, matches []interface{}, condition Condition) bool {
	for _, match := range matches {
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
		fields := make([]interface{}, 0, 4)
		if f, ok := condition[keyField].(string); ok {
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

func numberSatisfied(k json.Number, condition Condition) bool { // nolint: gocyclo
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
