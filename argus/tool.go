package argus

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/big"
	"os"
)

type signature struct {
	R *big.Int `json:"r"`
	S *big.Int `json:"s"`
}

func Sign(hash []byte, private *ecdsa.PrivateKey) ([]byte, error) {
	r, s, err := ecdsa.Sign(rand.Reader, private, hash)
	if err != nil {
		return nil, err
	}
	sr := &signature{
		R: r,
		S: s,
	}
	buf, err := json.Marshal(sr)
	if err != nil {
		return nil, err
	}
	return buf, err
}

func Verify(hash []byte, public *ecdsa.PublicKey, sig string) error {
	sr := &signature{}
	err := json.Unmarshal([]byte(sig), sr)
	if err != nil {
		return err
	}
	if ecdsa.Verify(public, hash, sr.R, sr.S) {
		return nil
	}
	return errors.New("Verify Failed")
}

//StringToPublicKey convert string to rsa or ecdsa public key
func StringToPublicKey(text string) (interface{}, error) {
	buf, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return nil, err
	}
	return x509.ParsePKIXPublicKey(buf)
}

//PublicKeyToString convert rsa or ecdsa public key to string
func PublicKeyToString(pubk interface{}) (string, error) {
	buf, err := x509.MarshalPKIXPublicKey(pubk)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf), nil
}

//StringToPrivateKey convert string to rsa or ecdsa public key
func StringToPrivateKey(text string) (interface{}, error) {
	buf, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return nil, err
	}
	return x509.ParsePKCS8PrivateKey(buf)
}

//PrivateKeyToString convert rsa or ecdsa public key to string
func PrivateKeyToString(pubk interface{}) (string, error) {
	buf, err := x509.MarshalPKCS8PrivateKey(pubk)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf), nil
}

func NewKey(pubFile string, priFile string) error {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}
	publicKey := privateKey.PublicKey
	prvBuf, err := PrivateKeyToString(privateKey)
	if err != nil {
		return err
	}
	pubBuf, err := PublicKeyToString(&publicKey)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(pubFile, []byte(pubBuf), os.ModePerm)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(priFile, []byte(prvBuf), os.ModePerm)
}
