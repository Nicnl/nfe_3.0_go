package deterministic_tar

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"mime"
	"path/filepath"
)

func GinTarStream(c *gin.Context, path string) error {
	// Precalculate the size of the tar file
	expectedSize, files, err := A_PrecalculateTarSize(path)
	if err != nil {
		return err
	}

	// Send the HTTP headers
	tarName := filepath.Base(path) + ".tar"
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", fmt.Sprintf("%d", expectedSize))
	c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{
		"filename": tarName,
	}))

	// Create and stream the tar on the fly
	err = B_tar_stream(c.Writer, files, expectedSize)
	if err != nil {
		return err
	}

	return nil
}
