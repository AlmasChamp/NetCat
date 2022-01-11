package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var dict = map[string]string{
	"red":    "красный",
	"green":  "зеленый",
	"blue":   "синий",
	"yellow": "желтый",
}

var (
	users       = make(map[net.Conn]string)
	tempHistory = []byte{}
)

func main() {

	// Считывает рисунок с файла "pengue.txt"
	pengue, err := ioutil.ReadFile("pengue.txt")
	if err != nil {
		log.Fatal(err)
	}
	// Иницифлизируем Мютекс для синхронизаций процессов
	var mut sync.Mutex

	// Порт по умолчанию
	port := "2525"
	args := os.Args[1:]

	// Проверем порт
	if len(args) != 0 {
		if validPort(args[0]) == true && len(args) == 1 {
			port = args[0]
		} else {
			log.Fatal("[USAGE]: ./TCPChat $port")
		}
	}

	// Подключаем порт
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer listener.Close()

	fmt.Println("Server is listening port " + port)

	// Слушаем порт, ждём подключение клиентов
	for {
		conn, err := listener.Accept()
		if err != nil {
			conn.Close()
			continue
		}

		conn.Write([]byte(pengue))
		conn.Write([]byte("\n[ENTER YOUR NAME]:"))
		// Запускаем горутину для обработки подключения
		go handleConnection(conn, &mut)
	}

}

// Обработка подключения
func handleConnection(conn net.Conn, mut *sync.Mutex) {

	newconn := bufio.NewReader(conn)

	// Считываем имя клиента и проверяем
	name := ""
	for {
		name, _ = newconn.ReadString('\n')
		if isValid(name[:len(name)-1]) == false {
			conn.Write([]byte("\n[ENTER YOUR NAME]:"))
		} else {
			break
		}
	}
	// Переменные (Только так хватило ума)
	tempTime := time.Now()
	newTime := strings.Split(tempTime.String(), ".")
	finalTime := "[" + newTime[0] + "]"

	finalyName := "[" + name[:len(name)-1] + "]:"
	gretings := " has joined our chat..."
	parting := " has left our chat..."

	// Отправляем историю чата
	conn.Write(tempHistory)

	//Вносим имя клиента и приветсвует собеседников
	//Используем Мутекс для синхрнизаций процессов (это механизм, позволяющий обеспечить целостность какого-либо ресурса (файл, данные в памяти)
	mut.Lock()
	users[conn] = string(name)
	for user, value := range users {
		if user != conn {
			user.Write([]byte("\n"))
			user.Write([]byte(name[:len(name)-1] + gretings + "\n"))
			user.Write([]byte(finalTime + "[" + value[:len(value)-1] + "]:"))
		}

	}
	mut.Unlock()

	defer conn.Close()

	for {

		// Инициализируем время
		tempTime = time.Now()
		newTime = strings.Split(tempTime.String(), ".")
		finalTime = "[" + newTime[0] + "]"

		conn.Write([]byte(finalTime + finalyName))

		message := ""
		var err error

		// Получаем сообщение от клиента и проверяем его
		for {
			message, err = newconn.ReadString('\n')
			if err != nil {
				err = errors.New("Error")
				break
			} else if isValid(message) == false {
				conn.Write([]byte(finalTime + finalyName + "Enter your message:"))
			} else {
				break
			}
		}

		// Информируем участкников чата об отсоединений клиента
		if err != nil {
			for user, value := range users {
				if user != conn {
					user.Write([]byte("\n" + name[:len(name)-1] + parting + "\n"))
					user.Write([]byte(finalTime + "[" + value[:len(value)-1] + "]:"))
				}
			}
			break
		}

		
		mut.Lock()
		// Оправляем сообщение всем участникам чата
		for user, value := range users {

			if user != conn {
				user.Write([]byte("\n" + finalTime + finalyName + message))
				user.Write([]byte(finalTime + "[" + value[:len(value)-1] + "]:"))
			}

		}
		mut.Unlock()

		// Сохраняем историю переписки
		tempHistory = append(tempHistory, finalTime+finalyName+message...)

		// Записываем историю в файл
		ioutil.WriteFile("history.txt", tempHistory, 0777)

	}
}

func validPort(p string) bool {
	for i := 0; i < len(p); i++ {
		if p[i] >= '0' && p[i] <= '9' {
			continue
		}
		return false
	}
	return true
}

func isValid(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] != 10 && s[i] != 32 {
			return true
		}
	}
	return false
}
