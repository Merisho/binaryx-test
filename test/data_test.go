package test

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/merisho/binaryx-test/api"
)

var rnd = rand.New(rand.NewSource(time.Now().Unix()))

func DefaultSignupRequest() api.SignupRequest {
	return api.SignupRequest{
		Email: fmt.Sprintf("test%d@example.com", rnd.Int()),
		FirstName: "Test",
		LastName: "User",
		Password: "12345678",
	}
}
