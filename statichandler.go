package lemon

import (
	"lemon/utils"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"
)

type StaticFileHandler struct {
	RequestHandler
	ModifiedTime time.Time
	CACHEMAXAGE  int
	root         string
	Extension    string
}

func (sh *StaticFileHandler) Initialize(params Dictionary) {
	sh.CACHEMAXAGE = 86400 * 365 * 10
	sh.root = params["path"].(string)
}

func (sh *StaticFileHandler) Get(Path string) {
	sh.Extension = filepath.Ext(Path)
	if sh.Request.Method() != "GET" && sh.Request.Method() != "HEAD" {
		http.NotFound(sh.ResponseWriter, sh.Request.Request)
		return
	}
	//requestPath := path.Clean(sh.Request.Url())
	//file := path.Join(sh.application.AbsWorkPath, sh.application.StaticPath, Path)
	file := path.Join(sh.root, Path)
	fileStat, err := os.Stat(file)
	if err != nil {
		http.NotFound(sh.ResponseWriter, sh.Request.Request)
		return
	}

	sh.ModifiedTime = fileStat.ModTime()
	sh.SetHeaders()
	var contentEncoding string
	if sh.application.IsGzip {
		contentEncoding = getAcceptEncodingZip(sh.Request.Request)
		memzipfile, err := openMemZipFile(file, contentEncoding)
		if err != nil {
			http.NotFound(sh.ResponseWriter, sh.Request.Request)
			return
		}

		if contentEncoding == "gzip" {
			sh.SetHeader("Content-Encoding", "gzip")
		} else if contentEncoding == "deflate" {
			sh.SetHeader("Content-Encoding", "deflate")
		} else {
			sh.SetHeader("Content-Length", strconv.FormatInt(fileStat.Size(), 10))
		}

		http.ServeContent(sh.ResponseWriter, sh.Request.Request, file, fileStat.ModTime(), memzipfile)

	} else {
		http.ServeFile(sh.ResponseWriter, sh.Request.Request, file)
	}

}

func (sh *StaticFileHandler) SetHeaders() {
	sh.SetHeader("Accept-Ranges", "bytes")

	sh.SetHeader("Last-Modified", string(utils.AppendTime([]byte{}, sh.ModifiedTime)))
	sh.SetHeader("Date", string(utils.AppendTime([]byte{}, time.Now().UTC())))
	ContentType := sh.GetContentType()
	if len(ContentType) != 0 {
		sh.SetHeader("Content-Type", ContentType)
	}

	//	CacheTime := sh.GetCaheTime()
	//	if CacheTime > 0 {
	//		expiresTime := time.Now().UTC().Add(time.Duration(CacheTime) * time.Second)
	//		exipiresTimeString := string(utils.AppendTime([]byte{}, expiresTime))
	//		sh.SetHeader("Expires", exipiresTimeString)
	//	} else {
	//		sh.SetHeader("Cache-Control", "max-age=0")
	//	}
}

func (sh *StaticFileHandler) GetCaheTime() int {
	version := sh.GetArgument("v")
	if len(version) == 0 {
		return 0
	} else {
		return sh.CACHEMAXAGE
	}
}

func (sh *StaticFileHandler) GetContentType() string {
	ctype := mime.TypeByExtension(sh.Extension)
	return ctype
}
