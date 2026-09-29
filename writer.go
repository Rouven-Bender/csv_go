package csv_go

import (
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"unsafe"
)

type Writer[T any] struct {
	csvwriter csv.Writer
}

func (w *Writer[T]) SetComma(r rune) {
	w.csvwriter.Comma = r
}
func (w *Writer[T]) UseCRLF(b bool) {
	w.csvwriter.UseCRLF = b
}
func (w *Writer[T]) Flush() {
	w.csvwriter.Flush()
}
func (w *Writer[T]) Close() error {
	w.csvwriter.Flush()
	return w.csvwriter.Error()
}
func (w *Writer[T]) Error() error {
	return w.csvwriter.Error()
}

func NewWriter[T any](w io.Writer) *Writer[T] {
	return &Writer[T]{
		csvwriter: *csv.NewWriter(w),
	}
}

func (w *Writer[T]) Write(record *T) error {
	rt := reflect.TypeOf(*record)

	strs := make([]string, furthestColumn[T]())

	for i := range rt.NumField() {
		field := rt.Field(i)
		if !field.IsExported() {
			continue
		}

		csvTag := field.Tag.Get("csv")
		if csvTag == "" {
			continue
		}

		idx := columnNameToNumber(csvTag)
		if !(idx < len(strs)) {
			return fmt.Errorf("field out of range")
		}
		uptr := (unsafe.Pointer(uintptr(unsafe.Pointer(record))+field.Offset))
		switch field.Type {
		case reflect.TypeFor[int]():
			ptr := (*int)(uptr)
			strs[idx] = strconv.FormatInt(int64(*ptr), 10)
		case reflect.TypeFor[string]():
			ptr := (*string)(uptr)
			strs[idx] = *ptr
		case reflect.TypeFor[float64]():
			ptr := (*float64)(uptr)
			strs[idx] = strconv.FormatFloat(*ptr, 'f', 2, 64)
		case reflect.TypeFor[float32]():
			ptr := (*float32)(uptr)
			f := float64(*ptr)
			strs[idx] = strconv.FormatFloat(f, 'f', 2, 32)
		default:
			return fmt.Errorf("can't marshal field of type: %v", field.Type)
		}
	}
	err := w.csvwriter.Write(strs)
	if err != nil {
		return err
	}

	return nil
}

func furthestColumn[T any]() int {
	biggestLength := 0
	rt := reflect.TypeFor[T]()
	for i := range rt.NumField() {
		field := rt.Field(i)
		if !field.IsExported() {
			continue
		}

		csvTag := field.Tag.Get("csv")
		if csvTag == "" {
			continue
		}
		
		lengthNeeded := columnNameToNumber(csvTag)+1
		
		if lengthNeeded > biggestLength {
			biggestLength = lengthNeeded
		}
	}
	return biggestLength
}
