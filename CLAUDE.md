# CLAUDE.md

## Project Overview

A Go desktop app (Fyne GUI) that converts shopping mall order Excel files to 3PL (Third-Party Logistics) shipping instruction Excel files. Provides two main features: transformation and validation.

## Tech Stack

- Language: Go 1.24+
- GUI: Fyne v2 (`fyne.io/fyne/v2`)
- Excel read/write: excelize (`github.com/xuri/excelize/v2`)
- Requires CGO (Fyne dependency)

## Project Structure

- `main.go` - Fyne GUI with two tabs (transform, validate), custom theme, file open helper
- `order.go` - `SourceOrder` and `ShippingOrder` types, Excel reading functions
- `transform.go` - Order transformation logic (source → shipping, excluding cancelled)
- `validate.go` - Validation logic comparing source and target data, formatted output
- `excel.go` - Excel writing for shipping orders, duplicate filename handling
- `transform_test.go` - Tests for reading, transformation, and Excel round-trip
- `validate_test.go` - Tests for validation logic and output formatting
- `testdata/source-order-example.xlsx` - Test fixture with sample order data
- `form/` - Excel form templates (source and target formats)
- `Makefile` - Linux/Windows cross-compilation

## Build & Run

```bash
go run .           # Run from source
make               # Build for Linux + Windows (output in dist/)
make linux         # Linux only
make windows       # Windows only (requires x86_64-w64-mingw32-gcc)
```

## Test

```bash
go test ./...
```

## Coding Conventions

- User-facing messages (UI, errors) are written in Korean
- Code comments follow English Go doc style
- Source Excel column mapping is defined as constants in `order.go`
- Target sheet name matches the 3PL format: "이지어드민 양식"
