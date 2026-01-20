package auth

type User struct {
	email string
	password string
	username string
	creationDate string
	admin bool
	// For fine grained RBAC, we need more params here.
}