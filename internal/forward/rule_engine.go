package forward

import (
	"sl651-platform/internal/model"
)

type RuleEngine struct{}

func NewRuleEngine() *RuleEngine {
	return &RuleEngine{}
}

func (re *RuleEngine) Match(rule *model.ForwardRule, data *model.DeviceData) bool {
	// 1. Device ID Filter
	if len(rule.Filter.DeviceIDs) > 0 {
		found := false
		for _, id := range rule.Filter.DeviceIDs {
			if id == data.DeviceID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// 2. Data Type Filter
	if len(rule.Filter.DataTypes) > 0 {
		found := false
		for _, dt := range rule.Filter.DataTypes {
			if dt == data.DataType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// 3. Condition Filter (SLT 324 requirement usually involves specific values)
	if len(rule.Filter.Conditions) > 0 {
		// Map values for easy lookup
		values := make(map[string]model.DataValue)
		for _, p := range data.Values {
			values[p.Tag] = p.Value
		}

		for _, cond := range rule.Filter.Conditions {
			val, ok := values[cond.Field]
			if !ok {
				return false // Field missing, condition fails
			}
			if !re.evaluate(cond, val) {
				return false
			}
		}
	}

	return true
}

func (re *RuleEngine) evaluate(cond model.Condition, actual model.DataValue) bool {
	switch cond.Operator {
	case model.OperatorEq:
		return re.compare(actual, cond.Value) == 0
	case model.OperatorGt:
		return re.compare(actual, cond.Value) > 0
	case model.OperatorLt:
		return re.compare(actual, cond.Value) < 0
	case model.OperatorGte:
		return re.compare(actual, cond.Value) >= 0
	case model.OperatorLte:
		return re.compare(actual, cond.Value) <= 0
	case model.OperatorNe:
		return re.compare(actual, cond.Value) != 0
	}
	return false
}

func (re *RuleEngine) compare(a, b model.DataValue) int {
	// Simple comparison for floats/ints
	va := re.toFloat(a)
	vb := re.toFloat(b)
	if va == nil || vb == nil {
		return 0 // Incomparable
	}
	if *va > *vb {
		return 1
	}
	if *va < *vb {
		return -1
	}
	return 0
}

func (re *RuleEngine) toFloat(v model.DataValue) *float64 {
	if v.Float != nil {
		return v.Float
	}
	if v.Int != nil {
		f := float64(*v.Int)
		return &f
	}
	return nil
}
