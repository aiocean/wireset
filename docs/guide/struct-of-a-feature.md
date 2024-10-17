# Structure of a Feature

In a well-organized application, a feature represents a self-contained unit of functionality.  Think of it like a building block.  A robust feature structure ensures maintainability, scalability, and ease of development.

Let's imagine we're building a feature for an e-commerce platform, specifically focusing on "Product Reviews."

## 1. Feature Core

At the heart of our feature lies the core struct, acting as the central hub connecting all components.

**feature/product_reviews/product_reviews.go**

```go

type ProductReviewsCore struct {
    // Command Handlers (Process actions that change application state)
    AddReviewCmdHandler    *command.AddReviewHandler
    ModerateReviewCmdHandler *command.ModerateReviewHandler

    // Event Handlers (React to events within the application)
    ReviewAddedEventHandler     *event.ReviewAddedHandler
    ReviewModeratedEventHandler  *event.ReviewModeratedHandler

    // API Handlers (Manage HTTP endpoints for external interaction)
    ReviewsAPIHandler *api.ReviewsHandler

    // Middleware (Address cross-cutting concerns like authentication)
    AuthMiddleware *middleware.ProductReviewsAuthMiddleware

    // Processors (Handle asynchronous tasks)
    EventProcessor   *cqrs.EventProcessor
    CommandProcessor *cqrs.CommandProcessor

    // Registries (For organized registration of handlers and routes)
    HttpRegistry *fiberapp.Registry
}
```

**Explanation:**

- **Command Handlers:**  These process commands like "AddReview" or "ModerateReview," which directly modify the application's state.
- **Event Handlers:** They react to events, such as a new review being added, without directly changing the application state.
- **API Handlers:** Manage HTTP endpoints (e.g., `/products/{id}/reviews`) to interact with the feature from outside.
- **Middleware:**  Handles tasks like authentication, ensuring only authorized users can perform certain actions.
- **Processors:** Manage the asynchronous processing of commands and events.
- **Registries:** Provide a structured way to register API routes and other handlers.

## 2. Initialization

Each feature needs an `Init()` method to set up its components.

**feature/product_reviews/product_reviews.go**

```go
func (f *ProductReviewsCore) Init() error {
    // Register Command Handlers
    if err := f.CommandProcessor.AddHandlers(
        f.AddReviewCmdHandler,
        f.ModerateReviewCmdHandler,
    ); err != nil {
        return err
    }

    // Register Event Handlers
    if err := f.EventProcessor.AddHandlers(
        f.ReviewAddedEventHandler,
        f.ReviewModeratedEventHandler,
    ); err != nil {
        return err
    }

    // Add Middleware
    f.HttpRegistry.AddHttpMiddleware("/reviews/*", f.AuthMiddleware.Handle) 

    // Register API Handlers
    f.HttpRegistry.AddHttpHandlers(
        &fiberapp.HttpHandler{
            Method:   fiber.MethodPost,
            Path:     "/products/:id/reviews",
            Handlers: []fiber.Handler{f.ReviewsAPIHandler.AddReview},
        },
        // ... more routes
    )

    return nil
}
```

**Explanation:**

- This method registers command and event handlers, adds middleware to specific routes, and registers API endpoints.

## 3. Wire Configuration

Dependency injection is crucial for maintainability. We'll use Google Wire to manage dependencies.

**feature/product_reviews/product_reviews.go**

```go
var DefaultWireset = wire.NewSet(
    wire.Struct(new(ProductReviewsCore), "*"), 

    command.NewAddReviewHandler,
    command.NewModerateReviewHandler,

    event.NewReviewAddedHandler,
    event.NewReviewModeratedHandler,

    // usually, you will never have to reimplement the authz middleware, it's already defined in the shopifyapp wireset
    middleware.NewProductReviewsAuthMiddleware,

    api.NewReviewsHandler,
)
```

**Explanation:**

- This configuration tells Wire how to create instances of our feature's components and their dependencies.

## 4. Sub-components

### 4.1. Command Handlers

**feature/product_reviews/command/add_review.go**

```go
type AddReviewHandler struct {
    // ... dependencies (e.g., repositories)
}

func (h *AddReviewHandler) HandlerName() string {
    return "AddReviewHandler"
}

func (h *AddReviewHandler) NewCommand() interface{} {
    return &model.AddReviewCmd{}
}

func (h *AddReviewHandler) RegisterBus(commandBus *cqrs.CommandBus, eventBus *cqrs.EventBus) {
    h.commandBus = commandBus
    h.eventBus = eventBus
}

func (h *AddReviewHandler) Handle(ctx context.Context, raw interface{}) error {
    cmd := raw.(*model.AddReviewCmd) 
    // ... logic to add the review (e.g., database interaction)
    return nil 
}
```

**Explanation:**

- Command handlers receive a command object and contain the logic to execute that command.

### 4.2. Event Handlers

**feature/product_reviews/event/review_added.go**

```go
type ReviewAddedHandler struct {
    // ... dependencies
}

func (h *ReviewAddedHandler) HandlerName() string {
    return "ReviewAddedHandler"
}

func (h *ReviewAddedHandler) NewEvent() interface{} {
    return &model.ReviewAddedEvt{}
}

func (h *ReviewAddedHandler) RegisterBus(commandBus *cqrs.CommandBus, eventBus *cqrs.EventBus) {
    h.commandBus = commandBus
    h.eventBus = eventBus
}

func (h *ReviewAddedHandler) Handle(ctx context.Context, event interface{}) error {
    evt := event.(*model.ReviewAddedEvt)
    // ... logic to react to the event (e.g., send a notification)
    return nil
}
```

**Explanation:**

- Event handlers receive an event object and perform actions in response to that event.

### 4.3. API Handlers

**feature/product_reviews/api/reviews.go**

```go
type ReviewsHandler struct {
    // ... dependencies
}

func (h *ReviewsHandler) AddReview(ctx *fiber.Ctx) error {
    // ... logic to handle the API request (e.g., validate input, call a command handler)
    return ctx.JSON(fiber.Map{"message": "Review added successfully"})
}
```

**Explanation:**

- API handlers manage HTTP requests, often interacting with command handlers to modify data.

## 5. Models, Configuration, Repositories

- **Models (`feature/product_reviews/models/models.go`):** Define data structures used within the feature (e.g., `Review` struct).
- **Configuration:**  Features might have configuration settings (e.g., maximum review length).
- **Repositories:**  Abstract interactions with databases or external services.

## Best Practices

- **Separation of Concerns:** Each component should have a single, well-defined responsibility.
- **Dependency Injection:** Use Wire for testability and modularity.
- **Error Handling:**  Handle errors gracefully and provide meaningful feedback.
- **Naming Conventions:**  Use consistent naming for clarity.
- **Registration:**  Centralize handler and route registration in the `Init()` method.
- **Encapsulation:** Keep feature internals private and expose only necessary functionality. 

By following this structure, you create features that are:

- **Modular:**  Easy to understand and maintain as they are self-contained.
- **Reusable:**  Potentially reusable in other parts of your application or even other projects.
- **Testable:**  Dependency injection makes it easier to write unit tests. 
