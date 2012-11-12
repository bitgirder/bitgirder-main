package json

import (
    "fmt"
    "errors"
)

func libError( msg string ) error {
    return errors.New( "mingle/codec/json: " + msg )
}

func libErrorf( tmpl string, argv ...interface{} ) error {
    return fmt.Errorf( "mingle/codec/json: " + tmpl, argv... )
}
