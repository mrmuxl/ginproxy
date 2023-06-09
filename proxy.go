package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-contrib/sessions"
	"io"
	"strings"

	//"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

const DEBUG = 0

const MyServer = "127.0.0.1"
const DbServer = "192.168.165.11"

// const MyServer = "192.168.165.11"
const baseURL = "http://" + MyServer

func HandleProxy(c *gin.Context) {
	session := sessions.Default(c)
	var otp OTPAuth
	cReqBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Println(err)
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(cReqBody))
	if c.Request.URL.RawQuery == "method=login" {
		//log.Println("RequestURI:",c.Request.RequestURI)
		//authorization:=c.PostForm("authorization")
		//timezone:=c.PostForm("login.timezon")
		//province:=c.PostForm("province")
		//city:=c.PostForm("city")
		//rectangle:=c.PostForm("rectangle")
		//trustdo_type:=c.PostForm("trustdo_type")
		login_username := c.PostForm("login_username")
		login_password := c.PostForm("login_password")
		//login_validatePwdStrength:=c.PostForm("login_validatePwdStrength")
		//random:=c.PostForm("random")
		//fontSize:=c.PostForm("fontSize")
		//screenWidth:=c.PostForm("screenWidth")
		//screenHeight:=c.PostForm("screenHeight")
		////
		//log.Println(authorization,timezone,province,city,rectangle,login_username,trustdo_type,login_password,login_validatePwdStrength,random,fontSize,screenWidth,screenHeight)

		//c.Redirect(302,"/auth/otp/start")
		token, err := GetToken()
		if err != nil {
			log.Printf("Get Token Error:%v\n", err)
		}
		//通过OA的API校验用户名和密码的正确性
		v, err := VerifyUser(login_username, login_password, token)
		if err != nil {
			log.Printf("VerfiyUser Error:%v\n", err)
		}
		if v == true {
			//开始注册
			session.Set("login_username", login_username)
			err := session.Save()
			if err != nil {
				return
			}
			log.Printf("Right Way!")
			if err := otp.GetUser(login_username); err != nil {
				log.Printf("otp.GetUser Error:%v\n", err)
				otp.Name = login_username
				otp.Password = login_password
				otp.Seed = ""
				otp.Status = 0
				if err := otp.Save(); err != nil {
					log.Printf("otp.Save Error:%v\n otp:%v\n", err, otp)
				}
				c.Redirect(302, "/auth/otp/start")
			} else {
				if login_password != otp.Password { //此时的登陆密码已经经过校验了，所以本地数据库存储密码过期，故更新。
					log.Println("Update Password !")
					if err := otp.PasswordUpdate(login_password); err != nil {
						log.Printf("otp.PasswordUpdate Error:%v\n", err)
					}
				}
				//检查用户有没有Seed，如果没有开始认证流程
				if otp.Seed == "" {
					c.Redirect(302, "/auth/otp/start")

				} else if otp.Status == 2 {
					//白名单，检测账号的权限状态，值为2直接定向到首页。
					if err := SeeyonLogin(otp.Name, otp.Password, session); err != nil {
						log.Println("SeeyonLogin:", err)
						//登录失败跳转到首页
						c.Redirect(302, "/")
						return
					}
				} else {
					//seed不为空,重定向开始认证，这里可以优化逻辑
					c.Redirect(302, "/auth/otp/auth")

				}
			}
		} else {
			//用户名和密码不正确,直接走原始路径
			log.Println("Are you sure!")
			c.Request.Body = io.NopCloser(bytes.NewBuffer(cReqBody))
			ReverseProxy(c)
		}

	} else {
		log.Println("NOP!")
		//log.Println("RequestURI:",c.Request.RequestURI)
		mycookie := session.Get("mycookie")
		if cookie, ok := mycookie.(string); ok {
			c.Request.Header.Set("Cookie", cookie)
		}
		if c.Request.URL.RawQuery == "method=logout" {
			session.Clear()
			err := session.Save()
			if err != nil {
				return
			}
		}
		ReverseProxy(c)
	}

}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func joinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}
	// Same as singleJoiningSlash, but uses EscapedPath to determine
	// whether a slash should be added
	apath := a.EscapedPath()
	bpath := b.EscapedPath()

	aslash := strings.HasSuffix(apath, "/")
	bslash := strings.HasPrefix(bpath, "/")

	switch {
	case aslash && bslash:
		return a.Path + b.Path[1:], apath + bpath[1:]
	case !aslash && !bslash:
		return a.Path + "/" + b.Path, apath + "/" + bpath
	}
	return a.Path + b.Path, apath + bpath
}

func ReverseProxy(c *gin.Context) {
	target := baseURL + ":81"
	targetUrl, err := url.Parse(target)
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(targetUrl)
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.URL.Scheme = targetUrl.Scheme
		req.URL.Host = targetUrl.Host
		req.URL.Path, req.URL.RawPath = joinURLPath(targetUrl, req.URL)
		if targetUrl.RawQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetUrl.RawQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetUrl.RawQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
		//req.Header.Set("X-FORWARDED-FOR", c.ClientIP())
		req.Header.Set("X-REAL-IP", c.ClientIP())
		req.Header.Set("X-FORWARD-HOST", targetUrl.Host)
		if DEBUG == 1 {
			dumpreq, err := httputil.DumpRequest(req, true)
			if err != nil {
				log.Printf("dump req error: %v\n", err)
			}
			log.Printf("[DEBUG] req request: %q", dumpreq)
		}
	}
	if DEBUG == 1 {
		dumpc, err := httputil.DumpRequest(c.Request, true)
		if err != nil {
			log.Printf("[DUBEG]dump c.Request error:%v\n", err)
		}
		log.Printf("[DEBUG] c.Request: %q", dumpc)

	}
	proxy.ModifyResponse = func(response *http.Response) error {
		if DEBUG == 1 {
			dumpresp, err := httputil.DumpResponse(response, false)
			if err != nil {
				log.Printf("[DEBUG] resp err:%v\n", err)
			}
			log.Printf("[DEBUG] Resp: %q", dumpresp)
		}
		return nil
	}

	proxy.ServeHTTP(c.Writer, c.Request)

}

type RestAuth struct {
	UserName string `json:"userName"`
	Password string `json:"password"`
}

type TokenResult struct {
	Id          string `json:"id"`
	BindingUser string `json:"bindingUser"`
}

func GetToken() (restToken string, err error) {
	var token TokenResult
	tokenURL := baseURL + ":81/seeyon/rest/token"
	client := http.Client{}

	jsondata, err := json.Marshal(&RestAuth{
		UserName: "restmfa",
		Password: "fabf1e55-b624-4c11-b61b-9d4fba537e13",
	})
	if err != nil {
		log.Printf("Json Marshal Error:%v\n", err)
		return "", err
	}
	req, err := http.NewRequest("POST", tokenURL, bytes.NewBuffer(jsondata))
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(req.Body)
	if err != nil {
		log.Printf("Get Token Error:%v\n", err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Http Post Error:%v\n", err)
		return "", err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Token ReadAll Error:%v\n", err)
		return "", err
	}
	if err := json.Unmarshal(body, &token); err != nil {
		log.Printf("Json Unmarshal Error%v\n", err)
		return "", err
	}
	restToken = token.Id
	if restToken != "" {
		return restToken, nil
	} else {
		return "", errors.New("Unknown Error,Mybe the token is nil")
	}
}

func VerifyUser(login_username, login_password, token string) (bool, error) {
	VerifyURL := baseURL + ":81/seeyon/rest/orgMember/effective/loginName/" + url.QueryEscape(login_username) + "?password=" + url.QueryEscape(login_password) + "&token=" + token
	//log.Printf("URL:%s\n", VerifyURL)
	resp, err := http.Get(VerifyURL)
	if err != nil {
		log.Printf("Verify loginName:%v\n", err)
		return false, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Request VerifyURl Error:%v\n", err)
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Request VerifyURl Error:%v\n", err)
		return false, err
	}
	if string(body) == "true" {
		return true, nil
	} else {
		return false, errors.New("unknown Error")
	}
}
