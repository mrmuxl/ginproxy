package main

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
)

func DumpRquestMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//data,err:=httputil.DumpRequestOut(ctx.Request,true)
		//if err!=nil{
		//	log.Println(err.Error())
		//}
		data, err := ctx.GetRawData()
		if err != nil {
			fmt.Println(err.Error())
		}
		log.Printf("RquestHeader%v\n", ctx.Request.Header)
		log.Printf("RawRquest: %v\n", string(data))

		ctx.Request.Body = io.NopCloser(bytes.NewBuffer(data)) // 关键点
		ctx.Next()
	}
}
