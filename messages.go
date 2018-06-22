package validate

// Messages for the checkers; this can be changed for i18n.
var (
	MessageRequired   = "must be set"
	MessageDomain     = "must be a valid domain"
	MessageEmail      = "must be a valid email address"
	MessageIPv4       = "must be a valid IPv4 address"
	MessageHexColor   = "must be a valid color code"
	MessageLenLonger  = "must be longer than %d characters"
	MessageLenShorter = "must be shorter than %d characters"
	MessageExclude    = "cannot be ‘%s’"
	MessageInclude    = "must be one of ‘%s’"
	MessageInteger    = "must be a whole number"
	MessageBool       = "must be a boolean"
	MessageDate       = "must be a date as ‘%s’"
	MessageValid      = "must be a valid ‘%s’"
)

func getMessage(in []string, def string) string {
	switch len(in) {
	case 0:
		return def
	case 1:
		return in[0]
	default:
		panic("can only pass one message")
	}
}
