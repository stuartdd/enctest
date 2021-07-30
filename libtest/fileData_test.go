package libtest

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"stuartdd.com/lib"
)

var (
	content1     = []byte("{\"text\":\"This is one of those times\"}")
	content2     = []byte("{\"text\":\"This is NOT one of those times\"}")
	content3     = []byte("\n  {}")
	content4     = []byte("[]")
	content5     = []byte("\n  F9")
	password     = []byte("mysecretpassword")
	salt         = []byte("012345678901234567890123456789XX")
	encCon       = []byte{213, 253, 183, 173, 50, 97, 227, 225, 26, 110, 238, 149, 114, 30, 96, 193, 75, 219, 42, 32, 221, 0, 104, 228, 190, 188, 117, 129}
	testFileName = "TestData.txt"
	iterations   = 8 // 2,4,8,16...
)

func TestMain(m *testing.M) {
	/*
		Creat the test file in a known state
	*/
	resetTestFile(testFileName, content1)
	/*
		Run the tests
	*/
	os.Exit(m.Run())
}

func TestEncReload(t *testing.T) {
	resetTestFile(testFileName, encCon)
	fd, err := lib.NewFileDataEnc(testFileName, password, salt, iterations)
	if err != nil {
		fd.StoreContent()
		echoTestFile(testFileName)
		t.Errorf("err should be nil for file found. %s", err)
	}
	fd.SetContent(content2)
	if string(fd.GetContent()) != string(content2) {
		t.Errorf("Content setter is incorrect:'%s' != '%s'", string(fd.GetContent()), string(content2))
	}
	fd.StoreContent()
	fd2, err2 := lib.NewFileDataEnc(testFileName, password, salt, iterations)
	if err2 != nil {
		t.Error("err should be nil for file found")
	}
	if string(fd2.GetContent()) != string(content2) {
		t.Errorf("File content is incorrect %s != %s", string(fd2.GetContent()), string(content2))
	}
	if !fd2.IsRawJson() {
		t.Error("Should be decrypted!")
	}
	if !fd2.HasEncData() {
		t.Error("file should have key and salt")
	}

}
func TestEnc(t *testing.T) {
	resetTestFile(testFileName, content1)
	fd1, err1 := lib.NewFileData(testFileName)
	if err1 != nil {
		t.Error("err should be nil for file found")
	}
	if string(fd1.GetContent()) != string(content1) {
		t.Error("File content is incorrect")
	}
	if !fd1.IsRawJson() {
		t.Error("Json file should be recognised")
	}
	if fd1.HasEncData() {
		t.Error("file should not have key and salt")
	}

	err1 = fd1.StoreContentEncrypted(password, salt, iterations)
	if err1 != nil {
		t.Errorf("Store Enc failed. %s", err1)
	}
	if string(fd1.GetContent()) != string(content1) {
		t.Error("File content is incorrect")
	}
	if !fd1.IsRawJson() {
		t.Error("Json file should be recognised")
	}
	if !fd1.HasEncData() {
		t.Error("file should have key and salt")
	}
	fd2, err2 := lib.NewFileDataEnc(testFileName, password, salt, iterations)
	if err2 != nil {
		t.Error("err should be nil for file found")
	}
	if !fd2.IsRawJson() {
		t.Error("Should be decrypted!")
	}
	if string(fd2.GetContent()) != string(content1) {
		t.Error("Did not decrypt correctly!")
	}
	if !fd2.HasEncData() {
		t.Error("file should have key and salt")
	}
}

func TestJsonRec(t *testing.T) {
	resetTestFile(testFileName, content3)
	fd3, err3 := lib.NewFileData(testFileName)
	if err3 != nil {
		t.Error("err should be nil for file found")
	}
	if !fd3.IsRawJson() {
		t.Error("Json file should be recognised")
	}
	if fd3.HasEncData() {
		t.Error("file should not have key and salt")
	}
	resetTestFile(testFileName, content4)
	fd4, err4 := lib.NewFileData(testFileName)
	if err4 != nil {
		t.Error("err should be nil for file found")
	}
	if !fd4.IsRawJson() {
		t.Error("Json file should be recognised")
	}
	if fd4.HasEncData() {
		t.Error("file should not have key and salt")
	}
	resetTestFile(testFileName, content5)
	fd5, err5 := lib.NewFileData(testFileName)
	if err5 != nil {
		t.Error("err should be nil for file found")
	}
	if fd5.IsRawJson() {
		t.Error("Json file should NOT be recognised")
	}
	if fd5.HasEncData() {
		t.Error("file should not have key and salt")
	}
}

func TestConstructNotFound(t *testing.T) {
	_, err := lib.NewFileData("NotFound.txt")
	if err == nil {
		t.Error("err should not be nil for file not found")
	}
	_, err = lib.NewFileDataEnc("NotFoundEnc.txt", password, salt, iterations)
	if err == nil {
		t.Error("err should be nil for enc file found")
	}
}

func TestConstruct(t *testing.T) {
	resetTestFile(testFileName, content1)
	fd, err := lib.NewFileData(testFileName)
	if err != nil {
		t.Error("err should be nil for file found")
	}
	if fd.GetFileName() != testFileName {
		t.Errorf("File name is incorrect")
	}
	if string(fd.GetContent()) != string(content1) {
		t.Errorf("Content is incorrect")
	}
	if !fd.IsRawJson() {
		t.Errorf("Json file should be recognised")
	}
	if fd.HasEncData() {
		t.Error("file should not have key and salt")
	}
}

func TestStoreContent(t *testing.T) {
	resetTestFile(testFileName, content1)
	fd1, err1 := lib.NewFileData(testFileName)
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
	fd1.StoreContent()
	if !fd1.IsRawJson() {
		t.Error("Json file should be recognised")
	}
	if fd1.HasEncData() {
		t.Error("file should not have key and salt")
	}

	fd2, err2 := lib.NewFileData(testFileName)
	if err2 != nil {
		t.Error("err should be null for file found")
	}
	if string(fd2.GetContent()) != string(content2) {
		t.Errorf("Store failed to change content:'%s' != '%s'", string(fd2.GetContent()), string(content2))
	}
	if fd2.GetFileName() != testFileName {
		t.Error("Stored File name is incorrect")
	}
	if !fd2.IsRawJson() {
		t.Error("Json file should be recognised")
	}
	if fd2.HasEncData() {
		t.Error("file should not have key and salt")
	}

}

func resetTestFile(fileName string, content []byte) {
	err := ioutil.WriteFile(fileName, content, 0644)
	if err != nil {
		log.Fatalf("Failed to reset test file: %s. %s\n", fileName, err)
	}
}

func echoTestFile(fileName string) {
	dat, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatalf("Read file failed '%s': %s", fileName, err)
	}
	fmt.Print("Static data for encrypted test file. Copy and paste, replacing var encCon. \n{")
	for _, v := range dat {
		fmt.Printf("%d,", v)
	}
	fmt.Println("}")
}
