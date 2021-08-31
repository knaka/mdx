package mdx

import "log"

type mdxElem struct {
	depth int
}

type mdxElemMethods interface {
	Name() string
	Depth() int
}

var _ mdxElemMethods = (*mdxElem)(nil)

func (elem *mdxElem) Name() string {
	log.Fatal("Do not call me")
	return ""
}

func (elem *mdxElem) Depth() int {
	return elem.depth
}

type mdxLinkElem struct {
	mdxElem
	href string
}

type mdxCodeElem struct {
	mdxElem
	filepath string
}

func (elem *mdxLinkElem) Name() string {
	return "mdxlink"
}

func (elem *mdxCodeElem) Name() string {
	return "mdxcode"
}

type mdxTocElem struct {
	mdxElem
	pattern string
}

func (elem *mdxTocElem) Name() string {
	return "mdxtoc"
}
