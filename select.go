package belvedere

import "bytes"

type (
	SelectOptionType string
	SelectOption     interface {
		Conditions() string
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
)

var selectOptionTypeWhere = SelectOptionType("where")
var selectOptionTypeLimit = SelectOptionType("limit")

func (st SelectOptionType) Equal(t SelectOptionType) bool {
	return t.String() == st.String()
}

func (st SelectOptionType) String() string {
	return string(st)
}

func (w *where) Conditions() string {
	return w.conditions
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

func buildWhereClause(selectOptions []SelectOption) (string, []interface{}) {
	if len(selectOptions) == 0 {
		return "", nil
	}
	var buf bytes.Buffer
	var values []interface{}
	buf.WriteString(" WHERE ")
	for _, option := range selectOptions {
		t := option.Type()
		if t.Equal(selectOptionTypeWhere) {
			buf.WriteString(option.Conditions())
			for _, v := range option.Params() {
				values = append(values, v)
			}
		}
	}
	return buf.String(), values
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
