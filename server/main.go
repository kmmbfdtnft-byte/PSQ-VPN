package main

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"sync"

	"github.com/songgao/water"
)

const (
	udpPort   = 5000
	tunIP     = "10.0.0.1/24"
	clientNet = "10.1.0.0/24"
)

func main() {
	config := water.Config{DeviceType: water.TUN}
	ifce, err := water.New(config)
	if err != nil {
		log.Fatal("Ошибка создания TUN: ", err)
	}
	fmt.Printf("Сервер: TUN %s создан\n", ifce.Name())

	run("ip", "addr", "add", tunIP, "dev", ifce.Name())
	run("ip", "link", "set", "dev", ifce.Name(), "up")
	run("ip", "route", "add", clientNet, "dev", ifce.Name())
	fmt.Printf("Сервер: %s, маршрут до %s\n", tunIP, clientNet)

	addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", udpPort))
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Сервер: слушаем UDP :%d\n", udpPort)

	var (
		clientAddr   *net.UDPAddr
		clientAddrMu sync.Mutex
	)

	go func() {
		buf := make([]byte, 1500)
		for {
			n, err := ifce.Read(buf)
			if err != nil {
				log.Fatal("TUN read: ", err)
			}
			fmt.Printf("СЕРВЕР: прочитал %d байт из TUN\n", n)

			clientAddrMu.Lock()
			ca := clientAddr
			clientAddrMu.Unlock()
			if ca != nil {
				conn.WriteToUDP(buf[:n], ca)
			}
		}
	}()

	buf := make([]byte, 1500)
	for {
		n, udpAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Fatal("UDP read: ", err)
		}
		fmt.Printf("СЕРВЕР: получил %d байт из UDP\n", n)

		clientAddrMu.Lock()
		if clientAddr == nil {
			clientAddr = udpAddr
			fmt.Printf("Сервер: запомнил клиента %s\n", udpAddr)
		}
		clientAddrMu.Unlock()

		if _, err := ifce.Write(buf[:n]); err != nil {
			log.Fatal("TUN write: ", err)
		}
	}
}

func run(name string, args ...string) {
	cmd := exec.Command(name, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Fatalf("Команда %s %v упала: %v\n%s", name, args, err, out)
	}
}
