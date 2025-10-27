package pdf

import (
	"io"
)

type Page struct {
	p      *PDF
	parent *PageTree

	objectID

	contents []*Content
}

// Should this be a method on the page tree?
func (p *PDF) NewPage() *Page {
	page := &Page{
		p:        p,
		objectID: p.newObjectID(),
	}

	p.pages = append(p.pages, page)
	p.pageTree.AddPage(page)

	return page
}

func (page *Page) WriteTo(w io.Writer) (int64, error) {
	var written, n int64
	var err error

	n, err = page.objectID.writeStartTo(w)
	written += n
	if err != nil {
		return written, err
	}

	var contents any
	if len(page.contents) == 1 {
		contents = page.contents[0].Reference()
	} else if len(page.contents) > 1 {
		contentsArray := make([]any, 0, len(page.contents))
		for _, c := range page.contents {
			contentsArray = append(contentsArray, c.Reference())
		}
		contents = contentsArray
	}

	// write page dict...
	dict := map[string]any{
		"Type":      Name("Page"),
		"Parent":    page.parent.Reference(),
		"Resources": map[string]any{},
		"MediaBox":  []any{0, 0, 612, 792},
	}

	if contents != nil {
		dict["Contents"] = contents
	}

	n, err = WriteDictionaryTo(w, dict)
	written += n
	if err != nil {
		return written, err
	}

	n, err = page.objectID.writeEndTo(w)
	written += n
	if err != nil {
		return written, err
	}

	for _, c := range page.contents {
		n, err = c.WriteTo(w)
		written += n
		if err != nil {
			return written, err
		}
	}

	return written, nil
}
