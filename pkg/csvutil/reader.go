// Package csvutil decodes a CSV stream into a slice of typed, validated
// structs. It is a generic engine: callers describe the row shape with a
// struct whose fields carry `csv` tags (matched against the header by name)
// and optional `validate` tags (go-playground/validator), then call Decode.
//
//	type Row struct {
//	    Name string `csv:"name" validate:"required"`
//	    Age  int    `csv:"age"  validate:"gte=0"`
//	}
//	rows, err := csvutil.Decode[Row](file)
//
// The CSV parsing and struct mapping (header matching, BOM stripping, type
// conversion) is delegated to github.com/gocarina/gocsv; this package adds
// per-row validation, empty-input detection, and stable sentinel errors on
// top. It deliberately knows nothing about any business domain so it can be
// reused across services — concerns that are not "parse a CSV" (request size
// limits, batching, queueing) belong to the caller.
package csvutil

import (
	"errors"
	"fmt"
	"io"

	"github.com/go-playground/validator/v10"
	"github.com/gocarina/gocsv"
)

var (
	// ErrEmptyCSV reports that the input had no data rows (it was empty or
	// contained only a header).
	ErrEmptyCSV = errors.New("csv file is empty")

	// ErrInvalidFormat reports that the input could not be decoded into []T: a
	// malformed record, an inconsistent column count, a target type gocsv
	// cannot map, or a row that fails validation. Wrapped errors carry the
	// offending data-row number and detail.
	ErrInvalidFormat = errors.New("invalid csv format")
)

// validate is shared and safe for concurrent use, mirroring requestutils.
var validate = validator.New(validator.WithRequiredStructEnabled())

// Decode reads a CSV from r and returns one T per data row.
//
// The first record is the header; gocsv matches each struct field's `csv` tag
// to a header column by name (a UTF-8 BOM on the first column is stripped).
// A declared column that is absent from the header leaves its field at the
// zero value, so a `validate:"required"` tag is what turns a missing required
// column into an error.
//
// Every row is validated with its `validate` tags and decoding returns at the
// first failure (fail fast), so a caller can reject the whole upload rather
// than silently importing partial data.
func Decode[T any](r io.Reader) ([]T, error) {
	var rows []T
	if err := gocsv.Unmarshal(r, &rows); err != nil {
		if errors.Is(err, gocsv.ErrEmptyCSVFile) {
			return nil, ErrEmptyCSV
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidFormat, err)
	}

	if len(rows) == 0 {
		return nil, ErrEmptyCSV
	}

	for i, row := range rows {
		if err := validate.Struct(row); err != nil {
			// i is the 0-based data-row index; i+1 is its human row number
			// (excluding the header).
			return nil, fmt.Errorf("%w: row %d: %v", ErrInvalidFormat, i+1, err)
		}
	}

	return rows, nil
}
