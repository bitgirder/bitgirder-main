package reactor

import (
    mg "mingle"
    "bitgirder/objpath"
    "bitgirder/pipeline"
    "bitgirder/stack"
    "fmt"
//    "log"
)

type BindError struct {
    Path objpath.PathNode
    Message string
}

func ( e *BindError ) Error() string {
    return mg.FormatError( e.Path, e.Message )
}

func NewBindError( path objpath.PathNode, msg string ) *BindError {
    return &BindError{ Path: path, Message: msg }
}

func NewBindErrorf( 
    path objpath.PathNode, tmpl string, argv ...interface{} ) *BindError {

    return NewBindError( path, fmt.Sprintf( tmpl, argv... ) )
}

func failBinderType( ev ReactorEvent ) error {
    var typ interface { ExternalForm() string }
    switch v := ev.( type ) {
    case *ValueEvent: typ = mg.TypeOf( v.Val )
    case *ListStartEvent: typ = v.Type
    case *MapStartEvent: typ = mg.TypeSymbolMap
    case *StructStartEvent: typ = v.Type
    default: panic( libErrorf( "can't get type for: %T", ev ) )
    }
    return NewBindErrorf( ev.GetPath(), 
        "unhandled value: %s", typ.ExternalForm() )
}

type ValueProducer interface {
    ProduceValue( ee *EndEvent ) ( interface{}, error )
}

type FieldSetBinder interface {

    StartField( fse *FieldStartEvent ) ( BinderFactory, error )

    SetValue( fld *mg.Identifier, val interface{}, path objpath.PathNode ) error

    ValueProducer
}

type ListBinder interface {
    
    AddValue( val interface{}, path objpath.PathNode ) error

    NextBinderFactory() BinderFactory

    ValueProducer
}

type BinderFactory interface {

    BindValue( ve *ValueEvent ) ( interface{}, error )

    StartMap( mse *MapStartEvent ) ( FieldSetBinder, error )

    StartStruct( sse *StructStartEvent ) ( FieldSetBinder, error )

    StartList( lse *ListStartEvent ) ( ListBinder, error )
}

type BindReactor struct {
    val interface{}
    hasVal bool
    stk *stack.Stack
}

func NewBindReactor( bf BinderFactory ) *BindReactor {
    res := &BindReactor{ stk: stack.NewStack() }
    res.stk.Push( bf )
    return res
}

func ( br *BindReactor ) GetValue() interface{} {
    if ! br.hasVal { panic( libError( "binder has no value" ) ) }
    return br.val
}

func ( br *BindReactor ) InitializePipeline( pip *pipeline.Pipeline ) {
    EnsureStructuralReactor( pip )
    EnsurePathSettingProcessor( pip )
}

func ( br *BindReactor ) completeFieldValue( 
    val interface{}, ev ReactorEvent ) error {

    fld := br.stk.Pop().( *mg.Identifier )
    fsb := br.stk.Peek().( FieldSetBinder )
    return fsb.SetValue( fld, val, ev.GetPath() )
}

func ( br *BindReactor ) completeValue( 
    val interface{}, ev ReactorEvent ) error {

    if br.stk.IsEmpty() {
        br.val, br.hasVal = val, true
        return nil
    }
    switch v := br.stk.Peek().( type ) {
    case *mg.Identifier: return br.completeFieldValue( val, ev )
    case ListBinder: return v.AddValue( val, ev.GetPath() )
    }
    panic( libErrorf( "unhandled value recipient: %T", br.stk.Peek() ) )
}

func ( br *BindReactor ) nextBindFact() BinderFactory {
    top := br.stk.Peek()
    switch v := top.( type ) {
    case BinderFactory: 
        br.stk.Pop()
        return v
    case ListBinder: return v.NextBinderFactory()
    }
    panic( libErrorf( "unhandled stack element for nextBindFact(): %T", top ) )
}

func ( br *BindReactor ) processValue( ve *ValueEvent ) error {
    val, err := br.nextBindFact().BindValue( ve )
    if err != nil { return err }
    return br.completeValue( val, ve )
}

func ( br *BindReactor ) startFieldSet( fsb FieldSetBinder, err error ) error {
    if err != nil { return err }
    br.stk.Push( fsb )
    return nil
}

func ( br *BindReactor ) processMapStart( mse *MapStartEvent ) error {
    return br.startFieldSet( br.nextBindFact().StartMap( mse ) )
}

func ( br *BindReactor ) processStructStart( 
    sse *StructStartEvent ) error {

    return br.startFieldSet( br.nextBindFact().StartStruct( sse ) )
}

func ( br *BindReactor ) processFieldStart( fse *FieldStartEvent ) error {
    fsb := br.stk.Peek().( FieldSetBinder )
    bf, err := fsb.StartField( fse )
    if err != nil { return err }
    br.stk.Push( fse.Field )
    br.stk.Push( bf )
    return nil
}

func ( br *BindReactor ) processListStart( lse *ListStartEvent ) error {
    lb, err := br.nextBindFact().StartList( lse )
    if err != nil { return err }
    br.stk.Push( lb )
    return nil
}

func ( br *BindReactor ) processEnd( ee *EndEvent ) error {
    vp := br.stk.Pop().( ValueProducer )
    val, err := vp.ProduceValue( ee )
    if err != nil { return err }
    return br.completeValue( val, ee )
}

func ( br *BindReactor ) ProcessEvent( ev ReactorEvent ) error {
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

type valBindFact int

func ( f valBindFact ) BindValue( ve *ValueEvent ) ( interface{}, error ) {
    return ve.Val, nil
}

type valFieldSetBinder struct {
    m *mg.SymbolMap
    res interface{}
}

func ( fsb *valFieldSetBinder ) StartField( 
    fse *FieldStartEvent ) ( BinderFactory, error ) {

    return ValueBinderFactory, nil
}

func ( fsb *valFieldSetBinder ) SetValue( 
    fld *mg.Identifier, val interface{}, path objpath.PathNode ) error {

    fsb.m.Put( fld, val.( mg.Value ) )
    return nil
}

func ( fsb *valFieldSetBinder ) ProduceValue( 
    ee *EndEvent ) ( interface{}, error ) {

    return fsb.res, nil
}

func ( f valBindFact ) StartMap( 
    mse *MapStartEvent ) ( FieldSetBinder, error ) {

    res := mg.NewSymbolMap()
    return &valFieldSetBinder{ res, res }, nil
}

func ( f valBindFact ) StartStruct( 
    sse *StructStartEvent ) ( FieldSetBinder, error ) {

    res := mg.NewStruct( sse.Type )
    return &valFieldSetBinder{ res.Fields, res }, nil
}

type listBindFact struct { l *mg.List }

func ( lbf *listBindFact ) AddValue( 
    val interface{}, path objpath.PathNode ) error {

    lbf.l.AddUnsafe( val.( mg.Value ) )
    return nil
}

func ( lbf *listBindFact ) NextBinderFactory() BinderFactory {
    return ValueBinderFactory
}

func ( lbf *listBindFact ) ProduceValue( 
    ee *EndEvent ) ( interface{}, error ) {

    return lbf.l, nil
}

func ( f valBindFact ) StartList( lse *ListStartEvent ) ( ListBinder, error ) {
    return &listBindFact{ mg.NewList( lse.Type ) }, nil
}

var ValueBinderFactory BinderFactory = valBindFact( 1 )
