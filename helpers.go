package pacparser

// go-pacparser - golang bindings for pacparser library

import "net/url"

// Test URL used by IsValid()
const TestURL = "http://www.google.com/"

// Create a new pacparser instance associated with the passed PAC file contents
func New(pac string) *ParserInstance {
	// allocate instance
	inst := new(ParserInstance)
	// populate elements
	inst.pac = pac
	// return
	return inst
}

// Parse the PAC body associated with the instance and return true or false.
// Errors that may occur are stored in the instance and may be retrieved
// by a call to LastError()
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

// Execute the FindProxyForURL function in the associated PAC body
// and find the proxy return for the given URL string.  The host portion
// will be parsed out of the URL passed to the function.  The returned
// string may be "" or "undefined" in addition to a proper proxy return
// depending on the contents of the associated PAC body.
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

// Return the most recent error that occured in the instance.
func (inst *ParserInstance) LastError() error {
	return inst.err
}

// Shortcut function that combines Parse() and FindProxy() with
// a test URL to quickly validate PAC syntax and basic functionality.
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
