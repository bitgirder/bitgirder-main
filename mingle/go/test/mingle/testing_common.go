package mingle

import (
    "bitgirder/assert"
    "bitgirder/objpath"
)

type valPtrCheckMap map[ PointerId ] *ValuePointer

func checkEqualTimestamps( 
    expct Timestamp, act Value, a *assert.PathAsserter ) {

    if tmAct, ok := act.( Timestamp ); ok {
        a.Truef( expct.Compare( tmAct ) == 0, 
            "input time was %s, got: %s", expct, tmAct )
    } else {
        a.Fatalf( "expected time, got %T", act )
    }
}

func checkEqualMappedValuePointers( 
    expct, act *ValuePointer, a *assert.PathAsserter, chkMap valPtrCheckMap ) {

    if prev, ok := chkMap[ expct.Id ]; ok {
        a.Equalf( prev.Id, act.Id, 
            "expect value with id %d maps to %d, " +
            "but actual value has id %d: %s",
            expct.Id, prev.Id, act.Id, QuoteValue( act ) )
    } else {
        chkMap[ expct.Id ] = act
        checkEqualValues( expct.Val, act.Val, a, chkMap )
    }
}

func checkEqualValuePointers( 
    expct *ValuePointer,
    actVal Value,
    a *assert.PathAsserter,
    chkMap valPtrCheckMap ) {

    act, ok := actVal.( *ValuePointer )
    a.Truef( ok, "not a value pointer: %T", act )
    if chkMap == nil { 
        checkDirectlyEqual( expct, act, a ) 
        return
    }
    checkEqualMappedValuePointers( expct, act, a, chkMap )
}

func checkEqualLists( 
    expct *List, actVal Value, a *assert.PathAsserter, chkMap valPtrCheckMap ) {

    act, ok := actVal.( *List )
    a.Truef( ok, "not a list: %T", act )
    a.Descend( "(ListLen)" ).Equal( expct.Len(), act.Len() )
    la := a.StartList()
    for i, e := 0, expct.Len(); i < e; i++ {
        checkEqualValues( expct.Get( i ), act.Get( i ), la, chkMap )
        la = la.Next()
    }
}

func checkDirectlyEqual( expct, act Value, a *assert.PathAsserter ) {
    a.Equalf( expct, act, "expected %s (%T) but got %s (%T)",
        QuoteValue( expct ), expct, QuoteValue( act ), act )
}

func checkEqualMaps(
    expct *SymbolMap,
    actVal Value,
    a *assert.PathAsserter,
    chkMap valPtrCheckMap ) {

    act, ok := actVal.( *SymbolMap )
    a.Truef( ok, "not a map: %T", actVal )
    expctKeys, actKeys := SortIds( expct.GetKeys() ), SortIds( act.GetKeys() )
    a.Equalf( expctKeys, actKeys, "expected fields %s, got %s",
        idSliceToString( expctKeys ), idSliceToString( actKeys ) )
    for _, fld := range expctKeys {
        fldValExpct, fldValAct := expct.Get( fld ), act.Get( fld )
        checkEqualValues( fldValExpct, fldValAct, a.Descend( fld ), chkMap )
    }
}

func checkEqualStructs( 
    expct *Struct, 
    actVal Value, 
    a *assert.PathAsserter, 
    chkMap valPtrCheckMap ) {

    act, ok := actVal.( *Struct )
    a.Truef( ok, "not a struct: %T", actVal )
    a.Descend( "$type" ).Equal( expct.Type, act.Type )
    checkEqualMaps( expct.Fields, act.Fields, a, chkMap )
}

func checkEqualValues( 
    expct, act Value, a *assert.PathAsserter, chkMap valPtrCheckMap ) {

    switch v := expct.( type ) {
    case Timestamp: checkEqualTimestamps( v, act, a )
    case *ValuePointer: checkEqualValuePointers( v, act, a, chkMap )
    case *List: checkEqualLists( v, act, a, chkMap )
    case *Struct: checkEqualStructs( v, act, a, chkMap )
    case *SymbolMap: checkEqualMaps( v, act, a, chkMap )
    default: checkDirectlyEqual( expct, act, a )
    }
}

func equalValues( expct, act Value, f assert.Failer, chkMap valPtrCheckMap ) {
    a := assert.NewPathAsserter( f )
    checkEqualValues( expct, act, a, chkMap )
}

func EqualWireValues( expct, act Value, f assert.Failer ) {
    equalValues( expct, act, f, map[ PointerId ] *ValuePointer{} )
}

func EqualValues( expct, act Value, f assert.Failer ) {
    equalValues( expct, act, f, nil )
}

func EqualPaths( expct, act objpath.PathNode, a assert.Failer ) {
    ( &assert.Asserter{ a } ).Equalf( 
        expct, 
        act,
        "expected path %q but got %q", FormatIdPath( expct ),
            FormatIdPath( act ),
    )
}

func typeRef( s string ) TypeReference { return MustTypeReference( s ) }

var qname = MustQualifiedTypeName

func atomicRef( s string ) *AtomicTypeReference {
    return typeRef( s ).( *AtomicTypeReference )
}

var id = MustIdentifier
