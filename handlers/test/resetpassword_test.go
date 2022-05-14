package test

import (
	"encoding/json"
	mock_database "github.com/decadevs/shoparena/database/mocks"
	"github.com/decadevs/shoparena/handlers"
	"github.com/decadevs/shoparena/models"
	"github.com/decadevs/shoparena/router"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestSendForgotPasswordEMailHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDB := mock_database.NewMockDB(ctrl)
	mockMail := mock_database.NewMockMailer(ctrl)
	h := &handlers.Handler{DB: mockDB, Mail: mockMail}
	route, _ := router.SetupRouter(h)
	resetPassword := struct {
		Email string `json:"email"`
	}{
		Email: "test@testmail.com",
	}
	buyer := models.Buyer{
		User: models.User{
			Email:        "test@gmail.com",
			PasswordHash: "passwordHash",
		},
		Orders: nil,
	}
	secretString := os.Getenv("JWTSECRET")
	privateAPIKey := os.Getenv("MAILGUN_API_KEY")
	yourDomain := os.Getenv("DOMAIN_STRING")
	mockDB.EXPECT().FindBuyerByEmail("test@testmail.com").Return(&buyer, nil)
	mockMail.EXPECT().GenerateNonAuthToken("test@gmail.com", &secretString).Return("test@gmail.com", nil)
	Link := "<strong>Here is your reset <a href='http://localhost:5000/api/v1/password-reset?reset_token=test@gmail.com'>link</a></strong>"
	mockMail.EXPECT().SendMail("forgot Password", Link, "test@gmail.com", privateAPIKey, yourDomain).Return(nil)
	resetPasswordPayload, err := json.Marshal(resetPassword)
	if err != nil {
		log.Println(err)
		t.Fail()
	}
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/buyer/forgotpassword",
		strings.NewReader(string(resetPasswordPayload)))
	route.ServeHTTP(w, req)
	assert.Contains(t, w.Body.String(), "please", "check")
	assert.Equal(t, w.Code, http.StatusOK)
}
func TestForgotPasswordResetHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDB := mock_database.NewMockDB(ctrl)
	mockMail := mock_database.NewMockMailer(ctrl)
	h := &handlers.Handler{DB: mockDB, Mail: mockMail}
	route, _ := router.SetupRouter(h)
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("12345678"), bcrypt.DefaultCost)
	resetPassword := struct {
		NewPassword        string `json:"new_password"`
		ConfirmNewPassword string `json:"confirm_new_password"`
	}{
		NewPassword:        "123456789",
		ConfirmNewPassword: "123456789",
	}
	buyer := models.Buyer{
		User: models.User{
			Email:        "test@gmail.com",
			PasswordHash: string(passwordHash),
		},
		Orders: nil,
	}
	secretString := os.Getenv("JWTSECRET")
	mockMail.EXPECT().DecodeToken(gomock.Any(), secretString).Return(buyer.Email, nil)
	mockDB.EXPECT().FindBuyerByEmail("test@gmail.com").Return(&buyer, nil)
	mockDB.EXPECT().BuyerResetPassword("test@gmail.com", gomock.Any()).Return(&buyer, nil)
	resetPasswordPayload, err := json.Marshal(resetPassword)
	if err != nil {
		log.Println(err)
		t.Fail()
	}
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/buyer/forgotpassword-reset",
		strings.NewReader(string(resetPasswordPayload)))
	route.ServeHTTP(w, req)
	assert.Contains(t, w.Body.String(), "reset", "password")
	assert.Equal(t, w.Code, http.StatusOK)
}
