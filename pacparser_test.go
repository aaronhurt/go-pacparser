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
	pacFiles map[string]string
}

// get test pacfiles
func (s *pacparserTestSuite) readFiles(fpath string) map[string]string {
	fullpath := path.Join(s.basePath, fpath) // path to test files
	fileMap := make(map[string]string)       // return map of file names and content
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
		fileMap[file.Name()] = string(contents)
	}
	// return file list
	return fileMap
}

// initiate test suite
func (s *pacparserTestSuite) SetupSuite() {
	var err error // error holder
	// set base path and check error
	if s.basePath, err = os.Getwd(); err != nil {
		panic(err.Error())
	}
	// read files
	s.pacFiles = s.readFiles("test")
}

// run tests
func TestSuite(t *testing.T) {
	ts := new(pacparserTestSuite)
	suite.Run(t, ts)
}

// bad pacfile test
func (s *pacparserTestSuite) TestBad1() {
	// load pac bodyand parse
	pp := New(s.pacFiles["bad1.pac"])
	// assert on parse
	s.False(pp.Parse())
	// execute FindProxyForURL and log
	ok, pxy := pp.FindProxy("http://www.google.com/")
	s.T().Logf("ok: %t, pxy: %s", ok, pxy)
	// assert returns
	s.False(ok)
	s.Empty(pxy)
	// pull error and log
	lastError := pp.LastError()
	s.T().Logf("lastError: %s", lastError)
	// assert on error
	s.NotNil(lastError)
}

// bad pacfile test
func (s *pacparserTestSuite) TestBad2() {
	// load pac body
	pp := New(s.pacFiles["bad2.pac"])
	// assert on parse
	s.True(pp.Parse())
	// execute FindProxyForURL and log
	ok, pxy := pp.FindProxy("http://www.google.com/")
	s.T().Logf("ok: %t, pxy: %s", ok, pxy)
	// assert returns
	s.False(ok)
	s.Equal("undefined", pxy)
	// pull eror and log
	lastError := pp.LastError()
	s.T().Logf("lastError: %s", lastError)
	// assert on error
	s.NotNil(lastError)
}

// good pacfile
func (s *pacparserTestSuite) TestGood1() {
	var ok bool    // status return
	var pxy string // proxy line
	// init good instance
	pp := New(s.pacFiles["good1.pac"])
	// assert on parse
	s.True(pp.Parse())
	// exectute FindProxyForURL and log
	ok, pxy = pp.FindProxy("http://www.google.com/")
	s.T().Logf("http://www.google.com/ -> %s", pxy)
	// assert returns
	s.True(ok)
	s.NotEmpty(pxy)
	s.Nil(pp.LastError())
	// exectute FindProxyForURL and log
	ok, pxy = pp.FindProxy("http://test.local/")
	s.T().Logf("http://test.local/ -> %s", pxy)
	// assert returns
	s.True(ok)
	s.NotEmpty(pxy)
	s.Nil(pp.LastError())
	// exectute FindProxyForURL and log
	ok, pxy = pp.FindProxy("http://localhost/")
	s.T().Logf("http://localhost/ -> %s", pxy)
	// assert returns
	s.True(ok)
	s.NotEmpty(pxy)
	s.Nil(pp.LastError())
	// exectute FindProxyForURL and log
	ok, pxy = pp.FindProxy("http://www.abcdomain.com/")
	s.T().Logf("http://www.abcdomain.com/ -> %s", pxy)
	// assert returns
	s.True(ok)
	s.NotEmpty(pxy)
	s.Nil(pp.LastError())
}

// test with IsValid
func (s *pacparserTestSuite) TestValid() {
	// load pacfile
	pp := New(s.pacFiles["bad1.pac"])
	// check validity and log
	ok := pp.IsValid()
	s.T().Logf("bad1.pac IsValid: %t", ok)
	// assert result
	s.False(ok)
	// load pacfile
	pp = New(s.pacFiles["bad2.pac"])
	// check validity and log
	ok = pp.IsValid()
	s.T().Logf("bad2.pac IsValid: %t", ok)
	// assert result
	s.False(ok)
	// load pacfile
	pp = New(s.pacFiles["good1.pac"])
	// check validity and log
	ok = pp.IsValid()
	s.T().Logf("good1.pac IsValid: %t", ok)
	// assert result
	s.True(ok)
}

// benchmark parse
func BenchmarkParse(b *testing.B) {
	ts := new(pacparserTestSuite)
	ts.SetupSuite()
	pp := New(ts.pacFiles["good1.pac"])
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ok := pp.Parse()
		err := pp.LastError()
		if !ok || err != nil {
			panic(err.Error())
		}
	}
}

// benchmark find
func BenchmarkFind(b *testing.B) {
	ts := new(pacparserTestSuite)
	ts.SetupSuite()
	pp := New(ts.pacFiles["good1.pac"])
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ok, pxy := pp.FindProxy("http://www.google.com/")
		err := pp.LastError()
		if !ok || pxy == "" || err != nil {
			panic(err.Error())
		}
	}
}
