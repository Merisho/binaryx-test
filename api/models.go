package api

type SignupRequest struct {
	Email string `json:"email"`
	FirstName string `json:"firstName"`
	LastName string `json:"lastName"`
	Password string `json:"password"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SignupResponse struct {
	ID string `json:"id"`
	Email string `json:"email"`
	FirstName string `json:"firstName"`
	LastName string `json:"lastName"`
	Wallets []WalletResponse `json:"wallets"`
}

type WalletResponse struct{
	UserID string `json:"userId"`
	Address string `json:"address"`
	Currency string `json:"currency"`
	Balance string `json:"balance"`
}
