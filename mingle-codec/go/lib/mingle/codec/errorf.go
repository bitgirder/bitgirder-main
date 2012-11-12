package codec

import "fmt"

func errorf( tmpl string, argv ...interface{} ) error {
    return fmt.Errorf( "mingle/codec: " + tmpl, argv... )
}
