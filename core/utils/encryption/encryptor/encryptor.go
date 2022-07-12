package encryptor

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"

	"github.com/kamalshkeir/kago/core/settings"
	"github.com/kamalshkeir/kago/core/utils"
	"golang.org/x/crypto/scrypt"
)


func Encrypt(data string) (string,error) {
    keyEnv := settings.GlobalConfig.Secret
    if keyEnv == "" {
        keyEnv = utils.GenerateRandomString(32)
        settings.GlobalConfig.Secret = keyEnv
    }
    keyByte, salt, err := deriveKey([]byte(keyEnv), nil)
    if err != nil {
        return "",err
    }

    blockCipher, err := aes.NewCipher([]byte(keyByte))
    if err != nil {
        return "",err
    }

    gcm, err := cipher.NewGCM(blockCipher)
    if err != nil {
        return "",err
    }

    nonce := make([]byte, gcm.NonceSize())
    if _, err = rand.Read(nonce); err != nil {
        return "",err
    }

    ciphertext := gcm.Seal(nonce, nonce, []byte(data), nil)

    ciphertext = append(ciphertext, salt...)
    return hex.EncodeToString(ciphertext),nil
}

func Decrypt(data string) (string,error) {
    keyEnv := settings.GlobalConfig.Secret
    if keyEnv == "" {
        keyEnv = utils.GenerateRandomString(32)
        settings.GlobalConfig.Secret = keyEnv
    }
    var salt []byte
    dataByte,_ := hex.DecodeString(data)
    salt, dataByte = dataByte[len(dataByte)-32:], dataByte[:len(dataByte)-32]

    key, _, err := deriveKey([]byte(keyEnv), salt)
    if err != nil {
        return "",err
    }

    blockCipher, err := aes.NewCipher(key)
    if err != nil {
        return "",err
    }

    gcm, err := cipher.NewGCM(blockCipher)
    if err != nil {
        return "",err
    }

    nonce, ciphertext := dataByte[:gcm.NonceSize()], dataByte[gcm.NonceSize():]

    plaintext, err := gcm.Open(nil, []byte(nonce), []byte(ciphertext), nil)
    if err != nil {
        return "",err
    }

    return string(plaintext),nil
}

func deriveKey(password, salt []byte) ([]byte, []byte, error) {
    if salt == nil {
        salt = make([]byte, 32)
        if _, err := rand.Read(salt); err != nil {
            return nil, nil, err
        }
    }

    key, err := scrypt.Key(password, salt, 1<<14, 8, 1, 32)
    if err != nil {
        return nil, nil, err
    }

    return key, salt, nil
}