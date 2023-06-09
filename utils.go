package main

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"runtime"
	"time"
)

// generate a random secret of given length
func RandomSecret(length int) string {
	goVer := "1.19.9"
	localVer := runtime.Version()
	//适配golang 1.20版本
	if VersionCompare(goVer, localVer, "le") {
		rand.Seed(time.Now().UnixNano())
	} else {
		r := rand.NewSource(time.Now().UnixNano())
		r.Seed(time.Now().UnixNano())
	}
	letterRunes := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ234567")

	bytes := make([]rune, length)

	for i := range bytes {
		bytes[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(bytes)
}

func Md5sum(s string) string {
	m := md5.New()
	m.Write([]byte(s))
	return hex.EncodeToString(m.Sum(nil))
}
