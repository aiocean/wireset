package shopifysvc

import (
	"strings"

	"github.com/tidwall/gjson"
)

type Shop struct {
	ID                   string `json:"id" firestore:"id" bson:"id"`
	Name                 string `json:"name" firestore:"name" bson:"name"`
	Email                string `json:"email" firestore:"email" bson:"email"`
	CountryCode          string `json:"countryCode" firestore:"countryCode" bson:"countryCode"`
	Domain               string `json:"domain" firestore:"domain" bson:"domain"`
	MyshopifyDomain      string `json:"myshopifyDomain" firestore:"myshopifyDomain" bson:"myshopifyDomain"`
	TimezoneAbbreviation string `json:"timezoneAbbreviation" firestore:"timezoneAbbreviation" bson:"timezoneAbbreviation"`
	IanaTimezone         string `json:"ianaTimezone" firestore:"ianaTimezone" bson:"ianaTimezone"`
	CurrencyCode         string `json:"currencyCode" firestore:"currencyCode" bson:"currencyCode"`
}

type Product struct {
	ID string `json:"id" firestore:"id" bson:"id"`
}

func (e *GraphQLError) Error() string {
	messages := make([]string, len(e.Errors))
	for i, err := range e.Errors {
		messages[i] = err.Get("message").String()
	}
	return strings.Join(messages, "\n")
}

type GraphQLError struct {
	Errors []gjson.Result
}
