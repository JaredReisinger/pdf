package main

import (
	"log/slog"
	"os"

	"github.com/jaredreisinger/pdf"
)

var fontFile = "/mnt/c/Users/jared/AppData/Local/Microsoft/Windows/Fonts/YoungSerif-Regular.ttf"

// var fontFile = "/mnt/c/Users/jared/AppData/Local/Microsoft/Windows/Fonts/SpaceGrotesk-Regular.ttf"

// var fontFile = "/mnt/c/Users/jared/AppData/Local/Microsoft/FontCache/4/CloudFonts/Arial Nova/33045110977.ttf"

func main() {
	// TODO: turn this into a real PDF generation example!
	p := pdf.NewPDF()

	font := p.NewFont(fontFile)
	page := p.NewPage()
	page.UseFont(font)
	c := page.NewContent()

	c.WriteString(`BT
	/F0 12 Tf
	18 TL
	72 720 Td
	(This is some sample text.) Tj
	T* (And another line?) Tj
	T* 150 Tz (And a third line?) Tj
	T* 100 Tz (And a fourth line?) Tj
	T* (The quick brown fox jumps over the lazy dog.) Tj
ET`)

	f, err := os.Create("example.pdf")
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	defer f.Close()

	p.WriteTo(f)
}
