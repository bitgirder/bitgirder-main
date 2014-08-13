package builtin

import (
    "fmt"
    "errors"
)

func libError( msg string ) error {
    return errors.New( "mingle/types/builtin: " + msg )
}

func libErrorf( tmpl string, argv ...interface{} ) error {
    return fmt.Errorf( "mingle/types/builtin: " + tmpl, argv... )
}
