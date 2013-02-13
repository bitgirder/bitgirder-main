package service

import (
    mg "mingle"
    "fmt"
)

type externalFormer interface { ExternalForm() string }

func errMsgNoSuchX( desc string, ef externalFormer ) string {
    return fmt.Sprintf( "no such %s: %s", desc, ef.ExternalForm() )
}

type NoSuchNamespaceError struct { Namespace *mg.Namespace }

func ( e *NoSuchNamespaceError ) Error() string {
    return errMsgNoSuchX( "namespace", e.Namespace )
}

type NoSuchServiceError struct { Service *mg.Identifier }

func ( e *NoSuchServiceError ) Error() string {
    return errMsgNoSuchX( "service", e.Service )
}

type NoSuchOperationError struct { Operation *mg.Identifier }

func ( e *NoSuchOperationError ) Error() string {
    return errMsgNoSuchX( "operation", e.Operation )
}

type AuthenticationMissingError struct {}

func ( e *AuthenticationMissingError ) Error() string {
    return "authentication missing"
}

var ErrAuthenticationMissing = &AuthenticationMissingError{}
