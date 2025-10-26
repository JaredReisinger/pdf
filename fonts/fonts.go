package fonts

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

func LoadDefaultFont() error {
	return LoadBytes(goregular.TTF)
}

func LoadTTF(file string) error {
	slog.Info("loading font", "file", file)

	fontData, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	return LoadBytes(fontData)
}

func LoadBytes(fontData []byte) error {
	fonts, err := opentype.ParseCollection(fontData)
	// fonts, err := sfnt.ParseCollection(fontData)
	if err != nil {
		return err
	}

	fontBuf := &sfnt.Buffer{}

	slog.Info("read font file", "count", fonts.NumFonts())

	for i := 0; i < fonts.NumFonts(); i++ {
		f, err := fonts.Font(i)
		if err != nil {
			return err
		}

		// namesX := map[string]string{}
		names := []slog.Attr{}
		for id := sfnt.NameIDCopyright; id <= sfnt.NameIDVariationsPostScriptPrefix; id++ {
			s, err := f.Name(fontBuf, id)
			if err != nil {
				if errors.Is(err, sfnt.ErrNotFound) {
					continue
				}
				return err
			}
			names = append(names, slog.Attr{Key: nameIDString(id), Value: slog.StringValue(s)})
			// namesX[nameIDString(id)] = s
		}

		slog.LogAttrs(context.Background(), slog.LevelInfo, "font names", names...)

		ppem := fixed.Int26_6(f.UnitsPerEm())
		// need to pass "ppem" (points per em?) and font hinting (?)
		metrics, err := f.Metrics(fontBuf, ppem, font.HintingNone)
		if err != nil {
			return err
		}

		slog.Info("font metrics",
			"height", metrics.Height,
			"ascent", metrics.Ascent,
			"descent", metrics.Descent,
			"xheight", metrics.XHeight,
			"capheight", metrics.CapHeight,
			"caretslope.x", metrics.CaretSlope.X,
			"caretslope.y", metrics.CaretSlope.Y,
		)

		slog.Info("font data",
			"glyphs", f.NumGlyphs(),
		)

		// try calculating font sizes...
		face, err := opentype.NewFace(f, &opentype.FaceOptions{
			Size:    10,
			DPI:     100,
			Hinting: font.HintingNone,
		})
		if err != nil {
			return err
		}

		for _, s := range []string{"Hello", "world", "this", "is", "awesome!"} {
			dx := font.MeasureString(face, s)
			slog.Info("measurement", "s", s, "dx", dx)
		}

	}

	return nil
}

// too bad that there's not an existing formatter for this!
func nameIDString(id sfnt.NameID) string {
	names := []string{
		"Copyright",
		"Family",
		"Subfamily",
		"UniqueIdentifier",
		"Full",
		"Version",
		"PostScript",
		"Trademark",
		"Manufacturer",
		"Designer",
		"Description",
		"VendorURL",
		"DesignerURL",
		"License",
		"LicenseURL",
		"TypographicFamily",
		"TypographicSubfamily",
		"CompatibleFull",
		"SampleText",
		"PostScriptCID",
		"WWSFamily",
		"WWSSubfamily",
		"LightBackgroundPalette",
		"DarkBackgroundPalette",
		"VariationsPostScriptPrefix",
	}

	if id < 0 || int(id) >= len(names) {
		return "(UNKNOWN)"
	}
	return names[id]
}
