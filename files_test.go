package validate

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"reflect"
	"testing"
)

var (
	testImageDir  = "test_images/"
	png2000x2000  = "test_png_2000_2000.png"
	jpeg2000x2000 = "test_jpeg_2000_2000.jpg"
	gif2000x2000  = "test_gif_2000_2000.gif"
	png1000x2000  = "test_png_1000_1000.png"
	jpeg1000x2000 = "test_jpeg_1000_1000.jpg"
	gif1000x2000  = "test_gif_1000_1000.gif"
)

func TestImageValidation(t *testing.T) {
	//Test Images
	jpegFile, pngFile, gifFile := getTestImages(2000, 2000)

	textFile, err := os.Open(makeOtherFiles("text_1.txt", "New text"))

	if err != nil {
		panic(err.Error())
	}

	tests := []struct {
		val        func(Validator)
		wantErrors map[string][]string
	}{
		{
			func(v Validator) { v.Image("k", jpegFile, "") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Image("k", pngFile, "") },
			make(map[string][]string),
		},
		{
			func(v Validator) { v.Image("k", gifFile, "") },
			make(map[string][]string),
		},

		{
			func(v Validator) { v.Image("k", textFile, "") },
			map[string][]string{"k": {"must be an image"}},
		},
		{
			func(v Validator) { v.Image("k", textFile, "Error") },
			map[string][]string{"k": {"Error"}},
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

//Test and Confirm Images Formats
func TestImageFormatValidation(t *testing.T) {
	//Test Images
	jpegFile, pngFile, gifFile := getTestImages(2000, 2000)

	textFile, err := os.Open(makeOtherFiles("text_1.txt", "New text"))

	if err != nil {
		panic(err.Error())
	}

	tests := []struct {
		testname   string
		val        func(Validator)
		wantErrors map[string][]string
	}{
		{
			"jpeg ok",
			func(v Validator) { v.ImageFormat("k", jpegFile, "JPEG", "") },
			make(map[string][]string),
		},
		{
			"png ok",
			func(v Validator) { v.ImageFormat("k", pngFile, "PNG", "") },
			make(map[string][]string),
		},
		{
			"gif ok",
			func(v Validator) { v.ImageFormat("k", gifFile, "gif", "") },
			make(map[string][]string),
		},
		//Wrong Image Format
		{
			"jpeg in, png wanted",
			func(v Validator) { v.ImageFormat("k", jpegFile, "PNG", "") },
			map[string][]string{"k": {"must be an image of 'PNG' format"}},
		},
		{
			"png in, jpeg wanted",
			func(v Validator) { v.ImageFormat("k", pngFile, "JPEG", "") },
			map[string][]string{"k": {"must be an image of 'JPEG' format"}},
		},
		{
			"gif in, png wanted",
			func(v Validator) { v.ImageFormat("k", gifFile, "PNG", "") },
			map[string][]string{"k": {"must be an image of 'PNG' format"}},
		},

		{
			"textfile in, png wanted",
			func(v Validator) { v.ImageFormat("k", textFile, "PNG", "") },
			map[string][]string{"k": {"must be an image of 'PNG' format"}},
		},
		{
			"textfile in, png wanted, custom error",
			func(v Validator) { v.ImageFormat("k", textFile, "PNG", "Error") },
			map[string][]string{"k": {"Error"}},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			v := New()
			tt.val(v)

			if !reflect.DeepEqual(v.Errors, tt.wantErrors) {
				t.Errorf("\nname:%s \nout:  %#v\nwant: %#v\n", tt.testname, v.Errors, tt.wantErrors)
			}
		})
	}
}

func TestImageDimensionValidation(t *testing.T) {

	//Test Images
	jpegFile, pngFile, gifFile := getTestImages(2000, 2000)

	//Create JP 1000x2000
	jpegFile1000x2000, err := os.Open(makeTestImage("JPEG", jpeg1000x2000, 1000, 2000))

	if err != nil {
		panic(err.Error())
	}

	//Create test Text File
	textFile, err := os.Open(makeOtherFiles("text_1.txt", "New text"))
	if err != nil {
		panic(err.Error())
	}

	tests := []struct {
		testname   string
		val        func(Validator)
		wantErrors map[string][]string
	}{
		{
			"jpeg ok",
			func(v Validator) { v.ImageDimension("k", jpegFile, 2000, 2000, "") },
			make(map[string][]string),
		},
		{
			"jpeg 1000x2000 ok",
			func(v Validator) { v.ImageDimension("k", jpegFile1000x2000, 1000, 2000, "") },
			make(map[string][]string),
		},
		{
			"png ok",
			func(v Validator) { v.ImageDimension("k", pngFile, 2000, 2000, "") },
			make(map[string][]string),
		},
		{
			"gif ok",
			func(v Validator) { v.ImageDimension("k", gifFile, 2000, 2000, "") },
			make(map[string][]string),
		},
		//Wrong Image dimension
		{
			"jpeg 2000x2000 in, 1000x2000 wanted",
			func(v Validator) { v.ImageDimension("k", jpegFile, 1000, 2000, "") },
			map[string][]string{"k": {"image dimension (W x H) must be '1000 x 2000' pixels"}},
		},
		{
			"png 2000x2000 in, 1000x2000 wanted",
			func(v Validator) { v.ImageDimension("k", pngFile, 1000, 2000, "") },
			map[string][]string{"k": {"image dimension (W x H) must be '1000 x 2000' pixels"}},
		},
		{
			"gif 2000x2000 in, 1000x2000 wanted",
			func(v Validator) { v.ImageDimension("k", gifFile, 1000, 2000, "") },
			map[string][]string{"k": {"image dimension (W x H) must be '1000 x 2000' pixels"}},
		},

		{
			"textfile in, png 1000x2000 wanted",
			func(v Validator) { v.ImageDimension("k", textFile, 1000, 2000, "") },
			map[string][]string{"k": {"must be an image", "image dimension (W x H) must be '1000 x 2000' pixels"}},
		},
		{
			"textfile in, png wanted, custom error",
			func(v Validator) { v.ImageDimension("k", textFile, 1000, 2000, "Error") },
			map[string][]string{"k": {"must be an image", "Error"}},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			v := New()
			tt.val(v)

			if !reflect.DeepEqual(v.Errors, tt.wantErrors) {
				t.Errorf("\nname:%s \nout:  %#v\nwant: %#v\n", tt.testname, v.Errors, tt.wantErrors)
			}
		})
	}
}

//Create Files
func getTestImages(w, h int) (*os.File, *os.File, *os.File) {
	jpegFile, err := os.Open(makeTestImage("JPEG", jpeg2000x2000, w, h))

	if err != nil {
		panic(err.Error())
	}
	//Create PNG
	pngFile, err := os.Open(makeTestImage("PNG", png2000x2000, w, h))

	if err != nil {
		panic(err.Error())
	}

	//Create GIF
	gifFile, err := os.Open(makeTestImage("GIF", gif2000x2000, w, h))

	if err != nil {
		panic(err.Error())
	}
	return jpegFile, pngFile, gifFile
}

//Make For a Test
func makeTestImage(format, name string, w, h int) string {

	newImage := image.NewRGBA(image.Rect(0, 0, w, h))
	fullName := testImageDir + name

	file, err := os.Create(fullName)
	if err != nil {
		panic("Error creating image: \n" + err.Error())
	}

	switch format {
	case "GIF":
		o := &gif.Options{NumColors: 10}
		gif.Encode(file, newImage, o)
		break
	case "JPEG":
		o := jpeg.Options{Quality: 80}
		jpeg.Encode(file, newImage, &o)
	default:
		png.Encode(file, newImage)
	}

	file.Close()

	return fullName
}

//Create other files types for testing
func makeOtherFiles(name, content string) string {
	fullName := testImageDir + name

	file, err := os.Create(fullName)

	if err != nil {
		panic("Error creating file: \n" + err.Error())
	}

	_, err = file.Write([]byte(content))

	if err != nil {
		panic("Error creating file: \n" + err.Error())
	}

	file.Close()

	return fullName
}
