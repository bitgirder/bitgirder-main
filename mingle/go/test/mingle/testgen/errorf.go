package testgen

import (
    "fmt"
    "errors"
)

func libError( msg string ) error {
    return errors.New( "mingle/testgen: " + msg )
}

func libErrorf( tmpl string, argv ...interface{} ) error {
    return fmt.Errorf( "mingle/testgen: " + tmpl, argv... )
}
