package libtest

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"stuartdd.com/lib"
)

var (
	content1 = []byte("This is one of those times")
	content2 = []byte("This is not one of those times")
	password = []byte("mysecretpassword")
	salt     = []byte("012345678901234567890123456789XX")

	testFile = "TestData.txt"
)

func TestMain(m *testing.M) {
	/*
		Creat the test file in a known state
	*/
	resetTestFile(testFile, content1)
	/*
		Run the tests
	*/
	os.Exit(m.Run())
}

func TestConstructErrors(t *testing.T) {
	_, err := lib.NewFileData("NotFound.txt")
	if err == nil {
		t.Error("err should not be nil for file not found")
	}
	_, err = lib.NewFileDataEnc("NotFoundEnc.txt", password, salt)
	if err == nil {
		t.Error("err should be nil for enc file found")
	}
}

func TestConstruct(t *testing.T) {
	resetTestFile(testFile, content1)
	fd, err := lib.NewFileData(testFile)
	if err != nil {
		t.Error("err should be nil for file found")
	}
	if fd.GetFileName() != testFile {
		t.Errorf("File name is incorrect")
	}
	if string(fd.GetSalt()) != "" {
		t.Errorf("Salt should be empty")
	}
	if string(fd.GetContent()) != string(content1) {
		t.Errorf("Content is incorrect")
	}
	if fd.IsEncrypted() {
		t.Errorf("File was NOT encrypted")
	}
}

func TestStoreContent(t *testing.T) {
	resetTestFile(testFile, content1)
	fd1, err1 := lib.NewFileData(testFile)
	if err1 != nil {
		t.Error("err should be null for file found")
	}
	if string(fd1.GetContent()) != string(content1) {
		t.Errorf("Failed to init content:'%s' != '%s'", string(fd1.GetContent()), string(content1))
	}

	fd1.SetContent(content2)
	if string(fd1.GetContent()) != string(content2) {
		t.Errorf("Content setter is incorrect:'%s' != '%s'", string(fd1.GetContent()), string(content2))
	}
	fd1.Store()

	fd2, err2 := lib.NewFileData(testFile)
	if err2 != nil {
		t.Error("err should be null for file found")
	}
	if string(fd2.GetContent()) != string(content2) {
		t.Errorf("Store failed to change content:'%s' != '%s'", string(fd2.GetContent()), string(content2))
	}
	if fd2.GetFileName() != testFile {
		t.Error("Stored File name is incorrect")
	}
	if string(fd2.GetSalt()) != "" {
		t.Error("Stored file Salt should be empty")
	}
	if fd2.IsEncrypted() {
		t.Error("Stored file was NOT encrypted")
	}
}

func resetTestFile(fileName string, content []byte) {
	err := ioutil.WriteFile(fileName, content, 0644)
	if err != nil {
		log.Fatalf("Failed to reset test file: %s. %s\n", fileName, err)
	}
}
