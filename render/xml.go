// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package render

import (
	"encoding/xml"
	"net/http"
)

// XML contains the given interface object.
type Xml struct {
	Data interface{}
}

var xmlContentType = []string{"application/xml; charset=utf-8"}

// Render (XML) encodes the given interface object and writes data with custom ContentType.
func (x Xml) Render(w http.ResponseWriter) error {
	return xml.NewEncoder(w).Encode(x.Data)
}

// WriteContentType (XML) writes XML ContentType for response.
func (x Xml) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, xmlContentType)
}
