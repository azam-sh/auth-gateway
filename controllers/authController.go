package controllers

import (
	"authgateway/initializers"
	"authgateway/models"
	"authgateway/token"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:7000/callback",
		ClientID:     "641954038333-g5g3b3ls6g4ois400mvbm35luue4mm91.apps.googleusercontent.com",
		ClientSecret: "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",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
	randomState = "random"
)

func Signup(c *gin.Context) {
	var body models.RequestUser

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
	user := models.User{Login: body.Login, Password: string(hash), FullName: body.FullName, Active: body.Active, RoleID: body.RoleID, Phone: body.Phone}

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

	token, err := token.GenerateToken(uint(user.ID), uint(user.RoleID))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

func Home(c *gin.Context) {
	var html = `<html><body><a href="/google/login">Google Log In</a></body></html>`
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

func GoogleLogin(c *gin.Context) {
	url := googleOauthConfig.AuthCodeURL(randomState)
	http.Redirect(c.Writer, c.Request, url, http.StatusTemporaryRedirect)
}

func Callback(c *gin.Context) {
	if c.Request.FormValue("state") != randomState {
		fmt.Println("state is not valid")
		http.Redirect(c.Writer, c.Request, "/", http.StatusTemporaryRedirect)
		return
	}

	googleToken, err := googleOauthConfig.Exchange(context.Background(), c.Request.FormValue("code"))
	if err != nil {
		fmt.Printf("could not get google token: %s\n", err.Error())
		http.Redirect(c.Writer, c.Request, "/", http.StatusTemporaryRedirect)
		return
	}

	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + googleToken.AccessToken)
	if err != nil {
		fmt.Printf("could not create get request: %s\n", err.Error())
		http.Redirect(c.Writer, c.Request, "/", http.StatusTemporaryRedirect)
		return
	}

	defer resp.Body.Close()
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("could not parse response: %s\n", err.Error())
		http.Redirect(c.Writer, c.Request, "/", http.StatusTemporaryRedirect)
		return
	}

	var client models.Client

	if err := json.Unmarshal(content, &client); err != nil {
		log.Println("Can not unmarshal JSON", err.Error())
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("qwerty"), 10)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to hash password",
		})
		return
	}

	var user models.User
	initializers.DB.First(&user, "login = ?", client.Email)

	if user.ID == 0 {

		randID := rand.Int63()

		newUser := models.User{
			ID:       randID,
			FullName: client.Name,
			Login:    client.Email,
			Active:   true,
			Password: string(hash),
			RoleID:   2,
		}

		result := initializers.DB.Create(&newUser)
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Failed to create user",
			})
			return
		}

		myToken, err := token.GenerateToken(uint(newUser.ID), uint(newUser.RoleID))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Could not generate token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"token": myToken,
		})
	} else {
		myToken, err := token.GenerateToken(uint(user.ID), uint(user.RoleID))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Could not generate token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"token": myToken,
		})
	}
}
