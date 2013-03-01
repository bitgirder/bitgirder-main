package assert

import (
    "testing"
    "errors"
    "fmt"
//    "log"
)

// store gotFailure separately from lastMsg in case the zero val for lastMsg
// ends up being a valid expected error message (not likely, but not ruling it
// out either)
type failHolder struct {
    lastMsg string
    gotFailure bool
    t *testing.T
}

// panic to the caller if f.gotFailure, since that indicates an error in the
// test in which a previous failure was not collected via failed()
func ( f *failHolder ) Fatal( args ...interface{} ) {
    if f.gotFailure { panic( f.lastMsg ) }
    f.lastMsg = fmt.Sprint( args... )
    f.gotFailure = true
}

func ( f *failHolder ) ok( call func() ) {
    call()
    if f.gotFailure { f.t.Fatalf( "Unexpected failure: %s", f.lastMsg ) }
}

// clears gotFailure on exit no matter how this method returns
func ( f *failHolder ) failed( expct string ) {
    if ! f.gotFailure { f.t.Fatalf( "No failure reported" ) }
    if expct != f.lastMsg { f.t.Fatalf( "%#v != %#v", expct, f.lastMsg ) }
    defer func() { f.gotFailure = false }()
}

func testAsserter( t *testing.T ) ( *Asserter, *failHolder ) {
    fh := &failHolder{ t: t }
    a := &Asserter{ fh }
    return a, fh
}

type s1 struct { I1, i2 int }

// Used below in a regression in which we were attempting to call .String() on
// nil references to types that implement Stringer()
type stringType struct { s string }
func ( st *stringType ) String() string { return st.s }

type testErr string
func ( e testErr ) Error() string { return string( e ) }

func TestBasicAsserts( t *testing.T ) {
    a, f := testAsserter( t )
    f.ok( func() { a.True( true ) } )
    a.True( false )
    f.failed( "Value is false" )
    f.ok( func() { a.False( false ) } )
    a.False( true )
    f.failed( "Value is true" )
    f.ok( func() { a.Equal( "a", "a" ) } )
    f.ok( func() { a.Equal( s1{ 1, 2 }, s1{ 1, 2 } ) } )
    f.ok( func() { a.Equal( &s1{ 1, 2 }, &s1{ 1, 2 } ) } )
    a.Equal( "a", "b" )
    // use strings in next test to check %#v formatting 
    f.failed( `expect( string: "a" ) != actual( string: "b" )` ) 
    a.Equal( s1{ 1, 2 }, s1{ 1, 3 } )
    f.failed(
        `expect( assert.s1: assert.s1{I1:1, i2:2} ) != ` +
            `actual( assert.s1: assert.s1{I1:1, i2:3} )` )
    a.Equal( &s1{ 1, 2 }, &s1{ 1, 3 } )
    f.failed( 
        `expect( *assert.s1: &assert.s1{I1:1, i2:2} ) != ` +
            `actual( *assert.s1: &assert.s1{I1:1, i2:3} )` )
    f.ok( func() { a.NotEqual( 1, 2 ) } )
    var nilVal *stringType
    a.Equal( nilVal, &stringType{ "abc" } )
    f.failed( `expect( *assert.stringType: <nil> ) != ` +
        `actual( *assert.stringType: abc )` )
    a.NotEqual( 1, 1 )
    f.failed( "'comp' and 'actual' are both 1" )
    err1, err2 := testErr( "err1" ), testErr( "err2" )
    f.ok( func() { a.EqualErrors( err1, err1 ) } )
    f.ok( func() { a.EqualErrors( nil, nil ) } )
    a.EqualErrors( nil, err1 )
    f.failed( "Expected no error but got assert.testErr: err1" )
    a.EqualErrors( err1, nil )
    f.failed( `Got no error but expected assert.testErr: err1` )
    a.EqualErrors( err1, err2 )
    f.failed( 
        `Expected assert.testErr with message "err1" but got ` +
        `assert.testErr with message "err2"`,
    )
    // loop through all variants of assert methods that take a format
    // string/args and ensure that those args are passed through correctly to
    // the fail method
    msg := "test-msg: %d"
    for i, blk := range []func( i int ) {
        func( i int ) { a.Truef( false, msg, i ) },
        func( i int ) { a.Falsef( true, msg, i ) },
        func( i int ) { a.Equalf( "a", "b", msg, i ) },
        func( i int ) { a.NotEqualf( "a", "a", msg, i ) },
        func( i int ) { a.Fatalf( msg, i ) },
    } {
        blk( i )
        f.failed( fmt.Sprintf( "test-msg: %d", i ) )
    }
}

func TestAssertError( t *testing.T ) {
    a, f := testAsserter( t )
    err1 := fmt.Errorf( "test-error" )
    funcErr := func() ( interface{}, error ) { return nil, err1 }
    funcOk := func() ( interface{}, error ) { return true, nil }
    gotCb := false
    errFunc := func( err error ) { 
        gotCb = true
        if err != err1 { a.Fatalf( "bad-error" ) }
    }
    f.ok( func() { a.AssertError( funcErr, errFunc ) } )
    if ! gotCb { t.Fatalf( "callback not triggered" ) }
    gotCb = false
    a.AssertError( funcOk, errFunc )
    if gotCb { t.Fatalf( "errFunc was called" ) }
    f.failed( "Expected error but call returned true" )
    err2 := fmt.Errorf( "test-error2" )
    funcErr2 := func() ( interface{}, error ) { return nil, err2 }
    a.AssertError( funcErr2, errFunc )
    f.failed( "bad-error" )
}

func TestAssertPanic( t *testing.T ) {
    a, f := testAsserter( t )
    err1 := fmt.Errorf( "test-error" )
    gotCb := false
    errFunc := func( val interface{} ) {
        gotCb = true
        if val != err1 { a.Fatalf( "bad-panic" ) }
    }
    f.ok( func() { a.AssertPanic( func() { panic( err1 ) }, errFunc ) } )
    if ! gotCb { t.Fatalf( "callback not triggered" ) }
    gotCb = false
    a.AssertPanic( func() {}, errFunc )
    if gotCb { t.Fatalf( "errFunc was called" ) }
    f.failed( "Call did not panic" )
    a.AssertPanic( func() { panic( fmt.Errorf( "test-error2" ) ) }, errFunc )
    f.failed( "bad-panic" )
}

// Don't exhaustively re-test all calls on the default asserter, just use this
// to check that it is wired properly
func TestDefaultFailer( t *testing.T ) {
    defer func() {
        if err := recover(); err == nil {
            t.Fatalf( "Expected panic" )
        } else { 
            expctErr := `expect( string: "a" ) != actual( string: "b" )`
            if err.( error ).Error() != expctErr { t.Fatal( err ) } 
        }
    }()
    Equal( "a", "b" )
}

func TestAsAsserter( t *testing.T ) {
    var errAct error
    a := AsAsserter( func( args ...interface{} ) { 
        errAct = errors.New( fmt.Sprint( args... ) )
    })
    a.Equal( 1, 2 )
    if errAct == nil { t.Fatalf( "Got no error" ) }
    if errAct.Error() != `expect( int: 1 ) != actual( int: 2 )` { 
        t.Fatal( errAct ) 
    }
}

func TestPathAsserter( t *testing.T ) {
    fh := &failHolder{ t: t }
    a := NewPathAsserter( fh )
    errStr := `expect( int: 1 ) != actual( int: 2 )`
    a.Equal( 1, 2 )
    fh.failed( errStr )
    a = a.Descend( "p1" )
    a.Equal( 1, 2 )
    fh.failed( "p1: " + errStr )
    a = a.Descend( "p2" ).StartList().Next().Next().Descend( "p3" )
    a.Equal( 1, 2 )
    fh.failed( "p1.p2[ 2 ].p3: " + errStr )
}
