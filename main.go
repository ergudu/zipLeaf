package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"os"
)

func main() {
	r := gin.Default()

	r.GET("/download", func(c *gin.Context) {
		zipName := "xxx/ooo.zip"
		f, err := os.Open(zipName)
		if err != nil {
			log.Fatal(err)
		}

		c.Writer.Header().Add("Content-Disposition", fmt.Sprintf("attachment;filename=%s", "window.zip"))
		c.Writer.Header().Set("Content-Type", "application/octet-stream")

		io.Copy(c.Writer, f)
	})

	log.Fatal(r.Run(":8080"))
}
