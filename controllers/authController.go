package controllers

import (
	"authgateway/initializers"
	"authgateway/models"
	"authgateway/token"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func Signup(c *gin.Context) {
	var body struct {
		FullName string `json:"fullName"`
		Login    string `json:"login"`
		Password string `json:"password"`
		RoleID   int64  `json:"roleId"`
		AccessID int64  `json:"accessId"`
		Active   bool   `json:"active"`
		Phone    string `json:"phone"`
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to hash password",
		})
		return
	}
	user := models.User{Login: body.Login, Password: string(hash), AccessID: body.AccessID, FullName: body.FullName, Active: body.Active, RoleID: body.RoleID, Phone: body.Phone}
	result := initializers.DB.Create(&user)

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user created!",
	})
}

func Login(c *gin.Context) {
	var body struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}
	var user models.User
	initializers.DB.First(&user, "login = ?", body.Login)

	if user.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid login or password",
		})
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid login or password",
		})
		return
	}

	token, err := token.GenerateToken(uint(user.ID))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}
