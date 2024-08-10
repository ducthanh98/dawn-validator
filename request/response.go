package request

import "time"

type LoginResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Token     string `json:"token"`
		UserID    string `json:"user_id"`
		ID        string `json:"_id"`
		Mobile    string `json:"mobile"`
		Email     string `json:"email"`
		Role      string `json:"role"`
		Firstname string `json:"firstname"`
		Lastname  string `json:"lastname"`
		Wallet    struct {
			ID               string `json:"_id"`
			Email            string `json:"email"`
			WalletAddress    string `json:"wallet_address"`
			WalletPrivateKey string `json:"wallet_private_key"`
			WalletDetails    struct {
				Message    string `json:"message"`
				Mnemonic   string `json:"Mnemonic"`
				Address    string `json:"Address"`
				PrivateKey string `json:"PrivateKey"`
			} `json:"wallet_details"`
			Active    bool      `json:"active"`
			CreatedAt time.Time `json:"createdAt"`
			UpdatedAt time.Time `json:"updatedAt"`
			V         int       `json:"__v"`
		} `json:"wallet"`
		ReferralCode string `json:"referralCode"`
	} `json:"data"`
	Servername string `json:"servername"`
}
