package serve

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"mime"
	"net"
	"net/http"
	"nfe_3.0_go/nfe/json_time"
	"nfe_3.0_go/nfe/transfer"
	"nfe_3.0_go/nfe/vfs"
	"time"
)

const averageTime = 333 * time.Millisecond

type Env struct {
	Vfs vfs.Vfs
}

func (env *Env) sendHeaders(c *gin.Context, t *transfer.Transfer) {
	c.Status(http.StatusOK)
	c.Header("Content-Length", fmt.Sprintf("%d", t.FileLength)) // Todo: sections
	c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{
		"filename": t.FileName,
	}))
}

func (env *Env) routineReadDisk(readerChannel chan []byte, f io.Reader, t *transfer.Transfer) {
	defer func() {
		recover()
		fmt.Println("Disk reader goroutine has terminated")
	}()

	var bufferSize = t.BufferSize

	for {
		b := make([]byte, bufferSize)

		readBytes, err := f.Read(b)
		if err != nil {
			t.CurrentState = transfer.StateInterruptedServer
			close(readerChannel)
			return
		}

		readerChannel <- b[:readBytes]
	}
}

func (env *Env) routineMeasureSpeed(speedChannel chan time.Duration, t *transfer.Transfer, c *gin.Context) {
	defer func() {
		recover()
		fmt.Println("Speed calculator goroutine has terminated")
	}()

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
			t.CurrentSpeed = int64(float64(sentPackets*t.BufferSize) / (float64(measureTime) / float64(time.Second)))
			//fmt.Println("B/s =>", t.CurrentSpeed)
			//fmt.Println("KB/s =>", t.CurrentSpeed/1000)
			//fmt.Println("MB/s =>", t.CurrentSpeed/1000/1000)

			measureTime = 0
			sentPackets = 0
		}

		if t.ShouldInterrupt {
			t.CurrentState = transfer.StateInterruptedAdmin
			close(speedChannel)

			defer c.Request.Body.Close()

			conn, _, err := c.Writer.Hijack()
			if err != nil {
				return
			}

			conn.(*net.TCPConn).SetLinger(0)
			conn.(*net.TCPConn).CloseRead()
			conn.(*net.TCPConn).CloseWrite()
			defer conn.Close()

			return
		}
	}
}

func (env *Env) ServeFile(c *gin.Context, t *transfer.Transfer) {
	// 1] Defer pour tout bien fermer proprement
	defer func() {
		defer func() { recover() }()
		c.Request.Body.Close()
	}()

	t.StartDate = json_time.JsonTime(time.Now())

	// Monosection fichier entier
	t.SectionLength = t.FileLength
	t.SectionStart = 0
	t.CurrentState = transfer.StateTransferring

	// 2] Defer pour informer le Transfer
	defer func() {
		// Todo: informer le Transfer
		t.EndDate = json_time.JsonTime(time.Now())

		if err := recover(); err != nil {
			fmt.Println("Http main serving goroutine has terminated forcefully:", err)

			if t.CurrentState == transfer.StateTransferring {
				t.CurrentState = transfer.StateInterruptedServer
			}
		} else {
			fmt.Println("Http main serving goroutine has terminated gracefully")
			t.CurrentState = transfer.StateFinished
		}
	}()

	fmt.Println("Has began to serve file") // Todo: informer le Transfer

	// Ouverture et obtention des infos du fichier
	f, err := env.Vfs.Open(t.FilePath)
	if err != nil {
		t.CurrentState = transfer.StateInterruptedServer
		panic(err)
	}
	defer f.Close()

	// Envoi des headers avec taille et nom du fichier
	env.sendHeaders(c, t)

	// Lancement de la routine de lecture du disque
	readerChannel := make(chan []byte, 100)
	defer close(readerChannel)
	go env.routineReadDisk(readerChannel, f, t)

	// Lancement de la routine de mesure de vitesse
	speedChannel := make(chan time.Duration)
	defer close(speedChannel)
	go env.routineMeasureSpeed(speedChannel, t, c)

	// Stream des données
	for {
		data, ok := <-readerChannel

		if !ok {
			break
		}

		start := time.Now()
		_, err := c.Writer.Write(data)
		diff := time.Since(start)
		t.Downloaded += int64(len(data))

		if err != nil {
			fmt.Println("ERROR WHEN WRITING LAST DATA")
			fmt.Println("t.Downloaded = ", t.Downloaded)
			fmt.Println("t.SectionLength = ", t.SectionLength)
			fmt.Println("err =", err)
			if t.Downloaded >= t.SectionLength {
				t.CurrentState = transfer.StateFinished
			} else {
				t.CurrentState = transfer.StateInterruptedClient
			}

			panic(err)
			return
		}

		//fmt.Println("Time:", int64(diff/time.Microsecond), "us")
		if t.CurrentSpeedLimitDelay != 0 && t.CurrentSpeedLimitDelay > diff {
			timeToWait := (t.CurrentSpeedLimitDelay - diff) * 95 / 100
			//fmt.Println("Client was too fast, waiting for", timeToWait/time.Microsecond, "us")
			time.Sleep(timeToWait)
			speedChannel <- diff + timeToWait
		} else {
			speedChannel <- diff
		}
	}

	t.CurrentState = transfer.StateFinished
}
