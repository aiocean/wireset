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

type Order struct {
	ID string
}

type OrderCreatedEvt struct {
	ShopID          string
	MyshopifyDomain string
	AccessToken     string
	Order           Order
}