# pdf

The `pdf` package represents the fundamental concepts from the ISO 32000-2:2020 standard. Any additional layout conveniences _**should**_ come from another sibling package.

## High- and low-level constructs

I'm still working on making this "holistically rational"; many features in a PDF file require objects to know about each other's IDs, which tacitly ties objects _to a specific PDF file_. I'd really like to defer that coupling to the ["last responsible moment"](https://blog.codinghorror.com/the-last-responsible-moment/) wherever possible.

Certainly, the idea of flowing text into boxes and calculating sizes does not at all depend on a specific PDF file... but making those calculations depends on a font, which has to be referenced. Similarly, the "idea of a Page" seems like an abstract concept, but at some point _something_ needs to be keeping track of all the Pages that you have, issuing object IDs, etc.

I've tried to avoid things like "creating the PDF object construct" until the last moment, but it's incredibly awkward. My current thought is that perhaps the data needed for object IDs and references are created when the high-level construct is created (i.e. creating a Page necessarily issues a new object ID from the PDF file), but that there's no separate "object as a type which knows the data it holds". Instead, there are "start object" and "end object" helpers that the higher-level constructs use when writing themselves.

## Serializing

In general, the lower-level PDF constructs implement the `io.WriterTo` interface, signalling that they know how to write themselves to an `io.Writer`.  Higher-level constructs, like the PDF file itself, use a wrapper around `io.Writer` that keeps track of byte offsets since several PDF implementation details relay on knowing the offset of PDF objects.
