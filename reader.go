package csv_go

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"iter"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

type Reader[T any] struct {
	reader csv.Reader
	germanDecimal bool
}

func NewReader[T any](r io.Reader) *Reader[T] {
	rr := &Reader[T]{
		reader: *csv.NewReader(r),
	}
	rr.reader.ReuseRecord = true
	return rr
}

func (re *Reader[T]) SetComma(r rune) {
	re.reader.Comma = r
}

// make float parsing work with 2,50
func (re *Reader[T]) SetGermanDecimal(b bool) {
	re.germanDecimal = b
}

func (re *Reader[T]) Iter() iter.Seq[*T] {
	return func(yield func(*T) bool) {
		for {
			record, err := re.Read()
			if err != nil && errors.Is(err, io.EOF) {
				break
			}
			if !yield(record) {
				return
			}
		}
	}
}

func (re *Reader[T]) SkipLine() error {
	_, err := re.reader.Read()
	return err
}

func (re *Reader[T]) Read() (record *T, err error) {
	record = new(T)
	csvRecord, err := re.reader.Read()
	if err != nil {
		return nil, err
	}
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

		idx := columnNameToNumber(csvTag)
		if !(idx < len(csvRecord)) {
			return nil, fmt.Errorf("selected field out of range")
		}
		uptr := unsafe.Add(unsafe.Pointer(record), int(field.Offset))
		switch field.Type {
		case reflect.TypeFor[string]():
			ptr := (*string)(uptr)
			*ptr = csvRecord[idx]
		case reflect.TypeFor[float64]():
			ptr := (*float64)(uptr)
			if re.germanDecimal {
				f, err := germanNumberStringToFloat(csvRecord[idx], 64)
				if err != nil {
					return nil, err
				}
				*ptr = f
			} else {
				f, err := strconv.ParseFloat(csvRecord[idx], 64)
				if err != nil {
					return nil, err
				}
				*ptr = f
			}
		case reflect.TypeFor[float32]():
			ptr := (*float32)(uptr)
			if re.germanDecimal {
				f, err := germanNumberStringToFloat(csvRecord[idx], 32)
				if err != nil {
					return nil, err
				}
				*ptr = float32(f)
			} else {
				f, err := strconv.ParseFloat(csvRecord[idx], 32)
				if err != nil {
					return nil, err
				}
				*ptr = float32(f)
			}
		case reflect.TypeFor[int]():
			ptr := (*int)(uptr)
			*ptr, err = strconv.Atoi(csvRecord[idx])
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("can't unmarshal type: %v\n", field.Type)
		}
	}
	return record, nil
}

func columnNameToNumber(name string) int {
	name = strings.ToUpper(name) // this is so fieldname a and A go to the same place
	sum := 0
	rs := []rune(name)
	for _, r := range rs {
		sum = sum*26 + int(r) - 64
	}
	return sum - 1
}

func germanNumberStringToFloat(in string, bitsize int) (float64, error) {
	v := strings.ReplaceAll(in, ".", "")
	v = strings.ReplaceAll(v, ",", ".")
	b, err := strconv.ParseFloat(v, bitsize)
	if err != nil {
		return -1, fmt.Errorf("GermanNumberString: %w", err)
	}
	return b, nil
}
