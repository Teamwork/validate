package validate

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
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
	jpeg1000x2000 = "test_jpeg_1000_1000.jpg"
)

//Test and Confirm Images Formats
func TestImageFormatValidation(t *testing.T) {
	//Test Images
	jpegFile, pngFile, gifFile := getTestImages(2000, 2000)

	textFile := prepareFileHeader(makeOtherFiles("text_1.txt", "text/plain", "New text"))

	tests := []struct {
		testname   string
		val        func(Validator)
		wantErrors map[string][]string
	}{
		{
			"jpeg ok",
			func(v Validator) { v.IsImage("k", jpegFile, "JPEG", "") },
			make(map[string][]string),
		},
		{
			"png ok",
			func(v Validator) { v.IsImage("k", pngFile, "PNG", "") },
			make(map[string][]string),
		},
		{
			"gif ok",
			func(v Validator) { v.IsImage("k", gifFile, "Gif", "") },
			make(map[string][]string),
		},

		//Wrong Image Format
		{
			"jpeg in, png wanted",
			func(v Validator) { v.IsImage("k", jpegFile, "PNG", "") },
			map[string][]string{"k": {"must be an image of 'PNG' format"}},
		},
		{
			"png in, jpeg wanted",
			func(v Validator) { v.IsImage("k", pngFile, "JPEG", "") },
			map[string][]string{"k": {"must be an image of 'JPEG' format"}},
		},
		{
			"gif in, png wanted",
			func(v Validator) { v.IsImage("k", gifFile, "PNG", "") },
			map[string][]string{"k": {"must be an image of 'PNG' format"}},
		},

		{
			"textfile in, png wanted",
			func(v Validator) { v.IsImage("k", textFile, "PNG", "") },
			map[string][]string{"k": {"must be an image of 'PNG' format"}},
		},
		{
			"textfile in, png wanted, custom error",
			func(v Validator) { v.IsImage("k", textFile, "PNG", "Error") },
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

func TestImageMaxDimensionValidation(t *testing.T) {
	//Test Images
	jpegFile, pngFile, gifFile := getTestImages(2000, 2000)

	jpegFile1000x2000 := prepareFileHeader(makeTestImage("image/jpeg", jpeg1000x2000, 1000, 2000))

	textFile := prepareFileHeader(makeOtherFiles("text_1.txt", "text/plain", "New text"))

	tests := []struct {
		testname   string
		val        func(Validator)
		wantErrors map[string][]string
	}{
		{
			"jpeg ok",
			func(v Validator) {
				v.ImageDimensions("k", jpegFile, &ImageDimension{2000, 2000}, nil, "")
			},
			make(map[string][]string),
		},
		{
			"jpeg 1000x2000 ok",
			func(v Validator) {
				v.ImageDimensions("k", jpegFile1000x2000, nil, &ImageDimension{1000, 2000}, "")
			},
			make(map[string][]string),
		},
		{
			"png ok",
			func(v Validator) {
				v.ImageDimensions("k", pngFile, &ImageDimension{2000, 2000}, &ImageDimension{2000, 2000}, "")
			},
			make(map[string][]string),
		},
		{
			"gif ok",
			func(v Validator) {
				v.ImageDimensions("k", gifFile, nil, &ImageDimension{2000, 2000}, "")
			},
			make(map[string][]string),
		},
		//Wrong Image dimension
		{
			"jpeg 2000x2000 in, 5000x5000 wanted",
			func(v Validator) {
				v.ImageDimensions("k", jpegFile, &ImageDimension{5000, 5000}, nil, "")
			},
			map[string][]string{"k": {"image dimension (W x H) cannot be less than '5000 x 5000' pixels"}},
		},
		{
			"png 2000x2000 in, 3000x500 wanted",
			func(v Validator) {
				v.ImageDimensions("k", pngFile, &ImageDimension{3000, 500}, &ImageDimension{3000, 1000}, "")
			},
			map[string][]string{"k": {"image dimension (W x H) must be between '3000 x 500' and '3000 x 1000' pixels"}},
		},
		{
			"gif 2000x2000 in, 1000x1000 max wanted",
			func(v Validator) {
				v.ImageDimensions("k", gifFile, nil, &ImageDimension{1000, 1000}, "")
			},
			map[string][]string{"k": {"image dimension (W x H) cannot be more than '1000 x 1000' pixels"}},
		},

		{
			"jpeg 1000x2000 in, with custome error",
			func(v Validator) {
				v.ImageDimensions("k", jpegFile1000x2000, &ImageDimension{3000, 2000}, nil, "Error")
			},
			map[string][]string{"k": {"Error"}},
		},

		{
			"textfile in, png 1000x1000 wanted",
			func(v Validator) {
				v.ImageDimensions("k", textFile, nil, &ImageDimension{1000, 1000}, "")
			},
			map[string][]string{"k": {"File is not an image. Only dimensions of image files can be determined."}},
		},
		{
			"textfile in, image wanted, custom error",
			func(v Validator) {
				v.ImageDimensions("k", textFile, nil, &ImageDimension{1000, 1000}, "Error")
			},
			map[string][]string{"k": {"File is not an image. Only dimensions of image files can be determined."}},
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

func TestFileSizeValidation(t *testing.T) {
	//Test Images
	jpegFile, pngFile, gifFile := getTestImages(2000, 2000)

	//Create test Text File
	textFile := prepareFileHeader(makeOtherFiles("text_1.txt", "text/plain", "New text"))

	tests := []struct {
		testname   string
		val        func(Validator)
		wantErrors map[string][]string
	}{
		{
			"jpeg ok",
			func(v Validator) { v.FileSize("k", jpegFile, jpegFile.Size, -1, "") },
			make(map[string][]string),
		},
		{
			"png ok",
			func(v Validator) { v.FileSize("k", pngFile, 0, -1, "") },
			make(map[string][]string),
		},
		{
			"gif ok",
			func(v Validator) { v.FileSize("k", gifFile, 0, 100000000, "") },
			make(map[string][]string),
		},
		{
			"text ok",
			func(v Validator) { v.FileSize("k", textFile, 2, 100000000, "") },
			make(map[string][]string),
		},
		{
			"text no min&max sizes",
			func(v Validator) { v.FileSize("k", textFile, -1, -1, "") },
			make(map[string][]string),
		},

		//Wrong File sizes
		{
			"jpeg needs twice the size",
			func(v Validator) { v.FileSize("k", jpegFile, 2*jpegFile.Size, -1, "") },
			map[string][]string{"k": {fmt.Sprintf("file size cannot be less than '%.1f'KB",
				bytesToKiloBytes(2*jpegFile.Size))}},
		},
		{
			"png 1000 bytes max",
			func(v Validator) { v.FileSize("k", pngFile, 100, 1000, "") },
			map[string][]string{"k": {fmt.Sprintf("file size cannot be larger than '%.1f'KB", bytesToKiloBytes(1000))}},
		},
		{
			"text 10 bytes max, custom error",
			func(v Validator) { v.FileSize("k", textFile, -1, 10, "Error") },
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

func TestFileMimeTypeValidation(t *testing.T) {
	//Test Images
	jpegFile, pngFile, gifFile := getTestImages(2000, 2000)

	//Create test Text File
	textFile := prepareFileHeader(makeOtherFiles("text_1.txt", "text/plain", "New text"))

	pdfFile := prepareFileHeader(makeOtherFiles("test_pdf.pdf", "application/pdf",
		" Lorem ipsum dolor sit amet, consectetur adipiscing elit."))

	tests := []struct {
		testname   string
		val        func(Validator)
		wantErrors map[string][]string
	}{
		{
			"jpeg ok",
			func(v Validator) { v.FileMimeType("k", jpegFile, "image/jpeg, image/png", "") },
			make(map[string][]string),
		},
		{
			"png ok",
			func(v Validator) { v.FileMimeType("k", pngFile, "image/png", "") },
			make(map[string][]string),
		},
		{
			"gif ok",
			func(v Validator) { v.FileMimeType("k", gifFile, "image/gif, image/png", "") },
			make(map[string][]string),
		},
		{
			"text ok",
			func(v Validator) { v.FileMimeType("k", textFile, "text/plain", "") },
			make(map[string][]string),
		},
		{
			"pdf ok",
			func(v Validator) { v.FileMimeType("k", pdfFile, "application/pdf", "") },
			make(map[string][]string),
		},

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

//Empty File struct to implement file interface for multipart
type emptyFile struct{}

//Mock empty reader for multipart file
func (f *emptyFile) Read(p []byte) (n int, err error) {
	return 0, nil
}

//Mock empty seek for multipart file
func (f *emptyFile) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
} //Mock empty close for multipart file
func (f *emptyFile) Close() error {
	return nil
} //Mock empty readAt for multipart file
func (f *emptyFile) ReadAt(p []byte, off int64) (n int, err error) {
	return 0, nil
}

func TestFileRequired(t *testing.T) {
	textFile := prepareFileHeader(makeOtherFiles("text_1.txt", "text/plain", "New text"))

	file, err := textFile.Open()
	defer file.Close()

	if err != nil {
		panic(err)
	}

	defer closeFiles(file.Close())

	tests := []struct {
		testname   string
		val        func(Validator)
		wantErrors map[string][]string
	}{
		{
			"text ok",
			func(v Validator) { v.Required("k", textFile, "") },
			make(map[string][]string),
		},
		{
			"File ok",
			func(v Validator) { v.Required("k", file) },
			make(map[string][]string),
		},
		{
			"Data required",
			func(v Validator) { v.Required("k", &multipart.FileHeader{}) },
			map[string][]string{"k": {"must be set"}},
		},

		{
			"Data required, custom error",
			func(v Validator) { v.Required("k", &multipart.FileHeader{}, "Error") },
			map[string][]string{"k": {"Error"}},
		},
		{
			"File empty",
			func(v Validator) { v.Required("k", &emptyFile{}) },
			map[string][]string{"k": {"must be set"}},
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
func getTestImages(w, h int) (*multipart.FileHeader, *multipart.FileHeader, *multipart.FileHeader) {
	// jpegFile, err := os.Open(makeTestImage("JPEG", jpeg2000x2000, w, h))

	jpegFile := prepareFileHeader(makeTestImage("image/jpeg", jpeg2000x2000, w, h))

	//Create PNG
	pngFile := prepareFileHeader(makeTestImage("image/png", png2000x2000, w, h))

	//Create GIF
	gifFile := prepareFileHeader(makeTestImage("image/gif", gif2000x2000, w, h))

	return jpegFile, pngFile, gifFile
}

//Prepare multipart header from File
//This creates file request and returns multipart Header for testing
func prepareFileHeader(req *http.Request) *multipart.FileHeader {

	err := req.ParseMultipartForm(10 << 20)
	if err != nil {
		panic("Cannot parse request object: " + err.Error())
	}

	_, header, err := req.FormFile("test_file")
	if err != nil {
		panic("Erro retrieving file: " + err.Error())
	}
	return header
}

//Make For a Test
func makeTestImage(format, name string, w, h int) *http.Request {

	newImage := image.NewRGBA(image.Rect(0, 0, w, h))
	fullName := testImageDir + name

	file, err := os.Create(fullName)
	if err != nil {
		panic("Error creating image: \n" + err.Error())
	}
	defer closeFiles(file.Close())

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

	return convertToRequest(fullName, format, file)
}

//Create other files types for testing
func makeOtherFiles(name, format, content string) *http.Request {
	fullName := testImageDir + name

	//Process PDF
	if format == "pdf" {
		pdf := gofpdf.New("P", "mm", "A4", "")
		pdf.AddPage()
		pdf.Text(20, 20, content)
		pdf.OutputFileAndClose(fullName)

		file, err := os.Open(fullName)
		if err != nil {
			panic("Error creating file: \n" + err.Error())
		}
		defer closeFiles(file.Close())

		return convertToRequest(fullName, format, file)
	}

	//Create test file on  file on Disk
	file, err := os.Create(fullName)
	if err != nil {
		panic("Error creating file: \n" + err.Error())
	}

	defer closeFiles(file.Close())

	_, err = file.Write([]byte(content))
	if err != nil {
		panic("Error creating file: \n" + err.Error())
	}

	return convertToRequest(fullName, format, file)
}

func convertToRequest(name, format string, file *os.File) *http.Request {
	_, err := file.Seek(0, 0)
	if err != nil {
		panic(err)
	}
	//Convert to Request
	var buff bytes.Buffer

	mw := multipart.NewWriter(&buff)
	//header
	hd := make(textproto.MIMEHeader)
	hd.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "test_file", name))
	hd.Set("Content-Type", format)

	formFile, err := mw.CreatePart(hd)

	if err != nil {
		panic("Error creating form file: " + err.Error())
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		panic(err)
	}
	//Copy Files to form file
	if _, err = io.Copy(formFile, file); err != nil {
		panic("Error copying form file: " + err.Error())
	}
	//Set Request Data
	req, err := http.NewRequest("POST", "localhost", &buff)
	if err != nil {
		panic("Error creating request object: " + err.Error())
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", mw.FormDataContentType())

	closeFiles(mw.Close())

	return req
}

//Close Files
func closeFiles(err error) {
	if err != nil {
		panic("Error closing file:" + err.Error())
	}
}
