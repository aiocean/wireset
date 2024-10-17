# Plan Repository Wireset

The Plan Repository Wireset provides a pre-configured setup for managing Shopify plan data in your applications using Go. It leverages Google Wire for dependency injection and includes the `PlanRepository` struct for handling plan-related operations.

## Features

- **Plan Management:** Provides the `PlanRepository` struct with methods for:
    - `GetPlan`: Retrieves a Shopify plan by ID.
    - `ListPlans`: Retrieves a list of all Shopify plans.
    - `CreatePlan`: Creates a new Shopify plan.
    - `UpdatePlan`: Updates an existing Shopify plan.
    - `DeletePlan`: Deletes a Shopify plan by ID.

## Declare the dependency provider


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
           repository.PlanRepoWireset,
       )
       return nil, nil, nil
   }
   ```

## Inject the dependency

Here's an example of how to use the `PlanRepository`:

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
    repository.PlanRepoWireset,
)

type FeatureExample struct {
    // ... other fields
    planRepository *repository.PlanRepository
}
```

## Inject the dependency

```go
// Example usage of PlanRepository
func (f *FeatureExample) MyFunction(ctx context.Context, planID string) error {
    // Get the plan from the repository
    plan, err := f.planRepository.GetPlan(ctx, planID)
    if err != nil {
        return fmt.Errorf("failed to get plan: %w", err)
    }

    // Do something with the plan
    fmt.Println("Plan ID:", plan.ID)
    fmt.Println("Plan Name:", plan.Name)

    return nil
}
```