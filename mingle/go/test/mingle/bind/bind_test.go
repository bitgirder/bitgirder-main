package bind

import (
    mg "mingle"
    "testing"
    "bitgirder/assert"
)

var mkTestId = mg.MakeTestId

func mkTestStruct( typNm string, flds ...interface{} ) *mg.Struct {
    return mg.MustStruct( "mingle:bind:test@v1/" + typNm, flds... )
}

type bindTestType1 struct {
    f1 int32
}

type testUnbindReactor struct { vb *mg.ValueBuilder }

func newTestUnbindReactor() *testUnbindReactor {
    return &testUnbindReactor{ vb: mg.NewValueBuilder() }
}

func ( ub *testUnbindReactor ) ProcessEvent( ev mg.ReactorEvent ) error {
    return ub.vb.ProcessEvent( ev )
}

func ( ub *testUnbindReactor ) newBindTestType1( 
    acc *mg.SymbolMapAccessor ) ( interface{}, error ) {

    res := &bindTestType1{}
    f1 := mkTestId( 1 )
    m := acc.GetMap()
    if m.HasField( f1 ) { 
        res.f1 = acc.MustGoInt32ById( f1 ) 
    } else { 
        return nil, mg.NewMissingFieldsError( nil, []*mg.Identifier{ f1 } ) 
    }
    return res, nil
}

func ( ub *testUnbindReactor ) GoValue() ( interface{}, error ) {
    val := ub.vb.GetValue()
    ms := val.( *mg.Struct )
    acc := mg.NewSymbolMapAccessor( ms.Fields, nil )
    switch ms.Type.Name.ExternalForm() {
    case "BindTestType1": return ub.newBindTestType1( acc )
    }
    panic( libErrorf( "unhandled unbind type: %s", ms.Type ) )
}

type testBinder struct {}

func newTestBinder() *testBinder { return &testBinder{} }

func ( tb *testBinder ) asMingleValue( goVal interface{} ) mg.Value {
    switch v := goVal.( type ) {
    case *bindTestType1: return mkTestStruct( "BindTestType1", "f1", v.f1 )
    }
    panic( libErrorf( "unhandled type: %T", goVal ) )
}

func ( tb *testBinder ) BindToReactor( goVal interface{},
                                       rct mg.ReactorEventProcessor,
                                       ctx *BinderContext ) error {

    val := tb.asMingleValue( goVal )
    return mg.VisitValue( val, rct )
}

func ( tb *testBinder ) NewUnbindReactor( ctx *BinderContext ) UnbindReactor {
    return newTestUnbindReactor()
}

func TestBinderInterface( t *testing.T ) {
    a := &assert.Asserter{ t }
    val := &bindTestType1{ f1: 1 }
    bndr := newTestBinder()
    mgValExpct := mkTestStruct( "BindTestType1", "f1", int32( 1 ) )
    vb := mg.NewValueBuilder()
    ctx := NewBinderContext()
    if err := bndr.BindToReactor( val, vb, ctx ); err != nil { a.Fatal( err ) }
    mg.EqualValues( mgValExpct, vb.GetValue(), a )    
    rct := bndr.NewUnbindReactor( ctx )
    if err := mg.VisitValue( vb.GetValue(), rct ); err != nil { a.Fatal( err ) }
    if val2, err := rct.GoValue(); err != nil { 
        a.Fatal( err ) 
    } else { a.Equal( val, val2 ) }
}

// the choice of error is somewhat arbitrary -- we just choose a missing fields
// error since that is representative of the type of error an unbinder might
// return. the main goal here is simply to test that the interfaces provide a
// clear way for an unbinder to return such an error in the normal reactor flow
func TestUnbindReactorErrorPropagation( t *testing.T ) {
    bndr := newTestBinder()
    ctx := NewBinderContext()
    rct := bndr.NewUnbindReactor( ctx )
    val := mkTestStruct( "BindTestType1" )
    var errAct error
    if errAct = mg.VisitValue( val, rct ); errAct == nil {
        var ubVal interface{}
        if ubVal, errAct = rct.GoValue(); errAct == nil {
            t.Fatalf( "got value: %v", ubVal )
        }
    }
    flds := []*mg.Identifier{ mg.MustIdentifier( "f1" ) }
    expctErr := mg.NewMissingFieldsError( nil, flds )
    pa := assert.NewPathAsserter( t )
    mg.AssertCastError( expctErr, errAct, pa )
}
