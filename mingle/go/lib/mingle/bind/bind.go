package bind

import (
    mg "mingle"
    mgRct "mingle/reactor"
    "bitgirder/objpath"
    "bitgirder/pipeline"
    "bitgirder/stack"
    "fmt"
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

type ValueProducer interface {
    ProduceValue( ee *mgRct.EndEvent ) ( interface{}, error )
}

type FieldSetBinder interface {

    StartField( fse *mgRct.FieldStartEvent ) ( BinderFactory, error )

    SetValue( fld *mg.Identifier, val interface{}, path objpath.PathNode ) error

    ValueProducer
}

type ListBinder interface {
    
    AddValue( val interface{}, path objpath.PathNode ) error

    NextBinderFactory() BinderFactory

    ValueProducer
}

type BinderFactory interface {

    BindValue( ve *mgRct.ValueEvent ) ( interface{}, error )

    StartMap( mse *mgRct.MapStartEvent ) ( FieldSetBinder, error )

    StartStruct( sse *mgRct.StructStartEvent ) ( FieldSetBinder, error )

    StartList( lse *mgRct.ListStartEvent ) ( ListBinder, error )
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
    mgRct.EnsureStructuralReactor( pip )
    mgRct.EnsurePathSettingProcessor( pip )
}

func ( br *BindReactor ) completeFieldValue( 
    val interface{}, ev mgRct.ReactorEvent ) error {

    fld := br.stk.Pop().( *mg.Identifier )
    fsb := br.stk.Peek().( FieldSetBinder )
    return fsb.SetValue( fld, val, ev.GetPath() )
}

func ( br *BindReactor ) completeValue( 
    val interface{}, ev mgRct.ReactorEvent ) error {

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

func ( br *BindReactor ) processValue( ve *mgRct.ValueEvent ) error {
    val, err := br.nextBindFact().BindValue( ve )
    if err != nil { return err }
    return br.completeValue( val, ve )
}

func ( br *BindReactor ) startFieldSet( fsb FieldSetBinder, err error ) error {
    if err != nil { return err }
    br.stk.Push( fsb )
    return nil
}

func ( br *BindReactor ) processMapStart( mse *mgRct.MapStartEvent ) error {
    return br.startFieldSet( br.nextBindFact().StartMap( mse ) )
}

func ( br *BindReactor ) processStructStart( 
    sse *mgRct.StructStartEvent ) error {

    return br.startFieldSet( br.nextBindFact().StartStruct( sse ) )
}

func ( br *BindReactor ) processFieldStart( fse *mgRct.FieldStartEvent ) error {
    fsb := br.stk.Peek().( FieldSetBinder )
    bf, err := fsb.StartField( fse )
    if err != nil { return err }
    br.stk.Push( fse.Field )
    br.stk.Push( bf )
    return nil
}

func ( br *BindReactor ) processListStart( lse *mgRct.ListStartEvent ) error {
    lb, err := br.nextBindFact().StartList( lse )
    if err != nil { return err }
    br.stk.Push( lb )
    return nil
}

func ( br *BindReactor ) processEnd( ee *mgRct.EndEvent ) error {
    vp := br.stk.Pop().( ValueProducer )
    val, err := vp.ProduceValue( ee )
    if err != nil { return err }
    return br.completeValue( val, ee )
}

func ( br *BindReactor ) ProcessEvent( ev mgRct.ReactorEvent ) error {
    if br.hasVal { return libError( "reactor already has a value" ) }
    switch v := ev.( type ) {
    case *mgRct.ValueEvent: return br.processValue( v )
    case *mgRct.MapStartEvent: return br.processMapStart( v )
    case *mgRct.StructStartEvent: return br.processStructStart( v )
    case *mgRct.FieldStartEvent: return br.processFieldStart( v )
    case *mgRct.ListStartEvent: return br.processListStart( v )
    case *mgRct.EndEvent: return br.processEnd( v )
    }
    return libErrorf( "unhandled event: %T", ev )
}
