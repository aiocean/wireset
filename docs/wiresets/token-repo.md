# Token Repository Wireset

The Token Repository Wireset provides a pre-configured setup for managing Shopify token data in your applications using Go. It leverages Google Wire for dependency injection and includes the `TokenRepository` struct for handling token-related operations.

## Features

- **Token Management:** Provides the `TokenRepository` struct with methods for:
    - `GetToken`: Retrieves a Shopify token by shop ID.
    - `SaveAccessToken`: Saves a Shopify access token for a shop.
- **Firestore Integration:** Uses Firestore as the underlying database for storing token data.
- **Dependency Injection:** Integrates seamlessly with Google Wire for easy dependency management.

## Usage


2. **Declare the dependency provider in your wire.go file:**

   ```go
   //go:build wireinject
   // +build wireinject

   package main

   import (
       "github.com/google/wire"
       // ... other imports
       "github.com/aiocean/wireset/repository" 
   )

   func InitializeServer() (server.Server, func(), error) {
       wire.Build(
           // ... other Wiresets
           repository.TokenRepoWireset,
       )
       return nil, nil, nil
   }
   ```

## Inject the dependency

Here's an example of how to use the `TokenRepository`:

```go
package example

import (
    "context"
    "fmt"

    "github.com/aiocean/wireset/model"
    "github.com/aiocean/wireset/repository"
)

var DefaultWireset = wire.NewSet(
    wire.Struct(new(FeatureExample), "*"),
    repository.TokenRepoWireset,
)

type FeatureExample struct {
    // ... other fields
    tokenRepository *repository.TokenRepository
}
```

## Usage

```go
// Example usage of TokenRepository
func (f *FeatureExample) MyFunction(ctx context.Context, shopID string) error {
    // Get the token from the repository
    token, err := f.tokenRepository.GetToken(ctx, shopID)
    if err != nil {
        return fmt.Errorf("failed to get token: %w", err)
    }

    // Do something with the token
    fmt.Println("Shop ID:", token.ShopID)
    fmt.Println("Access Token:", token.AccessToken)

    return nil
}
```