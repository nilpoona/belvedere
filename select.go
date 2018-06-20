package belvedere

import (
	"bytes"
	"errors"
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

func (l *limit) Conditions() (string, error) {
	return l.conditions, nil
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
	return selectOptionTypeWhere
}

func buildWhereClause(selectOptions []SelectOption) (string, []interface{}, error) {
	var buf bytes.Buffer
	var values []interface{}
	if len(selectOptions) == 0 {
		return "", values, nil
	}

	buf.WriteString(" WHERE ")
	for _, option := range selectOptions {
		t := option.Type()
		if t.Equal(selectOptionTypeWhere) {
			condition, err := option.Conditions()
			if err != nil {
				return "", values, err
			}
			buf.WriteString(condition)
			for _, v := range option.Params() {
				values = append(values, v)
			}
		}
	}
	return buf.String(), values, nil
}

func buildLimitClause(o SelectOption) (string, []interface{}, error) {
	if o == nil {
		return "", []interface{}{}, nil
	}

	if o.Type() != selectOptionTypeLimit {
		return "", []interface{}{}, errors.New("It is not a type limit")
	}

	conditions, err := o.Conditions()
	if err != nil {
		return "", []interface{}{}, err
	}

	params := o.Params()
	p := params[0]
	return conditions, []interface{}{p}, nil
}

func newSelectOption(optionFncs ...NewSelectOption) (wheres []SelectOption, limit SelectOption) {
	for _, optionFnc := range optionFncs {
		option := optionFnc()
		t := option.Type()
		if t == selectOptionTypeWhere {
			wheres = append(wheres, option)
		} else if t == selectOptionTypeLimit {
			limit = option
		}
	}

	return wheres, limit
}

func Where(conditions string, args ...interface{}) NewSelectOption {
	return func() SelectOption {
		return &where{
			conditions: conditions,
			args:       args,
		}
	}
}

func Limit(p int) NewSelectOption {
	return func() SelectOption {
		return &limit{
			conditions: " LIMIT ?",
			args: []interface{}{
				p,
			},
		}
	}
}

func IN(conditions string, args ...interface{}) NewSelectOption {
	return func() SelectOption {
		return &whereIn{
			conditions: conditions,
			args:       args,
		}
	}
}
