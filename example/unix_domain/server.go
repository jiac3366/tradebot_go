package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"syscall"
)

func main() {
	// 确保socket文件不存在
	socketPath := "/tmp/echo.sock"
	os.Remove(socketPath)

	// 创建 Unix Domain Socket 监听器
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	fmt.Printf("服务器正在监听: %s\n", socketPath)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("接受连接错误: %v\n", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// 打开要传输的文件
	file, err := os.Open("testfile.txt")
	if err != nil {
		log.Printf("打开文件错误: %v\n", err)
		return
	}
	defer file.Close()

	// 获取Unix domain socket的文件描述符
	unixConn := conn.(*net.UnixConn)
	socketFile, err := unixConn.File()
	if err != nil {
		log.Printf("获取socket文件描述符错误: %v\n", err)
		return
	}
	defer socketFile.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		log.Printf("获取文件信息错误: %v\n", err)
		return
	}

	// 使用splice或sendfile进行传输
	remaining := fileInfo.Size()
	var written int64
	for remaining > 0 {
		n, err := syscall.Sendfile(
			int(socketFile.Fd()),
			int(file.Fd()),
			&written,
			int(remaining),
		)
		if err != nil {
			if err == syscall.EAGAIN {
				continue
			}
			log.Printf("sendfile错误: %v\n", err)
			return
		}
		remaining -= int64(n)
		written += int64(n)
	}

	fmt.Println("文件传输完成")
}
