package objpath

import (
    "fmt"
    "errors"
)

func libError( msg string ) error {
    return errors.New( "bitgirder/objpath: " + msg )
}

func libErrorf( tmpl string, argv ...interface{} ) error {
    return fmt.Errorf( "bitgirder/objpath: " + tmpl, argv... )
}
