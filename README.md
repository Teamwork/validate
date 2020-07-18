[![Build Status](https://travis-ci.com/Teamwork/validate.svg?branch=master)](https://travis-ci.com/Teamwork/validate)
[![codecov](https://codecov.io/gh/Teamwork/validate/branch/master/graph/badge.svg?token=n0k8YjbQOL)](https://codecov.io/gh/Teamwork/validate)
[![GoDoc](https://godoc.org/github.com/Teamwork/validate?status.svg)](https://godoc.org/github.com/Teamwork/validate)

HTTP request parameter validation for Go.

Basic usage example:

	v := validate.New()
	v.Required("firstName", customer.FirstName)
	if v.HasErrors() {
		fmt.Println("Had the following validation errors:")
		for key, errors := range v.Errors {
			fmt.Printf("    %s: %s", key, strings.Join(errors))
		}
	}

See godoc for more info.

<!-- import "github.com/teamwork/validate" -->

Basic usage example: Image Validation with different formats

	func uploadFile(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(10 << 20)
		file, fileHeader, err := r.FormFile("myFile")
		if err != nil {
			fmt.Println("Error Retrieving the File")
			fmt.Println(err)
			return
		}
		defer file.Close()
		validator := validate.New()
		validator.IsImage("myFile", fileHeader, "") OR
		validator.IsImage("myFile", fileHeader, "jpeg") OR
		validator.IsImage("myFile", fileHeader, "jpeg, png")
		if validator.HasErrors() {
			fmt.Fprintf(w, validator.String())
		}
	}

For Image Dimensions (Width and Heights)	
	
	func uploadFile(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(10 << 20)
		file, fileHeader, err := r.FormFile("myFile")
		if err != nil {
			fmt.Println("Error Retrieving the File")
			fmt.Println(err)
			return
		}
		defer file.Close()
		validator := validate.New()
		minDim := validate.ImageDimension{Width: 200, Height: 300} //minimum dimension
		maxDim := validate.ImageDimension{Width: 300, Height: 350} //maximum dimension
		validator.ImageDimensions("myFile", fileHeader, &minDim, &maxDim, "")
		if validator.HasErrors() {
			fmt.Fprintf(w, validator.String())
		}
	}


You can also validate different file mime types.
For more information on mime types, visit:

https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types
https://www.iana.org/assignments/media-types/media-types.xhtml

File Mime type Validation:

	func uploadFile(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(10 << 20)
		file, fileHeader, err := r.FormFile("myFile")
		if err != nil {
			fmt.Println("Error Retrieving the File")
			fmt.Println(err)
			return
		}
		defer file.Close()
		validator := validate.New()
		validator.FileMimeType("myFile", fileHeader, "application/pdf", "")
		if validator.HasErrors() {
			fmt.Fprintf(w, validator.String())
		}	
	}

For File Sizes in bytes:

	func uploadFile(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(10 << 20)
		file, fileHeader, err := r.FormFile("myFile")
		if err != nil {
			fmt.Println("Error Retrieving the File")
			fmt.Println(err)
			return
		}
		defer file.Close()
		validator := validate.New()
		validator.FileSize("myFile", fileHeader, 100000, 200000, "")
		if validator.HasErrors() {
			fmt.Fprintf(w, validator.String())
		}
	}	