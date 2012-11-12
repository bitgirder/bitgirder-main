package assert

import(
    "reflect"
    "errors"
    "fmt"
    "bitgirder/objpath"
    "log"
)

type Failer interface {
    Fatal( args ...interface{} )
}

// Used to hold an arbitrary value that may at some point be converted to a
// string in an error message. Allows us to lazily only convert to string if
// actually generating an error message
type fmtVal struct { val interface{} }

func ( v fmtVal ) String() string {
    switch t := v.val.( type ) {
    case fmt.Stringer: return fmt.Sprintf( "%s", t )
    }
    return fmt.Sprintf( "%#v", v.val )
}

type Asserter struct {
    Failer
}

func ( a *Asserter ) Fatal( args ...interface{} ) { a.Failer.Fatal( args... ) }

func ( a *Asserter ) Fatalf( msg string, args ...interface{} ) {
    a.Fatal( fmt.Sprintf( msg, args... ) ) 
}

func ( a *Asserter ) Truef( val bool, msg string, args ...interface{} ) {
    if ! val { a.Fatalf( msg, args... ) }
}

func ( a *Asserter ) True( val bool ) { a.Truef( val, "Value is false" ) }

func ( a *Asserter ) Falsef( val bool, msg string, args ...interface{} ) {
    a.Truef( ! val, msg, args... )
}

func ( a *Asserter ) False( val bool ) { a.Falsef( val, "Value is true" ) }

func ( a *Asserter ) Equalf(
    expct, actual interface{}, msg string, args ...interface{} ) {
    if ! reflect.DeepEqual( expct, actual ) { a.Fatalf( msg, args... ) }
}

func ( a *Asserter ) Equal( expct, actual interface{} ) {
    a.Equalf( expct, actual, "expect( %T: %s ) != actual( %T: %s )",
        expct, fmtVal{ expct }, actual, fmtVal{ actual } )
}

func ( a *Asserter ) NotEqualf( 
    comp, actual interface{}, msg string, args ...interface{} ) {
    if reflect.DeepEqual( comp, actual ) { a.Fatalf( msg, args... ) }
}

func ( a *Asserter ) NotEqual( comp, actual interface{} ) {
    a.NotEqualf( 
        comp, actual, "'comp' and 'actual' are both %s", fmtVal{ comp } )
}

func ( a *Asserter ) AssertError(
    call func() ( interface{}, error ), errChk func( err error ) ) {
    if val, err := call(); err == nil {
        a.Fatalf( "Expected error but call returned %v", val )
    } else { errChk( err ) }
}

func ( a *Asserter ) AssertPanic(
    call func(), panicRecv func( val interface{} ) ) {
    defer func() { if val := recover(); val != nil { panicRecv( val ) } }()
    call()
    a.Fatalf( "Call did not panic" )
}

type assertFunc func( args ...interface{} )

func ( f assertFunc ) Fatal( args ...interface{} ) { f( args... ) }

func AsAsserter( f assertFunc ) *Asserter { return &Asserter{ f } }

type PathAsserter struct {
    *Asserter
    f Failer
    p objpath.PathNode
}

func ( pa *PathAsserter ) Fatal( args ...interface{} ) {
    args2 := args
    if pa.p != nil {
        args2 = make( []interface{}, 1, 1 + len( args ) )
        args2[ 0 ] = objpath.Format( pa.p, objpath.StringDotFormatter ) + ": "
        args2 = append( args2, args... )
    }
    pa.f.Fatal( args2... )
}

func makePathAsserter( f Failer, p objpath.PathNode ) *PathAsserter {
    res := &PathAsserter{ f: f, p: p }
    res.Asserter = &Asserter{ res }
    return res
}

func NewPathAsserter( f Failer ) *PathAsserter {
    return makePathAsserter( f, nil )
}

func ( pa *PathAsserter ) Descend( node interface{} ) *PathAsserter {
    p := pa.p
    if p == nil { p = objpath.RootedAt( node ) } else { p = p.Descend( node ) }
    return makePathAsserter( pa.f, p )
}

func ( pa *PathAsserter ) StartList() *PathAsserter {
    p := pa.p
    if p == nil { p = objpath.RootedAtList() } else { p = p.StartList() }
    return makePathAsserter( pa.f, p )
}

func ( pa *PathAsserter ) Next() *PathAsserter {
    p := pa.p.( *objpath.ListNode ).Next()
    return makePathAsserter( pa.f, p )
}

func ( pa *PathAsserter ) Printf( tmpl string, args ...interface{} ) {
    if pa.p == nil {
        log.Printf( tmpl, args... )
    } else {
        log.Printf( "%s: %s",
            objpath.Format( pa.p, objpath.StringDotFormatter ),
            fmt.Sprintf( tmpl, args... ),
        )
    }
}

var defl *Asserter

type PanicFailer struct {}

func ( p *PanicFailer ) Fatal( args ...interface{} ) { 
    panic( errors.New( fmt.Sprint( args... ) ) )
}

func init() { defl = &Asserter{ &PanicFailer{} } }

func Fatal( args ...interface{} ) { defl.Fatal( args... ) }

func Fatalf( msg string, args ...interface{} ) { defl.Fatalf( msg, args... ) }

func True( val bool ) { defl.True( val ) }

func Truef( val bool, msg string, args ...interface{} ) {
    defl.Truef( val, msg, args... )
}

func False( val bool ) { defl.False( val ) }

func Falsef( val bool, msg string, args ...interface{} ) {
    defl.Falsef( val, msg, args... )
}

func Equal( expct, actual interface{} ) { defl.Equal( expct, actual ) }

func Equalf( expct, actual interface{}, msg string, args ...interface{} ) { 
    defl.Equalf( expct, actual, msg, args... ) 
}

func NotEqual( comp, actual interface{} ) { defl.NotEqual( comp, actual ) }

func NotEqualf( comp, actual interface{}, msg string, args ...interface{} ) { 
    defl.NotEqualf( comp, actual, msg, args... ) 
}

func AssertError( 
    call func() ( interface{}, error ), errChk func( err error ) ) {
    defl.AssertError( call, errChk )
}

func AssertPanic( call func(), panicRecv func( val interface{} ) ) {
    defl.AssertPanic( call, panicRecv )
}
