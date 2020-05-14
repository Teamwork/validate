package validate

import (
	"image"

	"net/http"
	"os"
	"strings"
)

//Supported Image Formats/Mime Types
var (
	supportedImageFormats = map[string]string{
		"jpeg": "image/jpeg", "png": "image/png", "gif": "image/gif", "jpg": "image/jpeg",
	}
)

//getFileMimeType returns the
func getFileMimeType(file *os.File) (string, error) {
	file.Seek(0, 0)
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
