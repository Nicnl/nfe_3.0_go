package serve

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"mime"
	"net"
	"net/http"
	"nfe_3.0_go/nfe/json_time"
	"nfe_3.0_go/nfe/mimelist"
	"nfe_3.0_go/nfe/transfer"
	"nfe_3.0_go/nfe/vfs"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const averageTime = 333 * time.Millisecond

type Env struct {
	AuthBlobAdmin      []byte
	AuthBlobRegular    []byte
	JwtSecret          []byte
	Vfs                vfs.Vfs
	BasePath           string
	PasswordHashSalt   []byte
	NonAdminTimeLimit  int64
	NonAdminSpeedLimit int64
	DefaultSpeedLimit  int64
	DefaultTimeLimit   int64
	GlobSalt           []byte
	GlobUrlList        []byte
	GlobUrlDown        []byte

	transfersMux sync.Mutex
	transfers    map[string]*transfer.Transfer
}

func (e *Env) TransfersInit() {
	e.transfersMux.Lock()
	defer e.transfersMux.Unlock()

	e.transfers = map[string]*transfer.Transfer{}
}

func (e *Env) TransfersGet(key string) (*transfer.Transfer, bool) {
	e.transfersMux.Lock()
	defer e.transfersMux.Unlock()

	output, ok := e.transfers[key]

	return output, ok
}

func (e *Env) TransfersSet(key string, val *transfer.Transfer) {
	e.transfersMux.Lock()
	defer e.transfersMux.Unlock()

	e.transfers[key] = val
}

func (e *Env) TransfersCopy() map[string]*transfer.Transfer {
	e.transfersMux.Lock()
	defer e.transfersMux.Unlock()

	output := make(map[string]*transfer.Transfer, len(e.transfers))
	for k, v := range e.transfers {
		output[k] = v
	}

	return output
}

func (e *Env) TransfersDelete(check func(key string, val *transfer.Transfer) bool) {
	e.transfersMux.Lock()
	defer e.transfersMux.Unlock()

	for k, v := range e.transfers {
		if check == nil || check(k, v) {
			delete(e.transfers, k)
		}
	}
}

func (env *Env) routineReadDisk(buffers [chanBufferSize][MaxBufferSize]byte, readerChannel chan buffIdentifier, readReturnChannel chan buffIdentifier, f io.Reader, t *transfer.Transfer, until int64) {
	defer func() { readerChannel = nil }()
	defer func() { readReturnChannel = nil }()
	//fmt.Println("Start disk gouroutine with until =", until)
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Disk reader goroutine has terminated forcefully:", err)
		} else {
			fmt.Println("Disk reader goroutine has terminated gracefully")
		}
	}()

	var identifier buffIdentifier
	for {
		if until <= 0 {
			close(readerChannel)
			close(readReturnChannel)
			return
		}

		// Obtention de la prochaine zone mémoire libre utilisable
		identifier = <-readReturnChannel
		var b []byte

		// Calcul de la taille de buffers
		buffSize := t.BufferSize
		// Todo: voir si il faut pas remettre cette comparaison optionnelle
		//if buffSize > MaxBufferSize {
		//	buffSize = MaxBufferSize
		//}

		if buffSize > until {
			buffSize = until
			b = buffers[identifier.Index][:buffSize]
		} else {
			// Todo: checker si c'est mauvais pour les perfs de reslicer à la volée
			b = buffers[identifier.Index][:]
		}

		// Lecture des données depuis le disque
		readBytes, err := f.Read(b)
		if err != nil {
			t.CurrentState = transfer.StateInterruptedServer
			close(readerChannel)
			close(readReturnChannel)
			return
		}
		until -= int64(readBytes)
		//fmt.Println("Disk reader goroutine has read", readBytes, "bytes, until is now", until)

		if buffSize > int64(readBytes) {
			identifier.Size = readBytes
		} else {
			identifier.Size = MaxBufferSize
		}

		readerChannel <- identifier
	}
}

func (env *Env) routineMeasureSpeed(speedChannel chan time.Duration, t *transfer.Transfer, c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Speed calculator goroutine has terminated forcefully:", err)
		} else {
			fmt.Println("Speed calculator goroutine has terminated gracefully")
		}
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

func rangeNotSatisfiable(c *gin.Context, t *transfer.Transfer) {
	c.Status(http.StatusRequestedRangeNotSatisfiable)
	c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", 0, t.FileLength-1, t.FileLength))
	c.Header("Accept-Ranges", fmt.Sprintf("bytes"))
}

func detectRanges(c *gin.Context, t *transfer.Transfer, info os.FileInfo) (int64, int64, bool) {
	rangeHeader := c.GetHeader("Range")
	//fmt.Println("RAW RANGE HEADER :", rangeHeader)
	if rangeHeader == "" {
		t.SectionLength = t.FileLength
		t.SectionStart = 0

		c.Status(http.StatusOK)
		//c.Header("Accept-Ranges", fmt.Sprintf("bytes"))
		c.Header("Content-Length", fmt.Sprintf("%d", t.FileLength))
		c.Header("Content-Type", mime.FormatMediaType(mimelist.GetMime(t.FileName), map[string]string{
			"name": t.FileName,
		}))
		c.Header("Content-Disposition", mime.FormatMediaType("inline", map[string]string{
			"filename": t.FileName,
		}))

		c.Header("Content-Description", "File Transfer")

		return 0, t.FileLength, true // Todo : vérifier range de fin
	}

	// Pas de support des plusieurs ranges
	if strings.Contains(rangeHeader, ",") {
		rangeNotSatisfiable(c, t)
		return -1, -1, false
	}

	// Parse de la range du client
	splitUnitRange := strings.Split(rangeHeader, "=")
	if len(splitUnitRange) != 2 {
		rangeNotSatisfiable(c, t)
		return -1, -1, false
	}

	if strings.ToLower(splitUnitRange[0]) != "bytes" {
		rangeNotSatisfiable(c, t)
		return -1, -1, false
	}

	splitRange := strings.Split(splitUnitRange[1], "-")
	if len(splitRange) != 2 {
		rangeNotSatisfiable(c, t)
		return -1, -1, false
	}

	rangeStart, err := strconv.ParseInt(splitRange[0], 10, 64)
	if err != nil {
		rangeNotSatisfiable(c, t)
		return -1, -1, false
	}

	var rangeEnd int64
	if splitRange[1] != "" {
		rangeEnd, err = strconv.ParseInt(splitRange[1], 10, 64)
		if err != nil {
			rangeNotSatisfiable(c, t)
			return -1, -1, false
		}
	} else {
		rangeEnd = t.FileLength - 1
	}

	if rangeStart < 0 || rangeEnd < 0 || rangeStart >= rangeEnd { // Wtf, ça devrais jamais arriver mais bon
		rangeNotSatisfiable(c, t)
		return -1, -1, false
	}

	if rangeStart >= t.FileLength || rangeEnd > t.FileLength { // Le client demande trop loin
		rangeNotSatisfiable(c, t)
		return -1, -1, false
	}

	t.SectionLength = rangeEnd - rangeStart + 1
	t.SectionStart = rangeStart

	c.Status(http.StatusPartialContent)
	c.Header("Accept-Ranges", fmt.Sprintf("bytes"))
	c.Header("Content-Length", fmt.Sprintf("%d", t.SectionLength))
	c.Header("Content-Type", mime.FormatMediaType(mimelist.GetMime(t.FileName), map[string]string{
		"name": t.FileName,
	}))
	c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{
		"filename": t.FileName,
	}))
	c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", rangeStart, rangeEnd, t.FileLength))
	c.Header("Content-Description", "File Transfer")
	//fmt.Println("Ranges =",  fmt.Sprintf("bytes %d-%d/%d", rangeStart, rangeEnd, t.FileLength))

	return rangeStart, t.SectionLength, true
}

const MaxBufferSize = 50 * 1024
const chanBufferSize = 100

type buffIdentifier struct {
	Index int
	Size  int
}

func (env *Env) ServeFile(c *gin.Context, t *transfer.Transfer, subVfs vfs.Vfs) {
	// 1] Defer pour tout bien fermer proprement
	defer func() {
		defer func() { recover() }()
		c.Request.Body.Close()
	}()

	t.StartDate = json_time.JsonTime(time.Now())

	// Monosection fichier entier
	t.CurrentState = transfer.StateTransferring

	// 2] Defer pour informer le Transfer
	defer func() {
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

	//fmt.Println("Has began to serve file")

	// Ouverture et obtention des infos du fichier
	info, err := subVfs.Stat(t.FilePath)
	if err != nil {
		t.CurrentState = transfer.StateInterruptedServer
		panic(err)
	}

	// Envoi des headers avec taille et nom du fichier
	//env.sendHeaders(c, t)
	fileSeek, streamLength, shouldContinue := detectRanges(c, t, info)
	if !shouldContinue {
		t.CurrentState = transfer.StateInterruptedClient
		return
	}

	var f io.ReadCloser
	if fileSeek == 0 {
		f, err = subVfs.Open(t.FilePath)
	} else {
		f, err = subVfs.OpenSeek(t.FilePath, fileSeek)
	}
	if err != nil {
		t.CurrentState = transfer.StateInterruptedServer
		panic(err)
	}
	defer f.Close()

	// Préparation de l'espace mémoire
	readerChannel := make(chan buffIdentifier, chanBufferSize)
	readReturnChannel := make(chan buffIdentifier, chanBufferSize+1)
	buffers := [chanBufferSize][MaxBufferSize]byte{}
	for i := 0; i < chanBufferSize; i++ {
		readReturnChannel <- buffIdentifier{
			Index: i,
			Size:  MaxBufferSize,
		}
	}

	// Lancement de la routine de lecture du disque
	defer func() {
		defer func() { recover() }()
		close(readerChannel)
		readerChannel = nil
	}()
	defer func() {
		defer func() { recover() }()
		close(readReturnChannel)
		readReturnChannel = nil
	}()
	go env.routineReadDisk(buffers, readerChannel, readReturnChannel, f, t, streamLength)

	// Lancement de la routine de mesure de vitesse
	speedChannel := make(chan time.Duration)
	defer func() {
		defer func() { recover() }()
		close(speedChannel)
	}()
	go env.routineMeasureSpeed(speedChannel, t, c)

	// Stream des données
	for {
		identifier, ok := <-readerChannel

		if !ok {
			break
		}

		start := time.Now()

		var err error
		// Todo: voir si c'est pas mauvais pour les perfs de reslicer à la volée
		if MaxBufferSize == identifier.Size {
			/*wroteBytes*/ _, err = c.Writer.Write(buffers[identifier.Index][:])
		} else {
			/*wroteBytes*/ _, err = c.Writer.Write(buffers[identifier.Index][:identifier.Size])
		}
		t.Downloaded += int64(identifier.Size)
		readReturnChannel <- identifier // On renvoie l'identifier afin que la zone mémoire puisse être réutilisée
		diff := time.Since(start)
		//fmt.Println("Have wrote to client:", len(data), "aka", wroteBytes)

		if err != nil {
			//fmt.Println("ERROR WHEN WRITING LAST DATA")
			//fmt.Println("t.Downloaded = ", t.Downloaded)
			//fmt.Println("t.SectionLength = ", t.SectionLength)
			//fmt.Println("err =", err)
			if t.Downloaded >= t.SectionLength {
				t.CurrentState = transfer.StateFinished
			} else {
				t.CurrentState = transfer.StateInterruptedClient
				panic(err)
			}

			return
		}

		//fmt.Println("Time:", int64(diff/time.Microsecond), "us")
		if t.CurrentSpeedLimitDelay != 0 && t.CurrentSpeedLimitDelay > diff {
			timeToWait := t.CurrentSpeedLimitDelay - diff
			//fmt.Println("Client was too fast, waiting for", timeToWait/time.Microsecond, "us")
			time.Sleep(timeToWait)
			speedChannel <- diff + timeToWait
		} else {
			speedChannel <- diff
		}
	}

	t.CurrentState = transfer.StateFinished
}
