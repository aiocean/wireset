package model

type AuthResponse struct {
	Message           string `json:"message,omitempty"`
	AuthenticationUrl string `json:"authenticationUrl,omitempty"`
}

type ShopifyToken struct {
	ShopID      string `json:"shopId" firestore:"shopId"`
	AccessToken string `json:"accessToken" firestore:"accessToken"`
}