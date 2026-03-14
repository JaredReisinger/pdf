package pdf

import (
	"io"
	"strings"
)

type Content struct {
	p    *PDF
	page *Page
	objectID

	s strings.Builder
}

func (page *Page) NewContent() *Content {
	p := page.p
	content := &Content{
		p:        p,
		page:     page,
		objectID: p.newObjectID(),
	}

	page.contents = append(page.contents, content)

	return content
}

func (c *Content) WriteTo(w io.Writer) (int64, error) {
	var written, n int64
	var err error

	n, err = c.objectID.writeStartTo(w)
	written += n
	if err != nil {
		return written, err
	}

	// write content dict...
	b := []byte(c.s.String())
	dict := map[string]any{
		"Type":   Name("Content"),
		"Length": len(b),
		// "Parent":    p.parent.Reference(),
		// "Resources": map[string]any{},
		// "MediaBox":  []any{0, 0, 612, 792},
	}

	n, err = WriteDictionaryTo(w, dict)
	written += n
	if err != nil {
		return written, err
	}

	n, err = writeBytesSlicesTo(w,
		[]byte("\nstream\n"),
		b,
		[]byte("\nendstream"),
	)

	n, err = c.objectID.writeEndTo(w)
	written += n
	if err != nil {
		return written, err
	}

	return written, nil
}

// expose the strings.Builder helpers
func (c *Content) Write(b []byte) (int, error) {
	return c.s.Write(b)
}

func (c *Content) WriteString(s string) (int, error) {
	return c.s.WriteString(s)
}

func (c *Content) WriteRune(r rune) (int, error) {
	return c.s.WriteRune(r)
}

func (c *Content) WriteByte(b byte) error {
	return c.s.WriteByte(b)
}
