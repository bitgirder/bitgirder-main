package cast

import (
    "fmt"
    "errors"
)

func libError( msg string ) error {
    return errors.New( "mingle/cast: " + msg )
}

func libErrorf( tmpl string, argv ...interface{} ) error {
    return fmt.Errorf( "mingle/cast: " + tmpl, argv... )
}
