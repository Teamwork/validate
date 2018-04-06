package validate

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestMerge(t *testing.T) {
	cases := []struct {
		a, b, want map[string][]string
	}{
		{
			map[string][]string{},
			map[string][]string{},
			map[string][]string{},
		},
		{
			map[string][]string{"a": {"b"}},
			map[string][]string{},
			map[string][]string{"a": {"b"}},
		},
		{
			map[string][]string{},
			map[string][]string{"a": {"b"}},
			map[string][]string{"a": {"b"}},
		},
		{
			map[string][]string{"a": {"b"}},
			map[string][]string{"a": {"c"}},
			map[string][]string{"a": {"b", "c"}},
		},
		{
			map[string][]string{"a": {"b"}},
			map[string][]string{"q": {"c"}},
			map[string][]string{"a": {"b"}, "q": {"c"}},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			in := New()
			in.Errors = tc.a
			other := New()
			other.Errors = tc.b

			in.Merge(other)

			if !reflect.DeepEqual(tc.want, in.Errors) {
				t.Errorf("\nout:  %#v\nwant: %#v\n", in.Errors, tc.want)
			}
		})
	}
}

func TestString(t *testing.T) {
	cases := []struct {
		in   Validator
		want string
	}{
		{Validator{}, "<no errors>"},
		{Validator{map[string][]string{}}, "<no errors>"},

		{Validator{map[string][]string{
			"k": {"oh no"},
		}}, "k: oh no.\n"},
		{Validator{map[string][]string{
			"k": {"oh no", "more"},
		}}, "k: oh no, more.\n"},
		{Validator{map[string][]string{
			"k": {"oh no", "more", "even more"},
		}}, "k: oh no, more, even more.\n"},
		{Validator{map[string][]string{
			"k":  {"oh no", "more", "even more"},
			"k2": {"asd"},
		}}, "k: oh no, more, even more.\nk2: asd.\n"},
		{Validator{map[string][]string{
			"zxc": {"asd"},
			"asd": {"oh no", "more", "even more"},
		}}, "asd: oh no, more, even more.\nzxc: asd.\n"},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			out := tc.in.String()
			if out != tc.want {
				t.Errorf("\nout:  %#v\nwant: %#v\n", out, tc.want)
			}
		})
	}
}

func TestValidators(t *testing.T) {
	cases := []struct {
		val        func(Validator)
		wantErrors map[string][]string
	}{
		// Required
		{
			func(v Validator) {
				v.Required("firstName", "not empty")
				v.Required("age", 32)
			},
			make(map[string][]string),
		},
		{
			func(v Validator) {
				v.Required("lastName", "", "foo")
				v.Required("age", 0, "bar")
			},
			map[string][]string{"lastName": {"foo"}, "age": {"bar"}},
		},
		{
			func(v Validator) {
				v.Required("lastName", "")
				v.Required("age", 0)
			},
			map[string][]string{"lastName": {"must be set"}, "age": {"must be set"}},
		},
		{
			func(v Validator) {
				v.Required("email", "")
				v.Email("email", "")

				v.Required("email2", "asd")
				v.Email("email2", "asd")

				v.Required("email3", "asd@example.com")
				v.Email("email3", "asd@example.com")
			},
			map[string][]string{
				"email":  {"must be set"},
				"email2": {"must be a valid email address"},
			},
		},
		{
			func(v Validator) { v.Required("k", true) },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Required("k", false) },
			map[string][]string{"k": {"must be set"}},
		},

		// Len
		{
			func(v Validator) { v.Len("v", "w00t", 2, 5) },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Len("v", "w00t", 4, 0) },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Len("v", "w00t", 0, 4) },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Len("v", "w00t", 1, 2) },
			map[string][]string{"v": {"must be shorter than 2 characters"}},
		},
		{
			func(v Validator) { v.Len("v", "w00t", 1, 2, "foo") },
			map[string][]string{"v": {"foo"}},
		},
		{
			func(v Validator) { v.Len("v", "w00t", 16, 32) },
			map[string][]string{"v": {"must be longer than 16 characters"}},
		},

		// Exclude
		{
			func(v Validator) { v.Exclude("key", "val", []string{}) },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Exclude("key", "val", nil) },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Exclude("key", "val", []string{"valx"}) },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Exclude("key", "val", []string{"VAL"}) },
			map[string][]string{"key": {`cannot be ‘VAL’`}},
		},
		{
			func(v Validator) { v.Exclude("key", "val", []string{"VAL"}, "foo") },
			map[string][]string{"key": {`foo`}},
		},
		{
			func(v Validator) { v.Exclude("key", "val", []string{"hello", "val"}) },
			map[string][]string{"key": {`cannot be ‘val’`}},
		},

		// Include
		{
			func(v Validator) { v.Include("key", "val", []string{}) },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Include("key", "val", nil) },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Include("key", "val", []string{"valx"}) },
			map[string][]string{"key": {`must be one of ‘valx’`}},
		},
		{
			func(v Validator) { v.Include("key", "val", []string{"valx"}, "foo") },
			map[string][]string{"key": {`foo`}},
		},
		{
			func(v Validator) { v.Include("key", "val", []string{"VAL"}) },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Include("key", "val", []string{"hello", "val"}) },
			make(map[string][]string),
		},

		// Domain
		{
			func(v Validator) { v.Domain("v", "") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Domain("v", "example.com") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Domain("v", "example.com.test.asd") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Domain("v", "example-test.com") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Domain("v", "ﻢﻔﺗﻮﺣ.ﺬﺑﺎﺑﺓ") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Domain("v", "xn--pgbg2dpr.xn--mgbbbe5a") },
			make(map[string][]string),
		},

		{
			func(v Validator) { v.Domain("v", "one-label") },
			map[string][]string{"v": {"must be a valid domain"}},
		},
		{
			func(v Validator) { v.Domain("v", "one-label", "foo") },
			map[string][]string{"v": {"foo"}},
		},
		{
			func(v Validator) { v.Domain("v", "example.com:-)") },
			map[string][]string{"v": {"must be a valid domain"}},
		},
		{
			func(v Validator) { v.Domain("v", "ex ample.com") },
			map[string][]string{"v": {"must be a valid domain"}},
		},

		// HexColor
		{
			func(v Validator) { v.HexColor("v", "") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.HexColor("v", "#36a") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.HexColor("v", "#3a6ea5") },
			make(map[string][]string),
		},

		{
			func(v Validator) { v.HexColor("v", "fff") },
			map[string][]string{"v": {"must be a valid color code"}},
		},
		{
			func(v Validator) { v.HexColor("v", "#ff") },
			map[string][]string{"v": {"must be a valid color code"}},
		},
		{
			func(v Validator) { v.HexColor("v", "#fffffff") },
			map[string][]string{"v": {"must be a valid color code"}},
		},

		// Date
		{
			func(v Validator) { v.Date("k", "2017-11-14T13:37:00Z", time.RFC3339) },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Date("k", "2017-11-14", time.RFC3339) },
			map[string][]string{"k": {"must be a date as ‘2006-01-02T15:04:05Z07:00’"}},
		},
		{
			func(v Validator) { v.Date("k", "2017-11-14", time.RFC3339, "not valid") },
			map[string][]string{"k": {"not valid"}},
		},

		// Email
		// Don't need to extensively validate emails, we have tests for that in
		// the mailaddress package already.
		{
			func(v Validator) { v.Email("v", "") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Email("v", "martin@example.com") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Email("v", "martin") },
			map[string][]string{"v": {"must be a valid email address"}},
		},
		{
			func(v Validator) { v.Email("v", "martin", "foo") },
			map[string][]string{"v": {"foo"}},
		},

		// IPv4
		{
			func(v Validator) { v.IPv4("v", "") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.IPv4("v", "127.0.0.1") },
			make(map[string][]string),
		},

		{
			func(v Validator) { v.IPv4("v", "127.0.0.4/8") },
			map[string][]string{"v": {"must be a valid IPv4 address"}},
		},
		{
			func(v Validator) { v.IPv4("v", "127.0.0.4/8", "foo") },
			map[string][]string{"v": {"foo"}},
		},
		{
			func(v Validator) { v.IPv4("v", "127.1") }, // Technically correct but Go doesn't seem to like it.
			map[string][]string{"v": {"must be a valid IPv4 address"}},
		},
		{
			func(v Validator) { v.IPv4("v", "127.0.0.506") },
			map[string][]string{"v": {"must be a valid IPv4 address"}},
		},
		{
			func(v Validator) { v.IPv4("v", "127.") },
			map[string][]string{"v": {"must be a valid IPv4 address"}},
		},
		{
			func(v Validator) { v.IPv4("v", "asdf") },
			map[string][]string{"v": {"must be a valid IPv4 address"}},
		},
		{
			func(v Validator) { v.IPv4("v", "::1") },
			map[string][]string{"v": {"must be a valid IPv4 address"}},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			v := New()
			tc.val(v)

			if !reflect.DeepEqual(v.Errors, tc.wantErrors) {
				t.Errorf("\nout:  %#v\nwant: %#v\n", v.Errors, tc.wantErrors)
			}
		})
	}
}

func TestInteger(t *testing.T) {
	cases := []struct {
		val        func(Validator) int64
		want       int64
		wantErrors map[string][]string
	}{
		{
			func(v Validator) int64 { return v.Integer("k", "6") },
			6,
			make(map[string][]string),
		},
		{
			func(v Validator) int64 { return v.Integer("k", " 6 ") },
			6,
			make(map[string][]string),
		},
		{
			func(v Validator) int64 { return v.Integer("k", "0") },
			0,
			make(map[string][]string),
		},
		{
			func(v Validator) int64 { return v.Integer("k", "-1") },
			-1,
			make(map[string][]string),
		},
		{
			func(v Validator) int64 { return v.Integer("k", "1.2") },
			0,
			map[string][]string{"k": {"must be a whole number"}},
		},
		{
			func(v Validator) int64 { return v.Integer("k", "asd") },
			0,
			map[string][]string{"k": {"must be a whole number"}},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			v := New()
			i := tc.val(v)

			if !reflect.DeepEqual(v.Errors, tc.wantErrors) {
				t.Errorf("\nout:  %#v\nwant: %#v\n", v.Errors, tc.wantErrors)
			}

			if i != tc.want {
				t.Errorf("\nout:  %#v\nwant: %#v\n", i, tc.want)
			}
		})
	}
}

func TestBoolean(t *testing.T) {
	cases := []struct {
		val        func(Validator) bool
		want       bool
		wantErrors map[string][]string
	}{
		{
			func(v Validator) bool { return v.Boolean("k", "true") },
			true,
			make(map[string][]string),
		},
		{
			func(v Validator) bool { return v.Boolean("k", "on") },
			false,
			map[string][]string{"k": {"must be a boolean"}},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			v := New()
			i := tc.val(v)

			if !reflect.DeepEqual(v.Errors, tc.wantErrors) {
				t.Errorf("\nout:  %#v\nwant: %#v\n", v.Errors, tc.wantErrors)
			}

			if i != tc.want {
				t.Errorf("\nout:  %#v\nwant: %#v\n", i, tc.want)
			}
		})
	}
}
