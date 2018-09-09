package serve

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"mime"
	"net/http"
	"nfe_3.0_go/nfe/transfer"
	"os"
	"time"
)

const averageTime = 333 * time.Millisecond

func sendHeaders(c *gin.Context, fileName string, fileSize int64) {
	c.Status(http.StatusOK)
	c.Header("Content-Length", fmt.Sprintf("%d", fileSize))
	c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{
		"filename": fileName,
	}))
}

func routineReadDisk(readerChannel chan []byte, f *os.File, bufferSize int64) {
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
}

func routineMeasureSpeed(speedChannel chan time.Duration, t *transfer.Transfer) {
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
			//t.CurrentSpeed = int64(float64(sentPackets*t.BufferSize) / (float64(measureTime) / float64(time.Second)))
			//fmt.Println("B/s =>", t.CurrentSpeed)
			//fmt.Println("KB/s =>", t.CurrentSpeed/1000)
			//fmt.Println("MB/s =>", t.CurrentSpeed/1000/1000)

			t.CurrentSpeedMeasureTime = measureTime
			t.CurrentSpeedSentPackets = sentPackets

			measureTime = 0
			sentPackets = 0
		}
	}
}

func ServeFile(c *gin.Context, t *transfer.Transfer) {
	// 1] Defer pour tout bien fermer proprement
	defer c.Request.Body.Close()

	// 2] Defer pour informer le Transfer
	defer func() {
		// Todo: informer le Transfer

		if err := recover(); err != nil {
			fmt.Println("Http main serving goroutine has terminated forcefully:", err)
		} else {
			fmt.Println("Http main serving goroutine has terminated gracefully")
		}
	}()

	fmt.Println("Has began to serve file") // Todo: informer le Transfer

	// Ouverture et obtention des infos du fichier
	f, err := os.Open(t.FilePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Envoi des headers avec taille et nom du fichier
	sendHeaders(c, f.Name(), t.FileLength)

	// Lancement de la routine de lecture du disque
	readerChannel := make(chan []byte, 100)
	go routineReadDisk(readerChannel, f, t.BufferSize)

	// Lancement de la routine de mesure de vitesse
	speedChannel := make(chan time.Duration)
	defer close(speedChannel)
	go routineMeasureSpeed(speedChannel, t)

	// Stream des données
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
		if t.CurrentSpeedLimitDelay != 0 && t.CurrentSpeedLimitDelay > diff {
			timeToWait := (t.CurrentSpeedLimitDelay - diff) * 95 / 100
			fmt.Println("Client was too fast, waiting for", timeToWait/time.Microsecond, "us")
			time.Sleep(timeToWait)
			speedChannel <- diff + timeToWait
		} else {
			speedChannel <- diff
		}
	}
}
