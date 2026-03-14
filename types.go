package pdf

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"slices"
	"strings"

	"golang.org/x/image/math/fixed"
)

func WriteValueTo(w io.Writer, el any) (int64, error) {
	if writerTo, ok := el.(io.WriterTo); ok {
		return writerTo.WriteTo(w)
	}

	// just in case we have a *typed* nil value, be sure to serialize as null
	if el == nil {
		return retShim(w.Write([]byte("null")))
	}

	switch val := el.(type) {
	// case nil:
	// 	w.WriteString("null")
	case bool:
		return retShim(fmt.Fprintf(w, "%t", val))
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		return retShim(fmt.Fprintf(w, "%d", val))
	case float32, float64:
		// TODO: optimize formatting? Using %g means it *could* use exponent
		// notation, but in practice, we should never have values that big.
		return retShim(fmt.Fprintf(w, "%g", val))
	case fixed.Int26_6:
		// we convert to float64... dividing by 64.0 is the trick, as the
		// integer portion of left-shifted by 6 bits, effectively multiplied by
		// 64.
		return retShim(fmt.Fprintf(w, "%g", float64(val)/64.0))
	case string:
		return WriteStringTo(w, val)
	case Name:
		return WriteNameTo(w, string(val))
	case []any:
		return WriteArrayTo(w, val)
	case map[string]any:
		return WriteDictionaryTo(w, val)
	case []byte:
		return WriteByteArrayTo(w, val)
	default:
		slog.Error("unknown array type", "value", el)
		panic("NEED TO ADD TYPE FORMATTING")
	}
}

func WriteStringTo(w io.Writer, s string) (int64, error) {
	var written int64

	n, err := retShim(w.Write([]byte("(")))
	written += n
	if err != nil {
		return written, err
	}

	// We escape parens and reverse solidus (backslash), see 7.3.4.2.  Although
	// PDF *can* handle balanced parens, it might be easier for us to always
	// escape them?
	//
	// Do we want to escape newline, tab, etc? (I don't think we'll generate
	// them, really.)
	for _, c := range s {
		switch c {
		case '(', ')', '\\':
			n, err = retShim(fmt.Fprintf(w, `\%c`, c))
		default:
			if c > 255 {
				// need to handle big unicode? into octal
				panic("???")
			}
			n, err = retShim(fmt.Fprintf(w, "%c", c))
		}

		written += n
		if err != nil {
			return written, err
		}
	}

	n, err = retShim(w.Write([]byte(")")))
	written += n
	if err != nil {
		return written, err
	}

	return written, nil
}

func WriteNameTo(w io.Writer, s string) (int64, error) {
	var written int64

	n, err := retShim(w.Write([]byte("/")))
	written += n
	if err != nil {
		return written, err
	}

	for _, c := range s {
		if c >= '!' && c <= '~' && c != '#' {
			n, err = retShim(fmt.Fprintf(w, "%c", c))
			written += n
			if err != nil {
				return written, err
			}
			continue
		}

		// *only* use the low byte!
		n, err = retShim(fmt.Fprintf(w, "#%x", byte(c)))
		written += n
		if err != nil {
			return written, err
		}
	}

	return written, nil
}

func WriteArrayTo(w io.Writer, a []any) (int64, error) {
	var written int64

	n, err := retShim(w.Write([]byte("[")))
	written += n
	if err != nil {
		return written, err
	}

	for i, el := range a {
		if i > 0 {
			n, err = retShim(w.Write([]byte(" ")))
			written += n
			if err != nil {
				return written, err
			}
		}

		n, err = WriteValueTo(w, el)
		written += n
		if err != nil {
			return written, err
		}
	}

	n, err = retShim(w.Write([]byte("]")))
	written += n
	if err != nil {
		return written, err
	}

	return written, nil
}

// We *don't* count nesting levels... should we even include newlines/tabs?
func WriteDictionaryTo(w io.Writer, dict map[string]any) (int64, error) {
	var written int64

	n, err := retShim(w.Write([]byte("<<")))
	written += n
	if err != nil {
		return written, err
	}

	// Go intentionally randomizes the order of map iteration... do we want to
	// provide a normalized order?  We may want our own slices.Sorted() style
	// implementation to put things like /Type first!
	for i, k := range slices.SortedFunc(maps.Keys(dict), dictKeyCmp) {
		if i > 0 {
			// use \n\t for debugging...
			n, err = retShim(w.Write([]byte("\n\t")))
			// n, err = retShim(w.Write([]byte(" ")))
			written += n
			if err != nil {
				return written, err
			}
		}
		n, err = WriteNameTo(w, k)
		written += n
		if err != nil {
			return written, err
		}

		n, err = retShim(w.Write([]byte(" ")))
		written += n
		if err != nil {
			return written, err
		}

		n, err = WriteValueTo(w, dict[k])
		written += n
		if err != nil {
			return written, err
		}
	}

	n, err = retShim(w.Write([]byte(">>")))
	written += n
	if err != nil {
		return written, err
	}

	return written, nil
}

// dictKeyCmp is a comparison/sorting function for dictionary keys that puts
// certain well-known keys at the very beginning
func dictKeyCmp(a, b string) int {
	for _, s := range []string{"Type", "Subtype"} {
		if a == s {
			return -1
		}
		if b == s {
			return 1
		}
	}

	return strings.Compare(a, b)
}

func WriteByteArrayTo(w io.Writer, b []byte) (int64, error) {
	return retShim(fmt.Fprintf(w, "<%x>", b))
}

// Name is a utility type to flag a string specifically as a PDF "Name" value.
type Name string

type objectID struct {
	p          *PDF
	id         uint64
	generation uint // we never use generation; should we even have this member?
}

func (o objectID) writeStartTo(w io.Writer) (int64, error) {
	var written int64

	n, err := retShim(w.Write([]byte("\n")))
	written += n
	if err != nil {
		return written, err
	}

	// *if* we're writing a PDF, we need to track the offset!
	if ww, ok := w.(*writer); ok && o.p != nil {
		ww.p.objectOffsets[o.id] = ww.Offset()
	}

	n, err = retShim(fmt.Fprintf(w, "%d %d obj\n", o.id, o.generation))
	written += n
	if err != nil {
		return written, err
	}

	return written, nil
}

func (o objectID) writeEndTo(w io.Writer) (int64, error) {
	return retShim(w.Write([]byte("\nendobj\n")))
}

func (o objectID) Reference() io.WriterTo {
	return objectReference(o)
}

type objectReference objectID

// func (o *ObjectReference) Write(w Writer) {
func (o objectReference) WriteTo(w io.Writer) (int64, error) {
	return retShim(fmt.Fprintf(w, "%d %d R", o.id, o.generation))
}

// func (o objectReference) Reference() io.WriterTo {
// 	return o
// }

// type ObjectReference interface {
// 	Reference() io.WriterTo
// }

type Stream struct {
	Dictionary map[string]any
	buf        *bytes.Buffer
}

var _ io.WriterTo = &Stream{} // static interface implementation check

func NewStream(b []byte) *Stream {
	return &Stream{
		Dictionary: make(map[string]any, 1),
		buf:        bytes.NewBuffer(b),
	}
}

// See 7.3.8
func (s *Stream) WriteTo(w io.Writer) (int64, error) {
	var written int64

	b := s.buf.Bytes()

	// We need to auto-graft the "/Length" in, but only if it's not already
	// present. Example in PDF spec is with an object-reference length... but I
	// have no idea why one would ever want to do that in practice.
	if _, ok := s.Dictionary["Length"]; !ok {
		s.Dictionary["Length"] = len(b)
	}

	n, err := WriteDictionaryTo(w, s.Dictionary)
	written += n
	if err != nil {
		return written, err
	}

	n, err = writeBytesSlicesTo(w,
		[]byte("\nstream\n"),
		b,
		[]byte("\nendstream\n"),
	)
	written += n
	if err != nil {
		return written, err
	}

	return written, nil
}

// io.Writer.Write() returns (int, error), but io.WriterTo.Write() returns
// (int64, error). This is truly obnoxious. We *also* often need to write a
// sequence of things, and capturing the length each time is obnoxious... so we
// have a helper that manages both of the above.
func writeStuffTo(w io.Writer, writerTos ...io.WriterTo) (int64, error) {
	var written, n int64
	var err error

	for _, writerTo := range writerTos {
		// we don't use retShim here just to avoid the additional call
		n, err = writerTo.WriteTo(w)
		written += n
		if err != nil {
			return written, err
		}
	}
	return written, nil
}

type writeToFunc func(w io.Writer) (int64, error)

func writeFuncsTo(w io.Writer, writerToFuncs ...writeToFunc) (int64, error) {
	var written, n int64
	var err error

	for _, writerToFunc := range writerToFuncs {
		// we don't use retShim here just to avoid the additional call
		n, err = writerToFunc(w)
		written += n
		if err != nil {
			return written, err
		}
	}
	return written, nil
}

func writeBytesSlicesTo(w io.Writer, bs ...[]byte) (int64, error) {
	var written int64
	for _, b := range bs {
		// we don't use retShim here just to avoid the additional call
		n, err := w.Write(b)
		written += int64(n)
		if err != nil {
			return written, err
		}
	}
	return written, nil
}

func retShim(n int, err error) (int64, error) {
	return int64(n), err
}
