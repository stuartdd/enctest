package lib

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io/ioutil"

	"golang.org/x/crypto/scrypt"
)

type FileData struct {
	fileName string
	salt     []byte
	key      []byte
	content  []byte
	isEmpty  bool
}

var (
	emptyDataJson = "{\"timeStamp\": \"Fri Jul 30 21:25:10 BST 2021\",\"groups\": {\"Empty\":{}}}"
	encIterations = 64
)

func NewFileData(fName string) (*FileData, error) {
	if fName == "" {
		fd := FileData{fileName: "", content: []byte(emptyDataJson), key: make([]byte, 0), salt: make([]byte, 0), isEmpty: true}
		return &fd, nil
	}
	fd := FileData{fileName: fName, content: make([]byte, 0), key: make([]byte, 0), salt: make([]byte, 0), isEmpty: false}
	return &fd, fd.loadData()
}

func (r *FileData) DecryptContents(encKey, encSalt []byte) error {
	if r.isEmpty {
		return errors.New("Cannot decrypt empty content data")
	}
	cont, err := decrypt(encKey, encSalt, r.content)
	if err != nil {
		return err
	}
	r.key = encKey
	r.salt = encSalt
	r.content = cont
	return nil
}

func (r *FileData) StoreContentEncrypted(encKey, encSalt []byte) error {
	cont, err := encrypt(encKey, encSalt, r.content)
	if err != nil {
		return err
	}
	err = r.storeData(cont)
	r.key = encKey
	r.salt = encSalt
	return err
}

func (r *FileData) StoreContentUnEncrypted() error {
	r.key = make([]byte, 0)
	r.salt = make([]byte, 0)
	return r.storeData(r.content)
}

func (r *FileData) StoreContent() error {
	if r.HasEncData() {
		return r.StoreContentEncrypted(r.key, r.salt)
	} else {
		return r.storeData(r.content)
	}
}

func (r *FileData) IsRawJson() bool {
	p := 0
	for p = 0; p < (len(r.content) - 1); p++ {
		if r.content[p] > 32 {
			break
		}
	}
	return (r.content[p] == '{') || (r.content[p] == '[')
}

func (r *FileData) HasEncData() bool {
	if len(r.key) == 0 {
		return false
	}
	if len(r.salt) == 0 {
		return false
	}
	return true
}

func (r *FileData) GetFileName() string {
	return r.fileName
}

func (r *FileData) GetContent() []byte {
	return r.content
}

func (r *FileData) IsEmpty() bool {
	return r.isEmpty
}

func (r *FileData) SetContent(data []byte) {
	r.content = data
}

func (r *FileData) SetContentString(content string) {
	r.isEmpty = false
	r.SetContent([]byte(content))
}

func (r *FileData) storeData(data []byte) error {
	return ioutil.WriteFile(r.fileName, data, 0644)
}

func (r *FileData) loadData() error {
	dat, err := ioutil.ReadFile(r.fileName)
	if err != nil {
		return err
	}
	r.content = dat
	return nil
}

func decrypt(key, salt []byte, data []byte) ([]byte, error) {

	key, err := deriveKey(key, salt)
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

	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func encrypt(key, salt []byte, data []byte) ([]byte, error) {

	key, err := deriveKey(key, salt)
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

	return ciphertext, nil
}

func deriveKey(key, salt []byte) ([]byte, error) {
	if len(salt) == 0 {
		return nil, errors.New("deriveKey: salt was not provided")
	}
	if len(key) == 0 {
		return nil, errors.New("deriveKey: key was not provided")
	}
	key, err := scrypt.Key(key, salt, 16384*encIterations, 8, 1, 32)
	if err != nil {
		return nil, err
	}

	return key, nil
}
