package pacparser

// go-pacparser - golang bindings for pacparser library

import "net/url"

// test url for pac validation
const TestURL = "http://www.google.com/"

// create a new pacparser instance
func New(pac string) *ParserInstance {
	// allocate instance
	inst := new(ParserInstance)
	// populate elements
	inst.pac = pac
	// return
	return inst
}

// parse a pac body
func (inst *ParserInstance) Parse() bool {
	// build and populate request
	req := new(parsePacRequest)
	req.inst = inst
	req.resp = make(chan *parserResponse, 1)
	// send request
	parsePacChannel <- req
	// wait for response
	resp := <-req.resp
	// set instance error
	inst.err = resp.err
	// return
	return resp.status
}

// find proxy for given arguments
func (inst *ParserInstance) FindProxy(urlString string) (bool, string) {
	// parse host from url
	u, err := url.Parse(urlString)
	// check err
	if err != nil {
		inst.err = InvalidURL
		return false, ""
	}
	// build and populate request
	req := new(findProxyRequest)
	req.inst = inst
	req.url = u.String()
	req.host = u.Host
	req.resp = make(chan *parserResponse, 1)
	// send request
	findProxyChannel <- req
	// wait for response
	resp := <-req.resp
	// set instance error
	inst.err = resp.err
	// return
	return resp.status, resp.proxy
}

// return last error
func (inst *ParserInstance) LastError() error {
	return inst.err
}

// verify a pacfile
func (inst *ParserInstance) IsValid() bool {
	// parse pacfile and check return
	if !inst.Parse() {
		return false
	}
	// evaluate function
	ok, _ := inst.FindProxy(TestURL)
	// check return
	if !ok {
		return false
	}
	// default return
	return true
}
