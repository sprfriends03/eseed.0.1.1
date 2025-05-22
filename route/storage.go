package route

import (
	"app/env"
	"app/pkg/ecode"
	"fmt"
	"mime/multipart"
	"net/http"
	"path"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nhnghia272/gopkg"
)

type storage struct {
	*middleware
}

func init() {
	handlers = append(handlers, func(m *middleware, r *gin.Engine) {
		s := storage{m}

		v1 := r.Group("/storage/v1")
		v1.GET("/images/:filename", s.NoAuth(), s.v1_DownloadImage())
		v1.GET("/videos/:filename", s.NoAuth(), s.v1_DownloadVideo())
		v1.POST("/images", s.BearerAuth(), s.v1_UploadImage([]string{"image"}))
		v1.POST("/videos", s.BearerAuth(), s.v1_UploadVideo([]string{"video"}))
	})
}

// @Tags Storage
// @Summary Download Image
// @Security BearerAuth
// @Param filename path string true "filename"
// @Router /storage/v1/images/{filename} [get]
func (s storage) v1_DownloadImage() gin.HandlerFunc {
	return s.v1_Download("images")
}

// @Tags Storage
// @Summary Download Video
// @Security BearerAuth
// @Param filename path string true "filename"
// @Router /storage/v1/videos/{filename} [get]
func (s storage) v1_DownloadVideo() gin.HandlerFunc {
	return s.v1_Download("videos")
}

func (s storage) v1_Download(kind string) gin.HandlerFunc {
	return func(c *gin.Context) {
		filename := c.Param("filename")

		if c.GetHeader("If-None-Match") == filename {
			c.Status(http.StatusNotModified)
			return
		}

		c.Header("ETag", filename)
		c.Header("Cross-Origin-Resource-Policy", "cross-origin")
		c.Header("Cross-Origin-Opener-Policy", "cross-origin")

		if file, err := s.store.Storage.Download(c.Request.Context(), fmt.Sprintf("/%v/%v", kind, filename)); err != nil {
			c.Error(ecode.FileNotFound)
			return
		} else if info, _ := file.Stat(); info.Size == 0 {
			c.Error(ecode.FileNotFound)
			return
		} else {
			c.DataFromReader(http.StatusOK, info.Size, info.Metadata.Get("Content-Type"), file, map[string]string{"Cache-Control": info.Metadata.Get("Cache-Control")})
		}

		// if file, err := s.store.Db.Bucket.OpenDownloadStreamByName(fmt.Sprintf("/%v/%v", kind, filename)); err != nil {
		// 	c.Error(ecode.FileNotFound)
		// 	return
		// } else if info := file.GetFile(); info == nil {
		// 	c.Error(ecode.FileNotFound)
		// 	return
		// } else {
		// 	c.DataFromReader(http.StatusOK, info.Length, info.Metadata.Lookup("Content-Type").StringValue(), file, map[string]string{"Cache-Control": info.Metadata.Lookup("Cache-Control").StringValue()})
		// }
	}
}

// @Tags Storage
// @Summary Upload Image
// @Security BearerAuth
// @Param files formData file true "files"
// @Router /storage/v1/images [post]
func (s storage) v1_UploadImage(exts []string) gin.HandlerFunc {
	return s.v1_Upload("images", exts)
}

// @Tags Storage
// @Summary Upload Video
// @Security BearerAuth
// @Param files formData file true "files"
// @Router /storage/v1/videos [post]
func (s storage) v1_UploadVideo(exts []string) gin.HandlerFunc {
	return s.v1_Upload("videos", exts)
}

func (s storage) v1_Upload(kind string, exts []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		form, err := c.MultipartForm()
		if err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		files := form.File["files"]
		if len(files) == 0 {
			c.Error(ecode.FileNotFound)
			return
		}

		for i := range files {
			if !slices.ContainsFunc(exts, func(e string) bool { return strings.HasPrefix(files[i].Header.Get("Content-Type"), e) }) {
				c.Error(ecode.BadRequest)
				return
			}
		}

		results := gopkg.MapFunc(files, func(file *multipart.FileHeader) string {
			filename := fmt.Sprintf("/%v/%v%v", kind, uuid.NewString(), path.Ext(file.Filename))
			// metadata := options.GridFSUpload().SetMetadata(db.M{"Content-Type": file.Header.Get("Content-Type"), "Cache-Control": "public, max-age=2592000"})
			if stream, err := file.Open(); err != nil {
				panic(err)
			} else if _, err := s.store.Storage.Upload(c.Request.Context(), filename, stream, file.Size, file.Header.Get("Content-Type")); err != nil {
				panic(err)
			}
			//  else if _, err := s.store.Db.Bucket.UploadFromStream(filename, stream, metadata); err != nil {
			// 	panic(err)
			// }
			return fmt.Sprintf("%v%v", env.CdnUri, filename)
		})

		c.JSON(http.StatusOK, results)
	}
}
