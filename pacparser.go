package pacparser

// go-pacparser - golang bindings for pacparser library

import (
	"errors"
	"net"
	"os"
	"strings"
)

// #cgo LDFLAGS: -lpacparser
// #include <stdarg.h>
// #include <stdio.h>
// #include <strings.h>
// #include <pacparser.h>
//
// static char lastError[2048]  = "";
// static int  bufferPosition   = 0;
//
// int bufferErrors(const char *fmt, va_list argp) {
//   bufferPosition += vsnprintf(lastError+bufferPosition,
//                    sizeof(lastError)-bufferPosition, fmt, argp);
//   return bufferPosition;
// }
//
// char *getLastError() {
//   return (char *)lastError;
// }
//
// void resetLastError() {
//   bufferPosition = 0;
//   lastError[0] = '\0';
// }
//
import "C"

// MaxConcurrency maximum pending requests
const MaxConcurrency = 100

// ParserInstance is the core package instance that stores information
// for all functions contained within the package
type ParserInstance struct {
	pac  string // pac file body
	myip string // returned by myIpAddress() javascript function
	err  error  // last instance error
}

// Unexported common response struct
type parserResponse struct {
	status bool   // translated error from pacparser
	proxy  string // response from FindProxyForURL
	err    error  // last request error
}

// Unexported parse request struct
type parsePacRequest struct {
	inst *ParserInstance
	resp chan *parserResponse
}

// Unexported findProxy request struct
type findProxyRequest struct {
	inst *ParserInstance
	url  string // url argument to FindProxyForURL
	host string // host argument to FindProxyForURL
	resp chan *parserResponse
}

// internal variables
var parsePacChannel chan *parsePacRequest
var findProxyChannel chan *findProxyRequest
var myIpDefault string

// Process upstream error responses
func getLastError() error {
	var lines []string // error lines
	// pull and trim upstream error string
	str := strings.TrimSpace(C.GoString(C.getLastError()))
	// check string
	if str == "" {
		return nil
	}
	// reset upstream error buffer
	C.resetLastError()
	// split upstream message on newline
	for _, l := range strings.Split(str, "\n") {
		var split []string // temp split holder
		// remove library prefixes
		if strings.HasPrefix(strings.TrimSpace(l), "pacparser.c: ") {
			split = strings.SplitN(l, ": ", 3)
			if len(split) == 3 {
				l = split[2]
			}
		}
		// append lines and standardize output
		lines = append(lines, strings.TrimSuffix(strings.TrimSpace(l), "."))
	}
	// check length - remove last line
	if len(lines) > 1 {
		lines = lines[:len(lines)-1]
	}
	// rejoin and return as error
	return errors.New(strings.Join(lines, " "))
}

// Handler to ensure only one active request to the underlying library
func parseHandler() {
	// event loop
	for {
		select {
		// handle parse requests
		case req := <-parsePacChannel:
			// initialize pacparser library context
			C.pacparser_init()
			// deprecated function in newer library versions
			// and simply returns without taking any action
			C.pacparser_enable_microsoft_extensions()
			// build response
			resp := new(parserResponse)
			// parse pac contents and set error
			// upstream function returns 1 on success and 0 on failure
			resp.status = (int(C.pacparser_parse_pac_string(C.CString(req.inst.pac))) != 0)
			// set error
			resp.err = getLastError()
			// cleanup library context
			C.pacparser_cleanup()
			// send response
			req.resp <- resp
		// handle find requests
		case req := <-findProxyChannel:
			// initialize pacparser library context
			C.pacparser_init()
			// deprecated function in newer library versions
			// and simply returns without taking any action
			C.pacparser_enable_microsoft_extensions()
			// build response
			resp := new(parserResponse)
			// parse pac contents to ensure we are using the right body
			// upstream function returns 1 on success and 0 on failure
			resp.status = (int(C.pacparser_parse_pac_string(C.CString(req.inst.pac))) != 0)
			// set error
			resp.err = getLastError()
			// check response
			if resp.status && resp.err == nil {
				// set ip
				C.pacparser_setmyip((C.CString(req.inst.myip)))
				// find proxy
				resp.proxy = C.GoString(C.pacparser_find_proxy(C.CString(req.url), C.CString(req.host)))
				// set error
				resp.err = getLastError()
				// workaround for sometimes empty values
				if resp.proxy == "" {
					resp.proxy = "undefined"
				}
				// handle undefined proxy returns
				if resp.proxy == "undefined" {
					resp.status = false
					// set error if empty
					if resp.err == nil {
						resp.err = errors.New("Invalid proxy returned")
					}
				}
			}
			// cleanup library context
			C.pacparser_cleanup()
			// send response
			req.resp <- resp
		}
	}
}

// Initialize base parser library and start handler
func init() {
	// set error handler
	C.pacparser_set_error_printer(C.pacparser_error_printer(C.bufferErrors))
	// build channels
	parsePacChannel = make(chan *parsePacRequest, 100)
	findProxyChannel = make(chan *findProxyRequest, 100)
	// set default ip
	myIpDefault = "127.0.0.1"
	// attempt to find local hostname
	if host, err := os.Hostname(); err == nil {
		// attempt to resolve returned host
		if addrs, err := net.LookupIP(host); err == nil {
			// loop over resolved addresses
			for _, addr := range addrs {
				if !addr.IsLoopback() {
					// set default ip address
					myIpDefault = addr.String()
					// break after first valid address
					break
				}
			}
		}
	}
	// spawn handler
	go parseHandler()
}
