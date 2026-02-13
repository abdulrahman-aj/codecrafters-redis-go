package resp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
)

type Reader struct {
	reader *bufio.Reader
}

func NewReader(r io.Reader) *Reader {
	return &Reader{reader: bufio.NewReader(r)}
}

func (r *Reader) ReadValue() (any, error) {
	byte, err := r.reader.ReadByte()
	if err != nil {
		return nil, err
	}

	switch byte {
	case '*':
		return r.readArray()
	case '$':
		return r.readBulkString()
	default:
		return nil, ErrBadRequest
	}

}

func (r *Reader) readArray() ([]any, error) {
	arrayLength, err := readLength(r.reader)
	if err != nil {
		return nil, err
	}

	var ret []any
	for range arrayLength {
		val, err := r.ReadValue()
		if err != nil {
			return nil, err
		}
		ret = append(ret, val)
	}

	return ret, nil
}

func (r *Reader) readBulkString() (string, error) {
	strLength, err := readLength(r.reader)
	if err != nil {
		return "", err
	}

	bytes, err := readline(r.reader)
	if err != nil {
		return "", err
	}

	if len(bytes) != strLength {
		return "", fmt.Errorf("%w: length mismatch: header=%d, actual=%d for string %s", ErrBadRequest, strLength, len(bytes), bytes)
	}

	return string(bytes), nil
}

var (
	ErrBadRequest = errors.New("Bad Request")
)

func readline(reader *bufio.Reader) ([]byte, error) {
	var ret []byte
	for {
		bytes, isPrefix, err := reader.ReadLine()
		if err != nil {
			return nil, err
		}
		ret = append(ret, bytes...)
		if !isPrefix {
			break
		}
	}
	return ret, nil
}

func readLength(reader *bufio.Reader) (int, error) {
	lengthStr, err := readline(reader)
	if err != nil {
		return 0, err
	}

	length, err := strconv.Atoi(string(lengthStr))
	if err != nil {
		return 0, fmt.Errorf("expected length=%s to be an integer: %w", lengthStr, err)
	}

	return length, nil
}

func SimpleString(v string) []byte {
	return fmt.Appendf(nil, "+%s\r\n", v)
}

func SimpleError(v string) []byte {
	return fmt.Appendf(nil, "-%s\r\n", v)
}

func BulkString(v string) []byte {
	return fmt.Appendf(nil, "$%d\r\n%s\r\n", len(v), v)
}

func Integer(v int) []byte {
	return fmt.Appendf(nil, ":%d\r\n", v)
}

var NullBulkString = []byte("$-1\r\n")
