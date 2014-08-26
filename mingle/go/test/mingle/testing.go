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

type errorAssert struct {
    expct error
    act error
    *assert.PathAsserter
}

func ( ea errorAssert ) assertValueCast() {
    expct := ea.expct.( *ValueCastError )
    act, ok := ea.act.( *ValueCastError )
    ea.Truef( ok, "not a value cast error: %T", ea.act )
    ea.Descend( "Message" ).Equal( expct.Message, act.Message )
    ea.Descend( "Location" ).Equal( expct.Location, act.Location )
}

func ( ea errorAssert ) assertMissingFieldsError() {
    expct := ea.expct.( *MissingFieldsError )
    act, ok := ea.act.( *MissingFieldsError )
    ea.Truef( ok, "not a missing fields error: %T", ea.act )
    ea.Descend( "Message" ).Equal( expct.Message, act.Message )
    ea.Descend( "Location" ).Equal( expct.Location, act.Location )
    ea.Descend( "Fields" ).Equal( expct.Fields(), act.Fields() )
}

func AssertErrors( expct, act error, a *assert.PathAsserter ) {
    ea := errorAssert{ expct: expct, act: act, PathAsserter: a }
    switch expct.( type ) {
    case *ValueCastError: ea.assertValueCast()
    case *MissingFieldsError: ea.assertMissingFieldsError()
    default: ea.EqualErrors( ea.expct, ea.act )
    }
}

func assertReadScalar( expct Value, rd *BinReader, a *assert.PathAsserter ) {
    if tc, err := rd.ReadTypeCode(); err == nil {
        if act, err := rd.ReadScalarValue( tc ); err == nil {
            AssertEqualValues( expct, act, a )
        } else {
            a.Fatalf( "couldn't read act: %s", err )
        }
    } else {
        a.Fatalf( "couldn't get type code: %s", err )
    }
}

func AssertBinIoRoundtripRead(
    rd *BinReader, expct interface{}, a *assert.PathAsserter ) {

    switch v := expct.( type ) {
    case Value: assertReadScalar( v, rd, a )
    case *Identifier:
        if id, err := rd.ReadIdentifier(); err == nil { 
            a.True( v.Equals( id ) )
        } else { a.Fatal( err ) }
    case *Namespace:
        if ns, err := rd.ReadNamespace(); err == nil {
            a.True( v.Equals( ns ) )
        } else { a.Fatal( err ) }
    case TypeName:
        if nm, err := rd.ReadTypeName(); err == nil {
            a.True( v.Equals( nm ) )
        } else { a.Fatal( err ) }
    case TypeReference:
        if typ, err := rd.ReadTypeReference(); err == nil {
            a.Truef( v.Equals( typ ), "expct (%v) != act (%v)", v, typ )
        } else { a.Fatal( err ) }
    default: a.Fatalf( "Unhandled expct val: %T", expct )
    }
}
