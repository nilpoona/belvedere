package belvedere

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"time"
)

type (
	tableInfo struct {
		Name        string
		Pk          pk
		ColumnValue reflect.Value
		ColumnInfo  reflect.Type
	}

	pk struct {
		Name  string
		Index int
	}
)

func (p pk) SameName(name string) bool {
	return name == p.Name
}

func (p pk) SameIndex(index int) bool {
	return index == p.Index
}
func fieldPts(t reflect.Type, v reflect.Value) ([]interface{}, error) {
	var pts []interface{}
	for i := 0; i < t.NumField(); i++ {
		f := v.Field(i)
		// TODO: JSON Type
		if f.IsValid() {
			switch f.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				value := f.Int()
				pts = append(pts, &value)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
				value := f.Uint()
				pts = append(pts, &value)
			case reflect.Float32, reflect.Float64:
				value := f.Float()
				pts = append(pts, &value)
			case reflect.String:
				value := f.String()
				pts = append(pts, &value)
			case reflect.Bool:
				var value int
				pts = append(pts, &value)
			case reflect.Struct:
				i := f.Interface()
				if f.Type().String() == "time.Time" {
					if value, ok := i.(time.Time); ok {
						pts = append(pts, &value)
					} else {
						return nil, errors.New("cannot convert this type")
					}
				} else {
					return nil, errors.New("cannot convert this type")
				}
			default:
				fmt.Println(f.Kind())
			}
		}
	}

	return pts, nil
}

func (ti *tableInfo) FieldPts() ([]interface{}, error) {
	return fieldPts(ti.ColumnInfo, ti.ColumnValue)
}

func (ti *tableInfo) PkValue() (interface{}, error) {
	f := ti.ColumnValue.Field(ti.Pk.Index)
	if f.IsValid() {
		switch f.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return f.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return f.Uint(), nil
		case reflect.Float32, reflect.Float64:
			return f.Float(), nil
		case reflect.String:
			return f.String(), nil
		case reflect.Bool:
			var value int
			return value, nil
		case reflect.Struct:
			i := f.Interface()
			if f.Type().String() == "time.Time" {
				if value, ok := i.(time.Time); ok {
					return value, nil
				} else {
					return nil, errors.New("cannot convert this type")
				}
			} else {
				return nil, errors.New("cannot convert this type")
			}
		default:
			return nil, errors.New("cannot convert this type")
		}
	}

	return nil, errors.New("cannot convert this type")
}

func (ti *tableInfo) Values(excludePk bool) ([]interface{}, error) {
	var values []interface{}
	for i := 0; i < ti.ColumnInfo.NumField(); i++ {
		f := ti.ColumnValue.Field(i)
		if excludePk && ti.Pk.SameIndex(i) {
			continue
		}
		// TODO: JSON Type
		if f.IsValid() {
			switch f.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				values = append(values, f.Int())
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
				values = append(values, f.Uint())
			case reflect.Float32, reflect.Float64:
				values = append(values, f.Float())
			case reflect.String:
				values = append(values, f.String())
			case reflect.Bool:
				r := f.Bool()
				if r {
					values = append(values, 1)
				} else {
					values = append(values, 0)
				}
			case reflect.Struct:
				i := f.Interface()
				if f.Type().String() == "time.Time" {
					if value, ok := i.(time.Time); ok {
						values = append(values, value)
					} else {
						return nil, errors.New("cannot convert this type")
					}
				} else {
					return nil, errors.New("cannot convert this type")
				}
			default:
				fmt.Println(f.Kind())
			}
		}
	}

	return values, nil
}

func (ti *tableInfo) StatementString(excludePk bool) string {
	valuesNum := ti.ColumnValue.NumField()
	if excludePk {
		valuesNum = valuesNum - 1
	}

	var buf []byte
	for i := 0; i < valuesNum; i++ {
		buf = append(buf, '?')
		if i != valuesNum-1 {
			buf = append(buf, ',')
		}
	}

	return string(buf)
}

func (ti *tableInfo) SetValue(values []interface{}) {
	for i := 0; i < ti.ColumnInfo.NumField(); i++ {
		f := ti.ColumnInfo.Field(i)
		fv := ti.ColumnValue.Field(i)
		v := values[i]
		rv := reflect.ValueOf(v).Elem()
		if !f.Anonymous {
			fv.Set(rv)
		}
	}
}

// ColumnNames Retrieve comma-separated column names.
func (ti *tableInfo) ColumnNames(excludePk bool) string {
	var buf bytes.Buffer
	var columnNum int

	columnNum = ti.ColumnInfo.NumField()
	for i := 0; i < columnNum; i++ {
		columnName := camelToSnake(ti.ColumnInfo.Field(i).Name)
		if excludePk && ti.Pk.SameName(columnName) {
			continue
		}

		isLast := i == ti.ColumnInfo.NumField()-1
		if isLast {
			buf.WriteString(columnName)
		} else {
			buf.WriteString(columnName + ",")
		}
	}

	return buf.String()
}

// generateInsertQuery Generate an insert statement from the structure.
// If the field of the structure contains information about the column.
func newTableInfo(src interface{}) *tableInfo {
	srcType := reflect.TypeOf(src)
	tableName := getTableNameFromTypeName(srcType)
	v := reflect.Indirect(reflect.ValueOf(src))
	t := v.Type()

	var pkName string
	var pkIndex int
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		pk := field.Tag.Get("pk")
		if pk == "" {
			continue
		}

		pkName = field.Name
		pkIndex = i
	}

	pk := pk{
		Name:  camelToSnake(pkName),
		Index: pkIndex,
	}

	return &tableInfo{
		Name:        tableName,
		Pk:          pk,
		ColumnValue: v,
		ColumnInfo:  t,
	}
}
