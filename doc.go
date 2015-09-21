/*
Package pacparser provides Go bindings and thread safety for the pacparser library.

Notes

Package functions are called off of a pacparser instance that is instantiated
with the PAC data that should be used with all other calls from the instance.

	pp := pacparser.New(pacFileString)
	...

Errors that occur are stored in the instance and should be checked after
calling any function that may produce an error.  Additional function calls
can and in many cases will replace the instance error contents.  It is up to
the user to store and process instance errors as collected.

	lastError := pp.LastError()
	...

For more information please see the usage example code below.

	package main

	import (
		"io/ioutil"
		"log"

		"github.com/leprechau/go-pacparser"
	)

	func main() {
		// read in an example file
		data, err := ioutil.ReadFile("../test/good1.pac")

		// check error
		if err != nil {
			panic(err)
		}

		// create a pacparser instance
		pp := pacparser.New(string(data))

		// parse pacfile and check error
		if !pp.Parse() {
			log.Fatalf("Error parsing pacfile: %s\n", pp.LastError())
		}

		// find proxy for given url
		ok, proxy := pp.FindProxy("http://www.google.com")

		// check return
		if ok && pp.LastError() == nil {
			log.Printf("%s", proxy)
		} else {
			log.Fatalf(pp.LastError().Error())
		}
	}

*/
package pacparser
