package toolkit

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

func TestTools_RandomString(t *testing.T) {
	var testTools Tools
	randomString := testTools.RandomString(10)
	if len(randomString) != 10 {
		t.Errorf("Expected length of random string to be 10, but got %d", len(randomString))
	}
}

var uploadTests = []struct {
	name          string
	allowedTypes  []string
	renamefile    bool
	errorExpected bool
}{
	{"ValidFileType", []string{"image/jpeg", "image/png"}, false, false},
	{"InvalidFileType", []string{"image/gif"}, false, true},
	{"RenameFile", []string{"image/png"}, true, false},
}

func TestTools_UploadFiles(t *testing.T) {

	for _, e := range uploadTests {
		//set up a pipe to avoid bufferring
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer writer.Close()
			defer wg.Done()

			/// create a form data file
			part, err := writer.CreateFormFile("file", "./testdata/img.png")
			if err != nil {
				t.Errorf("Error creating form file: %v", err)
				return
			}
			f, err := os.Open("./testdata/img.png")
			if err != nil {
				t.Errorf("Error opening file: %v", err)
				return
			}
			defer f.Close()
			img, _, err := image.Decode(f)
			if err != nil {
				t.Errorf("Error decoding image: %v", err)
				return
			}

			err = png.Encode(part, img)
			if err != nil {
				t.Errorf("Error encoding image: %v", err)
				return
			}

		}()

		// read from the pipe which receives the data
		request := httptest.NewRequest(http.MethodPost, "/", pr)
		request.Header.Add("Content-Type", writer.FormDataContentType())
		var testTools Tools
		testTools.AllowedFileTypes = e.allowedTypes

		uploodedfiles, err := testTools.UploadFiles(request, "./testdata/uploads", e.renamefile)

		if err != nil && !e.errorExpected {
			t.Error(err)
		}

		if !e.errorExpected {
			file := fmt.Sprintf("./testdata/uploads/%s", uploodedfiles[0].NewFileName)
			if _, err := os.Stat(file); os.IsNotExist(err) {
				t.Errorf("%s: expected file to exist :%s", e.name, err.Error())
			}
			os.Remove(file)
		}

		if !e.errorExpected && err != nil {
			t.Errorf("%s: error expected but none receieved", e.name)
		}

		wg.Wait()
	}
}

func TestTools_UploadOneFile(t *testing.T) {

	// for _, e := range uploadTests {
	//set up a pipe to avoid bufferring
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)
	// wg := sync.WaitGroup{}
	// wg.Add(1)
	go func() {
		defer writer.Close()
		// defer wg.Done()

		/// create a form data file
		part, err := writer.CreateFormFile("file", "./testdata/img.png")
		if err != nil {
			t.Errorf("Error creating form file: %v", err)
			return
		}
		f, err := os.Open("./testdata/img.png")
		if err != nil {
			t.Errorf("Error opening file: %v", err)
			return
		}
		defer f.Close()
		img, _, err := image.Decode(f)
		if err != nil {
			t.Errorf("Error decoding image: %v", err)
			return
		}

		err = png.Encode(part, img)
		if err != nil {
			t.Errorf("Error encoding image: %v", err)
			return
		}

	}()

	// read from the pipe which receives the data
	request := httptest.NewRequest(http.MethodPost, "/", pr)
	request.Header.Add("Content-Type", writer.FormDataContentType())
	var testTools Tools
	// testTools.AllowedFileTypes = e.allowedTypes

	uploodedfile, err := testTools.UploadOneFile(request, "./testdata/uploads", true)

	if err != nil {
		t.Error(err)
	}

	file := fmt.Sprintf("./testdata/uploads/%s", uploodedfile.NewFileName)
	if _, err := os.Stat(file); os.IsNotExist(err) {
		t.Errorf(" expected file to exist :%s", err.Error())
	}
	os.Remove(file)

	// if !e.errorExpected && err != nil {
	// 	t.Errorf("%s: error expected but none receieved", e.name)
	// }

	// wg.Wait()
	// }
}

func TestTools_CreateDirIfnotExists(t *testing.T) {
	var testTool Tools

	dirName := "./testdata/myDir"

	err := testTool.CreateDirIfNotExist(dirName)

	if err != nil {
		t.Error(err)
	}

	err = testTool.CreateDirIfNotExist(dirName)

	if err != nil {
		t.Error(err)
	}
	os.Remove(dirName)
}

var slugTests = []struct {
	name          string
	s             string
	expected      string
	errorExpected bool
}{
	{name: "valid string", s: "Now is the time", expected: "now-is-the-time", errorExpected: false},
	{name: "Invalid string", s: "Now is the time", expected: "now-is.the.time", errorExpected: true},
	{name: "Invalid sluggify", s: "Now is the time", expected: "now-is.the.time", errorExpected: true},
	{name: "empty sluggify", s: "", expected: "", errorExpected: true},
}

func TestTools_Slugify(t *testing.T) {
	var testTool Tools

	for _, e := range slugTests {
		slug, err := testTool.Slugify(e.s)
		if err != nil && !e.errorExpected {
			t.Errorf("%s: error received when none expected: %s", e.name, err.Error())
		}

		if !e.errorExpected && slug != e.expected {
			t.Errorf("%s : wrong slug returned; expected %s but got %s", e.name, e.expected, slug)
		}
	}
}

func TestTools_DownloadStaticFile(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	var tool Tools

	tool.DownloadStaticFile(rr, req, "./testdata", "img.png", "pubbs.png")

	res := rr.Result()
	defer res.Body.Close()

	if res.Header["Content-Length"][0] != "534283" {
		t.Error("wrong content length of", res.Header["Content-Length"][0])
	}
}
