/*
Example program using the go-pacparser wrapper package.

This application will parse the test/good1.pac example file and evaluate the
contained FindProxyForURL function to return the PROXY line for
http://www.google.com and then exit.

Please view the source for more information.  The code is commented
and should provide a clear example usage of the package.
*/
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
