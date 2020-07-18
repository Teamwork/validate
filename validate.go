// Package validate provides simple validation for Go.
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
// You can set your own errors with v.Append():
//
//   if !condition {
//       v.Append("key", "must be a valid foo")
//   }
//
// Some validators return the parsed value, which makes it easier both validate
// and get a useful value at the same time:
//
//   v := validate.New()
//   id := v.Integer("id", c.Param("id"))
//   if v.HasErrors() {
//       return v
//   }
//   user := getUserByID(id)
package validate // import "github.com/teamwork/validate"

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net"
	"net/url"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

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

// Code returns the HTTP status code for the error. Satisfies the guru.coder
// interface in github.com/teamwork/guru.
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
// Errors from the subvalidation are merged with the top-level one, the keys are
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
func (v *Validator) Sub(key, subKey string, err error) {
	if err == nil {
		return
	}

	if subKey != "" {
		key = fmt.Sprintf("%s[%s]", key, subKey)
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
	case *string:
		if val == nil || strings.TrimSpace(*val) == "" {
			v.Append(key, msg)
		}
	case int:
		if val == int(0) {
			v.Append(key, msg)
		}
	case int64:
		if val == int64(0) {
			v.Append(key, msg)
		}
	case uint:
		if val == uint(0) {
			v.Append(key, msg)
		}
	case uint64:
		if val == uint64(0) {
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
	case []int64:
		if len(val) == 0 {
			v.Append(key, msg)
		}
	case []string:
		if len(val) == 0 {
			v.Append(key, msg)
			return
		}

		// Make sure there is at least one non-empty entry.
		nonEmpty := false
		for i := range val {
			if val[i] != "" { // Consider " " to be non-empty on purpose.
				nonEmpty = true
				break
			}
		}

		if !nonEmpty {
			v.Append(key, msg)
		}
	//Added Required for File
	case *multipart.FileHeader:
		_, err := val.Open()
		if val == nil || val.Size <= 0 || err != nil {
			v.Append(key, msg)
		}

	case multipart.File:
		b := make([]byte, 10)
		num, _ := val.Read(b)
		if num == 0 {
			v.Append(key, msg)
		}
	default:
		if vv := reflect.ValueOf(value); vv.Kind() == reflect.Ptr {
			if value == reflect.Zero(vv.Type()).Interface() {
				v.Append(key, msg)
			}
			return
		}
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

// Domain validates that the domain is valid.
//
// A domain must consist of at least two labels. So "com" or "localhost" – while
// technically valid domain names – are not accepted, whereas "example.com" or
// "me.localhost" are. For the overwhelming majority of applications this makes
// the most sense.
//
// This works for internationalized domain names (IDN), either as UTF-8
// characters or as punycode.
//
// Limitation: the RFC limits domain labels to 63 bytes, but this validation
// accepts labels up to 63 *characters*.
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

// URL validates that the string contains a valid URL.
//
// The URL may consist of a scheme, host, path, and query parameters. Only the
// host is required.
//
// The host is validated with the Domain() validation.
//
// If the scheme is not given "http" will be prepended.
func (v *Validator) URL(key, value string, message ...string) *url.URL {
	if value == "" {
		return nil
	}

	msg := getMessage(message, MessageURL)

	u, err := url.Parse(value)
	if err != nil && u == nil {
		v.Append(key, "%s: %s", msg, err)
		return nil
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
		return nil
	}

	if u.Host == "" {
		v.Append(key, msg)
		return nil
	}

	host := u.Host
	if h, _, err := net.SplitHostPort(u.Host); err == nil {
		host = h
	}

	if !validDomain(host) {
		v.Append(key, msg)
		return nil
	}

	return u
}

// Email validates if this email looks like a valid email address.
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

// Len sets the minimum and maximum length for a string in characters, not in
// bytes.
//
// A maximum of 0 indicates there is no upper limit.
func (v *Validator) Len(key, value string, min, max int, message ...string) {
	msg := getMessage(message, "")

	length := utf8.RuneCountInString(value)

	switch {
	case length < min:
		if msg != "" {
			v.Append(key, msg)
		} else {
			v.Append(key, fmt.Sprintf(MessageLenLonger, min))
		}
	case max > 0 && length > max:
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
// There are a great amount of writing conventions for phone numbers:
// https://en.wikipedia.org/wiki/National_conventions_for_writing_telephone_numbers
//
// This merely checks a field contains 5 to 20 characters "0123456789+\-() .",
// which is not very strict but should cover all conventions.
func (v *Validator) Phone(key, value string, message ...string) {
	if value == "" {
		return
	}

	msg := getMessage(message, MessagePhone)
	if !rePhone.MatchString(value) {
		v.Append(key, msg)
	}
}

// Range sets the minimum and maximum value of a integer.
//
// A maximum of 0 indicates there is no upper limit.
func (v *Validator) Range(key string, value, min, max int64, message ...string) {
	msg := getMessage(message, "")

	if value < min {
		if msg != "" {
			v.Append(key, msg)
		} else {
			v.Append(key, fmt.Sprintf(MessageRangeHigher, min))
		}
	}
	if max > 0 && value > max {
		if msg != "" {
			v.Append(key, msg)
		} else {
			v.Append(key, fmt.Sprintf(MessageRangeLower, max))
		}
	}
}

//IsImage checks if the file is an image.
//This only checks for JPEG, PNG or GIF  MIME Header:
//It accepts pointer to multipart.FileHeader and formats(string)
//Multiple formats can be stated and should be separated by comma(','), eg: "jpeg, png, gif"
//  For Example:
//	func uploadFile(w http.ResponseWriter, r *http.Request) {
//		r.ParseMultipartForm(10 << 20)
// 		file, fileHeader, err := r.FormFile("myFile")
// 		if err != nil {
// 			fmt.Println("Error Retrieving the File")
// 			fmt.Println(err)
// 			return
// 		}
// 		defer file.Close()
// 		validator := validate.New()
// 		validator.IsImage("myFile", fileHeader, "") OR
// 		validator.IsImage("myFile", fileHeader, "jpeg") OR
// 		validator.IsImage("myFile", fileHeader, "jpeg, png")
// 		if validator.HasErrors() {
// 			fmt.Fprintf(w, validator.String())
// 		}
//	}
func (v *Validator) IsImage(key string,
	fileHeader *multipart.FileHeader,
	formats string, message ...string) bool {
	uploadedTypes := fileHeader.Header["Content-Type"]
	msg := getMessage(message, "")

	for _, uploadedType := range uploadedTypes {
		if isFileImage(uploadedType, strings.ToLower(formats)) {
			return true
		}
	}

	//Append a different error if Image format is present
	if msg != "" {
		v.Append(key, msg)
	} else if formats != "" {
		v.Append(key, fmt.Sprintf(MessageImageFormat, formats))
	} else {
		v.Append(key, MessageNotAnImage)
	}
	return false
}

//ImageDimensions  validates the maximum and minimum dimensions of an image in Pixels
//minDimension speficies the lower limit of the image dimension while maxDimension specifies the upper limit
//You can use nil to replace any of minDimension or maxDimension to indicate no limit
//It accepts pointer to multipart.FileHeader as the first parameter
//Both minDimension & maxDimension are of pointer to ImageDimension
//  For Example:
//	func uploadFile(w http.ResponseWriter, r *http.Request) {
//		r.ParseMultipartForm(10 << 20)
// 		file, fileHeader, err := r.FormFile("myFile")
// 		if err != nil {
// 			fmt.Println("Error Retrieving the File")
// 			fmt.Println(err)
// 			return
// 		}
// 		defer file.Close()
// 		validator := validate.New()
// 		minDim := validate.ImageDimension{Width: 200, Height: 300}
// 		maxDim := validate.ImageDimension{Width: 300, Height: 350}
// 		validator.ImageDimensions("myFile", fileHeader, &minDim, &maxDim, "")
// 		if validator.HasErrors() {
// 			fmt.Fprintf(w, validator.String())
// 		}
//	}
func (v *Validator) ImageDimensions(
	key string,
	fileHeader *multipart.FileHeader,
	minDimension, maxDimension *ImageDimension,
	message ...string) {

	if maxDimension == nil && minDimension == nil {
		panic("You must specify either minimum dimension or maximum dimension!")
	}

	msg := getMessage(message, "")

	//Confirm Uploaded file is Image
	if !v.IsImage(key, fileHeader, "", "File is not an image. Only dimensions of image files can be determined.") {
		return
	}

	//Get real dimension
	realDim, err := getDimension(fileHeader)

	if err != nil {
		v.Append(key, fmt.Sprintf("Error getting uploaded image dimension. Please try again."))
		return
	}
	//Init Error Message
	var minDimErrorMsg, maxDimErrorMsg string = "", ""

	//Check for minimum dimension requirement
	if minDimension != nil && (realDim.Width < minDimension.Width || realDim.Height < minDimension.Height) {
		minDimErrorMsg = fmt.Sprintf(MessageImageMinDimension, minDimension.Width, minDimension.Height)
	}

	//Check for maximum dimension requirements
	if maxDimension != nil && (realDim.Width > maxDimension.Width || realDim.Height > maxDimension.Height) {
		maxDimErrorMsg = fmt.Sprintf(MessageImageMaxDimension, maxDimension.Width, maxDimension.Height)
	}

	//Append appropriate errors
	if msg != "" {
		v.Append(key, msg)
	} else if minDimErrorMsg != "" && maxDimErrorMsg != "" {
		v.Append(key, fmt.Sprintf(MessageImageDimension,
			minDimension.Width,
			minDimension.Height,
			maxDimension.Width,
			maxDimension.Height))
	} else if maxDimErrorMsg != "" {
		v.Append(key, maxDimErrorMsg)
	} else if minDimErrorMsg != "" {
		v.Append(key, minDimErrorMsg)
	}

}

//FileSize validates the minimum size ytes a file can be
//This can be used with any type of files as well as images
//It accepts pointer to multipart.FileHeader as the first parameter.
//Minimum and maximum sizes of files in integer
//Use '-1' for any size
//  For Example:
//	func uploadFile(w http.ResponseWriter, r *http.Request) {
//		r.ParseMultipartForm(10 << 20)
// 		file, fileHeader, err := r.FormFile("myFile")
// 		if err != nil {
// 			fmt.Println("Error Retrieving the File")
// 			fmt.Println(err)
// 			return
// 		}
// 		defer file.Close()
// 		validator := validate.New()
// 		validator.FileSize("myFile", fileHeader, 100000, 200000, "")
// 		if validator.HasErrors() {
// 			fmt.Fprintf(w, validator.String())
// 		}
//	}
func (v *Validator) FileSize(key string,
	fileHeader *multipart.FileHeader,
	minSizeInBytes, maxSizeInBytes int64,
	message ...string) {

	if minSizeInBytes == 0 && maxSizeInBytes == 0 {
		panic("File size cannot be zeros. Minimum or maximum size in bytes must be specified")
	}
	//Return if all negative
	if minSizeInBytes < 0 && maxSizeInBytes < 0 {
		return
	}

	msg := getMessage(message, "")
	//initialise error messages
	var minSizeErrMsg, maxSizeErrMsg = "", ""

	//For mimum size requirement
	if minSizeInBytes >= 0 && fileHeader.Size < minSizeInBytes {
		minSizeErrMsg = fmt.Sprintf(MessageFileMinSize, bytesToKiloBytes(minSizeInBytes))
	}

	//For maximum size requirement
	if maxSizeInBytes >= 0 && fileHeader.Size > maxSizeInBytes {
		maxSizeErrMsg = fmt.Sprintf(MessageFileMaxSize, bytesToKiloBytes(maxSizeInBytes))
	}

	//Append appropriate errors
	if msg != "" {
		v.Append(key, msg)
	} else if minSizeErrMsg != "" && maxSizeErrMsg != "" {
		v.Append(key,
			fmt.Sprintf(MessageFileSize, bytesToKiloBytes(minSizeInBytes),
				bytesToKiloBytes(maxSizeInBytes)))
	} else if minSizeErrMsg != "" {
		v.Append(key, minSizeErrMsg)
	} else if maxSizeErrMsg != "" {
		v.Append(key, maxSizeErrMsg)
	}

}

//FileMimeType validates file mime type
//It accepts pointer to multipart.FileHeader as the first parameter,  mime type(s)
//Multiple mimeType formats can be separated by comma Eg: "image/jpeg, text/csv, applications/pdf"
//For more information on mime types, visit:
//https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types
//https://www.iana.org/assignments/media-types/media-types.xhtml
//  For Example:
//	func uploadFile(w http.ResponseWriter, r *http.Request) {
//		r.ParseMultipartForm(10 << 20)
// 		file, fileHeader, err := r.FormFile("myFile")
// 		if err != nil {
// 			fmt.Println("Error Retrieving the File")
// 			fmt.Println(err)
// 			return
// 		}
// 		defer file.Close()
// 		validator := validate.New()
// 		validator.FileMimeType("myFile", fileHeader, "application/pdf", "")
// 		if validator.HasErrors() {
// 			fmt.Fprintf(w, validator.String())
// 		}
// }
func (v *Validator) FileMimeType(key string,
	fileHeader *multipart.FileHeader,
	requiredMimeTypes string,
	message ...string) {

	if requiredMimeTypes == "" {
		panic("Required mime type cannot be empty.")
	}
	//Get uploaded content type
	uploadedTypes := fileHeader.Header["Content-Type"]
	msg := getMessage(message, "")

	//Compare each with supllied Mime Type
	for _, uploadedType := range uploadedTypes {
		fmt.Println(uploadedTypes)
		//Return once there is  a match
		if isFileMimeTypeValid(uploadedType, requiredMimeTypes) {
			return
		}
	}

	//Append a different error if Image format is present
	if msg != "" {
		v.Append(key, msg)
	} else {
		v.Append(key, fmt.Sprintf(MessageFileMimeType, requiredMimeTypes))
	}
}
