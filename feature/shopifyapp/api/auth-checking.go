package api

/*
Tài liệu AuthHandler

AuthHandler chịu trách nhiệm quản lý xác thực và ủy quyền cho một ứng dụng Shopify. Nó cung cấp hai chức năng chính:

1. Phương thức Checkin:
   - Điểm cuối để kiểm tra trạng thái xác thực của các yêu cầu đến.
   - Xử lý cả yêu cầu đã xác thực và chưa xác thực.

2. Luồng Xác thực:
   a. Đối với yêu cầu không có xác thực:
      - Kiểm tra xem cửa hàng có tồn tại trong cơ sở dữ liệu không.
      - Nếu cửa hàng tồn tại, cung cấp liên kết đến ứng dụng trong trang quản trị Shopify.
      - Nếu cửa hàng không tồn tại, cung cấp liên kết để bắt đầu quá trình OAuth.

   b. Đối với yêu cầu đã xác thực:
      - Xác thực token JWT trong tiêu đề Authorization.
      - Lưu trữ cache các phản hồi xác thực thành công để tăng hiệu suất.
      - Xuất bản sự kiện cửa hàng đã check-in để xử lý thêm.

Các tính năng chính:
- Sử dụng JWT để quản lý phiên.
- Triển khai bộ nhớ đệm để giảm thiểu chi phí xác thực.
- Xử lý các tình huống khác nhau như truy cập ứng dụng trực tiếp và môi trường phát triển.
- Cung cấp URL xác thực phù hợp dựa trên trạng thái của cửa hàng.
- Tích hợp với kho lưu trữ cửa hàng để xác minh tên miền.
- Sử dụng event bus để xuất bản các sự kiện check-in của cửa hàng.

Handler đảm bảo truy cập an toàn vào ứng dụng Shopify đồng thời cung cấp trải nghiệm tích hợp mượt mà cho các cửa hàng mới và xác thực hiệu quả cho các cửa hàng hiện có.
*/

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aiocean/wireset/model"
	"github.com/aiocean/wireset/shopifysvc"
	goshopify "github.com/bold-commerce/go-shopify/v3"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

func (s *AuthHandler) sessionTokenKeyFunc(token *jwt.Token) (interface{}, error) {
	return []byte(s.ShopifyConfig.ClientSecret), nil
}

func (s *AuthHandler) Checkin(ctx *fiber.Ctx) error {
	authentication := strings.TrimPrefix(ctx.Get("authorization"), "Bearer ")

	if authentication == "" {
		// this case is rare, but it's possible
		// Happen when user access the app directly, or from development environment
		return s.handleNoAuth(ctx)
	}

	return s.handleAuth(ctx, authentication)
}

// handleNoAuth processes requests without authentication.
//
// This function handles cases where the authentication header is empty.
// It checks for the presence of a 'shop' query parameter and determines
// the appropriate authentication URL based on whether the shop exists
// in the database.
//
// Parameters:
//   - ctx: A pointer to the Fiber context containing the HTTP request and response.
//
// Returns:
//   - An error if any occurs during processing, otherwise nil.
//
// The function follows these steps:
// 1. Logs that the authentication header is empty.
// 2. Retrieves the 'shop' query parameter.
// 3. If 'shop' is empty, returns an unauthorized response with the app listing URL.
// 4. Formats the shop name and domain.
// 5. Checks if the shop domain exists in the database.
// 6. If the shop exists, returns an unauthorized response with the Shopify admin app URL.
// 7. If the shop doesn't exist, returns an unauthorized response with the OAuth authorization URL.
func (s *AuthHandler) handleNoAuth(ctx *fiber.Ctx) error {
	s.LogSvc.Info("authentication header is empty")
	shopQuery := ctx.Query("shop")
	if shopQuery == "" {
		s.LogSvc.Info("shop query parameter is empty")
		return ctx.Status(http.StatusOK).JSON(model.AuthResponse{
			Message:           "Unauthorized",
			AuthenticationUrl: s.ShopifyConfig.AppListingUrl,
		})
	}

	shopName := strings.TrimSuffix(shopQuery, ".myshopify.com")
	shopDomain := shopName + ".myshopify.com"

	isExists, err := s.ShopRepo.IsDomainExists(ctx.UserContext(), shopDomain)
	if err != nil {
	
		s.LogSvc.Error("error while checking shop domain", zap.Error(err))
		// this is critical error, we need to return 500 to the client, and have no destination to redirect
		// so we need to return the app listing url, so that the user can install the app
		return ctx.Status(http.StatusInternalServerError).JSON(model.AuthResponse{
			Message: "Internal server error",
			AuthenticationUrl: s.ShopifyConfig.AppListingUrl,
		})
	}

	if isExists {
		s.LogSvc.Info("Shop exists in the database")
		return ctx.Status(http.StatusOK).JSON(model.AuthResponse{
			Message:           "Authorized",
			AuthenticationUrl: "https://admin.shopify.com/store/" + shopName + "/apps/" + s.ShopifyConfig.ClientId,
		})
	}

	return ctx.Status(http.StatusOK).JSON(model.AuthResponse{
		Message:           "Unauthorized",
		AuthenticationUrl: authorizeUrl(shopName, s.ShopifyConfig),
	})
}

// authorizeUrl generates the OAuth authorization URL for a Shopify store.
//
// This function constructs the URL used to initiate the OAuth flow for a Shopify app.
// It takes the shop name and Shopify configuration as input and returns the full
// authorization URL as a string.
//
// Parameters:
//   - shopName: The name of the Shopify store (without the '.myshopify.com' suffix).
//   - shopifyConfig: A pointer to the Shopify configuration containing client ID and nonce.
//
// Returns:
//   - A string representing the complete OAuth authorization URL.
//
// The function performs the following steps:
// 1. Constructs the base URL for the shop using goshopify.ShopBaseUrl.
// 2. Sets the path to '/admin/oauth/authorize'.
// 3. Adds query parameters for 'client_id' and 'state' (nonce).
// 4. Encodes the query parameters and returns the final URL as a string.
//
// Note: This function ignores any error returned by url.Parse as it assumes
// the shop base URL will always be valid.
func authorizeUrl(shopName string, shopifyConfig *shopifysvc.Config) string {
	shopUrl, _ := url.Parse(goshopify.ShopBaseUrl(shopName))
	shopUrl.Path = "/admin/oauth/authorize"
	query := shopUrl.Query()
	query.Set("client_id", shopifyConfig.ClientId)
	query.Set("state", shopifyConfig.LoginNonce)
	shopUrl.RawQuery = query.Encode()
	return shopUrl.String()
}

func (s *AuthHandler) handleAuth(ctx *fiber.Ctx, authentication string) error {

	// this case have time to live, so dont worry about security
	if authResponse, ok := s.CacheSvc.Get(authentication); ok {
		return ctx.Status(http.StatusOK).JSON(authResponse)
	}

	/*
		{
			"iss": "<shop-name.myshopify.com/admin>",
			"dest": "<shop-name.myshopify.com>",
			"aud": "<client ID>",
			"sub": "<user ID>",
			"exp": "<time in seconds>",
			"nbf": "<time in seconds>",
			"iat": "<time in seconds>",
			"jti": "<random UUID>",
			"sid": "<session ID>"
			"sig": "<signature>"
		}
	*/
	var sessionClaim model.CustomJwtClaims
	sessionToken, err := jwt.ParseWithClaims(authentication, &sessionClaim, s.sessionTokenKeyFunc)
	if err != nil {
		s.LogSvc.Error("error parsing jwt sessionToken", zap.Error(err))
		return ctx.Status(http.StatusUnauthorized).JSON(model.AuthResponse{
			Message: "Unauthorized",
		})
	}

	// check if the sessionToken is valid, if not return unauthorized and give the link to the app listing, so that user can install the app
	if !sessionToken.Valid {
		s.LogSvc.Error("invalid jwt sessionToken")
		return ctx.Status(http.StatusOK).JSON(model.AuthResponse{
			Message:           "Unauthorized",
			AuthenticationUrl: s.ShopifyConfig.AppListingUrl,
		})
	}

	parsedMyshopifyDomain := strings.Split(sessionClaim.Dest, "/")[2]

	// Publish the event to the event bus, so that other features can handle the check-in event. Eg: Feature to sync the shop data, check the shop status, etc.
	if err := s.EventBus.Publish(ctx.UserContext(), &model.ShopCheckedInEvt{
		MyshopifyDomain: parsedMyshopifyDomain,
		SessionToken:    authentication,
	}); err != nil {
		s.LogSvc.Error("error publishing event", zap.Error(err))
	}

	authResponse := model.AuthResponse{
		Message: "Authorized",
	}

	// calculate the time to live for the cache, it is the difference between the expiration time and the current time, we should buffer some time to make sure the cache is still valid
	ttl := int64(sessionClaim.Exp) - time.Now().Unix() - 60 // 60 seconds buffer
	ttlDuration := time.Duration(ttl) * time.Second
	s.CacheSvc.SetWithTTL(authentication, authResponse, ttlDuration)

	return ctx.Status(http.StatusOK).JSON(authResponse)
}
