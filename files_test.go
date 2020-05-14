package validate

import (
	"encoding/csv"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"reflect"
	"testing"

	"github.com/jung-kurt/gofpdf"
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

	textFile, err := os.Open(makeOtherFiles("text_1.txt", "text", "New text"))

	if err != nil {
		panic(err.Error())
	}

	defer closeAllFiles(jpegFile, pngFile, gifFile, textFile)

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

	textFile, err := os.Open(makeOtherFiles("text_1.txt", "text", "New text"))

	if err != nil {
		panic(err.Error())
	}

	defer closeAllFiles(jpegFile, pngFile, gifFile, textFile)
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
	textFile, err := os.Open(makeOtherFiles("text_1.txt", "text", "New text"))
	if err != nil {
		panic(err.Error())
	}

	defer closeAllFiles(jpegFile, pngFile, gifFile, textFile)

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
			func(v Validator) { v.ImageDimension("k", gifFile, 1000, 2000, "Error") },
			map[string][]string{"k": {"Error"}},
		},

		{
			"textfile in, png 1000x2000 wanted",
			func(v Validator) { v.ImageDimension("k", textFile, 1000, 2000, "") },
			map[string][]string{"k": {"must be an image"}},
		},
		{
			"textfile in, png wanted, custom error",
			func(v Validator) { v.ImageDimension("k", textFile, 1000, 2000, "Error") },
			map[string][]string{"k": {"must be an image"}},
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

func TestImageMaxDimensionValidation(t *testing.T) {

	//Test Images
	jpegFile, pngFile, gifFile := getTestImages(2000, 2000)

	//Create JP 1000x2000
	jpegFile1000x2000, err := os.Open(makeTestImage("JPEG", jpeg1000x2000, 1000, 2000))

	if err != nil {
		panic(err.Error())
	}

	//Create test Text File
	textFile, err := os.Open(makeOtherFiles("text_1.txt", "text", "New text"))
	if err != nil {
		panic(err.Error())
	}

	defer closeAllFiles(jpegFile, pngFile, gifFile, textFile)

	tests := []struct {
		testname   string
		val        func(Validator)
		wantErrors map[string][]string
	}{
		{
			"jpeg ok",
			func(v Validator) { v.ImageMaxDimension("k", jpegFile, 2000, 2000, "") },
			make(map[string][]string),
		},
		{
			"jpeg 1000x2000 ok",
			func(v Validator) { v.ImageMaxDimension("k", jpegFile1000x2000, 3000, 3000, "") },
			make(map[string][]string),
		},
		{
			"png ok",
			func(v Validator) { v.ImageMaxDimension("k", pngFile, 2500, 2000, "") },
			make(map[string][]string),
		},
		{
			"gif ok",
			func(v Validator) { v.ImageMaxDimension("k", gifFile, 2000, 2500, "") },
			make(map[string][]string),
		},
		//Wrong Image dimension
		{
			"jpeg 2000x2000 in, 1000x2000 wanted",
			func(v Validator) { v.ImageMaxDimension("k", jpegFile, 1000, 2000, "") },
			map[string][]string{"k": {"image dimension (W x H) cannot be more than '1000 x 2000' pixels respectively"}},
		},
		{
			"png 2000x2000 in, 1000x2000 wanted",
			func(v Validator) { v.ImageMaxDimension("k", pngFile, 1000, 2000, "") },
			map[string][]string{"k": {"image dimension (W x H) cannot be more than '1000 x 2000' pixels respectively"}},
		},
		{
			"gif 2000x2000 in, 2000x1000 wanted",
			func(v Validator) { v.ImageMaxDimension("k", gifFile, 2000, 1000, "") },
			map[string][]string{"k": {"image dimension (W x H) cannot be more than '2000 x 1000' pixels respectively"}},
		},

		{
			"textfile in, png 1000x2000 wanted",
			func(v Validator) { v.ImageMaxDimension("k", textFile, 1000, 2000, "") },
			map[string][]string{"k": {"must be an image"}},
		},
		{
			"textfile in, png wanted, custom error",
			func(v Validator) { v.ImageMaxDimension("k", textFile, 1000, 2000, "Error") },
			map[string][]string{"k": {"must be an image"}},
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

func TestImageMinDimensionValidation(t *testing.T) {

	//Test Images
	jpegFile, pngFile, gifFile := getTestImages(2000, 2000)

	//Create JP 1000x2000
	jpegFile1000x2000, err := os.Open(makeTestImage("JPEG", jpeg1000x2000, 1000, 2000))

	if err != nil {
		panic(err.Error())
	}

	//Create test Text File
	textFile, err := os.Open(makeOtherFiles("text_1.txt", "text", "New text"))
	if err != nil {
		panic(err.Error())
	}

	defer closeAllFiles(jpegFile, pngFile, gifFile, textFile)

	tests := []struct {
		testname   string
		val        func(Validator)
		wantErrors map[string][]string
	}{
		{
			"jpeg ok",
			func(v Validator) { v.ImageMinDimension("k", jpegFile, 2000, 2000, "") },
			make(map[string][]string),
		},
		{
			"jpeg 1000x2000 ok",
			func(v Validator) { v.ImageMinDimension("k", jpegFile1000x2000, 1000, 2000, "") },
			make(map[string][]string),
		},
		{
			"png ok",
			func(v Validator) { v.ImageMinDimension("k", pngFile, 2000, 1000, "") },
			make(map[string][]string),
		},
		{
			"gif ok",
			func(v Validator) { v.ImageMinDimension("k", gifFile, 2000, 2000, "") },
			make(map[string][]string),
		},
		//Wrong Image dimension
		{
			"jpeg 2000x2000 in, 3000x2000 wanted",
			func(v Validator) { v.ImageMinDimension("k", jpegFile, 3000, 2000, "") },
			map[string][]string{"k": {"image dimension (W x H) cannot be less than '3000 x 2000' pixels respectively"}},
		},
		{
			"png 2000x2000 in, 2000x3000 wanted",
			func(v Validator) { v.ImageMinDimension("k", pngFile, 2000, 3000, "") },
			map[string][]string{"k": {"image dimension (W x H) cannot be less than '2000 x 3000' pixels respectively"}},
		},
		{
			"gif 2000x2000 in, 3000x3000 wanted",
			func(v Validator) { v.ImageMinDimension("k", gifFile, 3000, 3000, "Error") },
			map[string][]string{"k": {"Error"}},
		},

		{
			"textfile in, png 1000x2000 wanted",
			func(v Validator) { v.ImageMinDimension("k", textFile, 1000, 2000, "") },
			map[string][]string{"k": {"must be an image"}},
		},
		{
			"textfile in, png wanted, custom error",
			func(v Validator) { v.ImageMinDimension("k", textFile, 1000, 2000, "Error") },
			map[string][]string{"k": {"must be an image"}},
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

func TestFileMaxSizeValidation(t *testing.T) {
	//Test Images
	jpegFile, pngFile, _ := getTestImages(2000, 2000)

	//Create test Text File
	textFile, err := os.Open(makeOtherFiles("text_1.txt", "text", "New text"))
	if err != nil {
		panic(err.Error())
	}

	//Create test Text File
	csvFile, err := os.Open(makeOtherFiles("testcsv.csv", "csv", ""))
	if err != nil {
		panic(err.Error())
	}

	defer closeAllFiles(jpegFile, pngFile, csvFile, textFile)

	tests := []struct {
		testname   string
		val        func(Validator)
		wantErrors map[string][]string
	}{
		{
			"jpeg ok",
			func(v Validator) { v.FileMaxSize("k", jpegFile, calculateFileSize(jpegFile), "") },
			make(map[string][]string),
		},
		{
			"png ok",
			func(v Validator) { v.FileMaxSize("k", pngFile, calculateFileSize(pngFile), "") },
			make(map[string][]string),
		},
		{
			"text ok",
			func(v Validator) { v.FileMaxSize("k", textFile, calculateFileSize(textFile), "") },
			make(map[string][]string),
		},
		{
			"csv ok",
			func(v Validator) { v.FileMaxSize("k", csvFile, calculateFileSize(csvFile), "") },
			make(map[string][]string),
		},
		//Bigger Sizes
		{
			"text 2 of size",
			func(v Validator) { v.FileMaxSize("k", textFile, 2*calculateFileSize(textFile), "") },
			make(map[string][]string),
		},
		{
			"csv 2000KB",
			func(v Validator) { v.FileMaxSize("k", csvFile, 2000, "") },
			make(map[string][]string),
		},

		//Wrong File sizes
		{
			"jpeg 30% of size",
			func(v Validator) { v.FileMaxSize("k", jpegFile, 0.3*calculateFileSize(jpegFile), "") },
			map[string][]string{"k": {fmt.Sprintf("file size cannot be larger than '%.1f' KiloBytes", 0.3*calculateFileSize(jpegFile))}},
		},
		{
			"text 0.4 of size",
			func(v Validator) { v.FileMaxSize("k", pngFile, 0.4*calculateFileSize(pngFile), "") },
			map[string][]string{"k": {fmt.Sprintf("file size cannot be larger than '%.1f' KiloBytes", 0.4*calculateFileSize(pngFile))}},
		},
		{
			"text 0.3 of size",
			func(v Validator) { v.FileMaxSize("k", textFile, 0.3*calculateFileSize(textFile), "Error") },
			map[string][]string{"k": {"Error"}},
		},
		{
			"csv 1/2 of size",
			func(v Validator) { v.FileMaxSize("k", csvFile, 0.5*calculateFileSize(csvFile), "") },
			map[string][]string{"k": {fmt.Sprintf("file size cannot be larger than '%.1f' KiloBytes", 0.5*calculateFileSize(csvFile))}},
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

func TestFileMinSizeValidation(t *testing.T) {
	//Test Images
	jpegFile, pngFile, _ := getTestImages(2000, 2000)

	//Create test Text File
	textFile, err := os.Open(makeOtherFiles("text_1.txt", "text", "New text"))
	if err != nil {
		panic(err.Error())
	}

	//Create test Text File
	csvFile, err := os.Open(makeOtherFiles("testcsv.csv", "csv", ""))
	if err != nil {
		panic(err.Error())
	}

	defer closeAllFiles(jpegFile, pngFile, csvFile, textFile)

	tests := []struct {
		testname   string
		val        func(Validator)
		wantErrors map[string][]string
	}{
		{
			"jpeg ok",
			func(v Validator) { v.FileMinSize("k", jpegFile, calculateFileSize(jpegFile), "") },
			make(map[string][]string),
		},
		{
			"png ok",
			func(v Validator) { v.FileMinSize("k", pngFile, calculateFileSize(pngFile), "") },
			make(map[string][]string),
		},
		{
			"text ok",
			func(v Validator) { v.FileMinSize("k", textFile, calculateFileSize(textFile), "") },
			make(map[string][]string),
		},
		{
			"csv ok",
			func(v Validator) { v.FileMinSize("k", csvFile, calculateFileSize(csvFile), "") },
			make(map[string][]string),
		},
		//Smaller Sizes
		{
			"text 2 of size",
			func(v Validator) { v.FileMinSize("k", textFile, 0.5*calculateFileSize(textFile), "") },
			make(map[string][]string),
		},
		{
			"csv 0.005",
			func(v Validator) { v.FileMinSize("k", csvFile, 0.005, "") },
			make(map[string][]string),
		},

		//Wrong File sizes
		{
			"jpeg 3 of size",
			func(v Validator) { v.FileMinSize("k", jpegFile, 3*calculateFileSize(jpegFile), "") },
			map[string][]string{"k": {fmt.Sprintf("file size cannot be less than '%.1f' KiloBytes", 3*calculateFileSize(jpegFile))}},
		},
		{
			"text 4  of size",
			func(v Validator) { v.FileMinSize("k", pngFile, 4*calculateFileSize(pngFile), "") },
			map[string][]string{"k": {fmt.Sprintf("file size cannot be less than '%.1f' KiloBytes", 4*calculateFileSize(pngFile))}},
		},
		{
			"text 3 of size",
			func(v Validator) { v.FileMinSize("k", textFile, 3*calculateFileSize(textFile), "Error") },
			map[string][]string{"k": {"Error"}},
		},
		{
			"csv 5 of size",
			func(v Validator) { v.FileMinSize("k", csvFile, 5*calculateFileSize(csvFile), "") },
			map[string][]string{"k": {fmt.Sprintf("file size cannot be less than '%.1f' KiloBytes", 5*calculateFileSize(csvFile))}},
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

func TestFileMimeTypeValidation(t *testing.T) {
	//Test Images
	jpegFile, pngFile, _ := getTestImages(2000, 2000)

	//Create test Text File
	csvFile, err := os.Open(makeOtherFiles("testcsv.csv", "csv", ""))
	if err != nil {
		panic(err.Error())
	}

	pdfFile, err := os.Open(makeOtherFiles("test_pdf.pdf", "pdf", "Lorem ipsum dolor sit amet."))

	if err != nil {
		panic(err.Error())
	}

	defer closeAllFiles(jpegFile, pngFile, csvFile, pdfFile)

	tests := []struct {
		testname   string
		val        func(Validator)
		wantErrors map[string][]string
	}{
		{
			"jpeg ok",
			func(v Validator) { v.FileMimeType("k", jpegFile, "image/jpeg", "") },
			make(map[string][]string),
		},
		{
			"png ok",
			func(v Validator) { v.FileMimeType("k", pngFile, "image/png", "") },
			make(map[string][]string),
		},
		{
			"csv ok",
			func(v Validator) { v.FileMimeType("k", csvFile, "text/csv,application/octet-stream", "") },
			make(map[string][]string),
		},
		{
			"pdf ok",
			func(v Validator) { v.FileMimeType("k", pdfFile, "application/pdf", "") },
			make(map[string][]string),
		},

		//Wrong File sizes
		{
			"jpeg, in image/png, want image/jpeg",
			func(v Validator) { v.FileMimeType("k", pngFile, "image/jpeg", "") },
			map[string][]string{"k": {"must be a file of type 'image/jpeg'"}},
		},
		{
			"png, in image/png, want image/jpeg,application/octet-stream",
			func(v Validator) { v.FileMimeType("k", pngFile, "image/jpeg,application/octet-stream", "") },
			map[string][]string{"k": {"must be a file of type 'image/jpeg,application/octet-stream'"}},
		},
		{
			"jpeg, in image/jpeg, want application/pdf",
			func(v Validator) { v.FileMimeType("k", jpegFile, "application/pdf", "") },
			map[string][]string{"k": {"must be a file of type 'application/pdf'"}},
		},
		{
			"pdf, in  image/jpeg, want application/pd",
			func(v Validator) { v.FileMimeType("k", jpegFile, "application/pdf", "Error") },
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

//
//--------------------------------------------------------- HELPER FUNCTIONS ---------------------------
//
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
func makeOtherFiles(name, format, content string) string {
	fullName := testImageDir + name
	if format == "pdf" {
		pdf := gofpdf.New("P", "mm", "A4", "")
		pdf.AddPage()
		pdf.Text(20, 20, content)
		pdf.OutputFileAndClose(fullName)
		return fullName
	}

	file, err := os.Create(fullName)
	defer file.Close()
	if err != nil {
		panic("Error creating file: \n" + err.Error())
	}

	if format == "text" {
		_, err = file.Write([]byte(content))
		if err != nil {
			panic("Error creating file: \n" + err.Error())
		}
	}

	if format == "csv" {
		writer := csv.NewWriter(file)
		writer.WriteAll([][]string{{"one", "two", "three"}, {"1", "2", "3"}, {"4", "5", "6"}})
		writer.Flush()
	}

	return fullName
}

//calculateFileSize
func calculateFileSize(file *os.File) float64 {
	info, err := file.Stat()
	if err != nil {
		panic("Could not determine file size: \n" + err.Error())
	}
	return float64(info.Size()) / 1024
}

//Close all files used for testing
func closeAllFiles(files ...*os.File) {
	for _, file := range files {
		file.Close()
	}
}
