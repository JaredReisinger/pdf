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
	_ = page

	page = p.NewPage()
	_ = page

	page = p.NewPage()
	_ = page

	f, err := os.Create("example.pdf")
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	defer f.Close()

	p.WriteTo(f)
}
