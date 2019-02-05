layla
=====

layla is a layout and template language based on the xelf language framework.

It is primarily used as an exploration into the feasibility of the xelf packages, but should also
provides a simple layout templates for thermal label printers and html previews.

The idea is that layout definitions using the xelf are already templates. Reusing the std tools
and some custom expression we can build the layout node tree. The nodes proxy to custom go
structs, making it easy to work with.

Layla should grow to a useful (for me at least) package of its own. I would like it to support
   text, block, rect, ellipse, qrcode, barcode and image elements
   group, vbox, hbox and table layouts

There will someday be render packages for:
   tsc  Taiwan Semiconductor (TSC) label printer, specifically for the DA-200 printer
   t88  Epson receipt printer, specifically the TM-T88III
   html preview in HTML with barcode rendering using boombuler/barcode
   pdf  renderer using jung-kurt/gofpdf

License
-------

Copyright (c) Martin Schnabel. All rights reserved.
Use of the source code is governed by a BSD-style license that can found in the LICENSE file.
