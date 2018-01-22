// Package validate provides validation for HTTP request parameters.
//
// Basic usage example:
//
//   v := validate.New()
//   v.Required("firstName", customer.FirstName)
//   if v.HasErrors() {
//       fmt.Println("Had the following validation errors:")
//       for range key, errors := v.Errors {
//           fmt.Printf("    %v: %v", key, strings.Join(errors))
//       }
//   }
//
//  When using it from echo, you can use this helper to return the appropriate
//  JSON response:
//
//    if v.HasErrors() {
//        return httperr.Validation(c, v.Errors)
//    }
//
// All validators treat the input's zero type (empty string, 0, nil, etc.) as
// valid. If you want to make a parameter required use the Required() validator.
//
// The error text only includes a simple human description such as "must be set"
// or "must be a valid email". When adding new validations, make sure that they
// can be displayed properly when joined with comma's. A text such as "Error:
// this field must be high than 42" would look weird:
//
//   must be set, Error: this field must be high than 42
package validate // import "github.com/teamwork/validate"

import (
	"fmt"
	"net"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/teamwork/mailaddress"
)

// Validator hold the validation errors.
//
// Typically you shouldn't create this directly but use the New() function.
type Validator struct {
	Errors map[string][]string
	// TagName is the struct field tag that will be look for
	// when called validator.Validate(i)
	TagName string
}

// New makes a new Validator and ensures that it is properly initialized.
func New() Validator {
	v := Validator{}
	v.TagName = "validate"
	v.Errors = make(map[string][]string)
	return v
}

// Append a new error to the error list for this key.
func (v *Validator) Append(key, value string, format ...interface{}) {
	v.Errors[key] = append(v.Errors[key], fmt.Sprintf(value, format...))
}

// HasErrors reports if this validation has any errors.
func (v *Validator) HasErrors() bool {
	return len(v.Errors) > 0
}

// Merge errors from another validator in to this one.
func (v *Validator) Merge(other Validator) {
	for k, val := range other.Errors {
		v.Errors[k] = append(v.Errors[k], val...)
	}
}

// Strings representation shows either all errors or "<no errors>" if there are
// no errors.
func (v *Validator) String() string {
	if !v.HasErrors() {
		return "<no errors>"
	}

	// Make sure the order is always the same.
	keys := make(sort.StringSlice, len(v.Errors))
	i := 0
	for k := range v.Errors {
		keys[i] = k
		i++
	}
	sort.Sort(keys)

	s := ""
	for _, k := range keys {
		s += fmt.Sprintf("%v: %v.\n", k, strings.Join(v.Errors[k], ", "))
	}
	return s
}

// Required indicates that this value must not be the type's zero value.
//
// Currently supported types are string, int, int64, uint, and uint64. It will
// panic if the type is not supported.
func (v *Validator) Required(key string, value interface{}, message ...string) {
	msg := getMessage(message, MessageRequired)

	switch value.(type) {
	case string:
		if strings.TrimSpace(value.(string)) == "" {
			v.Append(key, msg)
		}
	case int:
		if value.(int) == 0 {
			v.Append(key, msg)
		}
	case int64:
		if value.(int64) == 0 {
			v.Append(key, msg)
		}
	case uint:
		if value.(uint) == 0 {
			v.Append(key, msg)
		}
	case uint64:
		if value.(uint64) == 0 {
			v.Append(key, msg)
		}
	case bool:
		if !value.(bool) {
			v.Append(key, msg)
		}
	default:
		panic(fmt.Sprintf("not a supported type: %T", value))
	}
}

var reValidDomain = regexp.MustCompile(`` +
	// Anchor
	`^` +

	// See RFC 1034, section 3.1, RFC 1035, secion 2.3.1
	//
	// - Only allow letters, numbers
	// - Max size of a single label is 63 characters (RFC specifies bytes, but that's
	//   not so easy to check AFAIK).
	// - Need at least two labels
	`[\p{L}\d-]{1,63}` + // Label
	`(\.[\p{L}\d-]{1,63})+` + // More labels

	// Anchor
	`$`,
)

// Exclude validates that the value is not in the exclude list.
//
// This list is matched case-insensitive.
func (v *Validator) Exclude(key, value string, exclude []string, message ...string) {
	msg := getMessage(message, "")

	value = strings.ToLower(value)
	for _, e := range exclude {
		if strings.ToLower(e) == value {
			if msg != "" {
				v.Append(key, msg)
			} else {
				v.Append(key, fmt.Sprintf(MessageExclude, e))
			}
			return
		}
	}
}

// Include validates that the value is in the include list.
//
// This list is matched case-insensitive.
func (v *Validator) Include(key, value string, include []string, message ...string) {
	if len(include) == 0 {
		return
	}

	value = strings.ToLower(value)
	for _, e := range include {
		if strings.EqualFold(e, value) {
			return
		}
	}

	msg := getMessage(message, "")
	if msg != "" {
		v.Append(key, msg)
	} else {
		v.Append(key, fmt.Sprintf(MessageInclude, strings.Join(include, ", ")))
	}
}

// Domain validates that the domain is valid. A domain must consist of at least
// two labels (so "com" in not valid, whereas "example.com" is).
func (v *Validator) Domain(key, value string, message ...string) {
	if value == "" {
		return
	}

	msg := getMessage(message, MessageDomain)
	if !reValidDomain.MatchString(value) {
		v.Append(key, msg)
	}
}

// Email gives an error if this does not look like a valid email address.
func (v *Validator) Email(key, value string, message ...string) mailaddress.Address {
	if value == "" {
		return mailaddress.Address{}
	}

	msg := getMessage(message, MessageEmail)
	addr, err := mailaddress.Parse(value)
	if err != nil {
		v.Append(key, msg)
	}
	return addr
}

// IPv4 validates that a string is a valid IPv4 address.
func (v *Validator) IPv4(key, value string, message ...string) net.IP {
	if value == "" {
		return net.IP{}
	}

	msg := getMessage(message, MessageIPv4)
	ip := net.ParseIP(value)
	if ip == nil || ip.To4() == nil {
		v.Append(key, msg)
	}
	return ip
}

var reValidHexColor = regexp.MustCompile(`(?i)^#[0-9a-f]{3,6}$`)

// HexColor validates if the string looks like a color as a hex triplet (e.g.
// #ffffff or #fff).
func (v *Validator) HexColor(key, value string, message ...string) {
	if value == "" {
		return
	}

	msg := getMessage(message, MessageHexColor)
	if !reValidHexColor.MatchString(value) {
		v.Append(key, msg)
	}
}

// Len sets the minimum and maximum length for a string.
//
// The maximum length can be 0, indicating there is no upper limit.
func (v *Validator) Len(key, value string, min, max int, message ...string) {
	msg := getMessage(message, "")

	switch {
	case len(value) < min:
		if msg != "" {
			v.Append(key, msg)
		} else {
			v.Append(key, fmt.Sprintf(MessageLenLonger, min))
		}
	case max > 0 && len(value) > max:
		if msg != "" {
			v.Append(key, msg)
		} else {
			v.Append(key, fmt.Sprintf(MessageLenShorter, max))
		}
	}
}

// Numeric checks if this looks like a numeric value.
//
// Right now this accepts while integers only; both positive and negative.
func (v *Validator) Numeric(key, value string, message ...string) {
	_, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		v.Append(key, getMessage(message, MessageNumeric))
	}
}

// Date checks if the string looks like a date in the given layout.
func (v *Validator) Date(key, value, layout string, message ...string) {
	msg := getMessage(message, "")
	_, err := time.Parse(layout, value)
	if err != nil {
		if msg != "" {
			v.Append(key, msg)
		} else {
			v.Append(key, fmt.Sprintf(MessageDate, layout))
		}
	}
}

// Validate looks into each field tag and validates according to:
//	type MyStruct struct {
//		Field  string `validate:"required"`
//		Field2 string `validate:"include:[a b c]"`
//		Field3 string `validate:"date:2006-01-02T15:04:05Z07:00"`
//		Field4 string `validate:"len:[5 9]"`
//	}
//	v.Validate(MyStruct{})
func (v *Validator) Validate(i interface{}) {
	valueof := reflect.ValueOf(i)
	typeof := reflect.TypeOf(i)
	for i := 0; i < valueof.NumField(); i++ {
		field := typeof.Field(i)
		if tag := field.Tag.Get(v.TagName); tag != "" {
			opts := parseOptions(tag)
			fieldInt := valueof.Field(i).Interface()

			switch e := fieldInt.(type) {
			case string:
				v.validateString(e, field.Name, opts)
				// wat. nothing else supported?
			}
		}
	}
}

func (v *Validator) validateString(s string, fieldName string, opts validateOpts) {
	if opts.Required && s == "" {
		v.Required(fieldName, s)
	}

	if s == "" {
		return
	}

	if opts.Domain {
		v.Domain(fieldName, s)
	}

	if opts.Email {
		v.Email(fieldName, s)
	}

	if opts.HexColor {
		v.HexColor(fieldName, s)
	}

	if opts.IPv4 {
		v.IPv4(fieldName, s)
	}

	if opts.Date != "" {
		v.Date(fieldName, s, opts.Date)
	}

	if opts.Numeric {
		v.Numeric(fieldName, s)
	}

	if len(opts.Include) != 0 {
		v.Include(fieldName, s, opts.Include)
	}

	if len(opts.Exclude) != 0 {
		v.Exclude(fieldName, s, opts.Exclude)
	}

	if opts.Len != nil {
		v.Len(fieldName, s, opts.Len.Min, opts.Len.Max)
	}
}

var lenRE = regexp.MustCompile(`len:\\[(\\d+)? (\\d+)?\\]`)
var dateRE = regexp.MustCompile(`date:(.+)`)
var includeRE = regexp.MustCompile(`include:\\[(.+( .+)*)\\]`)
var excludeRE = regexp.MustCompile(`exclude:\\[(.+( .+)*)\\]`)

func parseOptions(s string) validateOpts {
	opts := validateOpts{}

	tagOpts := strings.Split(s, ",")
	for _, tagOpt := range tagOpts {
		opts.Required = tagOpt == "required"
		opts.Domain = tagOpt == "domain"
		opts.Email = tagOpt == "email"
		opts.HexColor = tagOpt == "hex"
		opts.IPv4 = tagOpt == "ipv4"
		opts.Numeric = tagOpt == "numeric"

		if includeRE.MatchString(tagOpt) {
			matches := includeRE.FindAllStringSubmatch(s, -1)
			opts.Include = strings.Split(matches[0][1], " ")
		}

		if excludeRE.MatchString(tagOpt) {
			matches := excludeRE.FindAllStringSubmatch(s, -1)
			opts.Exclude = strings.Split(matches[0][1], " ")
		}

		if dateRE.MatchString(tagOpt) {
			matches := dateRE.FindAllStringSubmatch(s, -1)
			opts.Date = matches[0][1]
		}

		if lenRE.MatchString(tagOpt) {
			opts.Len = &struct{ Min, Max int }{}
			matches := lenRE.FindAllStringSubmatch(s, -1)

			var err error
			opts.Len.Min, err = strconv.Atoi(matches[0][1])
			if err != nil {
				opts.Len.Min = -1
			}
			opts.Len.Max, err = strconv.Atoi(matches[0][2])
			if err != nil {
				opts.Len.Max = -1
			}
		}
	}

	return opts
}

type validateOpts struct {
	Required bool
	Domain   bool
	Email    bool
	Numeric  bool
	HexColor bool
	IPv4     bool
	Len      *struct{ Min, Max int }
	Date     string
	Exclude  []string
	Include  []string
}
