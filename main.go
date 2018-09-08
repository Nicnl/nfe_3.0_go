package main

import (
	"fmt"
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

func main() {

	path := "/vmshare_hub/ISOs/Windows/Windows 10 - 1703/Win10_1703_French_x64.iso"

	http.HandleFunc("/lol.pngz", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("Http main serving goroutine has terminated forcefully:", err)
			} else {
				fmt.Println("Http main serving goroutine has terminated gracefully")
			}
		}()

		f, err := os.Open(path)
		if err != nil {
			panic(err)
		}

		info, err := f.Stat()
		if err != nil {
			panic(err)
		}

		w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))
		w.Header().Set("Content-Disposition", "attachment; filename=\"Win10_1703_French_x64.iso\"")
		w.WriteHeader(http.StatusOK)

		limiteVitesse := 400 * 1024
		bufferSize := 20 * 1000

		nbPackets := limiteVitesse / bufferSize
		timeBetweenPackets := time.Second / time.Duration(nbPackets)

		readerChannel := make(chan []byte, 100)
		go func(readerChannel chan []byte) {
			defer func() { recover(); fmt.Println("Reader goroutine has terminated") }()

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

		for {
			data, ok := <-readerChannel

			if !ok {
				break
			}

			start := time.Now()
			_, err := w.Write(data)
			diff := time.Since(start)

			if err != nil {
				close(readerChannel)
				panic(err)
				return
			}

			fmt.Println("Time:", int64(diff/time.Microsecond), "us")
			if timeBetweenPackets > diff {
				timeToWait := (timeBetweenPackets - diff) * 97 / 100
				fmt.Println("Client was too fast, waiting for", timeToWait/time.Microsecond, "us")
				time.Sleep(timeToWait)
			}
		}
	})

	err := http.ListenAndServe(":9000", nil)
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
