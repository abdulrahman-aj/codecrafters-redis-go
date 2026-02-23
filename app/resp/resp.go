package resp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store/lists"
	"github.com/codecrafters-io/redis-starter-go/app/util"
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
	case '+':
		return r.readSimpleString()
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

func (r *Reader) readSimpleString() (string, error) {
	bytes, err := readline(r.reader)
	if err != nil {
		return "", err
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

func Array(arr any) []byte {
	var ret []byte

	switch arr := arr.(type) {
	case lists.List:
		ret = fmt.Appendf(nil, "*%d\r\n", len(arr))
		for _, x := range arr {
			ret = append(ret, BulkString(x)...)
		}
	case []any:
		ret = fmt.Appendf(nil, "*%d\r\n", len(arr))
		for _, x := range arr {
			switch x := x.(type) {
			case string:
				ret = append(ret, BulkString(x)...)
			case []any:
				ret = append(ret, Array(x)...)
			case []byte: // already serialized
				ret = append(ret, x...)
			default:
				util.Fatal("resp.Array not implemented branch")
			}
		}
	case nil:
		ret = fmt.Appendf(nil, "*%d\r\n", 0)
	default:
		util.Fatal("resp.Array not implemented branch")
	}

	return ret
}

var (
	NullBulkString = []byte("$-1\r\n")
	NullArray      = []byte("*-1\r\n")
)
