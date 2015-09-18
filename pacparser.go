package pacparser

// go-pacparser - golang bindings for pacparser library

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
//   bufferPosition = vsnprintf(lastError+bufferPosition, sizeof(lastError)-bufferPosition, fmt, argp);
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
import "errors"
import "strings"

// Maximum pending requests
const MaxConcurrency = 100

// Core package instance that stores information
// for all functions contained within the package
type ParserInstance struct {
	pac string // pac file body
	err error  // last instance error
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

// Unexported package channels
var parsePacChannel chan *parsePacRequest
var findProxyChannel chan *findProxyRequest

// Error returned when FindProxy fails
var InvalidProxyReturn = errors.New("Invalid proxy return value")

// Error returned when we fail to parse the passed URL string
var InvalidURL = errors.New("Invalid URL")

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
		lines = append(lines, strings.TrimSpace(l))
	}
	// check length - remove last line
	if len(lines) > 1 {
		lines = lines[:len(lines)-1]
	}
	// rejoin and return as error
	return errors.New(strings.Join(lines, " -> "))
}

// Handler to ensure only one active request to the underlying library
func parseHandler() {
	// cleanup engine on exit
	defer C.pacparser_cleanup()

	// event loop
	for {
		select {
		// handle parse requests
		case req := <-parsePacChannel:
			// build response
			resp := new(parserResponse)
			// parse pac contents and set error
			// upstream function returns 1 on success and 0 on failure
			resp.status = (int(C.pacparser_parse_pac_string(C.CString(req.inst.pac))) != 0)
			// set error
			resp.err = getLastError()
			// send response
			req.resp <- resp
		// handle find requests
		case req := <-findProxyChannel:
			// build response
			resp := new(parserResponse)
			// parse pac contents to ensure we are using the right body
			// upstream function returns 1 on success and 0 on failure
			resp.status = (int(C.pacparser_parse_pac_string(C.CString(req.inst.pac))) != 0)
			// set error
			resp.err = getLastError()
			// check response
			if resp.status {
				// find proxy
				resp.proxy = C.GoString(C.pacparser_find_proxy(C.CString(req.url), C.CString(req.host)))
				// set error
				resp.err = getLastError()
				// check proxy
				if resp.proxy == "undefined" || resp.proxy == "" {
					resp.status = false
					resp.err = InvalidProxyReturn
				}
			}
			// send response
			req.resp <- resp
		}
	}
}

// Initialize base parser libary and start handler
func init() {
	// initialize pacparser library
	C.pacparser_init()
	C.pacparser_enable_microsoft_extensions()
	// set error handler
	C.pacparser_set_error_printer(C.pacparser_error_printer(C.bufferErrors))
	// build channels
	parsePacChannel = make(chan *parsePacRequest, 100)
	findProxyChannel = make(chan *findProxyRequest, 100)
	// spawn handler
	go parseHandler()
}
