package arukas

import (
	"fmt"
	"strings"

	"reflect"
	"time"

	"github.com/google/uuid"
)

// validateID validates id formats. id needs UUID format.
func validateID(label, value string) error {
	if value == "" {
		return nil
	}
	_, err := uuid.Parse(value)
	if err != nil {
		return malformatUUIDError(label)
	}
	return nil
}

// malformatUUIDError returns an error indicating that the ID format is illegal as a UUID
func malformatUUIDError(label string) error {
	return fmt.Errorf("%q is malformated as UUID", label)
}

// validateRequired validates value is not empty
func validateRequired(label string, value interface{}) error {
	if isEmpty(value) {
		return requiredError(label)
	}
	return nil
}

// requiredError returns an error indicating that value is required
func requiredError(label string) error {
	return fmt.Errorf("%q is required", label)
}

// validateRange validates value is in range of min and max
func validateRange(label string, value, min, max int) error {
	if !(min <= value && value <= max) {
		return outOfRangeError(label, min, max)
	}
	return nil
}

// outOfRangeError returns an error indicating that value is in out of range
func outOfRangeError(label string, min, max int) error {
	return fmt.Errorf("%q must be between %d and %d", label, min, max)
}

// valiateStrByteLen validates length of value's bytes is in range of min and max
func valiateStrByteLen(label, value string, min, max int) error {
	strLen := len(value)
	if !(min <= strLen && strLen <= max) {
		return strByteLenError(label, min, max)
	}
	return nil
}

// strByteLenError returns an error indicating that length of value's bytes is in out of range
func strByteLenError(label string, min, max int) error {
	return fmt.Errorf("%q must be between %d and %d bytes", label, min, max)
}

// validateInStrValues validates value is exists in specified values
func validateInStrValues(label, value string, values ...string) error {
	for _, v := range values {
		if value == v {
			return nil
		}
	}
	return inStrValuesError(label, values...)
}

// inStrValuesError return an error indicating that value is exists in specified values
func inStrValuesError(label string, values ...string) error {
	return fmt.Errorf("%q mest be in [%s]", label, strings.Join(values, "/"))
}

var numericZeros = []interface{}{
	int(0),
	int8(0),
	int16(0),
	int32(0),
	int64(0),
	uint(0),
	uint8(0),
	uint16(0),
	uint32(0),
	uint64(0),
	float32(0),
	float64(0),
}

// isEmpty is copied from github.com/stretchr/testify/assert/assetions.go
func isEmpty(object interface{}) bool {

	if object == nil {
		return true
	} else if object == "" {
		return true
	} else if object == false {
		return true
	}

	for _, v := range numericZeros {
		if object == v {
			return true
		}
	}

	objValue := reflect.ValueOf(object)

	switch objValue.Kind() {
	case reflect.Map:
		fallthrough
	case reflect.Slice, reflect.Chan:
		{
			return (objValue.Len() == 0)
		}
	case reflect.Struct:
		switch object.(type) {
		case time.Time:
			return object.(time.Time).IsZero()
		}
	case reflect.Ptr:
		{
			if objValue.IsNil() {
				return true
			}
			switch object.(type) {
			case *time.Time:
				return object.(*time.Time).IsZero()
			default:
				return false
			}
		}
	}
	return false
}
