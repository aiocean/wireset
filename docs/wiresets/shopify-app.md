# Shopify App Wireset

The Shopify App Wireset provides a pre-configured setup for building Shopify applications using Go. It leverages the power of Google Wire for dependency injection and includes common dependencies and functionalities required for Shopify app development.

## Features

- **Shopify Authentication and API Integration:** Provides seamless integration with the Shopify API, including authentication, API client initialization, and handling Shopify webhooks.
- **Command and Event Handling:** Implements a robust system for handling commands and events using Watermill, enabling asynchronous operations and event-driven architecture.
- **Websocket Support:** Offers built-in support for real-time communication with Shopify stores using Websockets.
- **Database Integration:** Includes integration with MongoDB for data persistence.
- **Tracing and Monitoring:** Integrates with Datadog for distributed tracing and monitoring.
- **Configuration Management:** Provides a centralized way to manage configuration settings from environment variables.

## Usage

3. **Declare your feature list:**

**main.go**

   ```go
   package main
   import (
       "github.com/aiocean/wireset"
       "github.com/aiocean/wireset/shopifyapp"
       // Import other required packages
   )
   func FeatureList(
       shopifyFeature *shopifyapp.FeatureCore,
   ) []server.Feature {
       return []server.Feature{
           shopifyFeature,
       }
   }
   ```

4. **Define your wire dependencies:**

**wire.go**

   ```go
   //go:build wireinject
   // +build wireinject

   package main

   import (
       "github.com/google/wire"
       "github.com/aiocean/wireset"
   )

   func InitializeServer() (server.Server, func(), error) {
       wire.Build(
           FeatureList,
           wireset.ShopifyApp,
           // Include other required Wiresets
       )
       return nil, nil, nil
   }
   ```

5. **Run the `wire` command:**

   ```bash
   wire gen ./...
   ```