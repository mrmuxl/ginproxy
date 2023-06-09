package main

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	//gindump "github.com/tpkeeper/gin-dump"
)

var store = cookie.NewStore([]byte("gin_seeyon"))

func main() {
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	f, _ := os.Create("main.log")
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
	router := gin.Default()
	router.Use(sessions.Sessions("sysessionid", store))
	//router.Use(DumpRquestMiddleware())
	router.Any("/", ReverseProxy)
	router.Any("/mobile_portal/*path", ReverseProxy)
	router.Any("/seeyon/*path", HandleProxy)
	router.StaticFS("/static", http.Dir("./static"))
	router.LoadHTMLGlob("tpl/*")
	g := router.Group("/auth")
	{
		g.GET("/otp/start", otpStart)
		g.POST("/otp/start", otpStart)
		g.GET("/otp/install", otpInstall)
		g.GET("/otp/bind", otpBind)
		g.POST("/otp/bind", otpBind)
		g.GET("/otp/auth", otpAuth)
		g.POST("/otp/auth", otpAuth)
		g.GET("/otp/redirect", otpRedirect)

	}
	err := router.Run(":80")
	if err != nil {
		panic(err)
	}
}
