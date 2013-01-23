package types

import (
    "fmt"
    "errors"
)

func libError( msg string ) error {
    return errors.New( "mingle/types: " + msg )
}

func libErrorf( tmpl string, argv ...interface{} ) error {
    return fmt.Errorf( "mingle/types: " + tmpl, argv... )
}
