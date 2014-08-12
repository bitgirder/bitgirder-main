package reactor

import (
    mg "mingle"
    "bitgirder/objpath"
    "bitgirder/pipeline"
    "bitgirder/stack"
    "errors"
    "fmt"
//    "log"
)

type BuilderErrorFactory func( path objpath.PathNode, msg string ) error

func defaultBuilderErrorFactory( path objpath.PathNode, msg string ) error {
    return errors.New( mg.FormatError( path, msg ) )
}

func failBuilderBadInput( ev ReactorEvent, errFact BuilderErrorFactory ) error {
    if errFact == nil { errFact = defaultBuilderErrorFactory }
    typ := TypeOfEvent( ev )
    msg := fmt.Sprintf( "unhandled value: %s", typ.ExternalForm() )
    return errFact( ev.GetPath(), msg )
}

type ValueProducer interface {
    ProduceValue( ee *EndEvent ) ( interface{}, error )
}

type FieldSetBuilder interface {

    StartField( fse *FieldStartEvent ) ( BuilderFactory, error )

    SetValue( fld *mg.Identifier, val interface{}, path objpath.PathNode ) error

    ValueProducer
}

type FieldSetBuilderStartFieldFunction func( 
    fse *FieldStartEvent ) ( BuilderFactory, error )

type FieldSetBuilderSetValueFunction func(
    fld *mg.Identifier, val interface{}, path objpath.PathNode ) error

type ListBuilder interface {
    
    AddValue( val interface{}, path objpath.PathNode ) error

    NextBuilderFactory() BuilderFactory

    ValueProducer
}

type BuilderFactory interface {

    BuildValue( ve *ValueEvent ) ( interface{}, error )

    StartMap( mse *MapStartEvent ) ( FieldSetBuilder, error )

    StartStruct( sse *StructStartEvent ) ( FieldSetBuilder, error )

    StartList( lse *ListStartEvent ) ( ListBuilder, error )
}

type BuildReactor struct {
    val interface{}
    hasVal bool
    stk *stack.Stack
    ErrorFactory BuilderErrorFactory
}

func NewBuildReactor( bf BuilderFactory ) *BuildReactor {
    res := &BuildReactor{ stk: stack.NewStack() }
    res.stk.Push( bf )
    return res
}

func ( br *BuildReactor ) HasValue() bool { return br.hasVal }

func ( br *BuildReactor ) GetValue() interface{} {
    if ! br.hasVal { panic( libError( "builder has no value" ) ) }
    return br.val
}

func ( br *BuildReactor ) InitializePipeline( pip *pipeline.Pipeline ) {
    EnsureStructuralReactor( pip )
    EnsurePathSettingProcessor( pip )
}

func ( br *BuildReactor ) completeFieldValue( 
    val interface{}, ev ReactorEvent ) error {

    fld := br.stk.Pop().( *mg.Identifier )
    fsb := br.stk.Peek().( FieldSetBuilder )
    return fsb.SetValue( fld, val, ev.GetPath() )
}

func ( br *BuildReactor ) completeValue( 
    val interface{}, ev ReactorEvent ) error {

    if br.stk.IsEmpty() {
        br.val, br.hasVal = val, true
        return nil
    }
    switch v := br.stk.Peek().( type ) {
    case *mg.Identifier: return br.completeFieldValue( val, ev )
    case ListBuilder: return v.AddValue( val, ev.GetPath() )
    }
    panic( libErrorf( "unhandled value recipient: %T", br.stk.Peek() ) )
}

func ( br *BuildReactor ) nextBuilderFact( 
    ev ReactorEvent ) ( BuilderFactory, error ) {

    top := br.stk.Peek()
    switch v := top.( type ) {
    case BuilderFactory: 
        br.stk.Pop()
        return v, nil
    case ListBuilder: 
        if lb := v.NextBuilderFactory(); lb != nil { return lb, nil }
        return nil, failBuilderBadInput( ev, br.ErrorFactory )
    }
    panic( libErrorf( "unhandled stack element: %T", top ) )
}

func ( br *BuildReactor ) processValue( ve *ValueEvent ) error {
    bf, err := br.nextBuilderFact( ve )
    if err != nil { return err }
    val, err := bf.BuildValue( ve )
    if err != nil { return err }
    return br.completeValue( val, ve )
}

func ( br *BuildReactor ) startFieldSet( 
    fsb FieldSetBuilder, err error ) error {

    if err != nil { return err }
    br.stk.Push( fsb )
    return nil
}

func ( br *BuildReactor ) processMapStart( mse *MapStartEvent ) error {
    bf, err := br.nextBuilderFact( mse )
    if err != nil { return err }
    return br.startFieldSet( bf.StartMap( mse ) )
}

func ( br *BuildReactor ) processStructStart( 
    sse *StructStartEvent ) error {

    bf, err := br.nextBuilderFact( sse )
    if err != nil { return err }
    return br.startFieldSet( bf.StartStruct( sse ) )
}

func ( br *BuildReactor ) processFieldStart( fse *FieldStartEvent ) error {
    fsb := br.stk.Peek().( FieldSetBuilder )
    bf, err := fsb.StartField( fse )
    if err != nil { return err }
    br.stk.Push( fse.Field )
    br.stk.Push( bf )
    return nil
}

func ( br *BuildReactor ) processListStart( lse *ListStartEvent ) error {
    bf, err := br.nextBuilderFact( lse )
    if err != nil { return err }
    lb, err := bf.StartList( lse )
    if err != nil { return err }
    br.stk.Push( lb )
    return nil
}

func ( br *BuildReactor ) processEnd( ee *EndEvent ) error {
    vp := br.stk.Pop().( ValueProducer )
    val, err := vp.ProduceValue( ee )
    if err != nil { return err }
    return br.completeValue( val, ee )
}

func ( br *BuildReactor ) ProcessEvent( ev ReactorEvent ) error {
    if br.hasVal { return libError( "reactor already has a value" ) }
    switch v := ev.( type ) {
    case *ValueEvent: return br.processValue( v )
    case *MapStartEvent: return br.processMapStart( v )
    case *StructStartEvent: return br.processStructStart( v )
    case *FieldStartEvent: return br.processFieldStart( v )
    case *ListStartEvent: return br.processListStart( v )
    case *EndEvent: return br.processEnd( v )
    }
    return libErrorf( "unhandled event: %T", ev )
}

type valBuildFact int

func ( f valBuildFact ) BuildValue( ve *ValueEvent ) ( interface{}, error ) {
    return ve.Val, nil
}

type valFieldSetBuilder struct {
    m *mg.SymbolMap
    res interface{}
}

func ( fsb *valFieldSetBuilder ) StartField( 
    fse *FieldStartEvent ) ( BuilderFactory, error ) {

    return ValueBuilderFactory, nil
}

func ( fsb *valFieldSetBuilder ) SetValue( 
    fld *mg.Identifier, val interface{}, path objpath.PathNode ) error {

    fsb.m.Put( fld, val.( mg.Value ) )
    return nil
}

func ( fsb *valFieldSetBuilder ) ProduceValue( 
    ee *EndEvent ) ( interface{}, error ) {

    return fsb.res, nil
}

func ( f valBuildFact ) StartMap( 
    mse *MapStartEvent ) ( FieldSetBuilder, error ) {

    res := mg.NewSymbolMap()
    return &valFieldSetBuilder{ res, res }, nil
}

func ( f valBuildFact ) StartStruct( 
    sse *StructStartEvent ) ( FieldSetBuilder, error ) {

    res := mg.NewStruct( sse.Type )
    return &valFieldSetBuilder{ res.Fields, res }, nil
}

type mgListBuilder struct { l *mg.List }

func ( lbf *mgListBuilder ) AddValue( 
    val interface{}, path objpath.PathNode ) error {

    lbf.l.AddUnsafe( val.( mg.Value ) )
    return nil
}

func ( lbf *mgListBuilder ) NextBuilderFactory() BuilderFactory {
    return ValueBuilderFactory
}

func ( lbf *mgListBuilder ) ProduceValue( 
    ee *EndEvent ) ( interface{}, error ) {

    return lbf.l, nil
}

func ( f valBuildFact ) StartList( 
    lse *ListStartEvent ) ( ListBuilder, error ) {

    return &mgListBuilder{ mg.NewList( lse.Type ) }, nil
}

var ValueBuilderFactory BuilderFactory = valBuildFact( 1 )
