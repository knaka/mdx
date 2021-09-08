package mdpp

import "log"

type mdppElem struct {
	depth int
}

type mdppElemMethods interface {
	Name() string
	Depth() int
}

var _ mdppElemMethods = (*mdppElem)(nil)

func (elem *mdppElem) Name() string {
	log.Fatal("Do not call me")
	return ""
}

func (elem *mdppElem) Depth() int {
	return elem.depth
}

type mdppLinkElem struct {
	mdppElem
	href string
}

type mdppCodeElem struct {
	mdppElem
	filepath string
}

func (elem *mdppLinkElem) Name() string {
	return "mdpplink"
}

func (elem *mdppCodeElem) Name() string {
	return "mdppcode"
}

type mdppIndexElem struct {
	mdppElem
	pattern string
}

func (elem *mdppIndexElem) Name() string {
	return "mdppindex"
}
