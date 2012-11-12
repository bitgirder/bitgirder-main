package iotests

import (
    "os"
    "log"
)

//type TestContext interface {
//    func Fatalf( fmtStr string, args ...interface{} )
//    func Fatal( args ...interface{} )
//}

// Each entry in entries should be nil, a string or an *os.File (nil so this is
// safe to call for declared but failed-to-be-initialized test files). If an
// *os.File, Close() will also be called. Current implementation silently
// discards any errors, on the assumption that this method is used by tests and,
// in cases when tests themselves are failing, it isn't helpful to introduce
// more logging about problems cleaning up their files.
func WipeEntries( entries ...interface{} ) {
    for _, e := range entries {
        var path string
        switch v := e.( type ) {
        case string: path = v
        case *os.File: {
            if v != nil {
                path = v.Name()
                v.Close()
            }
        }
        default: log.Printf( "Unrecognized wipe entry (%T): %v", v, v )
        }
        if path != "" { os.RemoveAll( path ) }
    }
}
