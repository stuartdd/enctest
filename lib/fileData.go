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
	content  []byte
}

func NewFileData(fName string) (*FileData, error) {
	fd := FileData{fileName: fName, content: make([]byte, 0), salt: make([]byte, 0)}
	return &fd, fd.load()
}

func NewFileDataEnc(fName string, encKey, encSalt []byte) (*FileData, error) {
	fd := FileData{fileName: fName, content: make([]byte, 0), salt: encSalt}
	return &fd, fd.loadEnc(encKey, encSalt)
}

func (r *FileData) StoreEnc(key []byte) error {
	cont, err := encrypt(key, r.salt, r.content)
	if err != nil {
		return err
	}
	err = r.storeData(cont)
	return err
}

func (r *FileData) Store() error {
	return r.storeData(r.content)
}

func (r *FileData) loadEnc(key, salt []byte) error {
	err := r.load()
	if err != nil {
		return err
	}
	r.content, err = decrypt(key, salt, r.content)
	return err
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

func (r *FileData) IsEncrypted() bool {
	return len(r.salt) > 0
}

func (r *FileData) GetFileName() string {
	return r.fileName
}

func (r *FileData) GetContent() []byte {
	return r.content
}

func (r *FileData) GetSalt() []byte {
	return r.salt
}

func (r *FileData) SetContent(data []byte) {
	r.content = data
}

func (r *FileData) SetSalt(data []byte) {
	r.salt = data
}

func decrypt(key, salt, data []byte) ([]byte, error) {

	key, _, err := deriveKey(key, salt)
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
	key, salt, err := deriveKey(key, nil)
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

func deriveKey(password, salt []byte) ([]byte, []byte, error) {
	if salt == nil {
		return nil, nil, errors.New("Salt was not provided")
	}

	key, err := scrypt.Key(password, salt, 1048576, 8, 1, 32)
	if err != nil {
		return nil, nil, err
	}

	return key, salt, nil
}
