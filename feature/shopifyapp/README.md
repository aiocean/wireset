## Config webhook

Xem danh sách webhook tại: https://shopify.dev/docs/api/webhooks

```toml
[webhooks]
api_version = "2024-07"

  [[webhooks.subscriptions]]
  topics = [ "app/uninstalled", "orders/create", "app_subscriptions/update", "products/delete", "products/update" ]
  uri = "https://home8080.aiocean.dev/webhooks"

  [[webhooks.subscriptions]]
  topics = [ "app/uninstalled", "orders/create" ]
  uri = "https://shipshield-production.up.railway.app/webhooks"
```
