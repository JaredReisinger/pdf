# pdf/fonts

Trying to figure out the font information provided in TrueType (`.ttf`). OpenType (`.otf`), and ??? fonts.  The font rendering doesn't matter as long as we can get the metrics to know how much space each glyph will occupy.

Looks like we can use the image-drawing font code from `golang.org/x/image/font`. It includes the ability to load/parse OpenType and TrueType fonts, and can calculate string measurements. That _should_ be enough to let us calculate layouts.