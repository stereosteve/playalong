package main

import (
	"crypto/tls"
	"io"
	"jamfu/repo"
	"jamfu/views"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/oklog/ulid/v2"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

func main() {

	repo.Dial()
	// todo: repo.Close() on signal

	// Start server
	e := setupEcho()

	isProd := false
	if isProd {
		// https://echo.labstack.com/docs/cookbook/auto-tls
		e.AutoTLSManager.Cache = autocert.DirCache("/var/www/.cache")
		autoTLSManager := autocert.Manager{
			Prompt: autocert.AcceptTOS,
			Cache:  autocert.DirCache("/var/www/.cache"),
			//HostPolicy: autocert.HostWhitelist("<DOMAIN>"),
		}
		s := http.Server{
			Addr:    ":443",
			Handler: e,
			TLSConfig: &tls.Config{
				GetCertificate: autoTLSManager.GetCertificate,
				NextProtos:     []string{acme.ALPNProto},
			},
			ReadTimeout: 30 * time.Second,
		}
		if err := s.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
			e.Logger.Fatal(err)
		}
	} else {
		e.Logger.Fatal(e.Start(":8080"))
	}
}

func setupEcho() *echo.Echo {
	// Initialize Echo server
	e := echo.New()
	e.HideBanner = true
	e.Debug = true

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", HomeHandler)
	e.GET("/create", func(c echo.Context) error {
		return Render(c, 200, views.Create())
	})
	e.POST("/upload", uploadAndTranscode)
	e.GET("/song/:id", serveSong)
	e.GET("/status", func(ctx echo.Context) error {
		return ctx.String(200, "OK")
	})
	e.Static("/", "public")

	return e
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
	songs, err := repo.ListSongs()
	if err != nil {
		return err
	}
	return Render(c, 200, views.Home(songs))
}

func serveSong(c echo.Context) error {
	song, err := repo.GetSong(c.Param("id"))
	if err != nil {
		return err
	}
	return Render(c, 200, views.Song(song))
}

func uploadAndTranscode(c echo.Context) error {

	id := ulid.Make().String()
	song := &repo.Song{
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

		song.Stems = append(song.Stems, repo.StemFile{
			Name: file.Filename,
			Path: path.Join("/uploads", id, stemName+".mp3"),
		})
	}

	if err := repo.SaveSong(song); err != nil {
		return err
	}
	// return c.JSON(200, song)
	return c.Redirect(302, "/song/"+song.ID)
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
