package pdf

import (
	"io"
)

type Catalog struct {
	p *PDF
	objectID
	// rootTree *PageTree
}

// should take parent?
func (p *PDF) NewCatalog() *Catalog {
	c := &Catalog{
		p:        p,
		objectID: p.newObjectID(),
	}

	return c
}

func (c *Catalog) WriteTo(w io.Writer) (int64, error) {
	var written, n int64
	var err error

	n, err = c.objectID.writeStartTo(w)
	written += n
	if err != nil {
		return written, err
	}

	n, err = WriteDictionaryTo(w, map[string]any{
		"Type":  Name("Catalog"),
		"Pages": c.p.pageTree.Reference(),
	})
	written += n
	if err != nil {
		return written, err
	}

	n, err = c.objectID.writeEndTo(w)
	written += n
	if err != nil {
		return written, err
	}

	return written, nil
}

// func (c *Catalog) AddRootPageTree(tree *PageTree) {
// 	c.rootTree = tree
// }
