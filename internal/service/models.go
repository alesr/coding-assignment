package service

type Credentials struct {
	Username string
	Password string
}

func (c Credentials) validate() error {
	if c.Username == "" {
		return ErrUsernameInvalid
	}

	if c.Password == "" {
		return ErrPasswordInvalid
	}
	return nil
}

type Token struct {
	AccessToken string
	TokenType   string
	ExpiresIn   int64
}
