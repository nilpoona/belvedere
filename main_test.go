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
