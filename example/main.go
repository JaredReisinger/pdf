package main

import (
	"log/slog"
	"os"

	"github.com/jaredreisinger/pdf"
)

func main() {
	// TODO: turn this into a real PDF generation example!
	p := pdf.NewPDF()

	// p.NewObject("This is a string.")
	page := p.NewPage()
	c := page.NewContent()

	c.WriteString(`BT
	/F13 12 Tf
	18 TL
	72 720 Td
	(This is some sample text.) Tj
	T* (And another line?) Tj
	T* 150 Tz (And a third line?) Tj
	T* 100 Tz (And a fourth line?) Tj
ET`)

	f, err := os.Create("example.pdf")
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	defer f.Close()

	p.WriteTo(f)
}
