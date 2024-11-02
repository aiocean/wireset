package model

import "github.com/aiocean/wireset/feature/shopifyapp/models"

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
type Subscription struct {
	GraphqlID string // admin_graphql_api_id
	Status            string
	Name string
	Plan              models.Plan
}

type AppSubscriptionUpdatedEvt struct {
	ShopID          string
	MyshopifyDomain string
	AccessToken     string
	Subscription    Subscription
}
