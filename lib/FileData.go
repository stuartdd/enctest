/*
 * Copyright (C) 2021 Stuart Davies (stuartdd)
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */
package lib

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/stuartdd/jsonParserGo/parser"

	"golang.org/x/crypto/scrypt"
)

type ImageRefType int

const (
	IMAGE_FILE_FOUND ImageRefType = iota
	IMAGE_NOT_FOUND
	IMAGE_URL
	IMAGE_GET_FAIL
	IMAGE_SUPPORTED
	IMAGE_NOT_SUPPORTED
)

type FileData struct {
	fileName    string
	postDataUrl string
	getDataUrl  string
	key         []byte
	content     []byte
	isEmpty     bool
	isEncrypted bool
}

/**
Changing any of this will prevent previous encryptions from being decrypted.
More encIterations will produce slower encryption and decryption times
Note that encIterations is multiplied by 1024
*/
var (
	supportedImageExtenstions = []string{".jpg", "png", ".svg"}
	encIterations             = 64                                         // Keep as power of 2.
	encSalt                   = []byte("SQhMXVt8rQED2MxHTHxmuZLMxdJz5DQI") // Keep as 32 randomly generated chars
)

func FileExists(fileName string) bool {
	if _, err := os.Stat(fileName); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

//
// Check the path to the image file.
// Do not do anything with fyne here. The lib module should NOT import fyne
//
func CheckImageFile(pathToImage string) (ImageRefType, string) {
	result := checkSupportedImage(pathToImage)
	if result == IMAGE_NOT_SUPPORTED {
		return result, ""
	}
	if FileExists(pathToImage) {
		return IMAGE_FILE_FOUND, ""
	}
	_, err := url.Parse(pathToImage)
	if err != nil {
		return IMAGE_NOT_FOUND, ""
	}
	err = checkImageUrl(pathToImage)
	if err != nil {
		return IMAGE_GET_FAIL, err.Error()
	}
	return IMAGE_URL, ""
}

func NewFileData(fName string, getUrl string, postUrl string) (*FileData, error) {
	fd := FileData{fileName: fName, getDataUrl: getUrl, postDataUrl: postUrl, content: make([]byte, 0), isEmpty: true, isEncrypted: false, key: make([]byte, 0)}
	return &fd, fd.loadData()
}

func (r *FileData) DecryptContents(encKey []byte) error {
	if r.IsEmpty() {
		return errors.New("cannot decrypt empty content data")
	}
	cont, err := decrypt(encKey, r.content)
	if err != nil {
		return err
	}
	r.key = encKey
	r.content = cont
	return nil
}

func (r *FileData) StoreContentEncrypted(encKey []byte, callbackWhenDone func()) error {
	cont, err := encrypt(encKey, r.content)

	if err != nil {
		return err
	}
	err = r.storeData(cont)
	if err != nil {
		return err
	}
	r.key = encKey
	callbackWhenDone()
	return nil
}

func (r *FileData) StoreContentUnEncrypted(callbackWhenDone func()) error {
	r.key = make([]byte, 0)
	err := r.storeData(r.content)
	if err != nil {
		return err
	}
	callbackWhenDone()
	return nil
}

func (r *FileData) StoreContentAsIs(callbackWhenDone func()) error {
	if r.HasEncData() {
		return r.StoreContentEncrypted(r.key, callbackWhenDone)
	} else {
		return r.StoreContentUnEncrypted(callbackWhenDone)
	}
}

func (r *FileData) RequiresDecryption() bool {
	return !r.IsRawJson()
}

func (r *FileData) IsRawJson() bool {
	p := 0
	for p = 0; p < (len(r.content) - 12); p++ {
		if r.content[p] == 't' { // t
			if r.content[p+1] == 'i' && // i
				r.content[p+2] == 'm' && // m
				r.content[p+3] == 'e' && // e
				r.content[p+4] == 'S' && // s
				r.content[p+5] == 't' && // t
				r.content[p+6] == 'a' && // a
				r.content[p+7] == 'm' && // m
				r.content[p+8] == 'p' && // p
				r.content[p+9] == '"' && // "
				r.content[p+10] == ':' { // :
				return true
			}
		}
	}
	return false
}

func (r *FileData) HasEncData() bool {
	return len(r.key) > 0
}

func (r *FileData) GetFileName() string {
	return r.fileName
}

func (r *FileData) GetContent() []byte {
	return r.content
}

func (r *FileData) GetContentString() string {
	return string(r.content)
}

func (r *FileData) IsEmpty() bool {
	return r.isEmpty
}

func (r *FileData) IsEncrypted() bool {
	return r.isEncrypted
}

func (r *FileData) SetContent(data []byte) {
	r.isEmpty = false
	r.content = data
}

func (r *FileData) storeData(data []byte) error {
	if r.postDataUrl != "" {
		_, err := parser.PostJsonBytes(fmt.Sprintf("%s/%s", r.postDataUrl, r.fileName), data)
		return err
	} else {
		return ioutil.WriteFile(r.fileName, data, 0644)
	}
}

func (r *FileData) loadData() error {
	var dat []byte
	var err error
	if r.getDataUrl != "" {
		dat, err = parser.GetJson(fmt.Sprintf("%s/%s", r.getDataUrl, r.fileName))
	} else {
		dat, err = ioutil.ReadFile(r.fileName)
	}
	if err != nil {
		return err
	}
	r.SetContent(dat)
	r.isEncrypted = !r.IsRawJson()
	return nil
}

func checkImageUrl(getUrl string) error {
	resp, err := http.Get(getUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to get data from server. Return Code: %d Url: %s", resp.StatusCode, getUrl)
	}
	return nil
}

func checkSupportedImage(fileName string) ImageRefType {
	fn := strings.ToLower(fileName)
	for _, ext := range supportedImageExtenstions {
		if strings.HasSuffix(fn, ext) {
			return IMAGE_SUPPORTED
		}
	}
	return IMAGE_NOT_SUPPORTED
}

func decrypt(key []byte, data []byte) ([]byte, error) {

	key, err := deriveKey(key)
	if err != nil {
		return nil, err
	}

	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return nil, err
	}

	dd, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return nil, err
	}

	nonce, ciphertext := dd[:gcm.NonceSize()], dd[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func encrypt(key, data []byte) ([]byte, error) {

	key, err := deriveKey(key)
	if err != nil {
		return nil, err
	}

	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	return []byte(base64.StdEncoding.EncodeToString(ciphertext)), nil
}

func deriveKey(key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, errors.New("deriveKey: key was not provided")
	}
	key, err := scrypt.Key(key, encSalt, 1024*encIterations, 8, 1, 32)
	if err != nil {
		return nil, err
	}

	return key, nil
}
