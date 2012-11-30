package mingle

import (
    "fmt"
    "errors"
)

func libError( msg string ) error {
    return errors.New( "mingle: " + msg )
}

func libErrorf( tmpl string, argv ...interface{} ) error {
    return fmt.Errorf( "mingle: " + tmpl, argv... )
}
