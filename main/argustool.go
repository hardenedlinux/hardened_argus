package main

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"fmt"
	"hardened_argus/argus"
	"io"
	"io/ioutil"
	"os"
)

func showUseage() {
	fmt.Println("-g publicFile privateFile ;generic public-private ECDSA key pair file")
	fmt.Println("-s privateFile srcFile ; sign file")
	fmt.Println("-v publickFile srcFile ; verify file")
}

func main() {
	if len(os.Args) != 4 {
		showUseage()
		return
	}
	if os.Args[1] == "-g" {
		err := argus.NewKey(os.Args[2], os.Args[3])
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("success")
		return
	}
	if os.Args[1] == "-v" {
		buf, err := ioutil.ReadFile(os.Args[2])
		if err != nil {
			fmt.Println(err)
			return
		}
		key, err := argus.StringToPublicKey(string(buf))
		if err != nil {
			fmt.Println(err)
			return
		}
		ecdsaKey := key.(*ecdsa.PublicKey)
		buf, err = ioutil.ReadFile(os.Args[3] + ".sig")
		if err != nil {
			fmt.Println(err)
			return
		}

		sigText := string(buf)

		fscript, err := os.Open(os.Args[3])
		if err != nil {
			fmt.Println(err)
			return
		}
		defer fscript.Close()

		chse := sha256.New()
		_, err = io.Copy(chse, fscript)
		if err != nil {
			fmt.Println(err)
			return
		}
		chs := chse.Sum(nil)
		err = argus.Verify(chs, ecdsaKey, sigText)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("verify ok")
		return
	}
	if os.Args[1] == "-s" {
		buf, err := ioutil.ReadFile(os.Args[2])
		if err != nil {
			fmt.Println(err)
			return
		}
		key, err := argus.StringToPrivateKey(string(buf))
		if err != nil {
			fmt.Println(err)
			return
		}
		ecdsaKey := key.(*ecdsa.PrivateKey)

		f, err := os.Open(os.Args[3])
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()

		chse := sha256.New()
		_, err = io.Copy(chse, f)
		if err != nil {
			fmt.Println(err)
			return
		}
		chs := chse.Sum(nil)
		sig, err := argus.Sign(chs, ecdsaKey)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = ioutil.WriteFile(os.Args[3]+".sig", sig, os.ModePerm)
		if err != nil {
			fmt.Println(err)
			return
		}
		return
	}
	showUseage()
}
