package license

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"time"
)

type License struct {
	MachineCode string `json:"machine_code"`
	ExpiresAt   string `json:"expires_at"`
	Customer    string `json:"customer"`
	IssuedAt    string `json:"issued_at"`
	Features    []string `json:"features"`
	Version     string   `json:"version"`
	Signature   string `json:"signature"`
}

type Status struct {
	Activated   bool     `json:"activated"`
	MachineCode string   `json:"machine_code"`
	Customer    string   `json:"customer,omitempty"`
	ExpiresAt   string   `json:"expires_at,omitempty"`
	IsExpired   bool     `json:"is_expired"`
	DaysLeft    int      `json:"days_left"`
}

const publicKeyPEM = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAteYr7ZWH0ZI+Rusf1UoW
os2cT/j/iosvzfts6w4dZdvCYf3J1j9nDYZeLgrKexOASmJL8frnlRVg6oXljaPT
qLzzwJjSbTI93wzxmVE+4zcwWsHiMsq4uonfEsJe1vS74/bYVgLJHO2c4wPgEuu/
yEuN6P9KwLh110Ef9krSvb8DnzZ7CVjrh6EbR8/JX/bsErUs5ZBfDfUrwpw9cGjy
ecc3/dD6Yp/7KGttpthkhzH5qheqXzbjITJvP3aYXKldl49xHjFIn0BbUnMp3mPw
LR5+EeE8wbvqOJUEC0QrJq3aZvAjpO8UCOI2yTab+kMEkjNQIXT5mlxoIKLdIxVm
9wIDAQAB
-----END PUBLIC KEY-----`

var publicKey *rsa.PublicKey
var licenseStore *License
const licenseFile = "data/license.json"

func init() {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		panic("invalid public key")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic("parse public key: " + err.Error())
	}
	var ok bool
	publicKey, ok = pub.(*rsa.PublicKey)
	if !ok {
		panic("not RSA public key")
	}
	// Try to load existing license
	data, err := os.ReadFile(licenseFile)
	if err == nil {
		if l, e := ParseAndVerify(data); e == nil {
			licenseStore = l
		}
	}
}

func GenerateMachineCode() (string, error) {
	var parts []string
	ifaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range ifaces {
			if iface.Flags&net.FlagUp != 0 && len(iface.HardwareAddr) > 0 {
				parts = append(parts, iface.HardwareAddr.String())
			}
		}
	}
	sort.Strings(parts)
	hostname, _ := os.Hostname()
	if hostname != "" {
		parts = append(parts, hostname)
	}
	if len(parts) == 0 {
		return "", fmt.Errorf("无法获取本机信息")
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	h := fmt.Sprintf("%X", sum[:8])
	return fmt.Sprintf("%s-%s-%s-%s", h[:4], h[4:8], h[8:12], h[12:16]), nil
}

func ParseAndVerify(data []byte) (*License, error) {
	lic := &License{}
	if err := json.Unmarshal(data, lic); err != nil {
		return nil, fmt.Errorf("文件格式错误")
	}
	if lic.Signature == "" {
		return nil, fmt.Errorf("缺少签名")
	}
	sig, err := base64.StdEncoding.DecodeString(lic.Signature)
	if err != nil {
		return nil, fmt.Errorf("签名格式错误")
	}
	payload, _ := json.Marshal(struct {
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
	hashed := sha256.Sum256(payload)
	if err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], sig); err != nil {
		return nil, fmt.Errorf("签名验证失败")
	}
	currentCode, _ := GenerateMachineCode()
	if lic.MachineCode != "" && lic.MachineCode != currentCode {
		return nil, fmt.Errorf("授权文件与本机不匹配")
	}
	if lic.ExpiresAt != "" {
		expiry, err := time.Parse("2006-01-02", lic.ExpiresAt)
		if err == nil && time.Now().After(expiry) {
			return nil, fmt.Errorf("授权已过期")
		}
	}
	return lic, nil
}

func SaveLicense(data []byte) (*License, error) {
	lic, err := ParseAndVerify(data)
	if err != nil {
		return nil, err
	}
	os.WriteFile(licenseFile, data, 0644)
	licenseStore = lic
	return lic, nil
}

func GetStatus() *Status {
	code, _ := GenerateMachineCode()
	s := &Status{Activated: licenseStore != nil, MachineCode: code}
	if licenseStore != nil {
		s.Customer = licenseStore.Customer
		s.ExpiresAt = licenseStore.ExpiresAt
		if licenseStore.ExpiresAt != "" {
			expiry, _ := time.Parse("2006-01-02", licenseStore.ExpiresAt)
			if !expiry.IsZero() {
				if time.Now().After(expiry) {
					s.IsExpired = true
				} else {
					s.DaysLeft = int(expiry.Sub(time.Now()).Hours() / 24)
				}
			}
		}
	}
	return s
}
