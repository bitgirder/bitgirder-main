package testing

import (
    "fmt"
    "errors"
)

func libError( msg string ) error {
    return errors.New( "mingle/codec/testing: " + msg )
}

func libErrorf( tmpl string, argv ...interface{} ) error {
    return fmt.Errorf( "mingle/codec/testing: " + tmpl, argv... )
}
