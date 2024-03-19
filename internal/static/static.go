package static

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"micli/internal/conf"
	"micli/pkg/util"
	"micli/public"

	"github.com/gin-gonic/gin"
)

var RawIndexHtml string

func InitIndex() {
	index, err := public.Public.ReadFile("dist/index.html")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			util.Log.Fatalf("index.html not exist, you may forget to put dist of frontend to public/dist")
		}
		util.Log.Fatalf("failed to read index.html: %v", err)
	}
	RawIndexHtml = string(index)
	port := conf.Cfg.Section("app").Key("PORT").MustString(":8080")
	basePath := fmt.Sprintf("http://localhost%s", port)
	replaceMap := map[string]string{
		"base_path: undefined": fmt.Sprintf("base_path: '%s'", basePath),
		"api: undefined":       fmt.Sprintf("api: '%s/api'", basePath),
	}
	for k, v := range replaceMap {
		RawIndexHtml = strings.Replace(RawIndexHtml, k, v, 1)
	}
}

func Static(r *gin.RouterGroup, noRoute func(handlers ...gin.HandlerFunc)) {
	InitIndex()
	folders := []string{"assets", "images", "streamer", "static"}
	r.Use(func(c *gin.Context) {
		for i := range folders {
			if strings.HasPrefix(c.Request.RequestURI, fmt.Sprintf("/%s/", folders[i])) {
				c.Header("Cache-Control", "public, max-age=15552000")
			}
		}
	})
	for i, folder := range folders {
		folder = "dist/" + folder
		sub, err := fs.Sub(public.Public, folder)
		if err != nil {
			util.Log.Fatalf("can't find folder: %s", folder)
		}
		r.StaticFS(fmt.Sprintf("/%s/", folders[i]), http.FS(sub))
	}

	noRoute(func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.Status(200)
		_, _ = c.Writer.WriteString(RawIndexHtml)
		c.Writer.Flush()
		c.Writer.WriteHeaderNow()
	})
}
