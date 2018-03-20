package belvedere

import (
	"testing"
	"time"
	"reflect"
)

type (
	User struct {
		ID uint32
		Name string
		Profile string
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	FooBar struct {}
	FooBarHoge struct {}
	FooBarHogeFuga struct{}
)

func TestTableInfo_Values(t *testing.T) {
	data := []struct {
		name string
		in   User
		want string
	}{
		{
			name: "get user table columns `id,name,profile,creatd_at,updated_at`",
			in: User{
				ID: uint32(2),
				Name: "foobar",
				Profile: "profile",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			want: "id,name,profile,created_at,updated_at",
		},
	}

	for _, d := range data {
		tableInfo := newTableInfo(&d.in)
		t.Log(tableInfo.Values())
		/*
		cnames := tableInfo.ColumnNames()
		if cnames != d.want {
			t.Errorf("The column names is not the value you expected expected: %s current value: %s", d.want, cnames)
		}
		*/
	}
}

func TestTableInfo_ColumnNames(t *testing.T) {
	data := []struct {
		name string
		in   User
		want string
	}{
		{
			name: "get user table columns `id,name,profile,creatd_at,updated_at`",
			in: User{},
			want: "id,name,profile,created_at,updated_at",
		},
	}

	for _, d := range data {
		t.Log(d.name)
		tableInfo := newTableInfo(d.in)
		cnames := tableInfo.ColumnNames()
		if cnames != d.want {
			t.Errorf("The column names is not the value you expected expected: %s current value: %s", d.want, cnames)
		}
	}
}

func TestGetTableNameFromTypeName(t *testing.T) {
	data := []struct {
		name string
		in   reflect.Type
		want string
	}{
		{
			name: "get table name user.",
			in: reflect.TypeOf(User{}),
			want: "user",
		},
		{
			name: "get table name foo_bar.",
			in: reflect.TypeOf(FooBar{}),
			want: "foo_bar",
		},
		{
			name: "get table name foo_bar_hoge.",
			in: reflect.TypeOf(FooBarHoge{}),
			want: "foo_bar_hoge",
		},
		{
			name: "get table name foo_bar_hoge_fuga.",
			in: reflect.TypeOf(FooBarHogeFuga{}),
			want: "foo_bar_hoge_fuga",
		},
	}

	for _, d := range data {
		t.Log(d.name)

		tableName := getTableNameFromTypeName(d.in)
		if tableName != d.want {
			t.Errorf("The table name is not the value you expected expected: %s current value: %s", d.want, tableName)
		}
	}

}
