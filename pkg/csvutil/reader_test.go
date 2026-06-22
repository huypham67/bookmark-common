package csvutil

import (
	"errors"
	"strings"
	"testing"
)

type person struct {
	Name  string `csv:"name" validate:"required,min=2"`
	Email string `csv:"email" validate:"required,email"`
}

func TestDecodeValid(t *testing.T) {
	in := "name,email\nAlice,alice@example.com\nBob,bob@example.com\n"

	rows, err := Decode[person](strings.NewReader(in))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2", len(rows))
	}
	if rows[0] != (person{Name: "Alice", Email: "alice@example.com"}) {
		t.Errorf("row 0 = %+v", rows[0])
	}
	if rows[1].Name != "Bob" {
		t.Errorf("row 1 name = %q, want Bob", rows[1].Name)
	}
}

func TestDecodeMapsByHeaderNameNotPosition(t *testing.T) {
	// Columns are in the opposite order from the struct declaration.
	in := "email,name\nalice@example.com,Alice\n"

	rows, err := Decode[person](strings.NewReader(in))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rows[0].Name != "Alice" || rows[0].Email != "alice@example.com" {
		t.Errorf("row 0 = %+v, want columns matched by name", rows[0])
	}
}

func TestDecodeIgnoresUntaggedAndExtraColumns(t *testing.T) {
	in := "name,email,note\nAlice,alice@example.com,hello\n"

	rows, err := Decode[person](strings.NewReader(in))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rows[0].Name != "Alice" {
		t.Errorf("unexpected row: %+v", rows[0])
	}
}

func TestDecodeStripsBOMFromHeader(t *testing.T) {
	in := "\uFEFFname,email\nAlice,alice@example.com\n"

	rows, err := Decode[person](strings.NewReader(in))
	if err != nil {
		t.Fatalf("unexpected error with BOM header: %v", err)
	}
	if rows[0].Name != "Alice" {
		t.Errorf("BOM not stripped: %+v", rows[0])
	}
}

func TestDecodeHandlesCRLFAndQuotedComma(t *testing.T) {
	in := "name,email\r\n\"Doe, John\",john@example.com\r\n"

	rows, err := Decode[person](strings.NewReader(in))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rows[0].Name != "Doe, John" {
		t.Errorf("quoted comma not preserved: %q", rows[0].Name)
	}
}

func TestDecodeEmptyInput(t *testing.T) {
	_, err := Decode[person](strings.NewReader(""))
	if !errors.Is(err, ErrEmptyCSV) {
		t.Errorf("got %v, want ErrEmptyCSV", err)
	}
}

func TestDecodeHeaderOnly(t *testing.T) {
	_, err := Decode[person](strings.NewReader("name,email\n"))
	if !errors.Is(err, ErrEmptyCSV) {
		t.Errorf("got %v, want ErrEmptyCSV", err)
	}
}

func TestDecodeMissingRequiredColumn(t *testing.T) {
	// The "email" column is absent; gocsv leaves Email zero and the required
	// validation tag turns that into an error.
	in := "name\nAlice\n"

	_, err := Decode[person](strings.NewReader(in))
	if !errors.Is(err, ErrInvalidFormat) {
		t.Fatalf("got %v, want ErrInvalidFormat", err)
	}
	if !strings.Contains(err.Error(), "Email") {
		t.Errorf("error should name the failing field: %v", err)
	}
}

func TestDecodeInconsistentColumnCount(t *testing.T) {
	// Second data row has an extra field relative to the header.
	in := "name,email\nAlice,alice@example.com\nBob,bob@example.com,extra\n"

	_, err := Decode[person](strings.NewReader(in))
	if !errors.Is(err, ErrInvalidFormat) {
		t.Fatalf("got %v, want ErrInvalidFormat", err)
	}
}

func TestDecodeValidationFailureReportsRow(t *testing.T) {
	// The 2nd data row has an invalid email.
	in := "name,email\nAlice,alice@example.com\nBob,not-an-email\n"

	_, err := Decode[person](strings.NewReader(in))
	if !errors.Is(err, ErrInvalidFormat) {
		t.Fatalf("got %v, want ErrInvalidFormat", err)
	}
	if !strings.Contains(err.Error(), "row 2") {
		t.Errorf("error should report the offending row: %v", err)
	}
}

func TestDecodeNonStructType(t *testing.T) {
	_, err := Decode[string](strings.NewReader("a\nb\n"))
	if !errors.Is(err, ErrInvalidFormat) {
		t.Errorf("got %v, want ErrInvalidFormat for non-struct target", err)
	}
}

type measurement struct {
	Label string  `csv:"label" validate:"required"`
	Count int     `csv:"count"`
	Ratio float64 `csv:"ratio"`
	Ok    bool    `csv:"ok"`
}

func TestDecodeNumericAndBoolFields(t *testing.T) {
	in := "label,count,ratio,ok\nwidget,42,3.14,true\n"

	rows, err := Decode[measurement](strings.NewReader(in))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := rows[0]
	if got.Label != "widget" || got.Count != 42 || got.Ratio != 3.14 || !got.Ok {
		t.Errorf("parsed row = %+v", got)
	}
}

func TestDecodeUnparseableNumber(t *testing.T) {
	in := "label,count,ratio,ok\nwidget,notanumber,1.0,true\n"

	_, err := Decode[measurement](strings.NewReader(in))
	if !errors.Is(err, ErrInvalidFormat) {
		t.Fatalf("got %v, want ErrInvalidFormat", err)
	}
}
