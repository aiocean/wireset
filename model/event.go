package model

type ShopInstalledEvt struct {
	MyshopifyDomain string
	AccessToken     string
	ShopID          string
}


type ShopCheckedInEvt struct {
	MyshopifyDomain string
	SessionToken    string
}

type ServerStartedEvt struct {
}
