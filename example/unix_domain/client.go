package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func main() {
	// 连接到Unix Domain Socket
	conn, err := net.Dial("unix", "/tmp/echo.sock")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// 创建输出文件
	outputFile, err := os.Create("received_file.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()

	// 接收数据并写入文件
	bytesWritten, err := io.Copy(outputFile, conn)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("接收完成，共写入 %d 字节\n", bytesWritten)
}
