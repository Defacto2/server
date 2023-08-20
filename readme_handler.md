# Handler readme

The `handler/` directory is where the Echo router, controller and its settings exist. In a traditional model-view-controller framework, this `handler/` directory would contain the controller.

This document lists possible new features and middleware that could be implemented into the existing Echo configuration.

---

### Unit tests

[This simple guide explains](https://echo.labstack.com/guide/testing/) how to test the fetching of a user by id from the database. If user is not found it returns 404 error with a message.

[List all routes for testing and debugging](https://echo.labstack.com/guide/routing/#list-routes).

[The list of middleware examples](https://github.com/labstack/echo/tree/master/middleware) contain `_test.go` test files.

---

### CRUD

User accounts example with create-read-update-delete, https://echo.labstack.com/cookbook/crud/

---

### Session ID

Session middleware https://echo.labstack.com/middleware/session/

Request ID Middleware https://echo.labstack.com/middleware/request-id/

---

### Uploads

[Upload a single file with parameters](https://echo.labstack.com/cookbook/file-upload/) example.

[Upload multiple files with parameters](https://echo.labstack.com/docs/cookbook/file-upload#upload-multiple-files-with-parameters) example.

---

### Timeouts

#### DO NOT USE

```
// ---------------------------------------------------------------------------------------------------------------
// WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING
// WARNING: Timeout middleware causes more problems than it solves.
// WARNING: This middleware should be first middleware as it messes with request Writer and could cause data race if
// 					it is in other position
```
[source code](https://github.com/labstack/echo/blob/master/middleware/timeout.go)


Timeout middleware for specific routes https://echo.labstack.com/middleware/timeout/

---

### Data validation

Request and validate data sourced from FORMS, queries etc.
https://echo.labstack.com/guide/request/#custom-binder

---

### Binding request data

#### Data Sources

Echo supports the following tags specifying data sources:

    query - query parameter
    param - path parameter (also called route)
    header - header parameter
    json - request body. Uses builtin Go json package for unmarshalling.
    xml - request body. Uses builtin Go xml package for unmarshalling.
    form - form data. Values are taken from query and request body. Uses Go standard library form parsing.

#### Data Types

When decoding the request body, the following data types are supported as specified by the Content-Type header:

    application/json
    application/xml
    application/x-www-form-urlencoded

When binding path parameter, query parameter, header, or form data, tags must be explicitly set on each struct field. However, JSON and XML binding is done on the struct field name if the tag is omitted. This is according to the behaviour of Go’s json package.

---

### API `response`

https://echo.labstack.com/guide/response/

#### OpenAPI

- https://github.com/deepmap/oapi-codegen
- https://threedots.tech/post/serverless-cloud-run-firebase-modern-go-application/#public-http-api

---

### Before and after hook responses

To handle data outside of the templates.

https://echo.labstack.com/guide/response/#hooks

---

### Custom context

HTTPS Server custom context that are used in handlers.

https://echo.labstack.com/guide/context/

---

### HTTP/2 Cleartext Server

https://echo.labstack.com/cookbook/http2/

---

### CORS

https://echo.labstack.com/middleware/cors/

https://echo.labstack.com/cookbook/cors/

---

### CSRF Middleware

https://echo.labstack.com/middleware/csrf/

---

### Sub-domains

http://html3.defacto2.net ?

https://echo.labstack.com/cookbook/subdomains/

---

## Proxy or no proxy?

#### IP config options

https://echo.labstack.com/guide/ip-address/

## Case 1. With no proxy

If you put no proxy (e.g.: directory facing to the internet), all you need to (and have to) see is IP address from network layer. Any HTTP header is untrustable because the clients have full control what headers to be set.

In this case, use echo.ExtractIPDirect().

## Case 2. With proxies using X-Forwarded-For header

`X-Forwared-For` (XFF) is the popular header to relay clients’ IP addresses. At each hop on the proxies, they append the request IP address at the end of the header.

## Case 3. With proxies using X-Real-IP header

`X-Real-IP` is another HTTP header to relay clients’ IP addresses, but it carries only one address unlike XFF.

---

## Event-driven applications

The intention is to enable two or more binaries to communicate with each other, ie `df2` and `server`. When `df2` updates, it could tell the `server` application to refresh some cached data stores such as the group statistics.

[Introducing Watermill - Go event-driven applications library](https://threedots.tech/post/introducing-watermill/)

https://watermill.io/

https://github.com/ThreeDotsLabs/watermill

https://github.com/ThreeDotsLabs/watermill/tree/master/_examples

---