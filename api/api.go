package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/merisho/binaryx-test/activerecord"
	"github.com/merisho/binaryx-test/service"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

const (
	fakeBTC = "fBTC"
	fakeETH = "fETH"
)

type Mode string
const (
	TestMode Mode = gin.TestMode
	DebugMode Mode = gin.DebugMode
	ReleaseMode Mode = gin.ReleaseMode
)

type Config struct {
	JWTSecret string
	APIMode Mode
}

func NewServer(config Config, activeRecordFactory activerecord.Facade, serviceWallets *service.Wallets) (*Server, error) {
	m := ReleaseMode
	if config.APIMode != "" {
		m = config.APIMode
	}

	gin.SetMode(string(m))

	s := &Server{
		config: config,
		activeRecords: activeRecordFactory,
		gin: gin.Default(),
		serviceWallets: serviceWallets,
	}

	s.initEndpoints()

	return s, nil
}

type Server struct {
	config Config
	gin *gin.Engine
	activeRecords activerecord.Facade
	serviceWallets *service.Wallets
}

func (s *Server) Gin() *gin.Engine {
	return s.gin
}

func (s *Server) initEndpoints() {
	s.gin.Handle(http.MethodPost, "/signup", s.signup)
}

func (s *Server) signup(ctx *gin.Context) {
	var req SignupRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	user, err := s.activeRecords.User().New(req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		switch err.(type) {
		case activerecord.ValidationError:
			ctx.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		default:
			ctx.AbortWithStatus(http.StatusInternalServerError)
			log.WithError(err).Error("could not create new user")
		}

		return
	}

	wallets, err := user.CreateWallets(fakeBTC, fakeETH)
	if err != nil {
		log.WithError(err).Error("could not create wallets")
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	for _, w := range wallets {
		serviceWallet := s.serviceWallets.Get(w.Currency())
		if serviceWallet == nil {
			log.Errorf("no service wallet for currency %s", w.Currency())
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		_, err := w.AcceptTransaction(serviceWallet, decimal.NewFromInt(100))
		if err != nil {
			log.WithError(err).Error("could not create transaction for wallet")
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}

	tx, err := s.activeRecords.Tx(ctx)
	if err != nil {
		log.WithError(err).Error("CRITICAL: could not begin transaction")
		ctx.AbortWithStatus(http.StatusInternalServerError)
	}

	err = user.Save(ctx)
	if err != nil {
		switch err.(type) {
		case activerecord.ConflictError:
			ctx.AbortWithStatusJSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
		default:
			ctx.AbortWithStatus(http.StatusInternalServerError)
			log.WithError(err).Error("could not save new user with wallets and transactions")
		}

		err = tx.Rollback(ctx)
		if err != nil {
			log.WithError(err).Error("CRITICAL: could not rollback transaction")
		}

		return
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.WithError(err).Error("CRITICAL: could not commit transaction")
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var walletsRes []WalletResponse
	for _, w := range wallets {
		walletsRes = append(walletsRes, WalletResponse{
			UserID:  w.UserID().String(),
			Address:  w.Address(),
			Currency: w.Currency(),
			Balance:  w.Balance().String(),
		})
	}

	ctx.JSON(http.StatusCreated, SignupResponse{
		ID: user.ID().String(),
		Email: user.Email(),
		FirstName: user.FirstName(),
		LastName: user.LastName(),
		Wallets: walletsRes,
	})
}
