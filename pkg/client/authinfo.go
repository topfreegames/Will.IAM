package client

// AuthInfo keep the information after the authentication is done.
type AuthInfo struct {
	code       int
	name       string
	token      string
	email      string
	body       []byte
	permission string
}

// Code return the code returned by william on check permisssion
// will always return 200 except you use an empty permission
func (a AuthInfo) Code() int { return a.code }

// Name return the service account name was authenticated with william
func (a AuthInfo) Name() string { return a.name }

// Token return the refresh token of the service account that make the permissions
func (a AuthInfo) Token() string { return a.token }

// Email return the email that was register to the service account
func (a AuthInfo) Email() string { return a.email }

// Body return the body returned by william
func (a AuthInfo) Body() []byte { return a.body }

// Permission return an strings of the permission used this authentication
func (a AuthInfo) Permission() string { return a.permission }
