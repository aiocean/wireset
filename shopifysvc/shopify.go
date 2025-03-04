package shopifysvc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/avast/retry-go"

	"github.com/google/wire"
	"go.uber.org/zap"

	"github.com/aiocean/wireset/cachesvc"
	"github.com/aiocean/wireset/configsvc"
	"github.com/pkg/errors"

	"github.com/tidwall/gjson"
)

const graphQlEndpointTemplate = "https://%s.myshopify.com/admin/api/%s/graphql.json"
const restEndpointTemplate = "https://%s.myshopify.com/admin/api/%s"

var DefaultWireset = wire.NewSet(
	NewShopifyService,
	NewShopifyApp,
)

type ShopifyService struct {
	ConfigService *configsvc.ConfigService
	ShopifyConfig *Config
	CacheSvc      *cachesvc.CacheService
	Logger        *zap.Logger
}

func NewShopifyService(
	configService *configsvc.ConfigService,
	shopifyConfig *Config,
	cacheSvc *cachesvc.CacheService,
	logger *zap.Logger,
) (*ShopifyService, func(), error) {
	cleanup := func() {

	}

	return &ShopifyService{
		ConfigService: configService,
		ShopifyConfig: shopifyConfig,
		CacheSvc:      cacheSvc,
		Logger:        logger.With(zap.Strings("tags", []string{"shopify"})),
	}, cleanup, nil
}

type ShopifyClient struct {
	ShopifyDomain string
	ShopifyConfig *Config
	ApiVersion    string
	AccessToken   string
	configSvc     *configsvc.ConfigService
	httpClient    *http.Client
	logger        *zap.Logger
}

func (s *ShopifyService) GetShopifyClient(shop, accessToken string) *ShopifyClient {
	shop = strings.Replace(shop, ".myshopify.com", "", -1)
	cacheKey := fmt.Sprintf("shopify_client_%s_%s", shop, accessToken)
	if client, ok := s.CacheSvc.Get(cacheKey); ok {
		return client.(*ShopifyClient)
	}

	client := ShopifyClient{
		ShopifyDomain: shop,
		AccessToken:   accessToken,
		ApiVersion:    s.ShopifyConfig.ApiVersion,
		configSvc:     s.ConfigService,
		ShopifyConfig: s.ShopifyConfig,
		logger:        s.Logger,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
	s.CacheSvc.SetWithTTL(cacheKey, client, 1*time.Hour)

	return &client
}
func (c *ShopifyClient) DoRestRequest(method, path string, body io.Reader) (*gjson.Result, error) {
	endpoint := fmt.Sprintf(restEndpointTemplate, c.ShopifyDomain, c.ApiVersion) + path
	req, err := http.NewRequest(method, endpoint, body)
	if err != nil {
		return nil, errors.Wrap(err, "DoRestRequest: failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Shopify-Access-Token", c.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "DoRestRequest: failed to do request")
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			c.logger.Error("DoRestRequest: failed to close response body", zap.Error(err))
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "DoRestRequest: failed to read response body")
	}

	data := gjson.ParseBytes(respBody)

	return &data, nil
}

type GraphQlRequest struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
}

func (c *ShopifyClient) DoGraphqlRequest(request *GraphQlRequest) (*gjson.Result, error) {
	var result *gjson.Result

	err := retry.Do(
		func() error {
			jsonPayload, err := json.Marshal(request)
			if err != nil {
				return errors.Wrap(err, "failed to marshal payload")
			}

			endpoint := fmt.Sprintf(graphQlEndpointTemplate, c.ShopifyDomain, c.ApiVersion)

			req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonPayload))
			if err != nil {
				return errors.Wrap(err, "failed to create request")
			}

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Shopify-Access-Token", c.AccessToken)

			resp, err := c.httpClient.Do(req)
			if err != nil {
				return errors.Wrap(err, "failed to do request")
			}

			defer func() {
				if err := resp.Body.Close(); err != nil {
					c.logger.Error("failed to close response body", zap.Error(err))
				}
			}()

			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return errors.Wrap(err, "failed to read response body")
			}

			if resp.StatusCode != http.StatusOK {
				return errors.Errorf("failed to do request, status code: %d, body: %s", resp.StatusCode, string(respBody))
			}

			parsedResult := gjson.GetManyBytes(respBody, "data", "errors")

			if parsedResult[1].Exists() {
				return &GraphQLError{Errors: parsedResult[1].Array()}
			}

			result = &parsedResult[0]
			return nil
		},
		retry.Attempts(3),
		retry.Delay(1*time.Second),
		retry.MaxDelay(5*time.Second),
		retry.OnRetry(func(n uint, err error) {
			c.logger.Warn("Retrying GraphQL request", zap.Uint("attempt", n), zap.Error(err))
		}),
	)

	if err != nil {
		return nil, errors.Wrap(err, "all retry attempts failed")
	}

	return result, nil
}

func (c *ShopifyClient) GetShopDetails() (*Shop, error) {
	requestBody := `{shop{
            id
            name
            email
            ianaTimezone
            timezoneOffset
            currencyCode
            myshopifyDomain
            primaryDomain {
                host
            }
        }}`

	response, err := c.DoGraphqlRequest(&GraphQlRequest{Query: requestBody})
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get shop details")
	}

	shopData := response.Get("shop")
	shopDetails := &Shop{
		ID:                   shopData.Get("id").String(),
		Name:                 shopData.Get("name").String(),
		Email:                shopData.Get("email").String(),
		CountryCode:          shopData.Get("countryCode").String(),
		Domain:               shopData.Get("primaryDomain.host").String(),
		MyshopifyDomain:      shopData.Get("myshopifyDomain").String(),
		TimezoneAbbreviation: shopData.Get("timezoneOffset").String(),
		IanaTimezone:         shopData.Get("ianaTimezone").String(),
		CurrencyCode:         shopData.Get("currencyCode").String(),
	}

	return shopDetails, nil
}

func (c *ShopifyClient) InstallScript(scriptUrl string) error {
	isInstalled, err := c.IsScriptInstalled(scriptUrl)
	if err != nil {
		return err
	}

	if isInstalled {
		return nil
	}

	requestBody := &GraphQlRequest{
		Query: `mutation scriptTagCreate($input: ScriptTagInput!) {
            scriptTagCreate(input: $input) {
                userErrors {
                    field
                    message
                }
                scriptTag {
                    src
                }
            }
        }`,
		OperationName: "scriptTagCreate",
		Variables: map[string]interface{}{
			"input": map[string]interface{}{
				"cache":        false,
				"displayScope": "ALL",
				"src":          scriptUrl,
			},
		},
	}

	if _, err := c.DoGraphqlRequest(requestBody); err != nil {
		return err
	}
	return nil
}

func (c *ShopifyClient) IsScriptInstalled(scriptUrl string) (bool, error) {
	requestBody := `{
		scriptTags(first: 10, src: "` + scriptUrl + `"){
			edges{
				node {
					src
				}
			}
		}
	}`
	response, err := c.DoGraphqlRequest(&GraphQlRequest{Query: requestBody})
	if err != nil {
		return false, err
	}

	total := response.Get("scriptTags.edges.#").Int()
	return total > 0, nil
}

func (c *ShopifyClient) InstallAppUninstalledWebhook() error {
	isInstalled, err := c.IsAppUninstalledWebhookInstalled()
	if err != nil {
		return err
	}

	if isInstalled {
		return nil
	}

	requestBody := &GraphQlRequest{
		Query: `mutation webhookSubscriptionCreate($input: WebhookSubscriptionInput!) {
			webhookSubscriptionCreate(input: $input) {
				userErrors {
					field
					message
				}
				webhookSubscription {
					id
				}
			}
		}`,
		Variables: map[string]interface{}{
			"input": map[string]interface{}{
				"topic":   "APP_UNINSTALLED",
				"format":  "JSON",
				"address": c.configSvc.ServiceUrl + "/webhook/shopify/app-uninstalled",
			},
		},
	}

	if _, err := c.DoGraphqlRequest(requestBody); err != nil {
		return errors.WithMessage(err, "failed to install app uninstalled webhook")
	}

	return nil
}

func (c *ShopifyClient) IsAppUninstalledWebhookInstalled() (bool, error) {
	requestBody := &GraphQlRequest{
		Query: `{
   webhookSubscriptions(first: 10, topic: APP_UNINSTALLED){
    edges{
     node {
      id
     }
    }
   }
  }`,
	}

	response, err := c.DoGraphqlRequest(requestBody)
	if err != nil {
		return false, errors.WithMessage(err, "failed to check if app uninstalled webhook is installed")
	}

	total := response.Get("webhookSubscriptions.edges.#").Int()
	return total > 0, nil
}

// GetCurrentTheme returns the current theme
func (c *ShopifyClient) GetCurrentTheme() (string, error) {
	requestBody := &GraphQlRequest{
		Query: `query($roles: [ThemeRole!]) {
			themes(first: 1, roles: $roles) {
				nodes {
					id
					name
					role
				}
			}
		}`,
		Variables: map[string]interface{}{
			"roles": []string{"MAIN"},
		},
	}
	
	response, err := c.DoGraphqlRequest(requestBody)
	if err != nil {
		return "", errors.WithMessage(err, "failed to get current theme")
	}

	themeData := response.Get("themes.nodes").Array()
	if len(themeData) == 0 {
		return "", errors.New("no theme found")
	}
	return themeData[0].Get("id").String(), nil
}

func (c *ShopifyClient) GetCurrentApplicationInstallationID() (string, error) {
	requestBody := &GraphQlRequest{
		Query: `{
        currentAppInstallation {
            id
        }
    }`,
	}

	response, err := c.DoGraphqlRequest(requestBody)
	if err != nil {
		return "", errors.Wrap(err, "failed to get current application installation ID")
	}

	installationID := response.Get("currentAppInstallation.id").String()
	return installationID, nil
}

// GetAppDataMetaField returns the value of the app data metafield
func (c *ShopifyClient) GetAppDataMetaField(ownerId, namespace, key string) (string, error) {
	requestBody := &GraphQlRequest{
		Query: `query GetAppDataMetafield($metafieldsQueryInput: [MetafieldsQueryInput!]!) {
            metafields(query: $metafieldsQueryInput) {
                edges {
                    node {
                        id
                        namespace
                        key
                        value
                    }
                }
            }
        }`,
		OperationName: "GetAppDataMetafield",
		Variables: map[string]interface{}{
			"metafieldsQueryInput": map[string]interface{}{
				"namespace": namespace,
				"key":       key,
				"ownerId":   ownerId,
			},
		},
	}

	response, err := c.DoGraphqlRequest(requestBody)
	if err != nil {
		return "", errors.Wrap(err, "failed to get app data metafield")
	}

	metafields := response.Get("metafields.edges.#.node")
	if metafields.Exists() {
		return metafields.Array()[0].Get("value").String(), nil
	}

	return "", nil
}

// GetShopMetaField accept ownerId, key
func (c *ShopifyClient) GetShopMetaField(namespace, key string) (string, error) {
	requestBody := &GraphQlRequest{
		Query: `query GetShopMetafield($namespace: String!, $key: String!) {
			shop {
				metafield(namespace: $namespace, key: $key) {
					value
				}
			}
		}`,
		Variables: map[string]interface{}{
			"namespace": namespace,
			"key":       key,
		},
	}

	response, err := c.DoGraphqlRequest(requestBody)
	if err != nil {
		return "", errors.Wrap(err, "failed to get shop metafield")
	}

	metafield := response.Get("shop.metafield")
	if metafield.Exists() {
		return metafield.Get("value").String(), nil
	}

	return "", nil
}

func (c *ShopifyClient) SetShopMetaField(ownerId, namespace, key, valueType, value string) error {
	requestBody := &GraphQlRequest{
		Query: `mutation CreateShopMetafield($metafieldsSetInput: [MetafieldsSetInput!]!) {
			metafieldsSet(metafields: $metafieldsSetInput) {
				metafields {
					id
					namespace
					key
					value
				}
				userErrors {
					field
					message
				}
			}
		}`,
		OperationName: "CreateShopMetafield",
		Variables: map[string]interface{}{
			"metafieldsSetInput": []map[string]interface{}{
				{
					"namespace": namespace,
					"key":       key,
					"type":      valueType,
					"value":     value,
					"ownerId":   ownerId,
				},
			},
		},
	}

	_, err := c.DoGraphqlRequest(requestBody)
	if err != nil {
		return errors.Wrap(err, "failed to create shop metafield")
	}

	return nil
}

func (c *ShopifyClient) SetAppDataMetaField(ownerId, namespace, key, valueType, value string) error {
	requestBody := &GraphQlRequest{
		Query: `mutation CreateAppDataMetafield($metafieldsSetInput: [MetafieldsSetInput!]!) {
            metafieldsSet(metafields: $metafieldsSetInput) {
                metafields {
                    id
                    namespace
                    key
                    value
                }
                userErrors {
                    field
                    message
                }
            }
        }`,
		OperationName: "CreateAppDataMetafield",
		Variables: map[string]interface{}{
			"metafieldsSetInput": []map[string]interface{}{
				{
					"namespace": namespace,
					"key":       key,
					"type":      valueType,
					"value":     value,
					"ownerId":   ownerId,
				},
			},
		},
	}

	_, err := c.DoGraphqlRequest(requestBody)
	if err != nil {
		return errors.Wrap(err, "failed to create app data metafield")
	}

	return nil
}

type Subscription struct {
	ID                        string
	Name                      string
	TrialDays                 int
	CurrentPeriodEnd          string
	Status                    string
	Test                      bool
	CurrentPeriodEndFormatted string
}

var ErrorSubscriptionNotFound = errors.New("subscription not found")

func (c *ShopifyClient) GetActiveSubscriptions() (*Subscription, error) {
	requestBody := &GraphQlRequest{
		Query: `{
		currentAppInstallation {
			activeSubscriptions{
				id
				name
				trialDays
				status
				test
				currentPeriodEnd
				}
			}
		}`,
	}

	response, err := c.DoGraphqlRequest(requestBody)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get subscription")
	}
	

	subscriptionData := response.Get("currentAppInstallation.activeSubscriptions.0")
	if !subscriptionData.Exists() {
		return nil, ErrorSubscriptionNotFound
	}

	subscription := &Subscription{
		ID:               subscriptionData.Get("id").String(),
		Name:             subscriptionData.Get("name").String(),
		TrialDays:        int(subscriptionData.Get("trialDays").Int()),
		CurrentPeriodEnd: subscriptionData.Get("currentPeriodEnd").String(),
		Status:           subscriptionData.Get("status").String(),
	}

	return subscription, nil
}

func (c *ShopifyClient) CreateSubscription(name string, price float32, interval string, returnUrl string, isTest bool) (*gjson.Result, error) {
	// returnUrl := "https://admin.shopify.com/store/" + c.ShopifyDomain + "/apps/" + c.ShopifyConfig.ClientId
	lineItems := []map[string]interface{}{
		{
			"plan": map[string]interface{}{
				"appRecurringPricingDetails": map[string]interface{}{
					"price": map[string]interface{}{
						"amount":       price,
						"currencyCode": "USD",
					},
					"interval": interval,
				},
			},
		},
	}

	requestBody := &GraphQlRequest{
		Query: `mutation AppSubscriptionCreate(
            $name: String!
            $lineItems: [AppSubscriptionLineItemInput!]!
            $returnUrl: URL!
			$test: Boolean!
        ) {
            appSubscriptionCreate(
                name: $name
                returnUrl: $returnUrl
                lineItems: $lineItems
                test: $test
            ) {
                userErrors {
                    field
                    message
                }
                appSubscription {
                    id
                }
                confirmationUrl
            }
        }`,
		Variables: map[string]interface{}{
			"name":      name,
			"returnUrl": returnUrl,
			"lineItems": lineItems,
			"test":      isTest,
		},
	}

	response, err := c.DoGraphqlRequest(requestBody)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create subscription")
	}

	return response, nil
}

func (c *ShopifyClient) GetChargeBillingInfo(chargeID string) (*gjson.Result, error) {
	url := fmt.Sprintf("https://%s/admin/api/%s/recurring_application_charges/%s.json", c.ShopifyDomain, c.ApiVersion, chargeID)
	response, err := c.DoRestRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get charge info")
	}

	return response, nil
}

func NormaizeShopifyDomain(shopifyDomain string) string {
	return strings.Replace(shopifyDomain, ".myshopify.com", "", -1) + ".myshopify.com"
}

func NormalizeShopifyName(shopifyName string) string {
	return strings.TrimSuffix(shopifyName, ".myshopify.com")
}

// GetOrdersCount returns the number of orders in a given time period
// If productId is provided, it will only count orders containing that product
func (c *ShopifyClient) GetOrdersCount(startDate, endDate string) (int64, error) {
	// Build the query filter
	queryFilter := fmt.Sprintf("created_at:>=%s AND created_at:<=%s", startDate, endDate)

	requestBody := &GraphQlRequest{
		Query: `query GetOrdersCount($query: String!) {
			ordersCount(query: $query) {
				count
			}
		}`,
		Variables: map[string]interface{}{
			"query": queryFilter,
		},
	}

	response, err := c.DoGraphqlRequest(requestBody)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get orders count")
	}

	count := response.Get("ordersCount.count").Int()
	return count, nil
}
