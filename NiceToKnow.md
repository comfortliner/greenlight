# NiceToKnow

## Five Authetication Options
  * __Basic authentication__
    * The client includes an `Authorization` header with every request containing their credentials. The credentials need to be in the format `username:password` and base-64 encoded.  
    For example `Authorization: Basic YWxpY2VAushb3husHJsiw88Hzbcjsuw8l02n=`  
    Simple to use for several clients.
  * Token authentication in general
    * Sometimes known as _bearer token authentication_. The token expires after a set period of time and the user need to resubmit their credentials again to get a new token.  
    For example `Authorization: Bearer <token>`  
    When the API receives this request, it checks that the token hasn't expired and examines the token value to determine who the user is.  
    Can be complicated for clients, because they will need to implement the necessary logic for caching tokens, monitoring and managing token expiry and periodically generationg new tokens.
    * __Stateful token authentication__
      * The token is stored server-side in a database. The value of the token is a high-entropy cryptographically-secure random string.
    * __Stateless token authentication__
      * The user ID and expiry time is encoded in the token itself.  
    There are a few different technologies: _JWT_ (well known), _PASETO_, _Branca_ and _nacl/secretbox_.
  * __API key authentication__
    * The user has a non-expiring secret key associated with their account.  
    For example `Authorization: Key <key>`  
    Similar to stateful token approach. The main difference is that the keys are permanent keys, rather than temporary tokens.
  * __OAuth 2.0 / OpenID Connect__
    * Information about your user and their passwords is stored by a third-party `identity provider (IP)`. `OAuth 2.0` is not an authentication protocol. If you want to implement authentication checks against an `IP`, you should use `OpenID Connect`.

### Simple rules-of-thumb:
  * If your API doesn't have 'real' user accounts with slow password hashes, then `HTTP basic authentication` can be good.
  * If you don't want to store user passwords yourself, all your users have accounts with a third-party identity provider that supports OpenID Connect, and your API is the back-end for a website, then use `OpenID Connect`.
  * If you required delegated authentication, such as when your API has a microservice architecture, then use `stateless token authentication`.
  * Otherwise use `API keys` or `stateful token authentication`.
    * Stateful token are a nice fit for APIs that act as the back-end for a website or single-page application.
    * API keys can be better for more 'general purpose' APIs because they're permanent and simpler for developers to use in ther application and scripts.

## GO and React in a monorepo

[github aesrael](https://github.com/aesrael/go-postgres-react-starter/tree/master)

[github ueokande](https://github.com/ueokande/go-react-boilerplate)

[Let's code React Gin Blog](https://letscode.blog/2021/06/25/react-gin-blog-1-19-golang-and-react-web-app-guide/)

[Embed ReactJS to a go binary](https://dev.to/pacholoamit/one-of-the-coolest-features-of-go-embed-reactjs-into-a-go-binary-41e9)
