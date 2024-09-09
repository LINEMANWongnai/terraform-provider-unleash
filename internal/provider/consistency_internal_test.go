package provider

import (
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func Test_isNullArrayAndExistingEmptyArray(t *testing.T) {
	type args[T any] struct {
		current []T
		before  []T
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want bool
	}
	tests := []testCase[string]{
		{
			name: "current is nil and before is nil",
			args: args[string]{
				current: nil,
				before:  nil,
			},
			want: false,
		},
		{
			name: "current is empty and before is nil",
			args: args[string]{
				current: []string{},
				before:  nil,
			},
			want: false,
		},
		{
			name: "current is empty and before is empty",
			args: args[string]{
				current: []string{},
				before:  []string{},
			},
			want: false,
		},
		{
			name: "current is nil and before is empty",
			args: args[string]{
				current: nil,
				before:  []string{},
			},
			want: true,
		},

		{
			name: "current is nil and before is not empty",
			args: args[string]{
				current: nil,
				before:  []string{"notempty"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNullArrayAndExistingEmptyArray(tt.args.current, tt.args.before); got != tt.want {
				t.Errorf("isNullArrayAndExistingEmptyArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isNullMapAndExistingEmptyMap(t *testing.T) {
	type args struct {
		value       map[string]types.String
		valueBefore map[string]types.String
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "value is nil and valueBefore is nil",
			args: args{
				value:       nil,
				valueBefore: nil,
			},
			want: false,
		},
		{
			name: "value is empty and valueBefore is nil",
			args: args{
				value:       make(map[string]types.String),
				valueBefore: nil,
			},
			want: false,
		},
		{
			name: "value is empty and valueBefore is empty",
			args: args{
				value:       make(map[string]types.String),
				valueBefore: make(map[string]types.String),
			},
			want: false,
		},
		{
			name: "value is nil and valueBefore is empty",
			args: args{
				value:       nil,
				valueBefore: make(map[string]types.String),
			},
			want: true,
		},
		{
			name: "value is nil and valueBefore is not empty",
			args: args{
				value:       nil,
				valueBefore: map[string]types.String{"notempty": types.StringValue("notempty")},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNullMapAndExistingEmptyMap(tt.args.value, tt.args.valueBefore); got != tt.want {
				t.Errorf("isNullMapAndExistingEmptyMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_tryUpdateToEmptyStringIfBeforeEmpty(t *testing.T) {
	type args struct {
		value       types.String
		valueBefore types.String
	}
	tests := []struct {
		name               string
		args               args
		shouldInvokeSetter bool
	}{
		{
			name: "value is null and valueBefore is null",
			args: args{
				value:       types.StringNull(),
				valueBefore: types.StringNull(),
			},
			shouldInvokeSetter: false,
		},
		{
			name: "value is empty and valueBefore is null",
			args: args{
				value:       types.StringValue(""),
				valueBefore: types.StringNull(),
			},
			shouldInvokeSetter: false,
		},
		{
			name: "value is empty and valueBefore is empty",
			args: args{
				value:       types.StringValue(""),
				valueBefore: types.StringValue(""),
			},
			shouldInvokeSetter: false,
		},
		{
			name: "value is null and valueBefore is empty",
			args: args{
				value:       types.StringNull(),
				valueBefore: types.StringValue(""),
			},
			shouldInvokeSetter: true,
		},
		{
			name: "value is null and valueBefore is not empty",
			args: args{
				value:       types.StringNull(),
				valueBefore: types.StringValue("not empty"),
			},
			shouldInvokeSetter: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := atomic.Pointer[types.String]{}
			tryUpdateToEmptyStringIfBeforeEmpty(tt.args.value, tt.args.valueBefore, func(value types.String) {
				p.Store(&value)
			})
			if tt.shouldInvokeSetter {
				if !p.Load().Equal(types.StringValue("")) {
					t.Errorf("tryUpdateToEmptyStringIfBeforeEmpty() = %v, empty", p.Load())
				}
			} else if p.Load() != nil {
				t.Errorf("tryUpdateToEmptyStringIfBeforeEmpty() = %v, nil", p.Load())
			}
		})
	}
}

func Test_tryUpdateToFalseIfBeforeFalse(t *testing.T) {
	type args struct {
		value       types.Bool
		valueBefore types.Bool
	}
	tests := []struct {
		name               string
		args               args
		shouldInvokeSetter bool
	}{
		{
			name: "value is null and valueBefore is null",
			args: args{
				value:       types.BoolNull(),
				valueBefore: types.BoolNull(),
			},
			shouldInvokeSetter: false,
		},
		{
			name: "value is false and valueBefore is null",
			args: args{
				value:       types.BoolValue(false),
				valueBefore: types.BoolNull(),
			},
			shouldInvokeSetter: false,
		},
		{
			name: "value is false and valueBefore is false",
			args: args{
				value:       types.BoolValue(false),
				valueBefore: types.BoolValue(false),
			},
			shouldInvokeSetter: false,
		},
		{
			name: "value is null and valueBefore is false",
			args: args{
				value:       types.BoolNull(),
				valueBefore: types.BoolValue(false),
			},
			shouldInvokeSetter: true,
		},
		{
			name: "value is null and valueBefore is true",
			args: args{
				value:       types.BoolNull(),
				valueBefore: types.BoolValue(true),
			},
			shouldInvokeSetter: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := atomic.Pointer[types.Bool]{}
			tryUpdateToFalseIfBeforeFalse(tt.args.value, tt.args.valueBefore, func(value types.Bool) {
				p.Store(&value)
			})
			if tt.shouldInvokeSetter {
				if !p.Load().Equal(types.BoolValue(false)) {
					t.Errorf("tryUpdateToFalseIfBeforeFalse() = %v, empty", p.Load())
				}
			} else if p.Load() != nil {
				t.Errorf("tryUpdateToFalseIfBeforeFalse() = %v, nil", p.Load())
			}
		})
	}
}
