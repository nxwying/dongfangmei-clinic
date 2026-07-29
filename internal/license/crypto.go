package license

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"strings"
)

const activateFile = "data/.activated"
const secret = "dfm-anti-pirate-2026"

func MachineCode() string {
	hostname, _ := os.Hostname()
	var addrs []string
	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp != 0 && len(iface.HardwareAddr) > 0 {
			addrs = append(addrs, iface.HardwareAddr.String())
		}
	}
	raw := strings.Join(addrs, "|")
	if raw == "" {
		raw = hostname
	}
	sum := sha256.Sum256([]byte(raw + "|" + hostname))
	h := fmt.Sprintf("%X", sum[:6])
	return fmt.Sprintf("%s-%s-%s", h[:4], h[4:8], h[8:12])
}

func GenerateCode(machine string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strings.ToUpper(strings.ReplaceAll(machine, "-", ""))))
	h := hex.EncodeToString(mac.Sum(nil))
	return strings.ToUpper(h[:12])
}

func IsActivated() bool {
	data, err := os.ReadFile(activateFile)
	if err != nil {
		return false
	}
	parts := strings.SplitN(strings.TrimSpace(string(data)), "|", 2)
	if len(parts) < 2 {
		return false
	}
	return GenerateCode(parts[0]) == parts[1]
}

func Activate(code string) error {
	machine := MachineCode()
	code = strings.ToUpper(strings.ReplaceAll(code, "-", ""))
	expected := GenerateCode(machine)
	if code != expected {
		return fmt.Errorf("解锁码错误")
	}
	return os.WriteFile(activateFile, []byte(machine+"|"+code), 0644)
}
