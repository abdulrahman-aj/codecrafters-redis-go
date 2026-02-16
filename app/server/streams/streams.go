package streams

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type Stream struct {
	entries []Entry
}

type Entry struct {
	Ms        int // Milliseconds
	Seq       int // Sequence Number
	Fields    []Field
	CreatedAt time.Time
}

func (e *Entry) Format() []any {
	kvs := []any{}

	for _, field := range e.Fields {
		kvs = append(kvs, field.Key)
		kvs = append(kvs, field.Value)
	}

	return []any{e.ID(), kvs}
}

func (e *Entry) ID() string {
	return fmt.Sprintf("%d-%d", e.Ms, e.Seq)
}

type Field struct {
	Key   string
	Value string
}

func (s *Stream) Append(entryID string, fields []Field, createdAt time.Time) (string, []byte) {
	ms, seq, err := s.generateID(entryID)
	if err != nil {
		return "", err
	}

	e := Entry{Ms: ms, Seq: seq, Fields: fields, CreatedAt: createdAt}
	s.entries = append(s.entries, e)

	return e.ID(), nil
}

func (s *Stream) Between(start, end string) ([]Entry, []byte) {
	var (
		startMs, startSeq int
		endMs, endSeq     int
		err               error
	)

	if start == "-" {
		startMs, startSeq = 0, 0
	} else if strings.Contains(start, "-") {
		startMs, startSeq, err = parseID(start)
	} else {
		startMs, err = strconv.Atoi(start)
	}

	if err != nil {
		return nil, errInvalidStreamID
	}

	if end == "+" {
		endMs, endSeq = math.MaxInt, math.MaxInt
	} else if strings.Contains(end, "-") {
		endMs, endSeq, err = parseID(end)
	} else {
		endMs, err = strconv.Atoi(end)
		endSeq = math.MaxInt
	}

	if err != nil {
		return nil, errInvalidStreamID
	}

	lb := s.lowerBound(startMs, startSeq)
	ub := s.upperBound(endMs, endSeq)
	if lb > ub {
		return nil, nil
	}

	return s.entries[lb:ub], nil
}

func (s *Stream) AfterTime(t time.Time) []Entry {
	i := sort.Search(len(s.entries), func(i int) bool {
		return s.entries[i].CreatedAt.After(t)
	})

	return s.entries[i:]
}

func (s *Stream) After(start string) ([]Entry, []byte) {
	ms, seq, err := parseID(start)
	if err != nil {
		return nil, errInvalidStreamID
	}

	ub := s.upperBound(ms, seq)
	return s.entries[ub:], nil
}

func (s *Stream) lowerBound(ms, seq int) int {
	return sort.Search(len(s.entries), func(i int) bool {
		e := s.entries[i]
		return e.Ms >= ms && e.Seq >= seq
	})
}

func (s *Stream) upperBound(ms, seq int) int {
	return sort.Search(len(s.entries), func(i int) bool {
		e := s.entries[i]
		return e.Ms > ms || e.Ms == ms && e.Seq > seq
	})
}

func (s *Stream) Len() int { return len(s.entries) }

func (s *Stream) generateID(entryID string) (int, int, []byte) {
	switch {
	case entryID == "*":
		return s.generateFullID()
	case strings.HasSuffix(entryID, "-*"):
		msStr, _ := strings.CutSuffix(entryID, "-*")

		ms, err := strconv.Atoi(msStr)
		if err != nil || ms < 0 {
			return 0, 0, errInvalidStreamID
		}

		return s.generatePartialID(ms)
	default:
		return s.validateID(entryID)
	}
}

func (s *Stream) generateFullID() (int, int, []byte) {
	return s.generatePartialID(int(time.Now().UnixMilli()))
}

func (s *Stream) generatePartialID(ms int) (int, int, []byte) {
	if len(s.entries) == 0 {
		if ms == 0 {
			return 0, 1, nil
		}
		return ms, 0, nil
	}

	last := s.entries[len(s.entries)-1]

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
	ms, seq, err := parseID(entryID)
	if err != nil || ms < 0 || seq < 0 {
		return 0, 0, errInvalidStreamID
	}

	if ms == 0 && seq == 0 {
		return 0, 0, errXaddZeroID
	}

	if len(s.entries) == 0 {
		return ms, seq, nil
	}

	if last := s.entries[len(s.entries)-1]; ms < last.Ms || ms == last.Ms && seq <= last.Seq {
		return 0, 0, errXaddEqualOrSmaller
	}

	return ms, seq, nil
}

func parseID(entryID string) (int, int, error) {
	var ms, seq int
	_, err := fmt.Sscanf(entryID, "%d-%d", &ms, &seq)
	return ms, seq, err
}

var (
	errInvalidStreamID    = resp.SimpleError("ERR Invalid stream ID specified as stream command argument")
	errXaddEqualOrSmaller = resp.SimpleError("ERR The ID specified in XADD is equal or smaller than the target stream top item")
	errXaddZeroID         = resp.SimpleError("ERR The ID specified in XADD must be greater than 0-0")
)
