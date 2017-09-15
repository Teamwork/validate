[![Build Status](https://travis-ci.org/Teamwork/validate.svg?branch=master)](https://travis-ci.org/Teamwork/validate)
[![codecov](https://codecov.io/gh/Teamwork/validate/branch/master/graph/badge.svg?token=n0k8YjbQOL)](https://codecov.io/gh/Teamwork/validate)
[![GoDoc](https://godoc.org/github.com/Teamwork/validate?status.svg)](https://godoc.org/github.com/Teamwork/validate)
[![Go Report Card](https://goreportcard.com/badge/github.com/Teamwork/validate)](https://goreportcard.com/report/github.com/Teamwork/validate)

Package validate implements basic validation for HTTP request parameters.

Basic usage example:

	v := validate.New()
	v.Required("firstName", customer.FirstName)
	if v.HasErrors() {
		fmt.Println("Had the following validation errors:")
		for range key, errors := v.Errors {
			fmt.Printf("    %v: %v", key, strings.Join(errors))
		}
	}

See godoc for more info.
