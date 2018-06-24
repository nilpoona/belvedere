package belvedere

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

type (
	SelectOptionType string
	OrderType        int
	SelectOption     interface {
		Conditions() (string, error)
		Type() SelectOptionType
		Params() []interface{}
	}

	SelectOptionMap map[SelectOptionType][]SelectOption

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

	order struct {
		conditions string
		oType      OrderType
	}

	offset struct {
		offset uint
	}

	and struct {
		newWheres []NewSelectOption
		som       SelectOptionMap
	}
)

var (
	selectOptionTypeWhere  = SelectOptionType("where")
	selectOptionTypeLimit  = SelectOptionType("limit")
	selectOptionTypeOrder  = SelectOptionType("order")
	selectOptionTypeOffset = SelectOptionType("offset")
)

const (
	OrderTypeDesc = OrderType(iota)
	OrderTypeAsc
)

func (o OrderType) String() string {
	if o == OrderTypeDesc {
		return "DESC"
	} else {
		return "ASC"
	}
}

func (som SelectOptionMap) Limit() SelectOption {
	if value, ok := som[selectOptionTypeLimit]; ok {
		return value[0]
	}

	return nil
}

func (som SelectOptionMap) Offset() SelectOption {
	if value, ok := som[selectOptionTypeOffset]; ok {
		return value[0]
	}

	return nil
}

func (som SelectOptionMap) Wheres() []SelectOption {
	if value, ok := som[selectOptionTypeWhere]; ok {
		return value
	}

	return nil
}

func (som SelectOptionMap) Order() SelectOption {
	if value, ok := som[selectOptionTypeOrder]; ok {
		return value[0]
	}

	return nil
}

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

// order
func (o *order) Conditions() (string, error) {
	return fmt.Sprintf(" ORDER BY `%s` %s", o.conditions, o.oType.String()), nil
}

func (o *order) Params() []interface{} {
	return []interface{}{}
}

func (o *order) Type() SelectOptionType {
	return selectOptionTypeOrder
}

// offset
func (o *offset) Conditions() (string, error) {
	return " OFFSET ?", nil
}

func (o *offset) Params() []interface{} {
	return []interface{}{
		o.offset,
	}
}

func (o *offset) Type() SelectOptionType {
	return selectOptionTypeOffset
}

// and
func (a *and) Conditions() (string, error) {
	var buf bytes.Buffer
	length := len(a.newWheres)
	if length == 0 {
		return "", nil
	}

	som := a.newSelectOptionMap()

	for i, w := range som.Wheres() {
		if w.Type() != selectOptionTypeWhere {
			return "", ErrDifferentOptionType
		}
		c, err := w.Conditions()
		if err != nil {
			return "", err
		}

		buf.WriteString(c)
		if i < length-1 {
			buf.WriteString(" AND ")
		}
	}

	return buf.String(), nil
}

func (a *and) newSelectOptionMap() SelectOptionMap {
	if len(a.som) > 0 {
		return a.som
	}

	som := newSelectOptionMap(a.newWheres...)
	a.som = som

	return som
}

func (a *and) Params() []interface{} {
	length := len(a.newWheres)
	if length == 0 {
		return []interface{}{}
	}

	var params []interface{}

	som := a.newSelectOptionMap()
	for _, w := range som.Wheres() {
		params = append(params, w.Params()...)
	}

	return params
}

func (a *and) Type() SelectOptionType {
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

func buildOrderClause(o SelectOption) (string, error) {
	if o == nil {
		return "", nil
	}

	return o.Conditions()
}

func buildOffsetClause(o SelectOption) (string, []interface{}, error) {
	if o == nil {
		return "", []interface{}{}, nil
	}

	conditions, _ := o.Conditions()
	return conditions, o.Params(), nil
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

func newSelectOptionMap(optionFncs ...NewSelectOption) SelectOptionMap {
	som := SelectOptionMap{}
	for _, optionFnc := range optionFncs {
		option := optionFnc()
		t := option.Type()
		var key SelectOptionType
		if t == selectOptionTypeWhere {
			key = selectOptionTypeWhere
		} else if t == selectOptionTypeLimit {
			key = selectOptionTypeLimit
		} else if t == selectOptionTypeOrder {
			key = selectOptionTypeOrder
		} else if t == selectOptionTypeOffset {
			key = selectOptionTypeOffset
		}

		som[key] = append(som[key], option)
	}

	return som
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

func Order(field string, oType OrderType) NewSelectOption {
	return func() SelectOption {
		return &order{
			conditions: field,
			oType:      oType,
		}
	}
}

func Offset(amount uint) NewSelectOption {
	return func() SelectOption {
		return &offset{
			offset: amount,
		}
	}
}

func And(neWheres ...NewSelectOption) NewSelectOption {
	return func() SelectOption {
		return &and{
			newWheres: neWheres,
		}
	}
}
