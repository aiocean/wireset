# Proxy to localhost

You need to use tunneling to access your local server from the internet. You can use [ngrok](https://ngrok.com/) for this purpose.

Then set the env PROXY_URL to the ngrok url.

Every request to the proxy url will be forwarded to your local server.

## Proxy replacement for fast development

Alternatively to using ngrok, you can configure the app to proxy requests to your local server.
This approach is useful for fast development, as it doesn't require setting up a tunnel.

To enable this feature, set the `PROXY_URL` environment variable to the URL of your local server.
For example, if your local server is running on `http://localhost:8080`, you would set the following environment variable:

```bash
export PROXY_URL=<your-ngrok-server-url>
```

With this configuration, all requests to the app will be proxied to your local server, except for requests to `/healthz`.
This allows you to test your code changes locally without having to redeploy the app.


## how it works

the `fiberapp` wireset will check if the `PROXY_URL` environment variable is set. if it is, the app will use the `proxy.Forward` function to proxy the request to the local server.
