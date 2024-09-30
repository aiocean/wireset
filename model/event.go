package model

import "time"

type ShopInstalledEvt struct {
	MyshopifyDomain string
	AccessToken     string
	ShopID          string
}

type ShopUninstalledEvt struct {
	MyshopifyDomain string
	ShopID          string
	UninstalledAt   time.Time
	Reason          string
}

type ShopCheckedInEvt struct {
	MyshopifyDomain string
	SessionToken    string
}

type ServerStartedEvt struct {
}
