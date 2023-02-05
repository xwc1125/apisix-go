// Package server
//
// @author: xwc1125
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	UNIX_SOCK_PIPE_PATH = "/Users/yijiaren/Downloads/unix/unixsock_test.sock" // socket file path，地址只能下划线及数字字母等
	IsHttp              = false
)

func main() {
	// Remove socket file
	os.Remove(UNIX_SOCK_PIPE_PATH)
	// Get unix socket address based on file path
	uaddr, err := net.ResolveUnixAddr("unix", UNIX_SOCK_PIPE_PATH)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Listen on the address
	unixListener, err := net.ListenUnix("unix", uaddr)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Close listener when close this function, you can also emit it because this function will not terminate gracefully
	defer unixListener.Close()

	fmt.Println("Waiting for asking questions ...")
	if IsHttp {
		router := gin.New()
		router.GET("/testGet", handlerGet)
		http.Serve(unixListener, router) // http监听
	} else {
		// Monitor request and process
		for {
			uconn, err := unixListener.AcceptUnix()
			if err != nil {
				fmt.Println(err)
				continue
			}

			// Handle request
			go handleConnection(uconn)
		}
	}
}

func handlerGet(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"resp": "ok",
	})
}

/*******************************************************
* Handle connection and request
* conn: conn handler
*******************************************************/
func handleConnection(conn *net.UnixConn) {
	// Close connection when finish handling
	defer func() {
		conn.Close()
	}()

	// Read data and return response
	data, err := parseRequest(conn)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%+v\tReceived from client: %s\n", time.Now(), string(data))
	time.Sleep(time.Duration(1) * time.Second) // sleep to simulate request process

	// Send back response
	sendResponse(conn, data)
}

/*******************************************************
* Parse request of unix socket
* conn: conn handler
*******************************************************/
func parseRequest(conn *net.UnixConn) ([]byte, error) {
	var reqLen uint32
	lenBytes := make([]byte, 4)
	if _, err := io.ReadFull(conn, lenBytes); err != nil {
		return nil, err
	}

	lenBuf := bytes.NewBuffer(lenBytes)
	if err := binary.Read(lenBuf, binary.BigEndian, &reqLen); err != nil {
		return nil, err
	}

	reqBytes := make([]byte, reqLen)
	_, err := io.ReadFull(conn, reqBytes)

	if err != nil {
		return nil, err
	}

	return reqBytes, nil
}

/*******************************************************
* Send response to client
* conn: conn handler
*******************************************************/
func sendResponse(conn *net.UnixConn, data []byte) {
	buf := new(bytes.Buffer)
	msglen := uint32(len(data))

	binary.Write(buf, binary.BigEndian, &msglen)
	data = append(buf.Bytes(), data...)
	fmt.Printf("%+v\tSend msg to client: %s\n", time.Now(), string(data))
	conn.Write(data)
}
