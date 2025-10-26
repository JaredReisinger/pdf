package pdf

import "io"

// Probably need an object/struct to represent a PDF file, structured with
// Attachments, Pages, etc.

func WriteToFile(w io.Writer) (int64, error) {
	var written int64

	// See 7.5.2, File header
	n, err := writeStuffTo(w,
		[]byte("%PDF-2.0\n"),
		// include a comment with 4 binary (> 128) bytes
		[]byte("%æìöú"), // haha! "aeiou"!
	)
	written += n
	if err != nil {
		return written, err
	}

	return written, nil
}

// PDF represents an entire PDF document
type PDF struct {
}

func (p *PDF) WriteTo(w io.Writer) (int64, error) {
	var written int64

	// wrap writer with counting writer so we can calculate offsets
	w2 := &writer{w, 0}

	n, err := p.writeHeader(w2)
	written += n
	if err != nil {
		return written, err
	}

	return written, nil

}

func (p *PDF) writeHeader(w *writer) (int64, error) {
	// See 7.5.2, File header
	return writeStuffTo(w,
		[]byte("%PDF-2.0\n"),
		// include a comment with 4 binary (> 128) bytes
		[]byte("%æìöú"), // haha! "aeiou"!
	)
}

// NOTE: We need a wrapper around the passed-in byte-/string-writer so that we
// can track offsets in some cases.  I *don't* think we ever need to go back to
// re-write a particular location (io.OffsetWriter), though.
//
// We *only* need this for actual file writing...

type writer struct {
	w      io.Writer
	offset int
}

// func NewCountingWriter(w io.Writer) *countingWriter {
// 	return &countingWriter{w, 0}
// }

func (w *writer) Offset() int {
	return w.offset
}

// func (w *countingWriter) WriteString(s string) (int, error) {
// 	n, err := w.w.WriteString(s)
// 	if err == nil {
// 		w.offset += n
// 	}
// 	return n, err
// }

func (w *writer) Write(b []byte) (int, error) {
	n, err := w.w.Write(b)
	if err == nil {
		w.offset += n
	}
	return n, err
}
