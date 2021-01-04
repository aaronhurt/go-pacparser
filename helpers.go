package pacparser

// go-pacparser - golang bindings for pacparser library

import (
	"errors"
	"net"
	"net/url"
	"strings"
)

// TestURL used by IsValid()
const TestURL = "http://www.google.com/"

// New creates a new pacparser instance associated with the passed PAC file contents
func New(pac string) *ParserInstance {
	// allocate instance
	inst := new(ParserInstance)
	// populate elements
	inst.pac = pac
	// set IP address to package default
	inst.myip = myIpDefault
	// return the instance
	return inst
}

// Parse the PAC body associated with the instance and return true or false.
// Errors that may occur are stored in the instance and may be retrieved
// by a call to LastError() and should be handled by the client BEFORE
// calling any additional instance functions that may overwrite the instance
// error state.
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

// FindProxy executes the FindProxyForURL function in the associated PAC body
// and find the proxy return for the given URL string.  The host portion
// will be parsed out of the URL passed to the function.  The returned
// string may be "" or "undefined" in addition to a proper proxy return
// depending on the contents of the associated PAC body.
func (inst *ParserInstance) FindProxy(urlString string) (bool, string) {
	// parse host from url
	u, err := url.Parse(urlString)
	// check err
	if err != nil {
		inst.err = err
		return false, ""
	}
	// build and populate request
	req := new(findProxyRequest)
	req.inst = inst
	req.url = u.String()
	req.host = strings.Split(u.Host, ":")[0]
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

// LastError returns the most recent error that occured in the instance.
func (inst *ParserInstance) LastError() error {
	return inst.err
}

// SetMyIp sets the IP address returned by the myIpAddress() javascript function
// when processing PAC scripts.  The package attempts to resolve the
// local system hostname and defaults to "127.0.0.1" if the local
// hostname is not resolvable.
func (inst *ParserInstance) SetMyIp(ipString string) error {
	if ip := net.ParseIP(ipString); ip != nil {
		inst.myip = ip.String()
		return nil
	}
	// ip didn't parse
	return errors.New("Invalid IP")
}

// MyIp returns the IP address used by the instance.
func (inst *ParserInstance) MyIp() string {
	return inst.myip
}

// Reset the instance IP address and error state to the default values.
func (inst *ParserInstance) Reset() {
	inst.err = nil
	inst.myip = myIpDefault
}

// IsValid provides a shortcut that combines Parse() and FindProxy() with
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
