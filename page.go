package pdf

import (
	"io"
)

type Page struct {
	p *PDF
	objectID

	// dict map[string]any
	parent *PageTree
}

func (p *PDF) NewPage() *Page {
	page := &Page{
		p:        p,
		objectID: p.newObjectID(),
		// dict: map[string]any{
		// 	"Type": Name("Page"),
		// },
	}

	p.pages = append(p.pages, page)
	p.pageTree.AddPage(page)

	return page
}

func (p *Page) WriteTo(w io.Writer) (int64, error) {
	var written, n int64
	var err error

	n, err = p.objectID.writeStartTo(w)
	written += n
	if err != nil {
		return written, err
	}

	// write page dict...
	dict := map[string]any{
		"Type":      Name("Page"),
		"Parent":    p.parent.Reference(),
		"Resources": map[string]any{},
		"MediaBox":  []any{0, 0, 612, 792},
	}

	n, err = WriteDictionaryTo(w, dict)
	written += n
	if err != nil {
		return written, err
	}

	n, err = p.objectID.writeEndTo(w)
	written += n
	if err != nil {
		return written, err
	}

	return written, nil
}
