package belvedere

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"
)

var repGetTableName = regexp.MustCompile(`^.+\.([^.]*?)$`)
var repUppercaseLetter = regexp.MustCompile(`([^A-Z])([A-Z])`)

type (
	QueryBuilder interface {
		Insert(ctx context.Context, src interface{}) error
	}

	// Belvedere query builder struct
	Belvedere struct {
		db *sql.DB
	}

	tableInfo struct {
		Name        string
		ColumnValue reflect.Value
		ColumnInfo  reflect.Type
	}
)

func (ti *tableInfo) Values() ([]interface{}, error) {
	var values []interface{}
	for i := 0; i < ti.ColumnInfo.NumField(); i++ {
		f := ti.ColumnValue.Field(i)
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

// ColumnNames Retrieve comma-separated column names.
func (ti *tableInfo) ColumnNames() string {
	var buf bytes.Buffer
	for i := 0; i < ti.ColumnInfo.NumField(); i++ {
		columnName := camelToSnake(ti.ColumnInfo.Field(i).Name)
		isLast := i == ti.ColumnInfo.NumField()-1
		if isLast {
			buf.WriteString(columnName)
		} else {
			buf.WriteString(columnName + ",")
		}
	}

	return buf.String()
}

func camelToSnake(str string) string {
	str = repUppercaseLetter.ReplaceAllString(str, `$1,$2`)
	return strings.ToLower(strings.Replace(str, ",", "_", -1))
}

func getTableNameFromTypeName(typeName reflect.Type) string {
	tableName := repGetTableName.ReplaceAllString(typeName.String(), "$1")
	return camelToSnake(tableName)
}

// generateInsertQuery Generate an insert statement from the structure.
// If the field of the structure contains information about the column.
func newTableInfo(src interface{}) *tableInfo {
	srcType := reflect.TypeOf(src)
	tableName := getTableNameFromTypeName(srcType)
	v := reflect.Indirect(reflect.ValueOf(src))
	t := v.Type()

	return &tableInfo{
		Name:        tableName,
		ColumnValue: v,
		ColumnInfo:  t,
	}
}

// Insert
func (b *Belvedere) Insert(ctx context.Context, src interface{}) error {
	tableInfo := newTableInfo(src)
	columnNames := tableInfo.ColumnNames()
	fmt.Println(columnNames)

	return nil
}

func NewBelvedere(driver, dataSorceName string) (QueryBuilder, error) {
	db, e := sql.Open(driver, dataSorceName)
	if e != nil {
		return nil, e
	}

	e = db.Ping()
	if e != nil {
		return nil, e
	}

	return &Belvedere{db: db}, nil
}
