package run

// AnySecurityScheme is implemented by all security scheme types.
type AnySecurityScheme interface {
	AuthType() string
}

// UserPasswordSecurity is a security scheme that uses a username and password for authentication.
type UserPasswordSecurity interface {
	UserPassword() (string, string)
}

// APIKeySecurity is a security scheme that uses an API key for authentication.
type APIKeySecurity interface {
	// APIKey returns the API key credential.
	APIKey() string
	// In returns the location of API key. For "apiKey" type the possible values are "user" and "password".
	In() string
}

//// X509Security is a security scheme that uses X.509 certificates for authentication and secure communication.
//type X509Security interface {
//	// TLSCert returns the TLS certificate, that includes the X509 certificate.
//	TLSCert() tls.Certificate
//}
//
//// HTTPAPIKeySecurity is a security scheme that uses an API key for authenticatio, but specifically for HTTP.
//type HTTPAPIKeySecurity interface {
//	// HTTPAPIKey returns the API key credential.
//	HTTPAPIKey() string
//	// Name returns the name of the header, query or cookie parameter to be used.
//	Name() string
//	// In returns the location of API key. For "httpApiKey" type the possible values are "header", "query" or "cookie".
//	In() string
//}
//
//// HTTPSecurity is a security scheme that uses the various HTTP authentication schemes. More details can be found in [RFC 7235].
////
//// [RFC 7235]: https://datatracker.ietf.org/doc/html/rfc7235#section-5.1
//type HTTPSecurity interface {
//	// Scheme returns the name of the HTTP authentication scheme
//	Scheme() string
//	// TODO: creds?
//}
//
//// HTTPBearerSecurity is a security scheme that uses Bearer tokens for authentication.
//type HTTPBearerSecurity interface {
//	HTTPSecurity
//	// BearerFormat returns a hint to the client to identify how the bearer token is formatted. See AsyncAPI spec.
//	BearerFormat() string
//	// TODO: creds?
//}
//
//// OpenIDConnectSecurity is a security scheme that uses OpenID Connect for authentication.
//type OpenIDConnectSecurity interface {
//	OpenIDConnectURL() string
//	Scopes() []string
//}
//
//type OAuth2Flow string
//
//const (
//	OAuth2FlowImplicit          OAuth2Flow = "implicit"
//	OAuth2FlowPassword          OAuth2Flow = "password"
//	OAuth2FlowClientCredentials OAuth2Flow = "clientCredentials"
//	OAuth2FlowAuthorizationCode OAuth2Flow = "authorizationCode"
//)
//
//// OAuth2Security is a security scheme that uses OAuth2 for authorization. Supports multiple OAuth2 flows.
//type OAuth2Security interface {
//	// ClientIDSecret returns the credentials: client ID and client secret.
//	ClientIDSecret() (string, string)
//	Flow() OAuth2Flow
//	AuthorizationURL() string
//	TokenURL() string
//	RefreshURL() string
//	AvailableScopes() map[string]string
//	Scopes() []string
//}
