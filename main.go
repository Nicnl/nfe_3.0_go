package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	nfeCrypt "nfe_3.0_go/nfe/crypt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func encode(path string) string {
	path = strings.TrimLeft(path, "/")
	if !strings.Contains(path, "/") {
		return nfeCrypt.GlobUnique([]byte(path))
	}

	fmt.Println("path =", path)

	return ""
}

const averageTime time.Duration = 333 * time.Millisecond

func main() {
	path := "/vmshare_hub/ISOs/Windows/Windows 10 - 1703/Win10_1703_French_x64.iso"

	r := gin.Default()

	r.GET("/download", func(c *gin.Context) {
		defer func() {
			c.Request.Body.Close()

			if err := recover(); err != nil {
				fmt.Println("Http main serving goroutine has terminated forcefully:", err)
			} else {
				fmt.Println("Http main serving goroutine has terminated gracefully")
			}
		}()

		fmt.Println("Has began to serve file")

		f, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		info, err := f.Stat()
		if err != nil {
			panic(err)
		}

		c.Status(http.StatusOK)
		c.Header("Content-Length", fmt.Sprintf("%d", info.Size()))
		c.Header("Content-Disposition", "attachment; filename=\"Win10_1703_French_x64.iso\"")

		var limiteVitesse int64 = 550 * 1024
		var bufferSize int64 = 20 * 1000

		nbPackets := limiteVitesse / bufferSize
		timeBetweenPackets := time.Second / time.Duration(nbPackets)

		readerChannel := make(chan []byte, 100)
		go func(readerChannel chan []byte) {
			defer func() { recover(); fmt.Println("Disk reader goroutine has terminated") }()

			for {
				b := make([]byte, bufferSize)

				readBytes, err := f.Read(b)
				if err != nil {
					close(readerChannel)
					return
				}

				readerChannel <- b[:readBytes]
			}
		}(readerChannel)

		speedChannel := make(chan time.Duration)
		defer close(speedChannel)
		go func(speedChannel chan time.Duration) {
			defer func() { recover(); fmt.Println("Speed calculator goroutine has terminated") }()

			var measureTime time.Duration = 0
			var sentPackets int64 = 0

			for {
				select {
				case duration, ok := <-speedChannel:
					if !ok {
						return
					}
					sentPackets += 1
					measureTime += duration
				case <-time.After(averageTime):
					measureTime += averageTime
				}

				if measureTime > averageTime {
					//bandwidth := float64(sentPackets * bufferSize) / ()
					bandwidth := float64(sentPackets*bufferSize) / (float64(measureTime) / float64(time.Second))
					fmt.Println("B/s =>", bandwidth)
					fmt.Println("KB/s =>", bandwidth/1000)
					fmt.Println("MB/s =>", bandwidth/1000/1000)

					measureTime = 0
					sentPackets = 0
				}
			}
		}(speedChannel)

		for {
			data, ok := <-readerChannel

			if !ok {
				break
			}

			start := time.Now()
			_, err := c.Writer.Write(data)
			diff := time.Since(start)

			if err != nil {
				close(readerChannel)
				panic(err)
				return
			}

			//fmt.Println("Time:", int64(diff/time.Microsecond), "us")
			if timeBetweenPackets > diff {
				//timeToWait := (timeBetweenPackets - diff) * 95 / 100
				//fmt.Println("Client was too fast, waiting for", timeToWait/time.Microsecond, "us")
				//time.Sleep(timeToWait)
				//speedChannel <- diff + timeToWait
				speedChannel <- diff
			} else {
				speedChannel <- diff
			}
		}
	})

	err := http.ListenAndServe(":9000", r)
	if err != nil {
		panic(err)
	}

	start := time.Now()
	time.Sleep(122 * time.Microsecond)
	elapsed := time.Since(start)
	fmt.Println(int64(elapsed/time.Millisecond), "ms")
	fmt.Println(int64(elapsed/time.Microsecond), "us")
	return

	erro := filepath.Walk("C:/", func(path string, info os.FileInfo, err error) error {
		fmt.Println(path)
		return nil
	})
	if erro != nil {
		panic(erro)
	}
}
