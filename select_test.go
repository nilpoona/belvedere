package belvedere

import "testing"

func TestWhereIn_Conditions(t *testing.T) {
	tests := []struct {
		name   string
		fn     string
		params []interface{}
		want   string
		err    error
	}{
		{
			name: "in",
			fn:   "id",
			params: []interface{}{
				1,
				2,
			},
			want: "id IN (?, ?)",
			err:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wiFnc := IN(tt.fn, tt.params...)
			wi := wiFnc()
			q, e := wi.Conditions()
			if q != tt.want {
				t.Errorf("whereIn.Conditions() result: %s expected value: %s", q, tt.want)
			}
			if e != tt.err {
				t.Errorf("whereIn.Conditions() err: %s expected value: %s", e, tt.err)
			}
		})
	}
}

func TestAnd_Conditions(t *testing.T) {
	tests := []struct {
		name   string
		wheres []NewSelectOption
		want   string
		err    error
	}{
		{
			name: "Specify two fields with an AND condition",
			wheres: []NewSelectOption{
				Where("age = ?", 1),
				Where("gender = ?", 'f'),
			},

			want: "age = ? AND gender = ?",
			err:  nil,
		},
		{
			name: "Specify three fields with an AND condition",
			wheres: []NewSelectOption{
				Where("age = ?", 1),
				Where("gender = ?", 'f'),
				IN("id", 1, 2, 3),
			},

			want: "age = ? AND gender = ? AND id IN (?, ?, ?)",
			err:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aFnc := And(tt.wheres...)
			a := aFnc()
			q, e := a.Conditions()
			if q != tt.want {
				t.Errorf("and.Conditions() result: %s expected value: %s", q, tt.want)
			}
			if e != tt.err {
				t.Errorf("and.Conditions() err: %s expected value: %s", e, tt.err)
			}
		})
	}
}
