package server

import "github.com/pquerna/otp/totp"

// TODO more of this

func NewTOTP(username string) error {
	_, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "fishbb",
		AccountName: username,
	})
	if err != nil {
		return err
	}
	return nil
}
