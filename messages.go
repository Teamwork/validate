package validate

// Messages for the checkers; this can be changed for i18n.
var (
	MessageRequired          = "must be set"
	MessageDomain            = "must be a valid domain"
	MessageURL               = "must be a valid url"
	MessageEmail             = "must be a valid email address"
	MessageIPv4              = "must be a valid IPv4 address"
	MessageHexColor          = "must be a valid color code"
	MessageLenLonger         = "must be longer than %d characters"
	MessageLenShorter        = "must be shorter than %d characters"
	MessageExclude           = "cannot be ‘%s’"
	MessageInclude           = "must be one of ‘%s’"
	MessageInteger           = "must be a whole number"
	MessageBool              = "must be a boolean"
	MessageDate              = "must be a date as ‘%s’"
	MessagePhone             = "must be a valid phone number"
	MessageRangeHigher       = "must be higher than %d"
	MessageRangeLower        = "must be lower than %d"
	MessageNotAnImage        = "must be an image"
	MessageImageFormat       = "must be an image of '%s' format"
	MessageImageDimension    = "image dimension (W x H) must be between '%d x %d' and '%d x %d' pixels"
	MessageImageMinDimension = "image dimension (W x H) cannot be less than '%d x %d' pixels"
	MessageImageMaxDimension = "image dimension (W x H) cannot be more than '%d x %d' pixels"
	MessageFileMimeType      = "must be a file of type '%s'"
	MessageFileSize          = "file size must be between '%.1f'KB and '%.1f'KB"
	MessageFileMaxSize       = "file size cannot be larger than '%.1f'KB"
	MessageFileMinSize       = "file size cannot be less than '%.1f'KB"
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
