package belvedere

import (
	"bytes"
	"fmt"
	"strings"
)

type (
	SelectOptionType string
	SelectOption     interface {
		Conditions() (string, error)
		Params() []interface{}
		Type() SelectOptionType
	}

	NewSelectOption       func() SelectOption
	CreateSelectOptionFnc func(conditions string, args ...interface{}) NewSelectOption

	where struct {
		conditions string
		args       []interface{}
	}

	limit struct {
		conditions string
		args       []interface{}
	}

	whereIn struct {
		conditions string
		args       []interface{}
	}
)

var selectOptionTypeWhere = SelectOptionType("where")
var selectOptionTypeLimit = SelectOptionType("limit")
var selectOptionTypeWhereIn = SelectOptionType("whereIn")

func (st SelectOptionType) Equal(t SelectOptionType) bool {
	return t.String() == st.String()
}

func (st SelectOptionType) String() string {
	return string(st)
}

// where
func (w *where) Conditions() (string, error) {
	return w.conditions, nil
}

func (w *where) Params() []interface{} {
	return w.args
}

func (w *where) Type() SelectOptionType {
	return selectOptionTypeWhere
}

func (l *limit) Conditions() string {
	return l.conditions
}

func (l *limit) Params() []interface{} {
	return l.args
}

func (l *limit) Type() SelectOptionType {
	return selectOptionTypeLimit
}

// in
func (wi *whereIn) Conditions() (string, error) {
	length := len(wi.Params())
	qms := make([]string, length)
	for i := 0; i < length; i++ {
		qms[i] = "?"
	}

	phs := strings.Join(qms, ", ")
	return fmt.Sprintf("%s IN (%s)", wi.conditions, phs), nil
}

func (wi *whereIn) Params() []interface{} {
	return wi.args
}

func (wi *whereIn) Type() SelectOptionType {
	return selectOptionTypeWhereIn
}

func buildWhereClause(selectOptions []SelectOption) (string, []interface{}, error) {
	var buf bytes.Buffer
	var values []interface{}
	buf.WriteString(" WHERE ")
	for _, option := range selectOptions {
		t := option.Type()
		if t.Equal(selectOptionTypeWhere) {
			condition, err := option.Conditions()
			if err != nil {
				return "", nil, err
			}
			buf.WriteString(condition)
			for _, v := range option.Params() {
				values = append(values, v)
			}
		}
	}
	return buf.String(), values, nil
}

func newSelectOption(optionFncs ...NewSelectOption) []SelectOption {
	options := make([]SelectOption, len(optionFncs))
	for i, optionFnc := range optionFncs {
		option := optionFnc()
		options[i] = option
	}

	return options
}

func Where(conditions string, args ...interface{}) NewSelectOption {
	return func() SelectOption {
		return &where{
			conditions: conditions,
			args:       args,
		}
	}
}

/*
func Limit(limit int) NewSelectOption {
	return func() SelectOption {
		return &limit{
			conditions: "LIMIT ?",
			args: []interface{}{
				limit,
			},
		}
	}
}
*/

func IN(conditions string, args ...interface{}) NewSelectOption {
	return func() SelectOption {
		return &whereIn{
			conditions: conditions,
			args:       args,
		}
	}
}
