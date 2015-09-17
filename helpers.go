package pacparser

// go-pacparser - golang bindings for pacparser library

// test payloads for pac validation
const TestURL = "http://www.google.com/"
const TestHost = "www.google.com"

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
	req.resp = make(chan *ParserResponse, 1)
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
func (inst *ParserInstance) FindProxy(url, host string) (bool, string) {
	// build and populate request
	req := new(findProxyRequest)
	req.inst = inst
	req.url = url
	req.host = host
	req.resp = make(chan *ParserResponse, 1)
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
	ok, pxy := inst.FindProxy(TestURL, TestHost)
	// check values
	if !ok || pxy == "" {
		return false
	}
	// default return
	return true
}
