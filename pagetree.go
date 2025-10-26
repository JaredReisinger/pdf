package pdf

import (
	"io"
)

type PageTree struct {
	p *PDF
	objectID
	// objID objectID
	kids []*Page
}

// should take parent?
func (p *PDF) NewPageTree() *PageTree {
	tree := &PageTree{
		p:        p,
		objectID: p.newObjectID(),
	}

	return tree
}

// func (p *PageTree) Reference() ObjectReference {
// 	return ObjectReference(p.objID)
// }

func (p *PageTree) WriteTo(w io.Writer) (int64, error) {
	var written, n int64
	var err error

	// p.dict["Count"] = len(p.kids)
	kidRefs := make([]any, 0, len(p.kids))

	// write all of the pages (first? last?)
	for _, kid := range p.kids {
		kid.parent = p
		n, err = kid.WriteTo(w)
		written += n
		if err != nil {
			return written, err
		}

		kidRefs = append(kidRefs, kid.objectID.Reference())
	}

	n, err = p.objectID.writeStartTo(w)
	written += n
	if err != nil {
		return written, err
	}

	n, err = WriteDictionaryTo(w, map[string]any{
		"Type":  Name("Pages"),
		"Kids":  kidRefs,
		"Count": len(p.kids),
	})
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

func (p *PageTree) AddPage(page *Page) {
	p.kids = append(p.kids, page)
}
