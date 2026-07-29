package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strings"
)

const secret = "dfm-anti-pirate-2026"

func main() {
	machine := flag.String("machine", "", "机器码")
	flag.Parse()
	if *machine == "" {
		fmt.Println("用法: keygen -machine XXXX-XXXX-XXXX")
		os.Exit(1)
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strings.ToUpper(strings.ReplaceAll(*machine, "-", ""))))
	h := hex.EncodeToString(mac.Sum(nil))
	fmt.Println(strings.ToUpper(h[:12]))
}
