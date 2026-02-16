package streams

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type Stream []Entry

type Entry struct {
	Ms     int // Milliseconds
	Seq    int // Sequence Number
	Fields []Field
}

func (e *Entry) ID() string {
	return fmt.Sprintf("%d-%d", e.Ms, e.Seq)
}

type Field struct {
	Key   string
	Value string
}

// GenerateID resolves a stream entry ID to (milliseconds, sequence).
// Valid formats: "*" | "ms-*" | "ms-seq"
// Validates that the ID exceeds the last entry (monotonic).
// Returns RESP error for invalid format or non-monotonic ID.
func (s Stream) GenerateID(entryID string) (int, int, []byte) {
	switch {
	case entryID == "*":
		ms, seq := s.generateFullID()
		return ms, seq, nil
	case strings.HasSuffix(entryID, "-*"):
		return s.generatePartialID(entryID)
	default:
		return s.validateID(entryID)
	}
}

func (s Stream) generateFullID() (int, int) {
	panic("not implemented")
}

func (s Stream) generatePartialID(entryID string) (int, int, []byte) {
	var ms int

	_, err := fmt.Sscanf(entryID, "%d-*", &ms)
	if err != nil || ms < 0 {
		return 0, 0, errInvalidStreamID
	}

	if len(s) == 0 {
		if ms == 0 {
			return 0, 1, nil
		}
		return ms, 0, nil
	}

	last := s[len(s)-1]

	switch {
	case last.Ms > ms:
		return 0, 0, errXaddEqualOrSmaller
	case last.Ms == ms:
		return ms, last.Seq + 1, nil
	default: // last.Ms < ms:
		return ms, 0, nil
	}
}

func (s Stream) validateID(entryID string) (int, int, []byte) {
	var ms, seq int
	if _, err := fmt.Sscanf(entryID, "%d-%d", &ms, &seq); err != nil || ms < 0 || seq < 0 {
		return 0, 0, errInvalidStreamID
	}

	if ms == 0 && seq == 0 {
		return 0, 0, errXaddZeroID
	}

	if len(s) == 0 {
		return ms, seq, nil
	}

	if last := s[len(s)-1]; ms < last.Ms || ms == last.Ms && seq <= last.Seq {
		return 0, 0, errXaddEqualOrSmaller
	}

	return ms, seq, nil
}

var (
	errInvalidStreamID    = resp.SimpleError("ERR Invalid stream ID specified as stream command argument")
	errXaddEqualOrSmaller = resp.SimpleError("ERR The ID specified in XADD is equal or smaller than the target stream top item")
	errXaddZeroID         = resp.SimpleError("ERR The ID specified in XADD must be greater than 0-0")
)
