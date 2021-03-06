package belvedere

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type (
	QueryBuilder interface {
		Insert(ctx context.Context, src interface{}) (sql.Result, error)
		Update(ctx context.Context, src interface{}) (sql.Result, error)
		SelectOne(ctx context.Context, dst interface{}) error
		Select(ctx context.Context, dst interface{}, options ...NewSelectOption) error
		Count(ctx context.Context, fn string, dst interface{}, options ...NewSelectOption) (int, error)
	}

	// Belvedere query builder struct
	Belvedere struct {
		db *sql.DB
	}
)

var repGetTableName = regexp.MustCompile(`^.+\.([^.]*?)$`)
var repUppercaseLetter = regexp.MustCompile(`([^A-Z])([A-Z])`)

func camelToSnake(str string) string {
	str = repUppercaseLetter.ReplaceAllString(str, `$1,$2`)
	return strings.ToLower(strings.Replace(str, ",", "_", -1))
}

func getTableNameFromTypeName(typeName reflect.Type) string {
	tableName := repGetTableName.ReplaceAllString(typeName.String(), "$1")
	return camelToSnake(tableName)
}

func (b *Belvedere) DB() *sql.DB {
	return b.db
}

// Insert
func (b *Belvedere) Insert(ctx context.Context, src interface{}) (sql.Result, error) {
	tableInfo := newTableInfo(src)
	columnNames := tableInfo.ColumnNames(true)
	values, e := tableInfo.Values(true)

	if e != nil {
		return nil, e
	}

	statementString := tableInfo.StatementString(true)
	q := fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", tableInfo.Name, columnNames, statementString)

	stmt, e := b.db.PrepareContext(ctx, q)

	if e != nil {
		return nil, e
	}

	result, e := stmt.ExecContext(ctx, values...)

	if e != nil {
		return nil, e
	}

	return result, nil
}

func buildUpdateQuery(tableName, columnNames string, whereClause string) string {
	cns := strings.Split(columnNames, ",")
	length := len(cns)
	var b []byte
	b = append(b, "UPDATE "...)
	b = append(b, tableName...)
	b = append(b, " SET "...)

	for i, cn := range cns {
		b = append(b, cn...)
		b = append(b, " = ?"...)
		if i < length-1 {
			b = append(b, ", "...)
		}
	}

	b = append(b, whereClause...)

	return string(b)
}

func (b *Belvedere) Update(ctx context.Context, src interface{}) (sql.Result, error) {
	tableInfo := newTableInfo(src)
	columnNames := tableInfo.ColumnNames(true)
	values, e := tableInfo.Values(true)

	if e != nil {
		return nil, e
	}

	var conditions []byte
	conditions = append(conditions, tableInfo.Pk.Name...)
	conditions = append(conditions, " = ?"...)

	pkv, err := tableInfo.PkValue()
	if err != nil {
		return nil, err
	}

	newWhere := Where(string(conditions), pkv)
	w := newWhere()

	whereClause, whereParams, err := buildWhereClause([]SelectOption{w})
	if err != nil {
		return nil, err
	}

	q := buildUpdateQuery(
		tableInfo.Name,
		columnNames,
		whereClause,
	)

	stmt, e := b.db.PrepareContext(ctx, q)

	if e != nil {
		return nil, e
	}

	params := append(values, whereParams...)
	result, e := stmt.ExecContext(ctx, params...)

	if e != nil {
		return nil, e
	}

	return result, nil
}

func (b *Belvedere) SelectOne(ctx context.Context, dst interface{}) error {
	tableInfo := newTableInfo(dst)
	q := fmt.Sprintf("SELECT * FROM %s", tableInfo.Name)

	var conditions []byte
	conditions = append(conditions, tableInfo.Pk.Name...)
	conditions = append(conditions, " = ?"...)

	pkv, err := tableInfo.PkValue()
	if err != nil {
		return err
	}

	newWhere := Where(string(conditions), pkv)
	w := newWhere()

	whereClause, whereParams, err := buildWhereClause([]SelectOption{w})
	if err != nil {
		return err
	}

	q = q + whereClause

	stmt, e := b.db.PrepareContext(ctx, q)
	if e != nil {
		return e
	}

	rows, e := stmt.QueryContext(ctx, whereParams...)
	if e != nil {
		return e
	}

	defer rows.Close()

	for rows.Next() {
		pts, e := tableInfo.FieldPts()
		if e != nil {
			return e
		}

		if e = rows.Scan(pts...); e != nil {
			return e
		}

		tableInfo.SetValue(pts)
	}

	return nil
}

func toSliceType(i interface{}) (reflect.Type, error) {
	t := reflect.TypeOf(i)
	if t.Kind() != reflect.Ptr {
		if t.Kind() == reflect.Slice {
			return nil, fmt.Errorf("belvedere: cannot SELECT into a non-pointer slice: %v", t)
		}
		return nil, nil
	}
	if t = t.Elem(); t.Kind() != reflect.Slice {
		return nil, nil
	}
	return t.Elem(), nil
}

func columnToFieldIndex(t reflect.Type, cols []string) ([][]int, error) {
	colToFieldIndex := make([][]int, len(cols))

	missingColNames := []string{}
	for x := range cols {
		colName := strings.ToLower(cols[x])
		field, found := t.FieldByNameFunc(func(fieldName string) bool {
			return colName == camelToSnake(fieldName)
		})

		if found {
			colToFieldIndex[x] = field.Index
		}
		if colToFieldIndex[x] == nil {
			missingColNames = append(missingColNames, colName)
		}
	}

	if len(missingColNames) > 0 {
		return colToFieldIndex, errors.New("missing column")

	}

	return colToFieldIndex, nil
}

func (b *Belvedere) Select(ctx context.Context, dst interface{}, options ...NewSelectOption) error {
	t, err := toSliceType(dst)

	if err != nil {
		return err
	}

	isPtr := t.Kind() == reflect.Ptr
	if isPtr {
		t = t.Elem()
	}

	tn := getTableNameFromTypeName(t)
	q := fmt.Sprintf("SELECT * FROM %s", tn)
	som := newSelectOptionMap(options...)
	whereClause, whereParams, err := buildWhereClause(som.Wheres())
	if err != nil {
		return err
	}

	limitClause, limitParams, err := buildLimitClause(som.Limit())
	if err != nil {
		return err
	}

	orderClause, _ := buildOrderClause(som.Order())
	offsetClause, offsetParams, _ := buildOffsetClause(som.Offset())

	groupByClause, groupByParams := buildGroupByClause(som.GroupBy())

	q = q + whereClause + orderClause + groupByClause + limitClause + offsetClause

	params := append(whereParams, limitParams...)
	params = append(params, offsetParams...)
	params = append(params, groupByParams...)

	var rows *sql.Rows
	if len(params) > 0 {
		stmt, err := b.db.PrepareContext(ctx, q)
		if err != nil {
			return err
		}

		rows, err = stmt.QueryContext(ctx, params...)
		if err != nil {
			return err
		}
	} else {
		rows, err = b.db.QueryContext(ctx, q)
		if err != nil {
			return err
		}
	}

	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	defer rows.Close()

	var colToFieldIndex [][]int
	colToFieldIndex, err = columnToFieldIndex(t, cols)
	if err != nil {
		return err
	}

	sliceValue := reflect.Indirect(reflect.ValueOf(dst))

	for rows.Next() {
		if rows.Err() != nil {
			return rows.Err()
		}

		v := reflect.New(t)

		dest := make([]interface{}, len(cols))
		for x := range cols {
			f := v.Elem()

			index := colToFieldIndex[x]
			f = f.FieldByIndex(index)
			target := f.Addr().Interface()

			dest[x] = target
		}

		err = rows.Scan(dest...)
		if err != nil {
			return err
		}

		if !isPtr {
			v = v.Elem()
		}
		sliceValue.Set(reflect.Append(sliceValue, v))
	}

	if sliceValue.IsNil() {
		sliceValue.Set(reflect.MakeSlice(sliceValue.Type(), 0, 0))
	}

	return nil
}

func (b *Belvedere) Count(ctx context.Context, fn string, dst interface{}, options ...NewSelectOption) (int, error) {
	tableInfo := newTableInfo(dst)
	q := fmt.Sprintf("SELECT COUNT(%s) AS `cnt` FROM %s", fn, tableInfo.Name)
	som := newSelectOptionMap(options...)
	whereClause, whereParams, err := buildWhereClause(som.Wheres())
	if err != nil {
		return 0, err
	}

	q = q + whereClause

	stmt, e := b.db.PrepareContext(ctx, q)
	if e != nil {
		return 0, e
	}

	rows, e := stmt.QueryContext(ctx, whereParams...)
	if e != nil {
		return 0, e
	}

	defer rows.Close()

	var cnt int
	for rows.Next() {
		if e = rows.Scan(&cnt); e != nil {
			return 0, e
		}
	}

	return cnt, nil
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
