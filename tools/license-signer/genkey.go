//go:build ignore
// +build ignore

package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func main() {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	// Save private key
	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	privFile, _ := os.Create("dev-private.pem")
	pem.Encode(privFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})
	privFile.Close()
	fmt.Println("✓ 私钥已生成: dev-private.pem")
	fmt.Println("  请妥善保管此文件，不要泄露！")

	// Save public key
	pubBytes, _ := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	pubFile, _ := os.Create("dev-public.pem")
	pem.Encode(pubFile, &pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	pubFile.Close()
	fmt.Println("✓ 公钥已生成: dev-public.pem")

	// Also print public key as Go constant
	pubPEM, _ := os.ReadFile("dev-public.pem")
	fmt.Println("\n=== 可用于嵌入代码的公钥常量为：")
	fmt.Printf("const publicKeyPEM = `%s`", string(pubPEM))
}
