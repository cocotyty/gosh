package main

import (
	"fmt"
	"github.com/elazarl/goproxy"
	"github.com/shadowsocks/shadowsocks-go/shadowsocks"
	"net/http"
	"net/url"
	"os"
	"os/exec"
)

const GOSH_ENV = "SS_SERVER"

func main() {
	// start ss http proxy
	startSSClient()
	// copy cmd to go
	cmd := exec.Command("go", os.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	wd, _ := os.Getwd()
	cmd.Dir = wd
	cmd.Env = append(os.Environ(), "HTTP_PROXY=http://127.0.0.1:30000", "HTTPS_PROXY=http://127.0.0.1:30000")
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
func startSSClient() {

	ssServer := os.Getenv(GOSH_ENV)

	if ssServer == "" {
		fmt.Println("MUST HAS " + GOSH_ENV + " in env")
		os.Exit(-1)
	}
	u, err := url.Parse(ssServer)
	if err != nil {
		fmt.Println(GOSH_ENV+" wrong", err)
		os.Exit(-1)
	}
	password, _ := u.User.Password()
	cipher, err := shadowsocks.NewCipher(u.User.Username(), password)
	if err != nil {
		fmt.Println(GOSH_ENV+" wrong", err)
		os.Exit(-1)
	}
	dialer, err := shadowsocks.NewDialer(u.Host, cipher)
	if err != nil {
		fmt.Println(GOSH_ENV+" wrong", err)
		os.Exit(-1)
	}
	server := goproxy.NewProxyHttpServer()

	server.ConnectDial = dialer.Dial
	server.Tr.Dial = dialer.Dial
	go listenHTTPProxy(server)
}
func listenHTTPProxy(server http.Handler) {
	err := http.ListenAndServe(":30000", server)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
