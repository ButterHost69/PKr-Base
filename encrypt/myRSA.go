package encrypt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"log"
	"path/filepath"

	"os"
)

var (
	KEY_SIZE = 4096
)

var (
	PRIVATE_KEYS_PATH = filepath.Join("tmp", "mykeys", "privatekey.pem")
)


func GenerateRSAKeys() (*rsa.PrivateKey, *rsa.PublicKey) {
	privateKey, err := rsa.GenerateKey(rand.Reader, KEY_SIZE)
	if err != nil {
		log.Println(" ~ Could not create Keys")
		return nil, nil
	}

	return privateKey, &privateKey.PublicKey
}

func ParsePrivateKeyToBytes(pkey *rsa.PrivateKey) []byte {
	pkeyBytes := x509.MarshalPKCS1PrivateKey(pkey)
	privatekey_pem_block := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: pkeyBytes,
		},
	)

	return privatekey_pem_block

}

func ParsePublicKeyToBytes(pbkey *rsa.PublicKey) []byte {
	pbkeyBytes := x509.MarshalPKCS1PublicKey(pbkey)
	publickey_pem_block := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: pbkeyBytes,
		},
	)

	return publickey_pem_block
}

func StorePrivateKeyInFile(filepath string, pkey *rsa.PrivateKey) error {
	private_pem_key := ParsePrivateKeyToBytes(pkey)

	if private_pem_key == nil {
		return errors.New("~ Private Key Could Not Be Converted To []Byte")
	}

	return os.WriteFile(filepath, private_pem_key, 0666)
}

func StorePublicKeyInFile(filepath string, pbkey *rsa.PublicKey) error {
	public_pem_key := ParsePublicKeyToBytes(pbkey)

	if public_pem_key == nil {
		return errors.New("~ Private Key Could Not Be Converted To []Byte")
	}

	return os.WriteFile(filepath, public_pem_key, 0666)
}

func DecryptData(cipherText string) (string, error) {
	block, _ := pem.Decode([]byte(loadPrivateKey()))
	if block == nil {
		log.Println("error in parsing the pem Block...")
		log.Println("Pls check if the provided Private Key is correct")
		return "", errors.New("error in retrieving the Pem Block")
	}

	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		log.Println("error in parsing the private key...")
		return "", err
	}

	hash := sha256.New()

	baseDecoded, _ := base64.StdEncoding.DecodeString(cipherText)
	label := []byte("")
	plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, privKey, []byte(baseDecoded), label)
	if err != nil {
		log.Println("error in decrypting cipher text...", err)
		return "", err
	}

	return string(plaintext), err
}

func EncryptData(data string, publicPemBock string) (string, error) {
	block, _ := pem.Decode([]byte(publicPemBock))
	if block == nil {
		log.Println("error in parsing the pem Block...")
		log.Println("Pls check if the provided Public Key is correct")
		return "", errors.New("error in retrieving the Pem Block")
	}

	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		log.Println("error in parsing the Public key...")
		return "", err
	}

	label := []byte("")
	hash := sha256.New()

	result, err := rsa.EncryptOAEP(hash, rand.Reader, publicKey, []byte(data), label)
	if err != nil {
		log.Println("error in encypting the messgae...")
		log.Println(err.Error())
		return "", err
	}

	// log.Printf("[Log] : Encryped Message: %s\n", string(result))

	base64Encrypted := base64.StdEncoding.EncodeToString(result)
	return base64Encrypted, nil
	// return string(result), nil
}

func GetPublicKey(path string) string {
	// file, err := os.OpenFile(KEYS_PATH, os.O_RDONLY, 0444)
	// if err != nil {
	// 	log.Println("error in loading public key")
	// 	log.Println(err.Error())

	// 	return ""
	// }
	key, err := os.ReadFile(path)
	if err != nil {
		log.Println("error in reading public key")
		log.Println(err.Error())

		return ""
	}
	return string(key)
}

func loadPrivateKey() string {
	key, err := os.ReadFile(PRIVATE_KEYS_PATH)
	if err != nil {
		log.Println("error in reading public key")
		log.Println(err.Error())

		return ""
	}
	return string(key)
}
