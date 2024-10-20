# Shop Repository Wireset

The Shop Repository Wireset provides a pre-configured setup for interacting with shop data in your Shopify applications using Go. It leverages the power of Google Wire for dependency injection and includes the `ShopRepository` struct for managing shop data.

## Features

- **Shop Data Management:** Provides the `ShopRepository` struct with methods for common shop data operations like:
    - `IsShopExists`: Checks if a shop exists.
    - `IsDomainExists`: Checks if a shop domain exists.
    - `Create`: Creates a new shop.
    - `Update`: Updates an existing shop.
    - `Get`: Retrieves a shop by ID.
    - `GetByDomain`: Retrieves a shop by domain.
    - `CountShops`: Returns the total number of shops.
    - `UpdateLastLogin`: Updates the last login time of a shop.
    - `UpdateStoreState`: Updates the state of a shop (not yet implemented).
- **Firestore Integration:** Uses Firestore as the underlying database for storing shop data.
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
           repository.ShopRepoWireset,
       )
       return nil, nil, nil
   }
   ```

## Inject the dependency

Here's an example of how to use the `ShopRepository` in your `FeatureExample`:

**feature/example/feature.go**

```go
package example

import (
    "context"
    "fmt"

    "github.com/aiocean/wireset/repository"
)

var DefaultWireset = wire.NewSet(
    wire.Struct(new(FeatureExample), "*"),
    repository.ShopRepoWireset,
)

type FeatureExample struct {
    // ... other fields
    shopRepository *repository.ShopRepository
}
```

## Usage

```go
// Example usage of ShopRepository
func (f *FeatureExample) MyFunction(ctx context.Context, shopID string) error {
    // Get the shop from the repository
    shop, err := f.shopRepository.Get(ctx, shopID)
    if err != nil {
        return fmt.Errorf("failed to get shop: %w", err)
    }

    // Do something with the shop data
    fmt.Println("Shop name:", shop.Name)

    return nil
}
```