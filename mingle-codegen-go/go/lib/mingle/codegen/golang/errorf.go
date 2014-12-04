package golang

import (
    "fmt"
    "errors"
)

func libError( msg string ) error {
    return errors.New( "mingle/codegen/golang: " + msg )
}

func libErrorf( tmpl string, argv ...interface{} ) error {
    return fmt.Errorf( "mingle/codegen/golang: " + tmpl, argv... )
}
