package main

import (
	"io"
	"jamfu/views"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/oklog/ulid/v2"
)

func main() {
	// Initialize Echo server
	e := echo.New()
	e.HideBanner = true
	e.Debug = true

	// Middleware
	// e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Upload and transcode route
	e.GET("/", HomeHandler)
	e.GET("/create", func(c echo.Context) error {
		return Render(c, 200, views.Create())
	})
	e.POST("/upload", uploadAndTranscode)

	e.Static("/", "public")

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}

// This custom Render replaces Echo's echo.Context.Render() with templ's templ.Component.Render().
func Render(ctx echo.Context, statusCode int, t templ.Component) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	if err := t.Render(ctx.Request().Context(), buf); err != nil {
		return err
	}

	return ctx.HTML(statusCode, buf.String())
}

func HomeHandler(c echo.Context) error {
	// todo: list playAlongs
	return Render(c, http.StatusOK, views.Home("asdf 123 111"))
}

type PlayAlong struct {
	ID    string
	Title string
	Genre string
	Key   string
	BPM   string
	Stems []StemFile
}
type StemFile struct {
	Name string
	Path string
}

func uploadAndTranscode(c echo.Context) error {

	id := ulid.Make().String()
	playAlong := &PlayAlong{
		ID:    id,
		Title: c.FormValue("title"),
		Genre: c.FormValue("genre"),
		Key:   c.FormValue("key"),
		BPM:   c.FormValue("bpm"),
	}

	form, err := c.MultipartForm()
	if err != nil {
		return err
	}

	os.MkdirAll(filepath.Join("public", "uploads", id), os.ModePerm)

	for _, file := range form.File["files"] {
		tmpFile, err := copyUploadToTempFile(file)
		if err != nil {
			return err
		}
		defer os.Remove(tmpFile.Name())

		uploadName := file.Filename
		stemName := strings.TrimSuffix(uploadName, filepath.Ext(uploadName))
		outputFile := filepath.Join("public", "uploads", id, stemName+".mp3")
		cmd := exec.Command(
			"ffmpeg",
			"-i", tmpFile.Name(),
			"-b:a", "192k",
			outputFile)
		err = cmd.Run()
		if err != nil {
			return err
		}

		playAlong.Stems = append(playAlong.Stems, StemFile{
			Name: file.Filename,
			Path: path.Join("uploads", id, stemName+".mp3"),
		})
	}

	// todo: save playAlong
	return c.JSON(200, playAlong)
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
