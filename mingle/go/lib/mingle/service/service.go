package service

type AuthenticationMissingError struct {}

func ( e *AuthenticationMissingError ) Error() string {
    return "authentication missing"
}

var ErrAuthenticationMissing = &AuthenticationMissingError{}
