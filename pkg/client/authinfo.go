package client

type AuthInfo struct {
	code       int
	name       string
	token      string
	email      string
	body       []byte
	permission string
}

func (a AuthInfo) Code() int          { return a.code }
func (a AuthInfo) Name() string       { return a.name }
func (a AuthInfo) Token() string      { return a.token }
func (a AuthInfo) Email() string      { return a.email }
func (a AuthInfo) Body() []byte       { return a.body }
func (a AuthInfo) Permission() string { return a.permission }
