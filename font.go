package pdf

import (
	"bytes"
	"io"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

const (
	// PDF spec numbers the bits starting at 1, not zero
	FontFlagFixedPitch  = 1 << iota // pos 1: 1
	FontFlagSerif                   // pos 2: 2
	FontFlagSymbolic                // pos 3: 4
	FontFlagScript                  // pos 4: 8
	_                               // pos 5: 16
	FontFlagNonSymbolic             // pos 6: 32
	FontFlagItalic                  // pos 7: 64
	_                               // pos 8: 128
	_                               // pos 9: 256
	_                               // pos 10: 512
	_                               // pos 11: 1024
	_                               // pos 12: 2048
	_                               // pos 13: 4096
	_                               // pos 14: 8092
	_                               // pos 15: 16384
	_                               // pos 16: 32768
	FontFlagAllCap                  // pos 17: 65536
	FontFlagSmallCap                // pos 18: 131072
	FontFlagForceBold               // pos 19: 262144
	_                               // pos 20: 524288
)

type Font struct {
	p *PDF

	objIDFont       objectID
	objIDDescriptor objectID
	objIDFontFile   objectID
	objIDCMap       objectID // only need once ever?

	data           []byte // do we really want this? it's the whole file!
	font           *sfnt.Font
	buf            sfnt.Buffer
	namePostScript string
	flags          int
	bbox           []any
	metrics        font.Metrics
	italicAngle    float64

	firstChar int
	lastChar  int
	widths    []any
}

// Should this be a method on the font tree?
func (p *PDF) NewFont(fontFile string) *Font {
	f := &Font{
		p:               p,
		objIDFont:       p.newObjectID(),
		objIDDescriptor: p.newObjectID(),
		objIDFontFile:   p.newObjectID(),
		// objIDCMap:       p.newObjectID(),
	}

	p.fonts = append(p.fonts, f)

	// We *could* wait to open the font file at write time...
	file, err := os.Open(fontFile)
	if err != nil {
		panic(err)
	}
	f.data, err = io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	coll, err := opentype.ParseCollection(f.data)
	if err != nil {
		panic(err)
	}

	if coll.NumFonts() > 1 {
		panic("only expecting one font")
	}

	f.font, err = coll.Font(0)
	if err != nil {
		panic(err)
	}

	// TODO: add fall-through for missing PostScript name!
	f.namePostScript, err = f.font.Name(&f.buf, sfnt.NameIDPostScript)
	if err != nil {
		panic(err)
	}

	// Fill out descriptor values?
	// slog.Info("font units per em", "upm", f.font.UnitsPerEm())
	// upem := fixed.I(int(f.font.UnitsPerEm()))
	// slog.Info("units per em", "raw", upem)
	// _ = upem
	// ppem := fixed.Int26_6((1000 / (float64(upem) / 64)) * 64)
	ppem := fixed.I(1000)

	rect, err := f.font.Bounds(&f.buf, ppem, font.HintingNone)
	if err != nil {
		panic(err)
	}

	f.bbox = []any{
		rect.Min.X,
		rect.Min.Y,
		rect.Max.X,
		rect.Max.Y,
	}

	f.metrics, err = f.font.Metrics(&f.buf, ppem, font.HintingNone)
	if err != nil {
		panic(err)
	}

	f.flags = FontFlagSerif | FontFlagNonSymbolic
	f.italicAngle = 0
	if f.font.PostTable() != nil {
		f.italicAngle = f.font.PostTable().ItalicAngle
	}

	// because of the way that Go parses found, we *always* start from zero
	f.firstChar = 0
	f.lastChar = f.font.NumGlyphs() - 1

	for i := f.firstChar; i <= f.lastChar; i++ {
		// _, advance, err := f.font.GlyphBounds(&f.buf, sfnt.GlyphIndex(i), ppem, font.HintingNone)
		advance, err := f.font.GlyphAdvance(&f.buf, sfnt.GlyphIndex(i), ppem, font.HintingNone)
		if err != nil {
			// look for ErrNotFound?
			panic(err)
		}
		f.widths = append(f.widths, advance)
	}

	return f
}

func (f *Font) Reference() io.WriterTo {
	return f.objIDFont.Reference()
	// return f.objIDDescriptor.Reference()
}

func (f *Font) WriteTo(w io.Writer) (int64, error) {
	return writeFuncsTo(w,
		f.writeFontObjTo,
		f.writeFontDescriptorTo,
		f.writeFontFileTo,
		// f.writeCMapTo,
	)
}

func (f *Font) writeFontObjTo(w io.Writer) (int64, error) {
	var written, n int64
	var err error

	n, err = f.objIDFont.writeStartTo(w)
	written += n
	if err != nil {
		return written, err
	}

	// write font dict...
	//
	// TODO: need to subtype based on actual font info!
	//
	// OR.... do we use /Type0 with a DescendantFont and ToUnicode map?
	dict := map[string]any{
		"Type":     Name("Font"),
		"Subtype":  Name("TrueType"),
		"BaseFont": Name(f.namePostScript),
		// "Encoding": Name("WinAnsiEncoding"), // ?
		"Encoding":       Name("Identity-H"), // ?
		"FirstChar":      f.firstChar,
		"LastChar":       f.lastChar,
		"Widths":         f.widths,
		"FontDescriptor": f.objIDDescriptor.Reference(),
		// "ToUnicode":      f.objIDCMap.Reference(),
	}

	n, err = WriteDictionaryTo(w, dict)
	written += n
	if err != nil {
		return written, err
	}

	n, err = f.objIDFont.writeEndTo(w)
	written += n
	if err != nil {
		return written, err
	}

	return written, nil
}

func (f *Font) writeFontDescriptorTo(w io.Writer) (int64, error) {
	var written, n int64
	var err error

	// and the descriptor...
	n, err = f.objIDDescriptor.writeStartTo(w)
	written += n
	if err != nil {
		return written, err
	}

	// write font dict...
	dict := map[string]any{
		"Type":        Name("FontDescriptor"),
		"FontName":    Name(f.namePostScript),
		"Flags":       f.flags,
		"FontBBox":    f.bbox,
		"ItalicAngle": f.italicAngle,
		"Ascent":      f.metrics.Ascent,
		"Descent":     f.metrics.Descent,
		"CapHeight":   f.metrics.CapHeight,
		"XHeight":     f.metrics.XHeight,
		"StemV":       0, // 100, // unknown1
		// TODO: FontFile2 or FontFile3 depending on font type!
		"FontFile2": f.objIDFontFile.Reference(), //?
	}

	n, err = WriteDictionaryTo(w, dict)
	written += n
	if err != nil {
		return written, err
	}

	n, err = f.objIDDescriptor.writeEndTo(w)
	written += n
	if err != nil {
		return written, err
	}

	return written, nil
}

func (f *Font) writeFontFileTo(w io.Writer) (int64, error) {
	var written, n int64
	var err error

	// extract the font from the font file... it's too bad we can't "pipe" this
	// over, but we're going to be subsetting it eventually anyway.
	fontData := bytes.Buffer{}
	n, err = f.font.WriteSourceTo(&f.buf, &fontData)
	if err != nil {
		return written, err
	}
	if n != int64(fontData.Len()) {
		panic("WTF?")
	}

	n, err = f.objIDFontFile.writeStartTo(w)
	written += n
	if err != nil {
		return written, err
	}

	// write font dict...
	dict := map[string]any{
		"Length": fontData.Len(),
		// Compress with FlateEncode?
	}

	n, err = WriteDictionaryTo(w, dict)
	written += n
	if err != nil {
		return written, err
	}

	n, err = writeBytesSlicesTo(w,
		[]byte("\nstream\n"),
		fontData.Bytes(),
		[]byte("\nendstream"),
	)
	written += n
	if err != nil {
		return written, err
	}

	n, err = f.objIDFontFile.writeEndTo(w)
	written += n
	if err != nil {
		return written, err
	}

	return written, nil
}

// func (f *Font) writeCMapTo(w io.Writer) (int64, error) {
// 	var written, n int64
// 	var err error

// 	n, err = f.objIDCMap.writeStartTo(w)
// 	written += n
// 	if err != nil {
// 		return written, err
// 	}

// 	// write font dict...
// 	dict := map[string]any{
// 		"Length": len(stdToUnicodeStream),
// 	}

// 	n, err = WriteDictionaryTo(w, dict)
// 	written += n
// 	if err != nil {
// 		return written, err
// 	}

// 	n, err = writeBytesSlicesTo(w,
// 		[]byte("\nstream\n"),
// 		[]byte(stdToUnicodeStream),
// 		[]byte("\nendstream"),
// 	)
// 	written += n
// 	if err != nil {
// 		return written, err
// 	}

// 	n, err = f.objIDCMap.writeEndTo(w)
// 	written += n
// 	if err != nil {
// 		return written, err
// 	}

// 	return written, nil
// }

// const stdToUnicodeStream = `/CIDInit /ProcSet findresource begin
// 12 dict begin
// begincmap
// /CIDSystemInfo <</Registry (Adobe) /Ordering (UCS) /Supplement 0>> def
// /CMapName /Adobe-Identity-UCS def
// /CMapType 2 def
// 1 begincodespacerange
// <0000> <FFFF>
// endcodespacerange
// 1 beginbfrange
// <0000> <FFFF> <0000>
// endbfrange
// endcmap
// CMapName currentdict /CMap defineresource pop
// end
// end`
