// Package validate provides validation for HTTP request parameters.
//
// Basic usage example:
//
//   v := validate.New()
//   v.Required("firstName", customer.FirstName)
//   if v.HasErrors() {
//       fmt.Println("Had the following validation errors:")
//       for key, errors := range v.Errors {
//           fmt.Printf("    %s: %s", key, strings.Join(errors))
//       }
//   }
//
// All validators treat the input's zero type (empty string, 0, nil, etc.) as
// valid. Use the Required() validator if you want to make a parameter required.
//
// All validators optionally accept a custom message as the last parameter:
//
//   v.Required("key", value, "you really need to set this")
//
// The error text only includes a simple human description such as "must be set"
// or "must be a valid email". When adding new validations, make sure that they
// can be displayed properly when joined with commas. A text such as "Error:
// this field must be higher than 42" would look weird:
//
//   must be set, Error: this field must be higher than 42
//
// You can set your own errors with v.Append("key", "message"):
//
//   if !condition {
//       v.Append("key", "must be a valid foo")
//   }
package validate // import "github.com/teamwork/validate"

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
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

// Code for the error. Satisfies the guru.coder interface in
// github.com/teamwork/guru.
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

// ErrorOrNil returns nil if there are no errors, or the Validator object if there are.
//
// This makes it a bit more elegant to return from a function:
//
//   if v.HasErrors() {
//       return v
//   }
//   return nil
//
// Can now be:
//
//   return v.ErrorOrNil()
func (v *Validator) ErrorOrNil() error {
	if v.HasErrors() {
		return v
	}
	return nil
}

// Sub allows to specific sub-validations.
//
// Errors from the subvalidation is merged with the top-level one, the keys are
// added as "top.sub" or "top[n].sub".
//
// If the error is not a Validator the text will be added as just the key name
// without subkey (i.e. the same as v.Append("key", "msg")).
//
// For example:
//
//   v := validate.New()
//   v.Required("name", customer.Name)
//
//   // e.g. "settings.domain"
//   v.Sub("settings", -1, customer.Settings.Validate())
//
//   // e.g. "addresses[1].city"
//   for i, a := range customer.Addresses {
//       a.Sub("addresses", i, c.Validate())
//   }
func (v *Validator) Sub(key string, n int, err error) {
	if err == nil {
		return
	}

	if n > -1 {
		key = fmt.Sprintf("%s[%d]", key, n)
	}

	sub, ok := err.(*Validator)
	if !ok {
		ss, ok := err.(Validator)
		if !ok {
			v.Append(key, err.Error())
			return
		}
		sub = &ss
	}
	if !sub.HasErrors() {
		return
	}

	for k, val := range sub.Errors {
		mk := fmt.Sprintf("%s.%s", key, k)
		v.Errors[mk] = append(v.Errors[mk], val...)
	}
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
	keys := make([]string, len(v.Errors))
	i := 0
	for k := range v.Errors {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		s := fmt.Sprintf("%s: %s.\n", k, strings.Join(v.Errors[k], ", "))
		b.WriteString(s)

	}
	return b.String()
}

// Required indicates that this value must not be the type's zero value.
//
// Currently supported types are string, int, int64, uint, and uint64. It will
// panic if the type is not supported.
func (v *Validator) Required(key string, value interface{}, message ...string) {
	msg := getMessage(message, MessageRequired)

	switch val := value.(type) {
	case string:
		if strings.TrimSpace(val) == "" {
			v.Append(key, msg)
		}
	case int, int64, uint, uint64:
		if val == 0 {
			v.Append(key, msg)
		}
	case bool:
		if !val {
			v.Append(key, msg)
		}
	case mailaddress.Address:
		if val.Address == "" {
			v.Append(key, msg)
		}
	case mailaddress.List:
		if len(val) == 0 {
			v.Append(key, msg)
		}
	default:
		panic(fmt.Sprintf("validate: not a supported type: %T", value))
	}
}

// Exclude validates that the value is not in the exclude list.
//
// This list is matched case-insensitive.
func (v *Validator) Exclude(key, value string, exclude []string, message ...string) {
	msg := getMessage(message, "")

	value = strings.TrimSpace(strings.ToLower(value))
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

	value = strings.TrimSpace(strings.ToLower(value))
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
	if !validDomain(value) {
		v.Append(key, msg)
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

func validDomain(v string) bool {
	return reValidDomain.MatchString(v)
}

// URL validates that the string contains a valid URL.  The scheme is optional.
// The URL may consist of a scheme, host, path, and query parameters.  Only the
// host is required.  If the scheme is not given it will be prepended.
func (v *Validator) URL(key, value string, message ...string) {
	if value == "" {
		return
	}

	msg := getMessage(message, MessageURL)

	u, err := url.Parse(value)
	if err != nil && u == nil {
		v.Append(key, "%s: %s", msg, err)
		return
	}

	// If we don't have a scheme the parse may or may not fail according to the
	// go docs. "Trying to parse a hostname and path without a scheme is invalid
	// but may not necessarily return an error, due to parsing ambiguities."
	if u.Scheme == "" {
		u.Scheme = "http"
		u, err = url.Parse(u.String())
	}

	if err != nil {
		v.Append(key, "%s: %s", msg, err)
		return
	}

	if u.Host == "" {
		v.Append(key, msg)
		return
	}

	if !validDomain(u.Host) {
		v.Append(key, msg)
		return
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

var rePhone = regexp.MustCompile(`^[0123456789+\-() .]{5,20}$`)

// Phone checks if the string looks like a valid phone number.
//
// https://en.wikipedia.org/wiki/National_conventions_for_writing_telephone_numbers
func (v *Validator) Phone(key, value string, message ...string) {
	if value == "" {
		return
	}

	msg := getMessage(message, MessagePhone)
	if !rePhone.MatchString(value) {
		v.Append(key, msg)
	}
}
