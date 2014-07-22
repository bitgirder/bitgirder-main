package mingle

import (
    "bitgirder/assert"
    "bitgirder/objpath"
    "fmt"
    "strconv"
    "time"
)

func MustTimestamp( s string ) Timestamp {
    tm, err := time.Parse( time.RFC3339Nano, s )
    if err != nil { panic( err ) }
    return Timestamp( tm )
}

func mkId( parts ...string ) *Identifier { return NewIdentifierUnsafe( parts ) }

func mkNs( ids ...*Identifier ) *Namespace {
    return &Namespace{ Version: ids[ 0 ], Parts: ids[ 1 : ] }
}

var mkDeclNm = NewDeclaredTypeNameUnsafe

func mkQn( ns *Namespace, nm *DeclaredTypeName ) *QualifiedTypeName {
    return &QualifiedTypeName{ Namespace: ns, Name: nm }
}

func ns1V1Qn( nm string ) *QualifiedTypeName {
    return mkQn( mkNs( mkId( "v1" ), mkId( "ns1" ) ), mkDeclNm( nm ) )
}

func ns1V1At( nm string ) *AtomicTypeReference {
    return &AtomicTypeReference{ Name: ns1V1Qn( nm ) }
}

func checkEqualTimestamps( 
    expct Timestamp, act Value, a *assert.PathAsserter ) {

    if tmAct, ok := act.( Timestamp ); ok {
        a.Truef( expct.Compare( tmAct ) == 0, 
            "input time was %s, got: %s", expct, tmAct )
    } else {
        a.Fatalf( "expected time, got %T", act )
    }
}

func checkEqualLists( expct *List, actVal Value, a *assert.PathAsserter ) {
    act, ok := actVal.( *List )
    a.Truef( ok, "not a list: %T", act )
    a.Descend( "(ListLen)" ).Equal( expct.Len(), act.Len() )
    la := a.StartList()
    for i, e := 0, expct.Len(); i < e; i++ {
        checkEqualValues( expct.Get( i ), act.Get( i ), la )
        la = la.Next()
    }
}

func checkDirectlyEqual( expct, act Value, a *assert.PathAsserter ) {
    a.Equalf( expct, act, "expected %s (%T) but got %s (%T)",
        QuoteValue( expct ), expct, QuoteValue( act ), act )
}

func checkEqualMapPairs( expct, act *SymbolMap, a *assert.PathAsserter ) {
    expctKeys, actKeys := SortIds( expct.GetKeys() ), SortIds( act.GetKeys() )
    a.Equalf( expctKeys, actKeys, "expected fields %s, got %s",
        idSliceToString( expctKeys ), idSliceToString( actKeys ) )
    for _, fld := range expctKeys {
        fldValExpct, fldValAct := expct.Get( fld ), act.Get( fld )
        checkEqualValues( fldValExpct, fldValAct, a.Descend( fld ) )
    }
}

func checkEqualMaps( expct *SymbolMap, actVal Value, a *assert.PathAsserter ) {
    act, ok := actVal.( *SymbolMap )
    a.Truef( ok, "not a map: %T", actVal )
    checkEqualMapPairs( expct, act, a )
}

func checkEqualStructs( expct *Struct, actVal Value, a *assert.PathAsserter ) {
    act, ok := actVal.( *Struct )
    a.Truef( ok, "not a struct: %T", actVal )
    a.Descend( "$type" ).Equal( expct.Type, act.Type )
    checkEqualMapPairs( expct.Fields, act.Fields, a )
}

func checkEqualValues( expct, act Value, a *assert.PathAsserter ) {
    switch v := expct.( type ) {
    case Timestamp: checkEqualTimestamps( v, act, a )
    case *List: checkEqualLists( v, act, a )
    case *Struct: checkEqualStructs( v, act, a )
    case *SymbolMap: checkEqualMaps( v, act, a )
    default: checkDirectlyEqual( expct, act, a )
    }
}

func AssertEqualValues( expct, act Value, f assert.Failer ) {
    checkEqualValues( expct, act, assert.NewPathAsserter( f ) )
}

func EqualPaths( expct, act objpath.PathNode, a assert.Failer ) {
    ( &assert.Asserter{ a } ).Equalf( 
        expct, 
        act,
        "expected path %q but got %q", FormatIdPath( expct ),
            FormatIdPath( act ),
    )
}

func MakeTestId( i int ) *Identifier { return mkId( fmt.Sprintf( "f%d", i ) ) }

func mustUint64( s string ) uint64 {
    res, err := strconv.ParseUint( s, 10, 64 )
    if ( err != nil ) { panic( err ) }
    return res
}

func startTestIdPath( elt interface{} ) objpath.PathNode {
    switch v := elt.( type ) {
    case int: return objpath.RootedAt( MakeTestId( v ) )
    case string: return objpath.RootedAtList().SetIndex( mustUint64( v ) )
    }
    panic( libErrorf( "unhandled elt: %T", elt ) )
}

func MakeTestIdPath( elts ...interface{} ) objpath.PathNode { 
    if len( elts ) == 0 { return nil }
    res := startTestIdPath( elts[ 0 ] )
    for i, e := 1, len( elts ); i < e; i++ {
        switch v := elts[ i ].( type ) {
        case int: res = res.Descend( MakeTestId( v ) ) 
        case string: res = res.StartList().SetIndex( mustUint64( v ) )
        default: panic( libErrorf( "unhandled elt: %T", v ) )
        }
    }
    return res
}

type CastErrorAssert struct {
    ErrExpect, ErrAct error
    *assert.PathAsserter
}

func ( cea CastErrorAssert ) FailActErrType() {
    cea.Fatalf(
        "Expected error of type %T but got %T: %s",
        cea.ErrExpect, cea.ErrAct, cea.ErrAct )
}

// Returns a path asserter that can be used further
func ( cea CastErrorAssert ) assertValueError( 
    expct, act ValueError ) *assert.PathAsserter {
    a := cea.Descend( "Err" )
    a.Descend( "Error()" ).Equal( expct.Error(), act.Error() )
    a.Descend( "Message()" ).Equal( expct.Message(), act.Message() )
    a.Descend( "Location()" ).Equal( expct.Location(), act.Location() )
    return a
}

func ( cea CastErrorAssert ) assertVcError() {
    if act, ok := cea.ErrAct.( *ValueCastError ); ok {
        cea.assertValueError( cea.ErrExpect.( *ValueCastError ), act )
    } else { cea.FailActErrType() }
}

func ( cea CastErrorAssert ) assertMissingFieldsError() {
    if act, ok := cea.ErrAct.( *MissingFieldsError ); ok {
        cea.assertValueError( cea.ErrExpect.( ValueError ), act )
    } else { cea.FailActErrType() }
}

func ( cea CastErrorAssert ) assertUnrecognizedFieldError() {
    if act, ok := cea.ErrAct.( *UnrecognizedFieldError ); ok {
        cea.assertValueError( cea.ErrExpect.( ValueError ), act )
    } else { cea.FailActErrType() }
}

func ( cea CastErrorAssert ) Call() {
    switch cea.ErrExpect.( type ) {
    case nil: cea.Fatal( cea.ErrAct )
    case *ValueCastError: cea.assertVcError()
    case *MissingFieldsError: cea.assertMissingFieldsError()
    case *UnrecognizedFieldError: cea.assertUnrecognizedFieldError()
    default: cea.Fatalf( "Unhandled Err type: %T", cea.ErrExpect )
    }
}

func AssertCastError( expct, act error, pa *assert.PathAsserter ) {
    ca := CastErrorAssert{ ErrExpect: expct, ErrAct: act, PathAsserter: pa }
    ca.Call()
}

func MustRegexRestriction( s string ) *RegexRestriction {
    rx, err := NewRegexRestriction( s )
    if err == nil { return rx }
    panic( err )
}
