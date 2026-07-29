package main

import (
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

//go:embed templates
var templatesFS embed.FS

//go:embed dev-private.pem
var privKeyFS embed.FS

const secret = "dfm-anti-pirate-2026"
var portStr = "9090"

type License struct {
	MachineCode string   `json:"machine_code"`
	ExpiresAt   string   `json:"expires_at"`
	Features    []string `json:"features"`
	Customer    string   `json:"customer"`
	IssuedAt    string   `json:"issued_at"`
	Version     string   `json:"version"`
	Signature   string   `json:"signature,omitempty"`
}

func main() {
	port := "9090"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	} else {
		// Try to find a free port
		for i := 0; i < 100; i++ {
			if !isPortInUse(port) {
				break
			}
			port = fmt.Sprintf("%d", 9091+i)
			portStr = port
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", serveIndex)
	mux.HandleFunc("/generate", handleGenerate)
	mux.HandleFunc("/generate-full", handleGenerateFull)
	mux.Handle("/download/", http.StripPrefix("/download/", http.FileServer(http.Dir("."))))

	url := fmt.Sprintf("http://localhost:%s/", portStr)
	fmt.Println("================================================")
	fmt.Println("  东芳美诊所 — 授权码生成工具 v2.0")
	fmt.Println("================================================")
	fmt.Println()
	fmt.Printf("  正在启动...\n")
	fmt.Printf("  浏览器打开：%s\n", url)
	fmt.Println()
	fmt.Println("  提示：关闭此窗口工具即停止运行")
	fmt.Println("================================================")
	fmt.Println()

	openBrowser(url)

	if err := http.ListenAndServe(":"+portStr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "启动失败: %v\n", err)
		os.Exit(1)
	}
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := fs.ReadFile(templatesFS, "templates/index.html")
	if err != nil {
		http.Error(w, "模板加载失败", http.StatusInternalServerError)
		return
	}
	t, err := template.New("index").Parse(string(tmpl))
	if err != nil {
		http.Error(w, "模板解析失败", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t.Execute(w, map[string]string{"BaseURL": "http://localhost:" + portStr})
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		MachineCode string `json:"machine_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}

	machine := strings.ToUpper(strings.ReplaceAll(req.MachineCode, "-", ""))
	if machine == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请输入机器码"})
		return
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(machine))
	code := strings.ToUpper(hex.EncodeToString(mac.Sum(nil))[:12])

	// Format as XXXX-XXXX
	formatted := ""
	for i, c := range code {
		if i > 0 && i%4 == 0 {
			formatted += "-"
		}
		formatted += string(c)
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"unlock_code": formatted,
		"format":      "HMAC-SHA256",
	})
}

func handleGenerateFull(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		MachineCode string `json:"machine_code"`
		Customer    string `json:"customer"`
		ExpireDays  int    `json:"expire_days"`
		Features    string `json:"features"`
		Format      string `json:"format"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}

	machine := strings.ToUpper(strings.ReplaceAll(req.MachineCode, "-", ""))
	if machine == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请输入机器码"})
		return
	}

	now := time.Now()
	issuedAt := now.Format("2006-01-02")
	expiresAt := "永久"
	if req.ExpireDays > 0 {
		expiresAt = now.AddDate(0, 0, req.ExpireDays).Format("2006-01-02")
	}

	features := []string{"all"}
	if req.Features != "" && req.Features != "all" {
		features = strings.Split(req.Features, ",")
	}

	lic := License{
		MachineCode: req.MachineCode,
		ExpiresAt:   expiresAt,
		Features:    features,
		Customer:    req.Customer,
		IssuedAt:    issuedAt,
		Version:     "1.0",
	}

	// Try to sign with private key if available
	privKeyData, err := privKeyFS.ReadFile("dev-private.pem")
	if err == nil {
		block, _ := pem.Decode(privKeyData)
		if block != nil {
			priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
			if err == nil {
				payloadBytes, _ := json.Marshal(License{
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
			}
		}
	}

	out, _ := json.MarshalIndent(lic, "", "  ")

	// Save to file
	filename := fmt.Sprintf("license_%s.json", strings.ReplaceAll(req.Customer, " ", "_"))
	os.WriteFile(filename, out, 0644)

	writeJSON(w, http.StatusOK, map[string]string{
		"content":      string(out),
		"download_url": "/download/" + filename,
	})
}

func isPortInUse(port string) bool {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return true
	}
	ln.Close()
	return false
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
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
