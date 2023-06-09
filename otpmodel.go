package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"time"
)

var DB *gorm.DB

func init() {
	DB = NewDB()
}

func Db() *gorm.DB {
	return DB
}

func NewDB() *gorm.DB {
	//链接数据库
	db, err := gorm.Open("mysql", "root:root@("+DbServer+")/zyoa?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}
	//defer conn.Close()
	//设置连接池
	//设置最大连接数
	db.DB().SetMaxOpenConns(100)
	//最大空闲连接数
	db.DB().SetMaxIdleConns(10)
	//设置链接的最大生命周期
	db.DB().SetConnMaxLifetime(time.Second * 300)
	//打开查询日志
	db.LogMode(true)
	if err := db.DB().Ping(); err != nil {
		fmt.Println("链接数据库失败", err)
	}
	return db
}

type OTPAuth struct {
	Name     string `form:"name" json:"name" gorm:"column:name;type:varchar(255);size:255" binding:"required"`
	Password string `form:"password" json:"password" gorm:"type:varchar(255)"  binding:"required"`
	Seed     string `form:"seed" json:"seed" gorm:"type:varchar(64)" binding:"required"`
	Status   int    `form:"status" json:"status" gorm:"type:tinyint" binding:"required"`
}

func (o *OTPAuth) TableName() string {
	return "otpauth"

}

func (o *OTPAuth) Save() error {
	if err := Db().Save(&o).Error; err != nil {
		return err
	}
	return nil
}

func (o *OTPAuth) GetUser(loginName string) error {
	if err := Db().Where("name=?", loginName).Find(&o).Error; err != nil {
		return err
	}
	return nil
}

func (o *OTPAuth) PasswordUpdate(password string) error {
	if err := Db().Model(&o).Where("name=?", o.Name).Update(OTPAuth{Password: password}).Error; err != nil {
		return err
	}
	return nil
}

func (o *OTPAuth) SaveSeed(loginName, seed string) error {
	if err := Db().Model(&o).Where("name=?", loginName).Update("seed", seed).Error; err != nil {
		return err
	}
	return nil
}
