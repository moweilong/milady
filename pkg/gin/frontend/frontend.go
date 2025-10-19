// Package frontend is a library for serving static files in a Gin web application.
// It supports local static files and embedding static files in binary.
package frontend

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type options struct {
	isUseEmbedFS    bool
	embedFS         embed.FS
	handleContentFn func(content []byte) []byte
	specifiedFile   map[string]struct{}
	is404ToHome     bool
}

func defaultOptions() *options {
	return &options{}
}

// Option set the jwt options.
type Option func(*options)

func (o *options) apply(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

// WithEmbedFS set embedFS to use embed.FS static resources.
func WithEmbedFS(efs embed.FS) Option {
	return func(o *options) {
		o.isUseEmbedFS = true
		o.embedFS = efs
	}
}

// WithHandleContent set function to handle content and specified files
func WithHandleContent(fn func(content []byte) []byte, files ...string) Option {
	return func(o *options) {
		o.handleContentFn = fn
		if len(files) > 0 {
			o.specifiedFile = make(map[string]struct{})
			for _, file := range files {
				o.specifiedFile[file] = struct{}{}
			}
		}
	}
}

// With404ToHome set 404 to home page
func With404ToHome() Option {
	return func(o *options) {
		o.is404ToHome = true
	}
}

// ------------------------------------------------------------------------------------

// FrontEnd is the frontend router configuration
type FrontEnd struct {
	//basePath string // custom prefix route path, default is empty.

	sourceDir string // directory where static files is located, e.g. dist, it's also gin's route path.

	isUseEmbedFS bool     // if true, use embed.FS, otherwise local static file.
	embedFS      embed.FS // embed.FS static resources.

	// only used for EmbedFS, e.g. config.js content, replace apiBaseUrl to backend api address.
	handleContentFn func(content []byte) []byte
	specifiedFile   map[string]struct{} // specified files to handle content, e.g. config.js

	// when request route notfound
	// true: redirect to index.html
	// false: returns 404, default is false.
	is404ToHome bool
}

// New create a new frontend, default use local static file, you can use WithEmbedFS to use embed.FS.
func New(sourceDir string, opts ...Option) *FrontEnd {
	if sourceDir == "" {
		sourceDir = "dist"
	}

	o := defaultOptions()
	o.apply(opts...)

	sourceDir = strings.Trim(sourceDir, "/")
	return &FrontEnd{
		//basePath:        o.basePath,
		sourceDir:       sourceDir,
		isUseEmbedFS:    o.isUseEmbedFS,
		embedFS:         o.embedFS,
		handleContentFn: o.handleContentFn,
		specifiedFile:   o.specifiedFile,
	}
}

// SetRouter set frontend router
func (f *FrontEnd) SetRouter(r *gin.Engine) error {
	// use embed file
	if f.isUseEmbedFS {
		if f.handleContentFn == nil {
			f.setEmbedFSRouter(r)
		} else {
			err := f.saveFSToLocal()
			if err != nil {
				return err
			}
			f.setLocalFileRouter(r)
		}
		return nil
	}

	// use local file
	f.setLocalFileRouter(r)
	return nil
}

func (f *FrontEnd) setEmbedFSRouter(r *gin.Engine) {
	if f.is404ToHome {
		homePage := fmt.Sprintf("%s/index.html", f.sourceDir)
		r.NoRoute(browserRefreshFS(f.embedFS, homePage)) // solve using history route 404 problem
	}

	relativePath := fmt.Sprintf("/%s/*filepath", f.sourceDir)
	r.GET(relativePath, func(c *gin.Context) {
		staticServer := http.FileServer(http.FS(f.embedFS))
		staticServer.ServeHTTP(c.Writer, c.Request)
	})
}

func (f *FrontEnd) setLocalFileRouter(r *gin.Engine) {
	routerPrefixPath := f.sourceDir
	if f.is404ToHome {
		homePage := fmt.Sprintf("%s/index.html", routerPrefixPath)
		r.NoRoute(browserRefresh(homePage)) // solve using history route 404 problem
	}

	relativePath := fmt.Sprintf("/%s/*filepath", routerPrefixPath)
	r.GET(relativePath, func(c *gin.Context) {
		localFileDir := f.sourceDir
		filePath := c.Param("filepath")
		c.File(localFileDir + filePath)
	})
}

func (f *FrontEnd) saveFSToLocal() error {
	_ = os.RemoveAll(f.sourceDir)
	time.Sleep(time.Millisecond * 10)

	// Walk through the embedded filesystem
	return fs.WalkDir(f.embedFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Create the corresponding directory structure locally
		localPath := path
		if d.IsDir() {
			err := os.MkdirAll(localPath, 0755)
			if err != nil {
				return err
			}
		} else {
			// Read the file from the embedded filesystem
			content, err := fs.ReadFile(f.embedFS, path)
			if err != nil {
				return err
			}

			// handle file content
			if len(f.specifiedFile) > 0 {
				for file := range f.specifiedFile {
					if strings.HasSuffix(path, file) {
						content = f.handleContentFn(content)
						break
					}
				}
			} else {
				content = f.handleContentFn(content)
			}

			// Save the content to the local file
			err = os.WriteFile(localPath, content, 0644)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// solve vue using history route 404 problem, for embed.FS
func browserRefreshFS(efs embed.FS, path string) func(c *gin.Context) {
	return func(c *gin.Context) {
		accept := c.Request.Header.Get("Accept")
		flag := strings.Contains(accept, "text/html")
		if flag {
			content, err := efs.ReadFile(path)
			if err != nil {
				c.Writer.WriteHeader(404)
				_, _ = c.Writer.WriteString("Not Found")
				return
			}
			c.Writer.WriteHeader(200)
			c.Writer.Header().Add("Accept", "text/html")
			_, _ = c.Writer.Write(content)
			c.Writer.Flush()
		}
	}
}

// solve vue using history route 404 problem, for system file
func browserRefresh(path string) func(c *gin.Context) {
	return func(c *gin.Context) {
		accept := c.Request.Header.Get("Accept")
		flag := strings.Contains(accept, "text/html")
		if flag {
			content, err := os.ReadFile(path)
			if err != nil {
				c.Writer.WriteHeader(404)
				_, _ = c.Writer.WriteString("Not Found")
				return
			}
			c.Writer.WriteHeader(200)
			c.Writer.Header().Add("Accept", "text/html")
			_, _ = c.Writer.Write(content)
			c.Writer.Flush()
		}
	}
}
