# pdf

The `pdf` package represents the fundamental concepts from the ISO 32000-2:2020 standard. Any additional layout conveniences _**should**_ come from another sibling package.

## Serializing and naming convention

There's a bit of ambiguity or conflict when it comes to naming various parts of the serialization process.  If you have a `Page`, or `ObjectReference`, and you want to serialize it to a file/bytes/pipe, the obvious name for the the method is `Write()`. This is _also_ the term used for feeding data to a write-to-able object.  Both of these make sense on their own:

```go
somePipe.Write(someBytes)
```

and 

```go
someObject.Write(somePipe)
```

Maybe "WriteTo" would be better? Oh, my!  There's actually an `io.WriterTo` interface?