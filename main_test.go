package main_test

import (
	"flag"
	"log"
)

//	func TestHome(t *testing.T) {
//		var router = SetupRouter(true)
//		w := httptest.NewRecorder()
//		req, _ := http.NewRequest("GET", "/", nil)
//		router.ServeHTTP(w, req)
//		assert.Equal(t, 200, w.Code)
//		assert.NotEqual(t, "", w.Body.String())
//	}
func init() {
	testFile := flag.String("test.testlogfile", "", "The test file to run")
	flag.Parse()
	if *testFile != "" {
		// Run the test file
		log.Printf("Running test file: %s", *testFile)
	}
}
