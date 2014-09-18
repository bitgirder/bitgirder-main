package reactor

import (
    "bitgirder/objpath"
    mg "mingle"
//    "log"
)

type BuildValueOkFunction func ( ve *ValueEvent ) ( interface{}, error, bool )

func NewBuildValueOkFunctionSequence( 
    funcs ...BuildValueOkFunction ) BuildValueOkFunction {

    return func( ve *ValueEvent ) ( interface{}, error, bool ) {
        for _, f := range funcs {
            if val, res, ok := f( ve ); ok { return val, res, ok }
        }
        return nil, nil, false
    }
}

type StructStartFunc func( sse *StructStartEvent ) ( FieldSetBuilder, error )

type FunctionsBuilderFactory struct {
    ErrorFactory BuilderErrorFactory
    ValueFunc BuildValueOkFunction
    ListFunc func( lse *ListStartEvent ) ( ListBuilder, error )
    MapFunc func( mse *MapStartEvent ) ( FieldSetBuilder, error )
    StructFunc StructStartFunc
}

func NewFunctionsBuilderFactory() *FunctionsBuilderFactory {
    return &FunctionsBuilderFactory{}
}

func ( bf *FunctionsBuilderFactory ) failBadInput( ev Event ) error {
    return failBuilderBadInput( ev, bf.ErrorFactory )
}

func ( bf *FunctionsBuilderFactory ) BuildValue( 
    ve *ValueEvent ) ( interface{}, error ) {

    if f := bf.ValueFunc; f != nil {
        if val, err, ok := f( ve ); ok { return val, err }
    }
    return nil, bf.failBadInput( ve )
}

func ( bf *FunctionsBuilderFactory ) implStartField(
    fsb FieldSetBuilder, 
    err error, 
    srcEv Event ) ( FieldSetBuilder, error ) {

    if fsb == nil && err == nil { return nil, bf.failBadInput( srcEv ) }
    return fsb, err
}

func ( bf *FunctionsBuilderFactory ) StartMap( 
    mse *MapStartEvent ) ( FieldSetBuilder, error ) {

    if f := bf.MapFunc; f != nil { 
        res, err := f( mse )
        return bf.implStartField( res, err, mse )
    }
    return nil, bf.failBadInput( mse )
}

func ( bf *FunctionsBuilderFactory ) StartStruct( 
    sse *StructStartEvent ) ( FieldSetBuilder, error ) {

    if f := bf.StructFunc; f != nil { 
        if res, err := f( sse ); ! ( res == nil && err == nil ) {
            return res, err
        }
    }
    return nil, bf.failBadInput( sse )
}

func ( bf *FunctionsBuilderFactory ) StartList( 
    lse *ListStartEvent ) ( ListBuilder, error ) {

    if bf.ListFunc != nil { 
        if lb, err := bf.ListFunc( lse ); lb != nil { return lb, err }
    }
    return nil, bf.failBadInput( lse )
}

type FunctionsListBuilder struct {
    
    Value interface{}

    NextFunc func() BuilderFactory

    AddFunc func( val interface{}, path objpath.PathNode ) error
}

func NewFunctionsListBuilder() *FunctionsListBuilder {
    return &FunctionsListBuilder{}
}

func mustCheckFunc( val interface{}, nm string ) {
    if val == nil { panic( libErrorf( "%s not set", nm ) ) }
}

func ( lb *FunctionsListBuilder ) AddValue( 
    val interface{}, path objpath.PathNode ) ( err error ) {

    mustCheckFunc( lb.AddFunc, "AddFunc" )
    return lb.AddFunc( val, path )
}

func ( lb *FunctionsListBuilder ) NextBuilderFactory() BuilderFactory {
    mustCheckFunc( lb.NextFunc, "NextFunc" )
    return lb.NextFunc()
}

func ( lb *FunctionsListBuilder ) ProduceValue( 
    ee *EndEvent ) ( interface{}, error ) {

    return lb.Value, nil
}

type StartFieldFunction func( path objpath.PathNode ) ( BuilderFactory, error )

type SetFieldValueFunction func( val interface{}, path objpath.PathNode ) error

type fieldBuilderFunctions struct {
    startField StartFieldFunction
    setValue SetFieldValueFunction
}

type FunctionsFieldSetBuilder struct {

    Value interface{}

    FinalValue func() interface{}

    flds *mg.IdentifierMap

    catchall struct { 
        start FieldSetBuilderStartFieldFunction
        set FieldSetBuilderSetValueFunction
    }
}

func NewFunctionsFieldSetBuilder() *FunctionsFieldSetBuilder {
    return &FunctionsFieldSetBuilder{ flds: mg.NewIdentifierMap() }
}

func ( fsb *FunctionsFieldSetBuilder ) RegisterCatchall(
    start FieldSetBuilderStartFieldFunction,
    set FieldSetBuilderSetValueFunction ) {
    
    fsb.catchall.start = start
    fsb.catchall.set = set
}

func ( fsb *FunctionsFieldSetBuilder ) RegisterField(
    fld *mg.Identifier, 
    startFunc StartFieldFunction, 
    setFunc SetFieldValueFunction ) {

    fsb.flds.Put( 
        fld, 
        &fieldBuilderFunctions{ 
            startField: startFunc, 
            setValue: setFunc,
        },
    )
}

func ( fsb *FunctionsFieldSetBuilder ) ProduceValue(
    ee *EndEvent ) ( interface{}, error ) {

    if fsb.FinalValue == nil { return fsb.Value, nil }
    return fsb.FinalValue(), nil
}

func ( fsb *FunctionsFieldSetBuilder ) SetValue(
    fld *mg.Identifier, val interface{}, path objpath.PathNode ) error {

    if v, ok := fsb.flds.GetOk( fld ); ok {
        return v.( *fieldBuilderFunctions ).setValue( val, path )
    }
    if f := fsb.catchall.set; f != nil { return f( fld, val, path ) }
    panic( libErrorf( "SetValue called for %s with no handlers", fld ) )
}

func ( fsb *FunctionsFieldSetBuilder ) StartField(
    fse *FieldStartEvent ) ( BuilderFactory, error ) {

    if val, ok := fsb.flds.GetOk( fse.Field ); ok {
        return val.( *fieldBuilderFunctions ).startField( fse.GetPath() )
    }
    if f := fsb.catchall.start; f != nil {
        if res, err := f( fse ); ! ( res == nil && err == nil ) {
            return res, err
        }
    }
    par := objpath.ParentOf( fse.GetPath() )
    return nil, mg.NewUnrecognizedFieldError( par, fse.Field )
}
