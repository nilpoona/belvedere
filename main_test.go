package belvedere

import (
	"context"
	"reflect"
	"testing"
	"time"
)

type (
	User struct {
		ID        uint64 `pk:"true"`
		Name      string
		Profile   string
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	FooBar         struct{}
	FooBarHoge     struct{}
	FooBarHogeFuga struct{}
)

func nowTime() time.Time {
	return time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)
}

func TestTableInfo_Values(t *testing.T) {
	mockNow := nowTime()
	data := []struct {
		name      string
		in        User
		want      []interface{}
		err       error
		excludePk bool
	}{
		{
			name: "get user table values",
			in: User{
				ID:        uint64(2),
				Name:      "foobar",
				Profile:   "profile",
				CreatedAt: mockNow,
				UpdatedAt: mockNow,
			},
			want: []interface{}{
				uint64(2),
				"foobar",
				"profile",
				mockNow,
				mockNow,
			},
			err:       nil,
			excludePk: false,
		},
		{
			name: "get user table values",
			in: User{
				ID:        uint64(2),
				Name:      "foobar",
				Profile:   "profile",
				CreatedAt: mockNow,
				UpdatedAt: mockNow,
			},
			want: []interface{}{
				"foobar",
				"profile",
				mockNow,
				mockNow,
			},
			err:       nil,
			excludePk: true,
		},
	}

	for _, d := range data {
		tableInfo := newTableInfo(&d.in)
		t.Log(d.name)
		values, e := tableInfo.Values(d.excludePk)
		if e != d.err {
			t.Errorf("The error is not the value you expected expected: %v current value: %v", d.err, e)
		}

		for i, v := range values {
			if v != d.want[i] {
				t.Errorf("The column values is not the value you expected expected: %v current value: %v", d.want[i], v)
			}
		}
	}
}

func TestTableInfo_ColumnNames(t *testing.T) {
	data := []struct {
		name      string
		in        User
		want      string
		excludePk bool
	}{
		{
			name:      "get user table columns `id,name,profile,creatd_at,updated_at`",
			in:        User{},
			want:      "id,name,profile,created_at,updated_at",
			excludePk: false,
		},
		{
			name:      "get user table columns `name,profile,creatd_at,updated_at`",
			in:        User{},
			want:      "name,profile,created_at,updated_at",
			excludePk: true,
		},
	}

	for _, d := range data {
		t.Log(d.name)
		tableInfo := newTableInfo(d.in)
		cnames := tableInfo.ColumnNames(d.excludePk)
		if cnames != d.want {
			t.Errorf("The column names is not the value you expected expected: %s current value: %s", d.want, cnames)
		}
	}
}

func TestBelvedere_SelectOne(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	b, e := NewBelvedere("mysql", "root:@/test?parseTime=true")
	if e != nil {
		t.Fatal(e)
	}

	//mockNow := nowTime()
	if e != nil {
		t.Fail()
	}

	dst := &User{}
	e = b.SelectOne(ctx, dst, Where("id = ?", 1))
	t.Log(dst)
}

func TestBelvedere_Select(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	b, e := NewBelvedere("mysql", "root:@/test?parseTime=true")
	if e != nil {
		t.Fatal(e)
	}

	//mockNow := nowTime()
	if e != nil {
		t.Fail()
	}

	var users []*User
	e = b.Select(
		ctx,
		&users,
		Limit(1),
	)
	if e != nil {
		t.Error(e)
	}

	for _, u := range users {
		t.Log(u.Name)
	}

}

func TestBelvedere_Count(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	b, e := NewBelvedere("mysql", "root:@/test?parseTime=true")
	if e != nil {
		t.Fatal(e)
	}

	if e != nil {
		t.Fail()
	}

	cnt, e := b.Count(ctx, "id", &User{})
	if e != nil {
		t.Error(e)
	}

	t.Log(cnt)
}

func TestBelvedere_Insert(t *testing.T) {
	mockNow := nowTime()
	data := []struct {
		name string
		in   User
		err  error
	}{
		{
			name: "insert user record.",
			in: User{
				Name:      "foo",
				Profile:   "foobar",
				CreatedAt: mockNow,
				UpdatedAt: mockNow,
			},
			err: nil,
		},
	}

	b, e := NewBelvedere("mysql", "root:root@/test")
	if e != nil {
		t.Fail()
	}

	ctx, _ := context.WithCancel(context.Background())

	for _, d := range data {
		t.Log(d.name)
		_, e := b.Insert(ctx, &d.in)
		if e != d.err {
			t.Errorf("The error is not the value you expected expected: %v current value: %v", d.err, e)
		}
		// tableName := getTableNameFromTypeName(d.in)
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
			in:   reflect.TypeOf(User{}),
			want: "user",
		},
		{
			name: "get table name foo_bar.",
			in:   reflect.TypeOf(FooBar{}),
			want: "foo_bar",
		},
		{
			name: "get table name foo_bar_hoge.",
			in:   reflect.TypeOf(FooBarHoge{}),
			want: "foo_bar_hoge",
		},
		{
			name: "get table name foo_bar_hoge_fuga.",
			in:   reflect.TypeOf(FooBarHogeFuga{}),
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
