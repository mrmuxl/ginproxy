package main

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
)

//register
//router := r.Group("/", ginBodyLogMiddleware())
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func ginBodyLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw
		c.Next()

		fmt.Println("Response body: " + blw.body.String())
	}
}

func ginBodyLogMiddleware1(c *gin.Context) {
	blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw
	c.Next()
	statusCode := c.Writer.Status()
	if statusCode >= 400 {
		//ok this is an request with error, let's make a record for it
		// now print body (or log in your preferred way)
		fmt.Println("Response body: " + blw.body.String())
	}
}
