package main

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

func otpStart(c *gin.Context) {
	session := sessions.Default(c)
	//cookie := session.Get("mycookie")
	//session.Delete("mycookie")
	//session.Save()
	if c.Request.Method == "GET" {
		login_username := session.Get("login_username")
		c.HTML(http.StatusOK, "otpStart.html", gin.H{
			"title":          "验证身份",
			"login_username": login_username,
		})

	} else {
		login_username := c.PostForm("login_username")
		login_password := c.PostForm("login_password")
		//log.Println(login_username, login_password)
		var otp OTPAuth
		if err := otp.GetUser(login_username); err != nil {
			c.HTML(http.StatusOK, "otpStart.html", gin.H{
				"title":          "验证身份",
				"login_username": login_username,
				"status":         true,
			})
		} else if otp.Password == login_password {
			c.Redirect(302, "/auth/otp/install")

		} else {
			c.HTML(http.StatusOK, "otpStart.html", gin.H{
				"title":          "验证身份",
				"login_username": login_username,
				"status":         true,
			})
		}
	}
}

func otpInstall(c *gin.Context) {
	session := sessions.Default(c)
	login_username := session.Get("login_username")
	c.HTML(http.StatusOK, "otpInstallApp.html", gin.H{
		"title":          "应用安装",
		"login_username": login_username,
		"qr_android":     "https://corp.zq.com/static/app/googleauthenticator.apk",
		"qr_ios":         "https://apps.apple.com/cn/app/google-authenticator/id388497605",
	})
}

func otpBind(c *gin.Context) {
	session := sessions.Default(c)
	login_username := session.Get("login_username")
	var otpauth OTPAuth
	var username string
	if name, ok := login_username.(string); ok {
		username = url.PathEscape(name)
	}
	var otpconf OTPConfig
	var secret string
	otpconf.WindowSize = 3
	otpconf.HotpCounter = 0
	if c.Request.Method == "GET" {
		//产生Seed
		secret = RandomSecret(16)
		otpconf.Secret = secret
		session.Set("seed", secret)
		session.Save()
		otpURL := otpconf.ProvisionURIWithIssuer(username, "lawtonfz")
		//log.Println("otpURL", otpURL, secret)
		if err := otpauth.GetUser(login_username.(string)); err != nil {
			log.Println(err)
			c.Redirect(302, "/")
		}
		if otpauth.Seed != "" {
			c.Redirect(302, "/")
		}
		c.HTML(http.StatusOK, "otpBind.html", gin.H{
			"title":          "App绑定",
			"login_username": login_username,
			"secret":         secret,
			"otpurl":         otpURL,
		})

	} else {
		//存储Seed
		otpCode := c.PostForm("otp_code")
		seed := session.Get("seed")
		if s, ok := seed.(string); ok {
			secret = s
			otpconf.Secret = s
			//log.Println("secret:", secret)
		}
		//log.Println("otpCode:", otpCode)
		if otpCode != "" && len(otpCode) == 6 {
			t, err := otpconf.Authenticate(otpCode)
			if err != nil {
				log.Println(err)
			}
			if t {
				//绑定成功,写Seed
				var otpauth OTPAuth
				if err := otpauth.SaveSeed(login_username.(string), secret); err != nil {
					log.Println("SaveSeed", err)
					c.Redirect(302, "/auth/otp/bind")
				}
				c.Redirect(302, "/auth/otp/redirect")

			} else {
				//bind失败
				log.Println("绑定失败")
				c.Redirect(302, "/auth/otp/bind")
			}

		} else {
			//otpCode不对
			log.Println("otpCode is wrong!")
			c.Redirect(302, "/auth/otp/bind")
		}

	}
}

func otpAuth(c *gin.Context) {
	session := sessions.Default(c)
	login_username := session.Get("login_username")
	var otpconf OTPConfig
	var otpauth OTPAuth
	otpconf.WindowSize = 3
	otpconf.HotpCounter = 0
	if c.Request.Method == "GET" {
		//
		c.HTML(http.StatusOK, "otpAuth.html", gin.H{
			"title":          "验证动态密码",
			"login_username": login_username,
		})

	} else {
		//校验动态密码
		otpCode := c.PostForm("otp_code")
		if username, ok := login_username.(string); ok {
			if err := otpauth.GetUser(username); err != nil {
				log.Println(err)
				//无用户名，重定向到/
				c.Redirect(302, "/")
			}
			seed := otpauth.Seed
			otpconf.Secret = seed
		}
		if otpCode != "" && len(otpCode) == 6 {
			t, err := otpconf.Authenticate(otpCode)
			if err != nil {
				log.Println(err)
			}
			if t {
				//true 登录
				//不允许seeyon1这个用户登录
				log.Println("status:", otpauth.Status)
				if otpauth.Status == 1 {
					c.Redirect(302, "/")
					return
				}
				if err := SeeyonLogin(otpauth.Name, otpauth.Password, session); err != nil {
					log.Println("SeeyonLogin:", err)
					//登录失败跳转到首页
					c.Redirect(302, "/")
					return
				}
				c.Redirect(302, "/seeyon/indexOpenWindow.jsp")
			} else {
				//false 重新输入
				c.Redirect(302, "/auth/otp/auth")
			}

		} else {
			//错误
			c.Redirect(302, "/auth/otp/auth")
		}

	}
}

func otpRedirect(c *gin.Context) {
	c.HTML(http.StatusOK, "otpRedirect.html", gin.H{"title": "页面跳转"})

}

func SeeyonLogin(login_username, login_password string, session sessions.Session) error {
	loginUrl := baseURL + ":81/seeyon/main.do?method=login"
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Println(err)
		return err
	}
	client := &http.Client{Jar: jar}
	var data = url.Values{}
	data.Add("login_username", login_username)
	data.Add("login_password", login_password)
	req, err := http.NewRequest("POST", loginUrl, strings.NewReader(data.Encode()))
	if err != nil {
		log.Printf("SeeyonLogin Request Error:%v\n", err)
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:90.0) Gecko/20100101 Firefox/90.1")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		log.Printf("Seeyon Post Error:%v\n", err)
		return err
	}
	session.Set("mycookie", resp.Request.Header.Get("Cookie"))
	session.Save()
	return nil
}
