package in_memory_content_serving

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"net/http"
	"nfe_3.0_go/nfe/mimelist"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func writeAll(w io.Writer, data []byte) error {
	totalWritten := 0

	for {
		written, err := w.Write(data[totalWritten:])
		if err != nil {
			return err
		}

		totalWritten += written
		if totalWritten >= len(data) {
			break
		}
	}

	return nil
}

func compressGzip(rawData []byte) ([]byte, error) {
	var buffer bytes.Buffer
	gz := gzip.NewWriter(&buffer)

	err := writeAll(gz, rawData)
	if err != nil {
		return nil, err
	}

	err = gz.Flush()
	if err != nil {
		return nil, err
	}

	err = gz.Close()
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func shouldCompress(req *http.Request) bool {
	if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") ||
		strings.Contains(req.Header.Get("Connection"), "Upgrade") {

		return false
	}

	extension := filepath.Ext(req.URL.Path)
	//fmt.Println("extension is =>", extension)
	if len(extension) < 4 { // fast path
		return true
	}

	switch extension {
	// Images compresées
	case ".png", ".gif", ".jpeg", ".jpg":
		return false

		// Archives
	case ".zip", ".gz":
		return false

		// Son
	case ".mp3":
		return false

	default:
		return true
	}
}

func generateHandler(file *staticFile, maxAge string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// https://developer.mozilla.org/fr/docs/Web/HTTP/Headers/ETag
		// Le client possède déjà la dernière version du fichier
		if strings.Contains(c.Request.Header.Get("If-None-Match"), file.Checksum) {
			//fmt.Println()
			//fmt.Println("##########################################################")
			//fmt.Println("YEEEEEEEEEEEEEEEEEEEH CACHE WIN")
			//fmt.Println("##########################################################")
			//fmt.Println()
			c.Status(http.StatusNotModified)
			return
		}

		c.Header("ETag", file.Checksum)
		c.Header("Content-Type", file.ContentType)
		c.Header("Cache-Control", "max-age="+maxAge)

		if shouldCompress(c.Request) {
			c.Header("Content-Encoding", "gzip")
			c.Header("Vary", "Accept-Encoding")
			c.Header("Content-Length", fmt.Sprint(len(file.DataGzip)))

			err := writeAll(c.Writer, file.DataGzip)
			if err != nil {
				panic(err)
			}

			c.Writer.Flush()
		} else {
			c.Header("Content-Length", fmt.Sprint(len(file.DataRaw)))

			err := writeAll(c.Writer, file.DataRaw)
			if err != nil {
				panic(err)
			}

			c.Writer.Flush()
		}
	}
}

type staticFile struct {
	DataRaw     []byte
	DataGzip    []byte
	Checksum    string
	ContentType string
}

func PopulateRouter(router *gin.Engine, scanPath string, expiration int) error {
	rawData := 0
	compressedData := 0

	maxAge := fmt.Sprintf("%d", expiration)

	err := filepath.Walk("./"+scanPath, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			return nil
		}

		if runtime.GOOS == "windows" {
			path = strings.Replace(path, "\\", "/", -1)
		}
		path = strings.TrimPrefix(path, "/")

		if path == "static/config.js" || path == "/static/config.js" {
			return nil
		}

		// Traitement de l'extension et du mimetype
		mimetype := mimelist.GetMime(f.Name())

		// Lecture des données et compression gzip
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		gzipped, err := compressGzip(data)
		if err != nil {
			return err
		}

		// Création du staticFile
		staticFile := staticFile{
			DataRaw:     data,
			DataGzip:    gzipped,
			Checksum:    fmt.Sprintf("%x", sha256.Sum256(data)),
			ContentType: mimetype,
		}

		data = nil
		gzipped = nil

		// Print de debug
		fmt.Println("Processing:", path)
		fmt.Println("  => File name:", f.Name())
		fmt.Println("  => Mimetype:", mimetype, "/", staticFile.ContentType)
		fmt.Println("  => Size:", len(staticFile.DataRaw)/1024, "KB, Gzipped:", len(staticFile.DataGzip)/1024, "KB")
		fmt.Println("  => Checksum:", staticFile.Checksum)
		fmt.Println()

		// Pour les stats
		rawData += len(staticFile.DataRaw)
		compressedData += len(staticFile.DataGzip)

		router.GET(path[len(scanPath):], generateHandler(&staticFile, maxAge))

		if strings.HasSuffix(path, "index.html") {
			router.GET(path[len(scanPath):len(path)-len("index.html")], generateHandler(&staticFile, maxAge))
		}

		return nil
	})
	if err != nil {
		return err
	}

	fmt.Println("Size of raw data:", rawData/1024/1024, "MB")
	fmt.Println("Size of compressed data:", compressedData/1024/1024, "MB")
	fmt.Println("  => Total allocated memory:", (rawData+compressedData)/1024/1024, "MB")
	return nil
}
