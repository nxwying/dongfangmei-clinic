package license

import (
	"crypto"
	"crypto/rand"
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

// License data structure
type License struct {
	MachineCode string   `json:"machine_code"`
	ExpiresAt   string   `json:"expires_at"`
	Features    []string `json:"features"`
	Customer    string   `json:"customer"`
	IssuedAt    string   `json:"issued_at"`
	Version     string   `json:"version"`
	Signature   string   `json:"signature"`
}

// Status is the public license info returned to the frontend
type Status struct {
	Activated   bool     `json:"activated"`
	MachineCode string   `json:"machine_code"`
	Customer    string   `json:"customer,omitempty"`
	ExpiresAt   string   `json:"expires_at,omitempty"`
	Features    []string `json:"features,omitempty"`
	IsExpired   bool     `json:"is_expired"`
	DaysLeft    int      `json:"days_left"`
}

// Development PEM public key — REPLACE with your own in production!
const publicKeyPEM = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAteYr7ZWH0ZI+Rusf1UoW
os2cT/j/iosvzfts6w4dZdvCYf3J1j9nDYZeLgrKexOASmJL8frnlRVg6oXljaPT
qLzzwJjSbTI93wzxmVE+4zcwWsHiMsq4uonfEsJe1vS74/bYVgLJHO2c4wPgEuu/
yEuN6P9KwLh110Ef9krSvb8DnzZ7CVjrh6EbR8/JX/bsErUs5ZBfDfUrwpw9cGjy
ecc3/dD6Yp/7KGttpthkhzH5qheqXzbjITJvP3aYXKldl49xHjFIn0BbUnMp3mPw
LR5+EeE8wbvqOJUEC0QrJq3aZvAjpO8UCOI2yTab+kMEkjNQIXT5mlxoIKLdIxVm
9wIDAQAB
-----END PUBLIC KEY-----`

var (
	licenseStore *License
	publicKey    *rsa.PublicKey
)

const licenseFilePath = "data/license.json"

func init() {
	// Parse the embedded public key
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		panic("failed to parse public key PEM")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic("failed to parse public key: " + err.Error())
	}
	var ok bool
	publicKey, ok = pub.(*rsa.PublicKey)
	if !ok {
		panic("public key is not RSA")
	}
}

// GenerateMachineCode creates a unique machine fingerprint
func GenerateMachineCode() (string, error) {
	parts := []string{}

	// 1. MAC addresses (sorted for consistency)
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range interfaces {
			if iface.Flags&net.FlagUp != 0 && len(iface.HardwareAddr) > 0 {
				parts = append(parts, iface.HardwareAddr.String())
			}
		}
	}
	sort.Strings(parts)

	// 2. Hostname
	hostname, err := os.Hostname()
	if err == nil {
		parts = append(parts, hostname)
	}

	if len(parts) == 0 {
		return "", fmt.Errorf("无法生成机器码")
	}

	raw := strings.Join(parts, "|")
	hash := sha256.Sum256([]byte(raw))

	// Format as XXXX-XXXX-XXXX-XXXX
	h := fmt.Sprintf("%X", hash[:8])
	return fmt.Sprintf("%s-%s-%s-%s",
		h[:4], h[4:8], h[8:12], h[12:16]), nil
}

// LoadLicense reads and verifies the license file from disk
func LoadLicense() (*License, error) {
	data, err := os.ReadFile(licenseFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // no license file = not activated
		}
		return nil, fmt.Errorf("读取授权文件失败: %w", err)
	}
	return ParseAndVerify(data)
}

// ParseAndVerify parses and cryptographically verifies a license
func ParseAndVerify(data []byte) (*License, error) {
	lic := &License{}
	if err := json.Unmarshal(data, lic); err != nil {
		return nil, fmt.Errorf("授权文件格式错误")
	}

	if lic.Signature == "" {
		return nil, fmt.Errorf("授权文件缺少签名")
	}

	// Verify RSA signature
	sig, err := base64.StdEncoding.DecodeString(lic.Signature)
	if err != nil {
		return nil, fmt.Errorf("签名格式错误")
	}

	// Build the signed payload (all fields except signature)
	payload := struct {
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
	}
	payloadBytes, _ := json.Marshal(payload)
	hashed := sha256.Sum256(payloadBytes)

	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], sig)
	if err != nil {
		return nil, fmt.Errorf("授权签名验证失败，授权文件可能已被篡改")
	}

	// Verify machine code matches
	currentCode, _ := GenerateMachineCode()
	if lic.MachineCode != "" && lic.MachineCode != currentCode {
		return nil, fmt.Errorf("授权文件绑定到其他电脑，与本机不匹配")
	}

	// Check expiry
	if lic.ExpiresAt != "" {
		expiry, err := time.Parse("2006-01-02", lic.ExpiresAt)
		if err == nil && time.Now().After(expiry) {
			return nil, fmt.Errorf("授权已过期（%s）", lic.ExpiresAt)
		}
	}

	licenseStore = lic
	return lic, nil
}

// SaveLicense saves an uploaded license to disk
func SaveLicense(data []byte) (*License, error) {
	lic, err := ParseAndVerify(data)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(licenseFilePath, data, 0644); err != nil {
		return nil, fmt.Errorf("保存授权文件失败: %w", err)
	}
	licenseStore = lic
	return lic, nil
}

// GetStatus returns the current public license status
func GetStatus() (*Status, error) {
	code, err := GenerateMachineCode()
	if err != nil {
		return nil, err
	}

	status := &Status{
		Activated:   licenseStore != nil,
		MachineCode: code,
	}

	if licenseStore != nil {
		status.Customer = licenseStore.Customer
		status.ExpiresAt = licenseStore.ExpiresAt
		status.Features = licenseStore.Features

		if licenseStore.ExpiresAt != "" {
			expiry, err := time.Parse("2006-01-02", licenseStore.ExpiresAt)
			if err == nil {
				now := time.Now()
				if now.After(expiry) {
					status.IsExpired = true
				} else {
					status.DaysLeft = int(expiry.Sub(now).Hours() / 24)
				}
			}
		}
	}

	return status, nil
}

// SignLicense creates a signed license (seller side)
func SignLicense(machineCode, expiresAt, customer string, features []string, privateKeyPEM string) ([]byte, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("无法解析私钥")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("私钥解析失败: %w", err)
	}

	lic := License{
		MachineCode: machineCode,
		ExpiresAt:   expiresAt,
		Features:    features,
		Customer:    customer,
		IssuedAt:    time.Now().Format("2006-01-02"),
		Version:     "1.0",
	}

	// Sign
	payload := struct {
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
	}
	payloadBytes, _ := json.Marshal(payload)
	hashed := sha256.Sum256(payloadBytes)

	sig, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hashed[:])
	if err != nil {
		return nil, fmt.Errorf("签名失败: %w", err)
	}

	lic.Signature = base64.StdEncoding.EncodeToString(sig)
	return json.MarshalIndent(lic, "", "  ")
}

// IsFeatureGranted checks if a specific feature is in the license
func IsFeatureGranted(feature string) bool {
	if licenseStore == nil {
		return false
	}
	for _, f := range licenseStore.Features {
		if f == feature || f == "all" {
			return true
		}
	}
	return false
}

// IsActivated returns whether a valid license is loaded
func IsActivated() bool {
	return licenseStore != nil
}
