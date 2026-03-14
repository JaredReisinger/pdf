package pdf

import (
	"crypto/rand"
	"fmt"
	"io"
	"sync/atomic"
)

// Probably need an object/struct to represent a PDF file, structured with

// PDF represents an entire PDF document
type PDF struct {
	version int // TBD: for flexible version support?

	objectCounter atomic.Uint64

	// Need to make this a separate type/class? There is always one and only one
	// document catalog.

	// Do "objects" exist before we start writing the file? They may not need
	// to.
	catalog  *Catalog
	pageTree *PageTree
	pages    []*Page
	fonts    []*Font

	// writing metadata
	objectOffsets  map[uint64]uint64
	lastXrefOffset uint64
}

func NewPDF() *PDF {
	p := &PDF{
		version: 2,
	}
	p.catalog = p.NewCatalog()
	p.pageTree = p.NewPageTree()
	// p.catalog.AddRootPageTree(p.pageTree)

	return p
}

// objects have to be numbered by the containing document!
func (p *PDF) newObjectID() objectID {
	o := objectID{
		p:          p,
		id:         p.objectCounter.Add(1),
		generation: 0,
	}

	return o
}

func (p *PDF) WriteTo(w io.Writer) (int64, error) {
	// set up some internal tracking structures...
	p.objectOffsets = make(map[uint64]uint64, 10)

	var written int64

	// wrap writer with counting writer so we can calculate offsets
	w2 := &writer{p, w, 0}

	n, err := p.writeHeader(w2)
	written += n
	if err != nil {
		return written, err
	}

	n, err = p.catalog.WriteTo(w2)
	written += n
	if err != nil {
		return written, err
	}

	n, err = p.pageTree.WriteTo(w2)
	written += n
	if err != nil {
		return written, err
	}

	for _, font := range p.fonts {
		n, err = font.WriteTo(w2)
		written += n
		if err != nil {
			return written, err
		}
	}

	// // TODO: other parts
	// n, err = p.writeObjects(w2)
	// written += n
	// if err != nil {
	// 	return written, err
	// }

	n, err = p.writeCrossReferenceTable(w2)
	written += n
	if err != nil {
		return written, err
	}

	n, err = p.writeFooter(w2)
	written += n
	if err != nil {
		return written, err
	}

	return written, nil

}

func (p *PDF) writeHeader(w *writer) (int64, error) {
	// See 7.5.2, File header
	return writeBytesSlicesTo(w,
		[]byte("%PDF-2.0\n"),
		// include a comment with 4 binary (> 128) bytes
		[]byte("%æìöú"), // haha! "aeiou"!
	)
}

// // not sure if we just iterate objects!
// func (p *PDF) writeObjects(w *writer) (int64, error) {
// 	var written int64

// 	for _, obj := range p.objects {
// 		n, err := obj.WriteTo(w)
// 		written += n
// 		if err != nil {
// 			return written, err
// 		}
// 	}

// 	return written, nil

// }

func (p *PDF) writeCrossReferenceTable(w *writer) (int64, error) {
	var written int64

	n, err := writeBytesSlicesTo(w, []byte("\n"))
	written += n
	if err != nil {
		return written, err
	}

	p.lastXrefOffset = w.Offset()
	n, err = writeBytesSlicesTo(w, []byte("xref\n"))
	written += n
	if err != nil {
		return written, err
	}

	// We always have a single cross-reference table, from 1 to N
	count := p.objectCounter.Load()
	n, err = retShim(fmt.Fprintf(w, "1 %d\n", count))
	written += n
	if err != nil {
		return written, err
	}

	for i := uint64(1); i <= count; i++ {
		n, err = retShim(fmt.Fprintf(w, "%010d 00000 n \n", p.objectOffsets[i]))
		written += n
		if err != nil {
			return written, err
		}
	}

	return written, nil
}

func (p *PDF) writeFooter(w *writer) (int64, error) {
	var written int64

	// calculate file identifier (14.4) -- MD5 of contents?
	// need an io.WriterAt to be able to go back and hash the file?
	// for now we create a random identifier
	fileIdentifier := make([]byte, 16)
	rand.Read(fileIdentifier)

	// See 7.5.5, File trailer
	trailer := map[string]any{
		"Size": p.objectCounter.Load(), // cross references
		"Root": p.catalog.Reference(),  // indirect dictionary reference
		"ID":   []any{fileIdentifier, fileIdentifier},
	}

	n, err := writeBytesSlicesTo(w, []byte("\ntrailer\n"))
	written += n
	if err != nil {
		return written, err
	}
	n, err = WriteDictionaryTo(w, trailer)
	written += n
	if err != nil {
		return written, err
	}

	n, err = writeBytesSlicesTo(w,
		[]byte("\nstartxref\n"),
		[]byte(fmt.Sprintf("%d", p.lastXrefOffset)),
	)
	written += n
	if err != nil {
		return written, err
	}

	return writeBytesSlicesTo(w,
		[]byte("\n%%EOF"),
	)
}

// NOTE: We need a wrapper around the passed-in byte-/string-writer so that we
// can track offsets in some cases.  I *don't* think we ever need to go back to
// re-write a particular location (io.OffsetWriter), though.
//
// We *only* need this for actual file writing...

type writer struct {
	p      *PDF
	w      io.Writer
	offset uint64
}

// func NewCountingWriter(w io.Writer) *countingWriter {
// 	return &countingWriter{w, 0}
// }

func (w *writer) Offset() uint64 {
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
		w.offset += uint64(n)
		// slog.Info("offset bump", "by", n, "to", w.offset)
	}
	return n, err
}
