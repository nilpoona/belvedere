package belvedere

import (
	"database/sql"
	"context"
	"reflect"
	"regexp"
	"strings"
)

var repGetTableName = regexp.MustCompile(`^.+\.([^.]*?)$`)

type QueryBuilder interface {
	Insert(ctx context.Context, src interface{}) error
}

// Belvedere query builder struct
type Belvedere struct { db *sql.DB }



func getTableNameFromTypeName(typeName reflect.Type) string {
	tableName := repGetTableName.ReplaceAllString(typeName.String(), "$1")
	tableName = regexp.MustCompile(`([^A-Z])([A-Z])`).ReplaceAllString(tableName, `$1,$2`)
	return strings.ToLower(strings.Replace(tableName, ",", "_", -1))
}

// generateInsertQuery Generate an insert statement from the structure.
// If the field of the structure contains information about the column.
func generateInsertQuery(src interface{}) string {
	srcType := reflect.TypeOf(src)
	tableName := getTableNameFromTypeName(srcType)
	return tableName
}

// Insert
func (b *Belvedere) Insert(ctx context.Context, src interface{}) error {
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

	return &Belvedere{ db: db }, nil
}
