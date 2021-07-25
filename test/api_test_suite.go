package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/merisho/binaryx-test/activerecord"
	"github.com/merisho/binaryx-test/api"
	"github.com/merisho/binaryx-test/service"
	log "github.com/sirupsen/logrus"
)

type Response struct {
	Code int
	Body []byte
}

type APITestSuite struct {
	server   *gin.Engine
	email    string
	password string
}

func (ts *APITestSuite) Setup() *api.Server {
	srv, _ := ts.createTestAPIServer()

	ts.server = srv.Gin()
	ts.email = fmt.Sprintf("test%d", time.Now().Unix())
	ts.password = "test12345"

	return srv
}

func (ts *APITestSuite) createTestAPIServer() (*api.Server, activerecord.Facade) {
	dbURL := os.Getenv("TEST_DB_URL")
	if dbURL == "" {
		dbURL = "postgres://admin:12345@localhost:5432/test?sslmode=disable"
	}

	db, err := pgxpool.Connect(context.Background(), dbURL)
	if err != nil {
		log.WithError(err).Fatal("could not connect to DB")
	}
	activeRecordFactory := activerecord.New(db)
	serviceWallets, err := service.NewWallets(activeRecordFactory)
	if err != nil {
		log.WithError(err).Fatal("could not connect to DB")
	}

	srv, err := api.NewServer(api.Config{
		JWTSecret: "test",
		APIMode:   api.TestMode,
	}, activeRecordFactory, serviceWallets)

	if err != nil {
		log.Fatal(err)
	}

	return srv, activeRecordFactory
}

func (ts *APITestSuite) Request(method, url string) *Request {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		log.WithError(err).Fatal("could not create a request")
	}

	return &Request{
		server:   ts.server,
		req:      req,
	}
}

type Request struct {
	server   *gin.Engine
	req      *http.Request
	resData  interface{}
	reqData  interface{}
}

func (r *Request) Do() Response {
	r.req.Body = r.prepareRequestData()

	w := httptest.NewRecorder()
	r.server.ServeHTTP(w, r.req)

	b := w.Body.Bytes()
	if r.resData != nil && len(b) > 0 {
		err := json.Unmarshal(b, r.resData)
		if err != nil {
			log.WithError(err).WithField("response", w.Body.String()).Fatal("could not unmarshal response")
		}
	}

	return Response{
		Code: w.Code,
		Body: w.Body.Bytes(),
	}
}

func (r *Request) prepareRequestData() io.ReadCloser {
	if r.reqData == nil {
		return nil
	}

	var res io.Reader
	switch d := r.reqData.(type) {
	case []byte:
		res = bytes.NewReader(d)
	case io.Reader:
		res = d
	default:
		data, err := json.Marshal(d)
		if err != nil {
			log.WithError(err).Fatal("could not marshal a request body")
		}

		res = bytes.NewReader(data)

		r.req.Header.Set("Content-Type", "application/json")
	}

	return io.NopCloser(res)
}

func (r *Request) WithBearerToken(token string) *Request {
	r.req.Header.Set("Authorization", "Bearer " + token)
	return r
}

func (r *Request) WithResponseData(v interface{}) *Request {
	r.resData = v
	return r
}

func (r *Request) WithRequestData(v interface{}) *Request {
	r.reqData = v
	return r
}
