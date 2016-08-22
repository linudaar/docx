package xml

import "strings"
import "io"

var (
	xsiURL     = "http://www.w3.org/2001/XMLSchema-instance"
	xsiNilName = Name{Space: xsiURL, Local: "nil"}
	xsiNilTrue = Attr{Name: xsiNilName, Value: "true"}
)

// Namespace adds a namespace declaration that will be encoded as an attribute of the root element.
func (enc *Encoder) Namespace(prefix, ns string) {
	p := &enc.p
	if p.attrPrefix == nil {
		p.attrPrefix = make(map[string]string)
		p.attrNS = make(map[string]string)
	}
	p.attrNS[prefix] = ns
	p.attrPrefix[ns] = prefix
}

// OptimizeNamespaces enables optimization of namespaces emitted by the encoder
func (enc *Encoder) OptimizeNamespaces(optimizeNS bool) {
	enc.p.optimizeNS = optimizeNS
}

// PrefixElements enables the use of namespace prefixes for XML elements.
func (enc *Encoder) PrefixElements(prefixElements bool) {
	enc.p.prefixElements = prefixElements
}

type charDataReader struct {
	*Decoder
	eof bool
}

func (cdr *charDataReader) Read(bs []byte) (n int, err error) {
	if cdr.eof {
		return 0, io.EOF
	}
	if len(bs) == 0 {
		return 0, nil
	}
	for {
		if n >= len(bs) {
			break
		}

		b, ok := cdr.getc()
		if !ok {
			err = cdr.err
			break
		}

		if b == '<' {
			// we reached the end of the chardata
			err = io.EOF
			cdr.ungetc(b) // put back the '<'
			cdr.eof = true
			break
		}

		bs[n] = b
		n++
	}
	return
}

// CharDataReader returns a reader that allows streaming of large character data chunks.
// (e.g. when downloading files)
func (d *Decoder) CharDataReader() io.Reader {
	return &charDataReader{Decoder: d}
}

func (p *printer) marshalNil(start StartElement) error {
	start.Attr = append(start.Attr, xsiNilTrue)
	start.Empty = true

	if p.optimizeNS {
		p.optimizeNamespace(&start)
	}

	if err := p.writeStart(&start, true); err != nil {
		return err
	}
	return p.cachedWriteError()
}

func (p *printer) addRootNamespaces(start *StartElement) {
	for k, v := range p.attrNS {
		start.Attr = append(start.Attr, Attr{Name: Name{Space: "", Local: "xmlns:" + k}, Value: v})
	}
	p.markPrefix()
}

func (p *printer) optimizeNamespace(start *StartElement) {
	if start.Name.Space != "" && len(p.tags) > 0 {
		for i := len(p.tags) - 1; i >= 0; i-- {
			tag := p.tags[i]
			if tag.Space == "" {
				continue
			}
			if tag.Space != start.Name.Space {
				return
			}
			// we use the same namespace as our parent. so do not emit it.
			start.Name.Space = ""
			return
		}
	}
}

func (p *printer) determinePrefix(start *StartElement) string {
	if start.Name.Space == "" {
		// If the element has no NS it inherits the namespace of its parent.
		// This is easy if the parent has a "xmlns" attribute set.
		// However, if it the parent has been prefixed, we also need to use that prefix
		// because prefixes will not be inherited!
		if len(p.tags) <= 0 {
			return ""
		}
		return p.tags[len(p.tags)-1].Prefix
	}
	if p.attrPrefix == nil {
		return ""
	}
	return p.attrPrefix[start.Name.Space]
}

func namespaceToPrefix(url string) string {
	// Pick a name. We try to use the final element of the path
	// but fall back to _.
	prefix := strings.TrimRight(url, "/")
	if i := strings.LastIndex(prefix, "/"); i >= 0 {
		prefix = prefix[i+1:]
	}

	// take the 1st lower case character from each part.
	parts := strings.FieldsFunc(prefix, func(r rune) bool {
		return r == '-' || r == ' ' || r == '.'
	})
	if len(parts) == 1 {
		return strings.ToLower(parts[0])
	}
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToLower(part[:1])
		}
	}
	return strings.Join(parts, "")
}
