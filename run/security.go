package run

// TODO: docs

type AnySecurityScheme = any

type UserPasswordSecurity interface {
	UserPassword() (string, string)
}

type APIKeySecurity interface {
	APIKey() string
}
