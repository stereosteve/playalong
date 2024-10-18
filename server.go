package main

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Initialize Echo server
	e := echo.New()

	e.Debug = true

	// Middleware
	// e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Upload and transcode route
	e.POST("/upload", uploadAndTranscode)

	e.Static("/", "public")

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}

// Handler for file upload and transcoding
func uploadAndTranscode(c echo.Context) error {
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}
	for _, file := range form.File["files"] {

		tmpFile, err := copyUploadToTempFile(file)
		if err != nil {
			return err
		}
		defer os.Remove(tmpFile.Name())

		outputFile := filepath.Join("public", "uploads", fmt.Sprintf("transcoded-%s.mp3", filepath.Base(tmpFile.Name())))

		fmt.Println(tmpFile.Name(), outputFile)

		cmd := exec.Command(
			"ffmpeg",
			"-i", tmpFile.Name(),
			"-b:a", "192k",
			outputFile)
		err = cmd.Run()
		if err != nil {
			return err
		}
	}

	return c.String(200, "OK")
}

func copyUploadToTempFile(file *multipart.FileHeader) (*os.File, error) {
	temp, err := os.CreateTemp("", "someupload")
	if err != nil {
		return nil, err
	}

	r, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer r.Close()

	_, err = io.Copy(temp, r)
	if err != nil {
		return nil, err
	}
	temp.Sync()
	temp.Seek(0, 0)

	return temp, nil
}
