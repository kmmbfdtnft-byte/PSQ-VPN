package main

import (
	"fmt"
	"log"
	"net"
	"os/exec"

	"github.com/songgao/water"
)

const (
	serverAddr  = "192.168.100.2:5000"
	tunIP       = "10.1.0.1/24"
	serverIPNet = "10.0.0.1/32"
)

func main() {
	config := water.Config{DeviceType: water.TUN}
	ifce, err := water.New(config)
	if err != nil {
		log.Fatal("Ошибка создания TUN: ", err)
	}
	fmt.Printf("Клиент: TUN %s создан\n", ifce.Name())

	run("ip", "addr", "add", tunIP, "dev", ifce.Name())
	run("ip", "link", "set", "dev", ifce.Name(), "up")
	run("ip", "route", "add", serverIPNet, "dev", ifce.Name())
	fmt.Printf("Клиент: %s, маршрут до %s\n", tunIP, serverIPNet)

	addr, _ := net.ResolveUDPAddr("udp", serverAddr)
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Клиент: подключён к %s\n", serverAddr)

	go func() {
		buf := make([]byte, 1500)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				log.Fatal("UDP read: ", err)
			}
			fmt.Printf("КЛИЕНТ: получил %d байт из UDP\n", n)
			if _, err := ifce.Write(buf[:n]); err != nil {
				log.Fatal("TUN write: ", err)
			}
		}
	}()

	buf := make([]byte, 1500)
	for {
		n, err := ifce.Read(buf)
		if err != nil {
			log.Fatal("TUN read: ", err)
		}
		fmt.Printf("КЛИЕНТ: прочитал %d байт из TUN\n", n)
		if _, err := conn.Write(buf[:n]); err != nil {
			log.Fatal("UDP write: ", err)
		}
	}
}

func run(name string, args ...string) {
	cmd := exec.Command(name, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Fatalf("Команда %s %v упала: %v\n%s", name, args, err, out)
	}
}
