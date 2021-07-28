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
}

func NewFileData(fName string) (*FileData, error) {
	fd := FileData{fileName: fName, content: make([]byte, 0), key: make([]byte, 0), salt: make([]byte, 0)}
	return &fd, fd.load()
}

func NewFileDataEnc(fName string, encKey, encSalt []byte) (*FileData, error) {
	fd := FileData{fileName: fName, content: make([]byte, 0), key: encKey, salt: encSalt}
	return &fd, fd.loadEnc(encKey, encSalt)
}

func (r *FileData) StoreContent() error {
	if r.HasEncData() {
		return r.StoreContentWithKey(r.key, r.salt)
	} else {
		return r.storeData(r.content)
	}
}

func (r *FileData) StoreContentWithKey(encKey, encSalt []byte) error {
	cont, err := encrypt(encKey, encSalt, r.content)
	if err != nil {
		return err
	}
	err = r.storeData(cont)
	r.key = encKey
	r.salt = encSalt
	return err
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

func (r *FileData) SetContent(data []byte) {
	r.content = data
}

func (r *FileData) loadEnc(key, salt []byte) error {
	if r.HasEncData() {
		err := r.load()
		if err != nil {
			return err
		}
		r.content, err = decrypt(key, salt, r.content)
		return err
	}
	return errors.New("salt or key are not defined")
}

func (r *FileData) storeData(data []byte) error {
	err := ioutil.WriteFile(r.fileName, data, 0644)
	return err
}

func (r *FileData) load() error {
	dat, err := ioutil.ReadFile(r.fileName)
	if err != nil {
		return err
	}
	r.content = dat
	return nil
}

func decrypt(key, salt, data []byte) ([]byte, error) {

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

func encrypt(key, salt, data []byte) ([]byte, error) {

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
	ciphertext = append(ciphertext, salt...)
	return ciphertext, nil
}

func deriveKey(password, salt []byte) ([]byte, error) {
	if len(salt) == 0 {
		return nil, errors.New("deriveKey: salt was not provided")
	}

	key, err := scrypt.Key(password, salt, 1048576, 8, 1, 32)
	if err != nil {
		return nil, err
	}

	return key, nil
}
