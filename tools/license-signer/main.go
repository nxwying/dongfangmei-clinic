package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

type License struct {
	MachineCode string   `json:"machine_code"`
	ExpiresAt   string   `json:"expires_at"`
	Features    []string `json:"features"`
	Customer    string   `json:"customer"`
	IssuedAt    string   `json:"issued_at"`
	Version     string   `json:"version"`
	Signature   string   `json:"signature"`
}

func main() {
	machineCode := flag.String("machine", "", "客户机器码")
	expiresAt := flag.String("expire", "", "过期日期 2006-01-02（空=永久）")
	customer := flag.String("customer", "", "客户名称")
	features := flag.String("features", "all", "授权功能，逗号分隔")
	privKeyFile := flag.String("key", "tools/license-signer/dev-private.pem", "私钥路径")
	output := flag.String("output", "license.json", "输出路径")
	flag.Parse()

	if *machineCode == "" || *customer == "" {
		fmt.Println("错误：必须指定 -machine 和 -customer")
		fmt.Println("\n用法:")
		fmt.Printf("  %s -machine XXXX-XXXX-XXXX-XXXX -customer '诊所名称' -expire 2027-12-31\n", os.Args[0])
		os.Exit(1)
	}

	privPEM, err := os.ReadFile(*privKeyFile)
	if err != nil {
		fmt.Printf("错误：读取私钥失败: %v\n", err)
		os.Exit(1)
	}

	block, _ := pem.Decode(privPEM)
	if block == nil {
		fmt.Println("错误：私钥格式无效")
		os.Exit(1)
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		fmt.Printf("错误：私钥解析失败: %v\n", err)
		os.Exit(1)
	}

	lic := License{
		MachineCode: *machineCode,
		ExpiresAt:   *expiresAt,
		Features:    strings.Split(*features, ","),
		Customer:    *customer,
		IssuedAt:    time.Now().Format("2006-01-02"),
		Version:     "1.0",
	}

	// Sign
	payloadBytes, _ := json.Marshal(struct {
		MachineCode string   `json:"machine_code"`
		ExpiresAt   string   `json:"expires_at"`
		Features    []string `json:"features"`
		Customer    string   `json:"customer"`
		IssuedAt    string   `json:"issued_at"`
		Version     string   `json:"version"`
	}{
		MachineCode: lic.MachineCode,
		ExpiresAt:   lic.ExpiresAt,
		Features:    lic.Features,
		Customer:    lic.Customer,
		IssuedAt:    lic.IssuedAt,
		Version:     lic.Version,
	})
	hashed := sha256.Sum256(payloadBytes)
	sig, _ := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hashed[:])
	lic.Signature = base64.StdEncoding.EncodeToString(sig)

	out, _ := json.MarshalIndent(lic, "", "  ")
	os.WriteFile(*output, out, 0644)

	var pretty map[string]interface{}
	json.Unmarshal(out, &pretty)
	delete(pretty, "signature")
	p, _ := json.MarshalIndent(pretty, "  ", "  ")

	fmt.Println("✓ 授权文件已生成!")
	fmt.Printf("  输出: %s\n", *output)
	fmt.Printf("  客户: %s\n", *customer)
	fmt.Printf("  机器码: %s\n", *machineCode)
	fmt.Printf("  过期: %s\n", *expiresAt)
	fmt.Printf("  功能: %s\n", *features)
	fmt.Printf("\n授权内容（不含签名）:\n  %s\n", string(p))
	fmt.Println("\n将此文件发给客户，在软件 系统设置→授权管理 中导入即可激活。")
}
