# Background

Or, "Another PDF creator? Really?"

For some context, I generate publications from data sources on a regular basis, and I've used two basic methods for this:

1. For relatively simple publications, I use Microsoft Word and the Mail Merge feature with the data in an Excel file. This works well enough, bur requires me to click through the mail merge process to create a new document with the merged data, often requires removing a bunch of section breaks that Word insists on injecting to separate each "item", then exporting to PDF.  In some cases, I need two different views on the data, which means managing _**two**_ mail merge template files, and merging the results at the end.

2. For a much more data-heavy publication—an annual roster of 200-300 people—I use a Go program to massage the data, then use Go templates to output TeX/LaTeX files. These files are processed (with XeTeX) into the final PDF. I love the typographical control and the ability to create an index... but it's a very "heavy" process and it's been so long since I've used TeX in any deep way that there are some weird layout cases that I just can't seem to get to work right.

I started looking at `codeberg.org/go-pdf/fpdf`, and it was enough to get me seriously considering a "generate directly to PDF" solution. At the same time, it's clear that `fpdf` has inherited some quirks from the original `FPDF` that I don't love. It was very much designed to handle basic text output, and other features have been grafted on without a holistic refactoring of the original codebase. Among others:

- The ASCII and UTF-8 codepaths are sometimes quite duplicative.  It matters _strongly_ which case you're in, for instance having `SplitLines()` for ASCII and `SplitText()` for UTF-8.

- There's really not much logic to help you lay out your document. You can get the width of a string, but for vertical flow, you generally have to actually "write" to the document and see where you end up.

- For a relatively simple change (I noticed that the logic that handled justified text for UTF-8 sometimes missed that it was being requested), it was incredibly difficult to write a test case without access to the deep internals of `fpdf`.

- Font handling is... quirky. Adding ASCII vs. UTF-8 fonts is very different, as is adding a font from a file vs. `io.Reader` or bytes.

On top of that, there were features I wanted to add:

- hyphenation when laying out text

- _maybe_ performing slight whitespace squeezing to optimize layout (badness calculations)

- _maybe_ handling whitespace between words and between sentences slightly differently (having intersentence spacing slightly larger than interword spacing)

- smarter column handling (send text to the first column, then the second, but determine the height that keeps the two columns close to even)

- the ability to calculate the complete results of laying out text before adding it to the document (this should also help with unit tests!)

- _??? others?_

If you squint, you can see that some of what I'm looking for is "TeX-lite" typographical control.  I started by trying to graft some of this on top of `fpdf`, but very quickly found that I really wanted to reach inside and get at the underlying layout calculations. Since it seemed like I was getting close to re-writing a bunch of layout calculations anyway, it made sense to step back and consider whether a ground-up redo made sense.

This project is an attempt to find that sweet spot between "only very basic PDF text handling" and "just use TeX, already".


## Notes

The PDF spec (ISO3200-2:2020), in section 7.7.3.3 (starting on page 104) defines the Page objects, which are a dictionary that includes a bunch of details specific to each page. Included in this is the content stream for the contents (yay!) and **UserUnit**, a PDF 1.6-added value that's a multiple of 1/72nd of an inch (a point) which is the default unit if unspecified.  It's not yet clear if graphics primitives use UserUnits or not, but I suspect they do.

I probably need some low-level "basic PDF types" classes.

