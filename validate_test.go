package validate

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/teamwork/mailaddress"
)

func TestRequiredInt(t *testing.T) {
	tests := []struct {
		a    interface{}
		want bool
	}{
		{0, true},
		{int64(0), true},
		{uint(0), true},
		{uint64(0), true},
		{1, false},
		{int64(1), false},
		{uint(1), false},
		{uint64(1), false},
	}

	for i, tt := range tests {
		name := fmt.Sprintf("%v", i)
		t.Run(name, func(t *testing.T) {
			v := New()
			v.Required(name, tt.a)
			if got := v.HasErrors(); got != tt.want {
				t.Errorf("\ngot:  %#v\nwant: %#v\n", got, tt.want)
			}
		})
	}
}

func TestMerge(t *testing.T) {
	tests := []struct {
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

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			in := New()
			in.Errors = tt.a
			other := New()
			other.Errors = tt.b

			in.Merge(other)

			if !reflect.DeepEqual(tt.want, in.Errors) {
				t.Errorf("\nout:  %#v\nwant: %#v\n", in.Errors, tt.want)
			}
		})
	}
}

func TestSub(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		v := New()
		v.Required("name", "")
		v.HexColor("color", "not a color")

		// Easy case
		s := New()
		s.Required("domain", "")
		s.Email("contactEmail", "not an email")
		v.Sub("setting", "", s.ErrorOrNil())

		// List
		addr1 := New()
		addr1.Required("city", "Bristol")
		v.Sub("addresses", "home", addr1)
		addr2 := New()
		addr2.Required("city", "")
		v.Sub("addresses", "office", addr2)

		// Non-Validator.
		v.Sub("other", "", errors.New("oh noes"))
		v.Sub("emails", "home", nil)
		v.Sub("emails", "office", errors.New("not an email"))

		// Sub with Sub.
		s1 := New()
		s2 := New()
		s2.Append("err", "very sub")
		s1.Sub("sub2", "", s2)
		v.Sub("sub1", "", s1)

		ls1 := New()
		ls2 := New()
		ls2.Append("err", "very sub")
		ls1.Sub("lsub2", "holiday", ls2)
		v.Sub("lsub1", "", ls1)

		want := map[string][]string{
			"lsub1.lsub2[holiday].err": []string{"very sub"},
			"sub1.sub2.err":            []string{"very sub"},
			"name":                     []string{"must be set"},
			"color":                    []string{"must be a valid color code"},
			"setting.domain":           []string{"must be set"},
			"setting.contactEmail":     []string{"must be a valid email address"},
			"addresses[office].city":   []string{"must be set"},
			"other":                    []string{"oh noes"},
			"emails[office]":           []string{"not an email"},
		}

		if d := cmp.Diff(v.Errors, want); d != "" {
			t.Errorf("(-got +want)\n:%s", d)
		}
	})
}

func TestString(t *testing.T) {
	tests := []struct {
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

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			out := tt.in.String()
			if out != tt.want {
				t.Errorf("\nout:  %#v\nwant: %#v\n", out, tt.want)
			}
		})
	}
}

func BenchmarkString(b *testing.B) {
	v := New()
	noOfErrors := 256
	const err = "Oh no!"
	for i := 0; i < noOfErrors; i++ {
		key := fmt.Sprintf("err%d", i)
		v.Append(key, err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = v.String()
	}
}

func TestValidators(t *testing.T) {
	tests := []struct {
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
		{
			func(v Validator) { v.Required("k", []string{}) },
			map[string][]string{"k": {"must be set"}},
		},
		{
			func(v Validator) { v.Required("k", *new([]string)) },
			map[string][]string{"k": {"must be set"}},
		},
		{
			func(v Validator) { v.Required("k", []string{""}) },
			map[string][]string{"k": {"must be set"}},
		},
		{
			func(v Validator) { v.Required("k", []string{"", "", ""}) },
			map[string][]string{"k": {"must be set"}},
		},
		{
			func(v Validator) { v.Required("k", []string{" "}) },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Required("k", []string{"", "", " "}) },
			make(map[string][]string),
		},

		// Required mailaddress
		{
			func(v Validator) {
				v.Required("k1", mailaddress.Address{})
				v.Required("k2", mailaddress.List{})
			},
			map[string][]string{"k1": {"must be set"}, "k2": {"must be set"}},
		},
		{
			func(v Validator) {
				v.Required("k1", mailaddress.Address{Address: "foo@example.com"})
				v.Required("k2", mailaddress.List{mailaddress.New("", "asd")})
			},
			make(map[string][]string),
		},

		// Len - string
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
			map[string][]string{"v": {"must be shorter than 2"}},
		},
		{
			func(v Validator) { v.Len("v", "w00t", 1, 2, "foo") },
			map[string][]string{"v": {"foo"}},
		},
		{
			func(v Validator) { v.Len("v", "w00t", 16, 32) },
			map[string][]string{"v": {"must be longer than 16"}},
		},
		{
			func(v Validator) { v.Len("v", "w00t", 16, 32, "foo") },
			map[string][]string{"v": {"foo"}},
		},

		// Len - slice
		{
			func(v Validator) { v.Len("foo", []int{1, 2}, 1, 3) },
			map[string][]string{},
		},
		{
			func(v Validator) { v.Len("foo", []int{1, 2}, 0, 3) },
			map[string][]string{},
		},
		{
			func(v Validator) { v.Len("foo", []int{1, 2}, 2, 0) },
			map[string][]string{},
		},
		{
			func(v Validator) { v.Len("foo", []int{}, 1, 0, "msg") },
			map[string][]string{"foo": []string{"msg"}},
		},
		{
			func(v Validator) { v.Len("foo", []int{}, 1, 0) },
			map[string][]string{"foo": []string{"must be longer than 1"}},
		},
		{
			func(v Validator) { v.Len("foo", []int{3, 4}, 0, 1, "msg") },
			map[string][]string{"foo": []string{"msg"}},
		},
		{
			func(v Validator) { v.Len("foo", []int{3, 4}, 0, 1) },
			map[string][]string{"foo": []string{"must be shorter than 1"}},
		},

		// Len - map
		{
			func(v Validator) { v.Len("foo", map[int]int{1: 1, 2: 2}, 1, 3) },
			map[string][]string{},
		},
		{
			func(v Validator) { v.Len("foo", map[int]int{1: 1, 2: 2}, 0, 3) },
			map[string][]string{},
		},
		{
			func(v Validator) { v.Len("foo", map[int]int{1: 1, 2: 2}, 2, 0) },
			map[string][]string{},
		},
		{
			func(v Validator) { v.Len("foo", map[int]int{}, 1, 0, "msg") },
			map[string][]string{"foo": []string{"msg"}},
		},
		{
			func(v Validator) { v.Len("foo", map[int]int{}, 1, 0) },
			map[string][]string{"foo": []string{"must be longer than 1"}},
		},
		{
			func(v Validator) { v.Len("foo", map[int]int{1: 1, 2: 2}, 0, 1, "msg") },
			map[string][]string{"foo": []string{"msg"}},
		},
		{
			func(v Validator) { v.Len("foo", map[int]int{1: 1, 2: 2}, 0, 1) },
			map[string][]string{"foo": []string{"must be shorter than 1"}},
		},

		// Len - array
		{
			func(v Validator) { v.Len("foo", [2]int{}, 1, 3) },
			map[string][]string{},
		},
		{
			func(v Validator) { v.Len("foo", [2]int{}, 0, 3) },
			map[string][]string{},
		},
		{
			func(v Validator) { v.Len("foo", [2]int{}, 2, 0) },
			map[string][]string{},
		},
		{
			func(v Validator) { v.Len("foo", [0]int{}, 1, 0, "msg") },
			map[string][]string{"foo": []string{"msg"}},
		},
		{
			func(v Validator) { v.Len("foo", [0]int{}, 1, 0) },
			map[string][]string{"foo": []string{"must be longer than 1"}},
		},
		{
			func(v Validator) { v.Len("foo", [2]int{}, 0, 1, "msg") },
			map[string][]string{"foo": []string{"msg"}},
		},
		{
			func(v Validator) { v.Len("foo", [2]int{}, 0, 1) },
			map[string][]string{"foo": []string{"must be shorter than 1"}},
		},

		// Len - channel
		{
			func(v Validator) {
				c := make(chan int, 3)
				c <- 1
				c <- 2
				v.Len("foo", c, 1, 3)
			},
			map[string][]string{},
		},
		{
			func(v Validator) {
				c := make(chan int, 3)
				c <- 1
				c <- 2
				v.Len("foo", c, 0, 3)
			},
			map[string][]string{},
		},
		{
			func(v Validator) {
				c := make(chan int, 3)
				c <- 1
				c <- 2
				v.Len("foo", c, 2, 0)
			},
			map[string][]string{},
		},
		{
			func(v Validator) { v.Len("foo", make(chan int), 1, 0, "msg") },
			map[string][]string{"foo": []string{"msg"}},
		},
		{
			func(v Validator) { v.Len("foo", make(chan int), 1, 0) },
			map[string][]string{"foo": []string{"must be longer than 1"}},
		},
		{
			func(v Validator) {
				c := make(chan int, 3)
				c <- 1
				c <- 2
				v.Len("foo", c, 0, 1, "msg")
			},
			map[string][]string{"foo": []string{"msg"}},
		},
		{
			func(v Validator) {
				c := make(chan int, 3)
				c <- 1
				c <- 2
				v.Len("foo", c, 0, 1)
			},
			map[string][]string{"foo": []string{"must be shorter than 1"}},
		},

		// Len - invalid
		{
			func(v Validator) { v.Len("foo", 1, 0, 0, "msg") },
			map[string][]string{"foo": []string{"cannot validate length of type int"}},
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

		// URL
		{
			func(v Validator) { v.URL("v", "") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.URL("v", "example.com") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.URL("v", "example.com.test.asd/testing.html") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.URL("v", "example-test.com") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.URL("v", "ﻢﻔﺗﻮﺣ.ﺬﺑﺎﺑﺓ") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.URL("v", "xn--pgbg2dpr.xn--mgbbbe5a") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.URL("v", "http://x.com") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.URL("v", "unknownschema://x.com?q=v&x=2%3Aa#frag") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.URL("v", "complex://x.com") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.URL("v", "http://sunbeam.teamwork.localhost:9000/bucket/1/avatar-1.jpeg") },
			make(map[string][]string),
		},

		{
			func(v Validator) { v.URL("v", "one-label") },
			map[string][]string{"v": {"must be a valid url"}},
		},
		{
			func(v Validator) { v.URL("v", "http://x") },
			map[string][]string{"v": {"must be a valid url"}},
		},
		{
			func(v Validator) { v.URL("v", "one-label", "foo") },
			map[string][]string{"v": {"foo"}},
		},
		{
			func(v Validator) { v.URL("v", "example.com:-)") },
			map[string][]string{"v": {"must be a valid url"}},
		},
		{
			func(v Validator) { v.URL("v", "ex ample.com") },
			map[string][]string{"v": {"must be a valid url: parse http://ex%20ample.com: invalid URL escape \"%20\""}},
		},
		{
			func(v Validator) { v.URL("v", "unknown_schema://x.com") },
			map[string][]string{"v": {"must be a valid url: parse unknown_schema://x.com: " +
				"first path segment in URL cannot contain colon"}},
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

		// Phone
		{
			func(v Validator) { v.Phone("v", "") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Phone("v", "12345123") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Phone("v", "(+31)-12345123") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Phone("v", "[+31]-12345123") },
			map[string][]string{"v": {"must be a valid phone number"}},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			v := New()
			tt.val(v)

			if !reflect.DeepEqual(v.Errors, tt.wantErrors) {
				t.Errorf("\nout:  %#v\nwant: %#v\n", v.Errors, tt.wantErrors)
			}
		})
	}
}

func TestInteger(t *testing.T) {
	tests := []struct {
		val        func(Validator) int64
		want       int64
		wantErrors map[string][]string
	}{
		{
			func(v Validator) int64 { return v.Integer("k", "") },
			0,
			make(map[string][]string),
		},
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

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			v := New()
			i := tt.val(v)

			if !reflect.DeepEqual(v.Errors, tt.wantErrors) {
				t.Errorf("\nout:  %#v\nwant: %#v\n", v.Errors, tt.wantErrors)
			}

			if i != tt.want {
				t.Errorf("\nout:  %#v\nwant: %#v\n", i, tt.want)
			}
		})
	}
}

func TestBoolean(t *testing.T) {
	tests := []struct {
		val        func(Validator) bool
		want       bool
		wantErrors map[string][]string
	}{
		{
			func(v Validator) bool { return v.Boolean("k", "on") },
			false,
			map[string][]string{"k": {"must be a boolean"}},
		},
		{
			func(v Validator) bool { return v.Boolean("k", "") },
			false,
			make(map[string][]string),
		},
		{
			func(v Validator) bool { return v.Boolean("k", "0") },
			false,
			make(map[string][]string),
		},
		{
			func(v Validator) bool { return v.Boolean("k", "n") },
			false,
			make(map[string][]string),
		},
		{
			func(v Validator) bool { return v.Boolean("k", "no") },
			false,
			make(map[string][]string),
		},
		{
			func(v Validator) bool { return v.Boolean("k", "f") },
			false,
			make(map[string][]string),
		},
		{
			func(v Validator) bool { return v.Boolean("k", "false") },
			false,
			make(map[string][]string),
		},
		{
			func(v Validator) bool { return v.Boolean("k", "FALSE") },
			false,
			make(map[string][]string),
		},
		{
			func(v Validator) bool { return v.Boolean("k", "1") },
			true,
			make(map[string][]string),
		},
		{
			func(v Validator) bool { return v.Boolean("k", "y") },
			true,
			make(map[string][]string),
		},
		{
			func(v Validator) bool { return v.Boolean("k", "yes") },
			true,
			make(map[string][]string),
		},
		{
			func(v Validator) bool { return v.Boolean("k", "t") },
			true,
			make(map[string][]string),
		},
		{
			func(v Validator) bool { return v.Boolean("k", "true") },
			true,
			make(map[string][]string),
		},
		{
			func(v Validator) bool { return v.Boolean("k", "TRUE") },
			true,
			make(map[string][]string),
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			v := New()
			i := tt.val(v)

			if !reflect.DeepEqual(v.Errors, tt.wantErrors) {
				t.Errorf("\nout:  %#v\nwant: %#v\n", v.Errors, tt.wantErrors)
			}

			if i != tt.want {
				t.Errorf("\nout:  %#v\nwant: %#v\n", i, tt.want)
			}
		})
	}
}

func TestErrorOrNil(t *testing.T) {
	tests := []struct {
		in   *Validator
		want error
	}{
		{&Validator{}, nil},
		{&Validator{Errors: map[string][]string{}}, nil},
		{
			&Validator{Errors: map[string][]string{"x": []string{"X"}}},
			&Validator{Errors: map[string][]string{"x": []string{"X"}}},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			got := tt.in.ErrorOrNil()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("\nout:  %#v\nwant: %#v\n", got, tt.want)
			}
		})
	}
}
