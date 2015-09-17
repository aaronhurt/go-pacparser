package pacparser

// go-pacparser - golang bindings for pacparser library

import (
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

// everything we need to run our tests
type pacparserTestSuite struct {
	suite.Suite
	basePath string
	pacFiles map[string][]byte
}

// get test pacfiles
func (s *pacparserTestSuite) readFiles(fpath string) map[string][]byte {
	fullpath := path.Join(s.basePath, fpath) // path to test files
	retVal := make(map[string][]byte)        // return map of file names and content
	// read all files and check error
	files, err := ioutil.ReadDir(fullpath)
	if err != nil {
		panic(err.Error())
	}
	// loop over read files
	for _, file := range files {
		contents, err := ioutil.ReadFile(path.Join(fullpath, file.Name()))
		if err != nil {
			panic(err.Error())
		}
		retVal[file.Name()] = contents
	}
	// return file list
	return retVal
}

// initiate test suite
func (s *pacparserTestSuite) SetupSuite() {
	var err error // error holder
	// set base path and check error
	if s.basePath, err = os.Getwd(); err != nil {
		panic(err.Error())
	}
	// read files
	s.pacFiles = s.readFiles("test_files")
}

// run tests
func TestSuite(t *testing.T) {
	ts := new(pacparserTestSuite)
	suite.Run(t, ts)
}

// bad pacfile test
func (s *pacparserTestSuite) TestBad1() {
	pp := New(string(s.pacFiles["bad1.pac"]))
	s.False(pp.Parse())
	ok, pxy := pp.FindProxy("http://www.google.com/", "www.google.com")
	s.T().Logf("ok: %t, pxy: %s", ok, pxy)
	s.False(ok)
	s.Empty(pxy)
	lastError := pp.LastError()
	s.T().Logf("lastError: %s", lastError)
	s.NotNil(lastError)
}

// bad pacfile test
func (s *pacparserTestSuite) TestBad2() {
	pp := New(string(s.pacFiles["bad2.pac"]))
	s.False(pp.Parse())
	ok, pxy := pp.FindProxy("http://www.google.com/", "www.google.com")
	s.T().Logf("ok: %t, pxy: %s", ok, pxy)
	s.False(ok)
	s.Empty(pxy)
	lastError := pp.LastError()
	s.T().Logf("lastError: %s", lastError)
	s.NotNil(lastError)
}

// good pacfile
func (s *pacparserTestSuite) TestGood1() {
	var ok bool    // status return
	var pxy string // proxy line
	pp := New(string(s.pacFiles["good1.pac"]))
	s.True(pp.Parse())
	ok, pxy = pp.FindProxy("http://www.google.com/", "www.google.com")
	s.T().Logf("http://www.google.com/, www.google.com -> %s", pxy)
	s.True(ok)
	s.NotEmpty(pxy)
	s.Nil(pp.LastError())
	ok, pxy = pp.FindProxy("http://test.local/", "test.local")
	s.T().Logf("http://test.local/, test.local -> %s", pxy)
	s.True(ok)
	s.NotEmpty(pxy)
	s.Nil(pp.LastError())
	ok, pxy = pp.FindProxy("http://localhost/", "localhost")
	s.T().Logf("http://localhost/, localhost -> %s", pxy)
	s.True(ok)
	s.NotEmpty(pxy)
	s.Nil(pp.LastError())
	ok, pxy = pp.FindProxy("http://www.abcdomain.com/", "www.abcdomain.com")
	s.T().Logf("http://www.abcdomain.com/, www.abcdomain.com -> %s", pxy)
	s.True(ok)
	s.NotEmpty(pxy)
	s.Nil(pp.LastError())
}
