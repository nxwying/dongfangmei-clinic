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
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

//go:embed dev-private.pem
//go:embed templates/index.html
var assets embed.FS

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
	port := "9090"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/generate", handleGenerate)
	http.HandleFunc("/assets/", func(w http.ResponseWriter, r *http.Request) {
		http.FileServer(http.FS(assets)).ServeHTTP(w, r)
	})

	fmt.Printf("\n  ╔═══════════════════════════════════════════╗\n")
	fmt.Printf("  ║   东芳美诊所 - 授权文件生成工具 v2.0      ║\n")
	fmt.Printf("  ╠═══════════════════════════════════════════╣\n")
	fmt.Printf("  ║  访问地址: http://localhost:%s          ║\n", port)
	fmt.Printf("  ╚═══════════════════════════════════════════╝\n\n")

	url := fmt.Sprintf("http://localhost:%s", port)
	go func() {
		time.Sleep(500 * time.Millisecond)
		openBrowser(url)
	}()

	http.ListenAndServe(":"+port, nil)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	tmpl, err := template.ParseFS(assets, "templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	tmpl.Execute(w, nil)
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	machineCode := strings.TrimSpace(r.FormValue("machine_code"))
	customer := strings.TrimSpace(r.FormValue("customer"))
	expiresAt := strings.TrimSpace(r.FormValue("expires_at"))
	features := strings.TrimSpace(r.FormValue("features"))
	if features == "" {
		features = "all"
	}

	if machineCode == "" || customer == "" {
		writeJSON(w, 400, map[string]string{"error": "机器码和客户名称为必填项"})
		return
	}

	priv := loadPrivateKey()
	if priv == nil {
		writeJSON(w, 500, map[string]string{"error": "私钥加载失败"})
		return
	}

	lic := License{
		MachineCode: machineCode,
		ExpiresAt:   expiresAt,
		Features:    strings.Split(features, ","),
		Customer:    customer,
		IssuedAt:    time.Now().Format("2006-01-02"),
		Version:     "1.0",
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
	sig, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hashed[:])
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "签名失败"})
		return
	}
	lic.Signature = base64.StdEncoding.EncodeToString(sig)

	out, _ := json.MarshalIndent(lic, "", "  ")

	if r.FormValue("download") == "1" {
		filename := fmt.Sprintf("license_%s.json", strings.ReplaceAll(customer, " ", "_"))
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		w.Write(out)
		return
	}

	writeJSON(w, 200, map[string]interface{}{
		"success":  true,
		"message":  "授权文件已生成",
		"customer": customer,
		"expires":  expiresAt,
		"content":  string(out),
	})
}

func loadPrivateKey() *rsa.PrivateKey {
	data, err := assets.ReadFile("dev-private.pem")
	if err != nil {
		return nil
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil
	}
	return priv
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func openBrowser(url string) {
	switch {
	case runtime.GOOS == "windows":
		execCmd("rundll32", "url.dll,FileProtocolHandler", url)
	case runtime.GOOS == "darwin":
		execCmd("open", url)
	default:
		execCmd("xdg-open", url)
	}
}

func execCmd(name string, args ...string) {
	proc, err := os.StartProcess(name, append([]string{name}, args...), &os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	})
	if err == nil {
		proc.Release()
	}
}
