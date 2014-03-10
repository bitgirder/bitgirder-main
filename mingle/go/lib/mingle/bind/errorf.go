package bind

import (
    "fmt"
    "errors"
)

func libError( msg string ) error {
    return errors.New( "mingle/bind: " + msg )
}

func libErrorf( tmpl string, argv ...interface{} ) error {
    return fmt.Errorf( "mingle/bind: " + tmpl, argv... )
}
