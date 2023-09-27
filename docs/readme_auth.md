# Authentication readme

## Options

When it comes to accessing online accounts and services, there are several authentication methods available on the web. These methods aim to verify the identity of users and ensure that only authorized individuals can access sensitive information. Each method has its own strengths and weaknesses, both in the complexity and the security of the of implementation.

---

### Basic

Basic auth middleware provides an HTTP basic authentication.

    For valid credentials it calls the next handler.
    For missing or invalid credentials, it sends “401 - Unauthorized” response.

- https://echo.labstack.com/middleware/basic-auth/

---

### Casbin Auth Middleware

    Dependencies
    Custom Configuration
    Configuration

Casbin is a powerful and efficient open-source access control library for Go. It provides support for enforcing authorization based on various models. 

- https://echo.labstack.com/middleware/casbin-auth/
- https://github.com/casbin/casbin
- https://casbin.org/

---

### JWT Middleware

    Custom Configuration
    Configuration
    Example

JWT provides a JSON Web Token (JWT) authentication middleware. Echo JWT middleware is located at https://github.com/labstack/echo-jwt
- https://echo.labstack.com/middleware/jwt/
- https://echo.labstack.com/cookbook/jwt/

---

### Key Auth Middleware

    Custom Configuration
    Configuration

Key auth middleware provides a key based authentication.

    For valid key it calls the next handler.
    For invalid key, it sends “401 - Unauthorized” response.
    For missing key, it sends “400 - Bad Request” response.

- https://echo.labstack.com/middleware/key-auth/

---