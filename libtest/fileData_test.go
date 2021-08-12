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
	encCon       = []byte{232, 93, 255, 102, 85, 218, 198, 139, 140, 254, 64, 121, 214, 100, 181, 58, 105, 1, 165, 36, 120, 204, 40, 254, 26, 74, 40, 241, 250, 240, 213, 187, 9, 42, 60, 114, 65, 60, 24, 10, 5, 34, 175, 214, 22, 53, 86, 239, 208, 82, 127, 147, 41, 217, 25, 134, 50, 176, 77, 70, 92, 30, 99, 215, 193}
	testFileName = "TempTestData.json"
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
	resetTestFile(testFileName, content1)
	fd1, err1 := lib.NewFileData(testFileName)
	if err1 != nil {
		t.Errorf("err should be nil for file found. %s", err1)
	}
	if !fd1.IsRawJson() {
		t.Errorf("Content should not be JSON:'%s'", string(fd1.GetContent()))
	}
	fd1.StoreContentAsIs()
	fd2, err2 := lib.NewFileData(testFileName)
	if err2 != nil {
		t.Error("err should be nil for file found")
	}
	if string(fd2.GetContent()) != string(fd1.GetContent()) {
		t.Errorf("File content is incorrect %s != %s", string(fd2.GetContent()), string(content2))
	}
	if !fd2.IsRawJson() {
		t.Error("Should be decrypted!")
	}
	if fd2.HasEncData() {
		t.Error("file should NOT have key")
	}
	fd2.StoreContentEncrypted(password)
	fd3, err3 := lib.NewFileData(testFileName)
	if err3 != nil {
		t.Error("err should be nil for file found")
	}
	if fd3.IsRawJson() {
		t.Error("Should be encrypted!")
	}
	if !fd2.HasEncData() {
		t.Error("file should have key")
	}
	err := fd3.DecryptContents(password)
	if err != nil {
		t.Error("Failed to decrypt")
	}
	if string(fd3.GetContent()) != string(fd1.GetContent()) {
		t.Errorf("File content is incorrect %s != %s", string(fd2.GetContent()), string(content2))
	}
	resetTestFile(testFileName, content1)
	err = fd3.StoreContentAsIs()
	if err != nil {
		t.Error("Failed to store as is")
	}
	fd4, err4 := lib.NewFileData(testFileName)
	if err4 != nil {
		t.Error("err should be nil for file found")
	}
	if fd4.IsRawJson() {
		t.Error("Should be encrypted!")
	}
	if fd4.HasEncData() {
		t.Error("file should not have key")
	}
	err = fd4.DecryptContents(password)
	if err != nil {
		t.Error("Failed to decrypt")
	}
	if string(fd4.GetContent()) != string(fd1.GetContent()) {
		t.Errorf("File content is incorrect %s != %s", string(fd2.GetContent()), string(content2))
	}
	if !fd4.HasEncData() {
		t.Error("file should have key")
	}

	err = fd4.StoreContentUnEncrypted()
	if err != nil {
		t.Error("Failed to store")
	}
	if string(fd4.GetContent()) != string(fd1.GetContent()) {
		t.Errorf("File content is incorrect %s != %s", string(fd2.GetContent()), string(content2))
	}
	if fd4.HasEncData() {
		t.Error("file should not have key")
	}

	fd5, err5 := lib.NewFileData(testFileName)
	if err5 != nil {
		t.Error("err should be nil for file found")
	}
	if !fd5.IsRawJson() {
		t.Error("Should be decrypted!")
	}
	if fd5.HasEncData() {
		t.Error("file should not have key")
	}
	if string(fd5.GetContent()) != string(fd1.GetContent()) {
		t.Errorf("File content is incorrect %s != %s", string(fd2.GetContent()), string(content2))
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
	err1 = fd1.StoreContentEncrypted(password)
	if err1 != nil {
		t.Errorf("Store Enc failed. %s", err1)
	}
	if string(fd1.GetContent()) != string(content1) {
		t.Error("Content is incorrect")
	}
	if !fd1.IsRawJson() {
		t.Error("Content should still be json")
	}
	if !fd1.HasEncData() {
		t.Error("file should have key")
	}
	fd2, err2 := lib.NewFileData(testFileName)
	if err2 != nil {
		t.Error("err should be nil for file found")
	}
	if fd2.IsRawJson() {
		t.Error("Should be encrypted!")
	}
	err3 := fd2.DecryptContents(password)
	if err3 != nil {
		t.Error("Failed to decrypt")
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
	_, err = lib.NewFileData("TestDataTypes.json")
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
	fd1.StoreContentAsIs()
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
