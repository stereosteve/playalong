package main

import (
	"encoding/json"
	"io"
	"jamfu/views"
	"log"
	"mime/multipart"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/oklog/ulid/v2"
	"go.etcd.io/bbolt"
)

var (
	db              *bbolt.DB
	PlayAlongBucket = []byte("PlayAlong")
)

func main() {
	{
		var err error
		db, err = bbolt.Open("my.db", 0600, nil)
		if err != nil {
			log.Fatal(err)
		}

		db.Update(func(tx *bbolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists(PlayAlongBucket)
			if err != nil {
				log.Fatal(err)
			}
			return nil
		})

		defer db.Close()
	}

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

	e.GET("/song/:id", serveSong)

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
	songs, err := listPlayAlongs()
	if err != nil {
		return err
	}
	return c.JSON(200, songs)
}

func serveSong(c echo.Context) error {
	song, err := getPlayAlong(c.Param("id"))
	if err != nil {
		return err
	}
	return c.JSON(200, song)
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

	if err := savePlayAlong(playAlong); err != nil {
		return err
	}
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

//
// DB STUFF
//

func savePlayAlong(playAlong *PlayAlong) error {
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(PlayAlongBucket)
		j, err := json.Marshal(playAlong)
		if err != nil {
			return err
		}
		return b.Put([]byte(playAlong.ID), j)
	})
}

func listPlayAlongs() ([]*PlayAlong, error) {
	all := []*PlayAlong{}
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(PlayAlongBucket)
		return b.ForEach(func(k, v []byte) error {
			var p *PlayAlong
			err := json.Unmarshal(v, &p)
			if err != nil {
				return err
			}
			all = append(all, p)
			return nil
		})
	})
	return all, err
}

func getPlayAlong(id string) (*PlayAlong, error) {
	var p *PlayAlong
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(PlayAlongBucket)
		v := b.Get([]byte(id))
		return json.Unmarshal(v, &p)
	})
	return p, err
}
