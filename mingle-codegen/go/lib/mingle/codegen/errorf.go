package codegen

import (
    "fmt"
    "errors"
)

func libError( msg string ) error {
    return errors.New( "mingle/codegen: " + msg )
}

func libErrorf( tmpl string, argv ...interface{} ) error {
    return fmt.Errorf( "mingle/codegen: " + tmpl, argv... )
}
