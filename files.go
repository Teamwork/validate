package validate

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"os"
	"strings"
)

//Supported Image Formats/Mime Types
var (
	supportedImageFormats = map[string]string{
		"jpeg": "image/jpeg", "png": "image/png", "gif": "image/gif",
	}
)

//ImageOptions defines the options associated with the Image file
//These include Minimum & Maxium Size, Minimum & Maxium  Dimensions and Image format
type ImageOptions struct {
	//For Images, use either JPEG, JPG, PNG or GIF
	Format string
	//Defines minimum and maximum size of file
	fileSize
	//Maximum Dimension(Width, Height)
	MaxDimension [2]int32
	//Minimum Dimension(Width, Height)
	MinDimension [2]int32
}

//FileOptions For Files
type FileOptions struct {
	//File Mime Types
	MimeType string
	//Defines minimum and maximum size of file
	fileSize
}

//Represent File Sze
type fileSize struct {
	//Maximum Size in KiloBytes. Must be an Integer
	MaxSize int32
	//Minimun Size in KiloBytes. Must be an Integer
	MinSize int32
}

//getFileMimeType returns the
func getFileMimeType(file *os.File) (string, error) {

	buff := make([]byte, 512)

	_, err := file.Read(buff)
	if err != nil {
		return "", err
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buff)

	return contentType, nil
}

//isFileImage confirms if this file is and Image of jpeg, png, gif
//format should be separated by comma
func isFileImage(file *os.File, format string) bool {
	calculatedFormat, err := getFileMimeType(file)

	if err != nil {
		return false
	}
	//Required format supplied
	if format != "" {
		//Check format is defined in supported format map
		requiredFormat, ok := supportedImageFormats[format]
		if ok && (requiredFormat == calculatedFormat) {
			return true
		}
		//Try splitting the required format in case of multiple formats
		formatsArray := strings.Split(format, ",")
		//Iterate through splitted formats
		for _, val := range formatsArray {
			requiredFormat, ok := supportedImageFormats[val]
			if ok && (requiredFormat == calculatedFormat) {
				return true
			}
		}
		//return false if not match is found
		return false
	}
	//Check if the file is an image
	for _, requiredFormat := range supportedImageFormats {
		if requiredFormat == calculatedFormat {
			return true
		}
	}

	return false
}

//getDimensions returns the dimensions of the uploaded image
func getDimension(file *os.File) (int, int, error) {

	img, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, err
	}

	return img.Width, img.Height, nil
}

//isFileMimeTypeValid confirms if the supplied mime type matches that of the image
func isFileMimeTypeValid(file *os.File, mimeType string) bool {
	//return for empty mimeType
	if mimeType == "" {
		return false
	}
	//find mimeType for the file
	calculatedMimeType, err := getFileMimeType(file)
	//return on error
	if err != nil {
		return false
	}

	//Split mimetype to individual values
	mimeTypeArray := strings.Split(mimeType, ",")

	//Check for single value mimetype
	if len(mimeTypeArray) == 0 && mimeType == calculatedMimeType {
		return true
	}

	//Iterate through splitted types to determine if mimeType matches
	for _, requiredMimeTypet := range mimeTypeArray {
		if requiredMimeTypet == calculatedMimeType {
			return true
		}
	}
	//return false on no match
	return false
}
