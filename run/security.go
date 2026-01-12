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
