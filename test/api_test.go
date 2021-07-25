package test

import (
	"strings"
	"testing"

	"github.com/merisho/binaryx-test/api"
	"github.com/stretchr/testify/suite"
)

func TestAPI(t *testing.T) {
	suite.Run(t, &FakeCoinsAPITestSuite{
		APITestSuite: &APITestSuite{},
	})
}

type FakeCoinsAPITestSuite struct {
	*APITestSuite
	suite.Suite
	testUserEmail string
	testUserPassword string
	testToken string
}

func (ts *FakeCoinsAPITestSuite) SetupSuite() {
	ts.Setup()
}

func (ts *FakeCoinsAPITestSuite) TestAPI() {
	ts.Run("sign up", ts.testSignUp)
	ts.Run("sign in", ts.testSignIn)
}

func (ts *FakeCoinsAPITestSuite) testSignUp() {
	ts.Run("success", ts.testSignUpSuccess)
	ts.Run("password is less than 8 characters", ts.testShortPassword)
	ts.Run("password is greater than 50 characters", ts.testLongPassword)
	ts.Run("email is invalid",  ts.testEmailIsInvalid)
	ts.Run("email domain does not exist",  ts.testEmailDomainDoesNotExist)
	ts.Run("first name or last name is invalid", ts.testInvalidNames)
}

func (ts *FakeCoinsAPITestSuite) testSignUpSuccess() {
	request := DefaultSignupRequest()
	ts.testUserEmail = request.Email
	ts.testUserPassword = request.Password

	var response api.SignupResponse
	res := ts.Request("POST", "/signup").
		WithRequestData(request).
		WithResponseData(&response).
		Do()
	ts.Equal(201, res.Code)
	userID := response.ID
	ts.NotEmpty(userID)
	ts.Len(response.Wallets, 2)

	for _, w := range response.Wallets {
		ts.Equal(userID, w.UserID)
		ts.Equal("100", w.Balance)
	}
}

func (ts *FakeCoinsAPITestSuite) testShortPassword() {
	request := DefaultSignupRequest()
	request.Password = "12345"

	var apiError api.ErrorResponse
	res := ts.Request("POST", "/signup").
		WithRequestData(request).
		WithResponseData(&apiError).
		Do()
	ts.Equal(400, res.Code)
	ts.Equal("invalid password", apiError.Error)
}

func (ts *FakeCoinsAPITestSuite) testLongPassword() {
	request := DefaultSignupRequest()
	request.Password = strings.Repeat("1", 100)

	var apiError api.ErrorResponse
	res := ts.Request("POST", "/signup").
		WithRequestData(request).
		WithResponseData(&apiError).
		Do()
	ts.Equal(400, res.Code)
	ts.Equal("invalid password", apiError.Error)
}

func (ts *FakeCoinsAPITestSuite) testEmailIsInvalid() {
	request := DefaultSignupRequest()
	request.Email = "q_1s12s_.@3fjjk@example.com"

	var apiError api.ErrorResponse
	res := ts.Request("POST", "/signup").
		WithRequestData(request).
		WithResponseData(&apiError).
		Do()
	ts.Equal(400, res.Code)
	ts.Equal("invalid email", apiError.Error)
}

func (ts *FakeCoinsAPITestSuite) testEmailDomainDoesNotExist() {
	request := DefaultSignupRequest()
	request.Email = "test@asdfqejnviersdvb.com"

	var apiError api.ErrorResponse
	res := ts.Request("POST", "/signup").
		WithRequestData(request).
		WithResponseData(&apiError).
		Do()
	ts.Equal(400, res.Code)
	ts.Equal("invalid email", apiError.Error)
}

func (ts *FakeCoinsAPITestSuite) testInvalidNames() {
	request := DefaultSignupRequest()
	request.FirstName = "123 Alexei"

	var apiError api.ErrorResponse
	res := ts.Request("POST", "/signup").
		WithRequestData(request).
		WithResponseData(&apiError).
		Do()
	ts.Equal(400, res.Code)
	ts.Equal("invalid first name", apiError.Error)

	request = DefaultSignupRequest()
	request.LastName = "Torunov 123"
	res = ts.Request("POST", "/signup").
		WithRequestData(request).
		WithResponseData(&apiError).
		Do()
	ts.Equal(400, res.Code)
	ts.Equal("invalid last name", apiError.Error)
}

func (ts *FakeCoinsAPITestSuite) testSignIn() {
	ts.Run("retrieve token", ts.testRetrieveToken)
}

func (ts *FakeCoinsAPITestSuite) testRetrieveToken() {
	req := api.TokenRequest{
		Email: ts.testUserEmail,
		Password: ts.testUserPassword,
	}

	var tokenRes api.TokenResponse
	res := ts.Request("POST", "/token").
		WithRequestData(req).
		WithResponseData(&tokenRes).
		Do()
	ts.Equal(200, res.Code)
	ts.NotEmpty(tokenRes.Token)
	ts.NotEmpty(tokenRes.ExpiresAt)

	ts.testToken = tokenRes.Token
}
