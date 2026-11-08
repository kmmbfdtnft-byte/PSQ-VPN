package main

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/songgao/water"
)

func main() {
	// 1. Создаём TUN-устройство
	config := water.Config{
		DeviceType: water.TUN,
	}
	ifce, err := water.New(config)
	if err != nil {
		log.Fatal("Ошибка создания TUN: ", err)
	}
	fmt.Printf("Ура! Создан TUN-интерфейс: %s\n", ifce.Name())

	// 2. Настраиваем IP-адрес на интерфейсе
	// Дадим ему адрес 10.0.0.1 с маской /24
	cmd := exec.Command("ip", "addr", "add", "10.0.0.1/24", "dev", ifce.Name())
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Fatalf("Ошибка настройки IP: %v\n%s", err, out)
	}
	fmt.Println("IP-адрес 10.0.0.1/24 назначен")

	// 3. Поднимаем интерфейс (включаем его)
	cmd = exec.Command("ip", "link", "set", "dev", ifce.Name(), "up")
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Fatalf("Ошибка поднятия интерфейса: %v\n%s", err, out)
	}
	fmt.Println("Интерфейс поднят. Слушаем пакеты... (Ctrl+C для выхода)")

	// 4. Бесконечный цикл чтения пакетов
	buf := make([]byte, 1500) // MTU стандартного Ethernet
	count := 0
	for {
		n, err := ifce.Read(buf)
		if err != nil {
			log.Fatal("Ошибка чтения: ", err)
		}
		count++
		fmt.Printf("Пакет #%d: прочитано %d байт\n", count, n)

		// Покажем первые 20 байт пакета в hex-формате
		preview := n
		if preview > 20 {
			preview = 20
		}
		fmt.Printf("   Первые байты: % x\n", buf[:preview])
	}
}

