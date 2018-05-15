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
//        return guru.Validation(c, v.Errors)
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
	"encoding/json"
	"fmt"
	"net"
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
	Errors map[string][]string `json:"errors"`
}

// New makes a new Validator and ensures that it is properly initialized.
func New() Validator {
	v := Validator{}
	v.Errors = make(map[string][]string)
	return v
}

// Error interface.
func (v Validator) Error() string { return v.String() }

// Code for the error. Satisfies the guru.coder interface.
func (v Validator) Code() int { return 400 }

// ErrorJSON for reporting errors as JSON.
func (v Validator) ErrorJSON() ([]byte, error) { return json.Marshal(v) }

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

// Integer checks if this looks like an integer (i.e. a whole number).
func (v *Validator) Integer(key, value string, message ...string) int64 {
	if value == "" {
		return 0
	}

	i, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		v.Append(key, getMessage(message, MessageInteger))
	}
	return i
}

// Boolean checks if this looks like a boolean value.
func (v *Validator) Boolean(key, value string, message ...string) bool {
	if value == "" {
		return false
	}

	switch strings.ToLower(value) {
	case "1", "y", "yes", "t", "true":
		return true
	case "0", "n", "no", "f", "false":
		return false
	}
	v.Append(key, getMessage(message, MessageBool))
	return false
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
