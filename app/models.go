package app

type authenticateRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (c *authenticateRequest) validate() error {
	if c.Username == "" {
		return ErrInvalidUsername
	}

	if c.Password == "" {
		return ErrInvalidPassword
	}
	return nil
}

type authenticaResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expired_in"`
}

type sumRequest any

type sumResponse struct {
	Sum string `json:"sum"`
}
