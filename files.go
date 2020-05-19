package validate

import (
	"fmt"
	"image"

	//For image encoding
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"mime/multipart"

	"strings"
)

//Supported Image Formats/Mime Types
var (
	supportedImageFormats = map[string]string{
		"jpeg": "image/jpeg", "png": "image/png", "gif": "image/gif", "jpg": "image/jpeg",
	}
)

//ImageDimension represents width and height of an image dimension in pixels
//This is required by image dimension validation
type ImageDimension struct {
	Width  int
	Height int
}

//isFileImage confirms if this file is and Image of jpeg, png, gif
//format should be separated by comma
func isFileImage(uploadedType, format string) bool {

	//Required format supplied
	if format != "" {
		//Check format is defined in supported format map
		requiredFormat, ok := supportedImageFormats[format]

		if ok && (strings.TrimSpace(requiredFormat) == uploadedType) {
			return true
		}
		//Try splitting the required format in case of multiple formats
		formatsArray := strings.Split(format, ",")
		//Iterate through splitted formats
		for _, val := range formatsArray {
			requiredFormat, ok := supportedImageFormats[val]
			if ok && (strings.TrimSpace(requiredFormat) == uploadedType) {
				return true
			}
		}
		//return false if not match is found
		return false
	}
	//Check if the file is an image
	for _, requiredFormat := range supportedImageFormats {

		if requiredFormat == uploadedType {
			return true
		}
	}

	return false
}

//getDimensions returns the dimensions of the uploaded image
func getDimension(fileHeader *multipart.FileHeader) (*ImageDimension, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("Error getting image dimension." + err.Error())
	}
	//Reset File
	file.Seek(0, 0)
	// buf := bufio.NewReader(file)

	img, _, err := image.DecodeConfig(file)

	if err != nil {
		fmt.Println(err.Error(), file)
		return nil, fmt.Errorf("Error getting image dimension." + err.Error())
	}

	return &ImageDimension{img.Width, img.Height}, nil
}

//isFileMimeTypeValid confirms if the supplied mime type matches that of the image
func isFileMimeTypeValid(uploadedMimeType, requiresMimeType string) bool {
	//Split mimetype to individual values
	mimeTypeArray := strings.Split(requiresMimeType, ",")
	//Check for single value mimetype
	if len(mimeTypeArray) == 0 && uploadedMimeType == strings.TrimSpace(requiresMimeType) {
		return true
	}

	//Iterate through splitted types to determine if mimeType matches
	for _, mimeType := range mimeTypeArray {
		if strings.TrimSpace(mimeType) == uploadedMimeType {
			return true
		}
	}
	//return false on no match
	return false
}

//Convert bytes to kilobytes
func bytesToKiloBytes(byteData int64) float64 {
	return math.Ceil(float64(byteData) / 1024)
}
