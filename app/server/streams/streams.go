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

func (s *Stream) Append(entryID string, fields []Field) (string, []byte) {
	ms, seq, err := s.generateID(entryID)
	if err != nil {
		return "", err
	}

	e := Entry{Ms: ms, Seq: seq, Fields: fields}
	s.entries = append(s.entries, e)

	return e.ID(), nil
}

func (s *Stream) Query(start, end string) ([]Entry, []byte) {
	var (
		startMs, startSeq int
		endMs, endSeq     int
		err               error
	)

	if start == "-" {
		startMs, startSeq = 0, 0
	} else if strings.Contains(start, "-") {
		startMs, startSeq, err = parseFullID(start)
	} else {
		startMs, err = strconv.Atoi(start)
	}

	if err != nil {
		return nil, errInvalidStreamID
	}

	if end == "+" {
		endMs, endSeq = math.MaxInt, math.MaxInt
	} else if strings.Contains(end, "-") {
		endMs, endSeq, err = parseFullID(end)
	} else {
		endMs, err = strconv.Atoi(end)
		endSeq = math.MaxInt
	}

	if err != nil {
		return nil, errInvalidStreamID
	}

	lowerBound := sort.Search(len(s.entries), func(i int) bool {
		e := s.entries[i]
		return e.Ms >= startMs && e.Seq >= startSeq
	})

	upperBound := sort.Search(len(s.entries), func(i int) bool {
		e := s.entries[i]
		return e.Ms > endMs || e.Ms == endMs && e.Seq > endSeq
	})

	if lowerBound > upperBound {
		return nil, nil
	}

	return s.entries[lowerBound:upperBound], nil
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
	ms, seq, err := parseFullID(entryID)
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

func parseFullID(entryID string) (int, int, error) {
	var ms, seq int
	_, err := fmt.Sscanf(entryID, "%d-%d", &ms, &seq)
	return ms, seq, err
}

var (
	errInvalidStreamID    = resp.SimpleError("ERR Invalid stream ID specified as stream command argument")
	errXaddEqualOrSmaller = resp.SimpleError("ERR The ID specified in XADD is equal or smaller than the target stream top item")
	errXaddZeroID         = resp.SimpleError("ERR The ID specified in XADD must be greater than 0-0")
)
