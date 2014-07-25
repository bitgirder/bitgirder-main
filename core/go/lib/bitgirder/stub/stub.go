package stub

import (
    "fmt"
    "runtime"
)

func Unimplemented() error {
    pc, _, _, _ := runtime.Caller( 1 )
    fc := runtime.FuncForPC( pc )
    return fmt.Errorf( "unimplemented: %s", fc.Name() )
}
