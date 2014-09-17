package reactor

import (
    "fmt"
    "errors"
)

func libError( msg string ) error {
    return errors.New( "mingle/reactor: " + msg )
}

func libErrorf( tmpl string, argv ...interface{} ) error {
    return fmt.Errorf( "mingle/reactor: " + tmpl, argv... )
}
