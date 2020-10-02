# OAuth2 Command Line

This is a small command-line helping me working on OAuth2 servers:

- `authorize`: implements the [OAuth 2 Authorization Code flow][RFC_CODE_FLOW].
   It keeps a cache of the value entered (except the client secret) to make it easier to replay the same process multiple times.

   Several options are available as output. By default it returns the whole JSON response but we can pass flags to get only the Accss-Token or ID-Token.
   It's useful in cases where we want to pipe this to a JWT decoder for example.

[RFC_CODE_FLOW]: https://tools.ietf.org/html/rfc6749#section-4.1
