package model

type ShopLoggedInEvt struct {
	ShopID          string
	MyshopifyDomain string
	AccessToken     string
}

type ShopWithoutSubscriptionFoundEvt struct {
	ShopID          string
	MyshopifyDomain string
	AccessToken     string
}
