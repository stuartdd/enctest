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
	"strconv"
	"strings"
	"time"

	"github.com/stuartdd2/JsonParser4go/parser"

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

type BackupFileDef struct {
	ref        string
	path       string
	sep        string
	pre        string
	mask       string
	post       string
	max        int
	tempSource string
}

type FileData struct {
	fileName      string
	postDataUrl   string
	getDataUrl    string
	backupFileDef *BackupFileDef
	key           []byte
	content       []byte
	isEmpty       bool
	isEncrypted   bool
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

func NewBackupFileDef(path, sep, pre, mask, post string, max int64) *BackupFileDef {
	return &BackupFileDef{path: path, sep: sep, pre: pre, mask: mask, post: post, max: int(max)}
}

func (r *BackupFileDef) CleanFiles() error {
	list2, _ := r.ListFiles()
	if len(list2) > int(r.max) {
		for i := 0; i < len(list2)-r.max; i++ {
			fmt.Printf("Remove file %s\n", list2[i])
			err := os.Remove(fmt.Sprintf("%s%s%s", r.path, r.sep, list2[i]))
			if err != nil {
				return fmt.Errorf("could not delete backup file %s%s%s.\nError: %s", r.path, r.sep, list2[i], err.Error())
			}
		}
	}
	return nil
}

func (r *BackupFileDef) SetTempSource(temp string) {
	r.tempSource = temp
}

func (r *BackupFileDef) GetTempSource() string {
	if r.tempSource == "" {
		return ""
	}
	return fmt.Sprintf("%s%s%s", r.path, r.sep, r.tempSource)
}

func (r *BackupFileDef) ListFiles() ([]string, error) {
	list, err := ioutil.ReadDir(r.path)
	if err != nil {
		return nil, fmt.Errorf("could not read contents of backup path:'%s'.\nError: %s", r.path, err.Error())
	}
	list2 := make([]string, 0)
	for _, v := range list {
		if strings.HasPrefix(v.Name(), r.pre) && strings.HasSuffix(v.Name(), r.post) {
			list2 = append(list2, v.Name())
		}
	}
	return list2, nil
}

func (r *BackupFileDef) Init(dataName string) error {
	r.ref = dataName
	r.path = strings.TrimSpace(r.path)
	r.sep = strings.TrimSpace(r.sep)
	r.pre = strings.TrimSpace(r.pre)
	r.mask = strings.TrimSpace(r.mask)
	r.post = strings.TrimSpace(r.post)

	if r.path == "" {
		return nil
	}
	if !FileExists(r.path) {
		return fmt.Errorf("backup definition field '%s.path':'%s'. Path does not exist", dataName, r.path)
	}
	if r.sep == "" || r.pre == "" || r.mask == "" || r.post == "" {
		return fmt.Errorf("backup definition fields 'sep', 'pre', 'mask' and 'post' are ALL required for correct backup")
	}
	if r.max <= 0 {
		return fmt.Errorf("backup definition fields '%s.max':'%d' cannot be less than 1", dataName, r.max)
	}
	_, err := r.ListFiles()
	if err != nil {
		return fmt.Errorf("backup definition: %s", err.Error())
	}
	return nil
}

func (r *BackupFileDef) IsRequired() bool {
	return r.path != ""
}

func (r *BackupFileDef) ComposeFullName() string {
	return fmt.Sprintf("%s%s%s", r.path, r.sep, r.ComposeFileName())
}

func (r *BackupFileDef) ComposeTemplateName() string {
	return fmt.Sprintf("%s%c%s%s%s", r.path, os.PathSeparator, r.pre, r.mask, r.post)
}

func (r *BackupFileDef) ComposeFileName() string {
	mfn := r.mask
	if strings.Contains(mfn, "%d") {
		mfn = strings.ReplaceAll(mfn, "%d", time.Now().Format("2006-01-02"))
	}
	if strings.Contains(mfn, "%h") {
		mfn = strings.ReplaceAll(mfn, "%h", strPad2(time.Now().Hour()))
	}
	if strings.Contains(mfn, "%m") {
		mfn = strings.ReplaceAll(mfn, "%m", strPad2(time.Now().Minute()))
	}
	if strings.Contains(mfn, "%s") {
		mfn = strings.ReplaceAll(mfn, "%s", strPad2(time.Now().Second()))
	}
	return fmt.Sprintf("%s%s%s", r.pre, mfn, r.post)
}

func NewFileData(fName string, backupFileDef *BackupFileDef, getUrl string, postUrl string) (*FileData, error) {
	if backupFileDef == nil {
		backupFileDef = NewBackupFileDef("", "", "", "", "", 1)
	}
	fd := FileData{fileName: fName, backupFileDef: backupFileDef, getDataUrl: getUrl, postDataUrl: postUrl, content: make([]byte, 0), isEmpty: true, isEncrypted: false, key: make([]byte, 0)}
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
	var err error
	if r.postDataUrl != "" {
		_, err = parser.PostJsonBytes(fmt.Sprintf("%s/%s", r.postDataUrl, r.fileName), data)
	} else {
		err = ioutil.WriteFile(r.fileName, data, 0644)
	}
	if r.backupFileDef.IsRequired() {
		mfn := r.backupFileDef.ComposeFullName()
		err2 := ioutil.WriteFile(mfn, data, 0644)
		if err2 != nil {
			fmt.Printf("Error writing Backup file [%s]. Error message: %s\n", mfn, err2)
		}
		err3 := r.backupFileDef.CleanFiles()
		if err3 != nil {
			fmt.Printf("Error deleting Backup file [%s]. Error message: %s\n", mfn, err3)
		}
	}
	return err
}

func (r *FileData) loadData() error {
	var dat []byte
	var err error
	if r.backupFileDef.GetTempSource() != "" {
		dat, err = ioutil.ReadFile(r.backupFileDef.GetTempSource())
	} else {
		if r.getDataUrl != "" {
			dat, err = parser.GetJson(fmt.Sprintf("%s/%s", r.getDataUrl, r.fileName))
		} else {
			dat, err = ioutil.ReadFile(r.fileName)
		}
	}
	if err != nil {
		return err
	}
	r.SetContent(dat)
	r.isEncrypted = !r.IsRawJson()
	return nil
}

func strPad2(i int) string {
	if i < 10 {
		return "0" + strconv.Itoa(i)
	}
	return strconv.Itoa(i)
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
