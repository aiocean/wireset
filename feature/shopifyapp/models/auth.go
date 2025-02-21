package models


type AuthData struct {
	AccessToken     string
	MyshopifyDomain string
	ShopID          string
	Iss             string
	Dest            string
	Aud             string
	Sub             string
	Exp             int
	Nbf             int
	Iat             int
	Jti             string
	Sid             string
}
