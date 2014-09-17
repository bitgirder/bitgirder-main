package stack

import (
    "fmt"
    "errors"
)

func libError( msg string ) error {
    return errors.New( "bitgirder/stack: " + msg )
}

func libErrorf( tmpl string, argv ...interface{} ) error {
    return fmt.Errorf( "bitgirder/stack: " + tmpl, argv... )
}
