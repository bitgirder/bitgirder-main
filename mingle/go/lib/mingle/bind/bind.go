package bind

import (
    mg "mingle"
    "fmt"
    "bitgirder/pipeline"
    "bitgirder/stack"
    "bitgirder/objpath"
//    "log"
)

func formatError( path objpath.PathNode,
                  f func( objpath.PathNode ) string,
                  msg string ) string {
    if path == nil { return msg }
    return fmt.Sprintf( "%s: %s", f( path ), msg )
}

type BindError struct {
    path objpath.PathNode
    msg string
}

func NewBindError( path objpath.PathNode, msg string ) *BindError {
    return &BindError{ path: path, msg: msg }
}

func NewBindErrorf( 
    path objpath.PathNode, tmpl string, args ...interface{} ) *BindError {

    return NewBindError( path, fmt.Sprintf( tmpl, args... ) )
}

func ( be *BindError ) Error() string {
    f := func( path objpath.PathNode ) string {
        return objpath.Format( path, objpath.StringDotFormatter )
    }
    return formatError( be.path, f, be.msg )
}

type ValueBinding interface { 
    BoundValue( path objpath.PathNode ) ( mg.Value, error ) 
}

type ListBinding interface { 
    HasNextBinding() bool
    NextBinding( path objpath.PathNode ) ( interface{}, error ) 
} 

type FieldSetBinding interface {

    HasNextBinding() bool

    NextFieldBinding( 
        path objpath.PathNode ) ( *mg.Identifier, interface{}, error )
}

type StructBinding interface {
    FieldSetBinding
    BoundType( path objpath.PathNode ) ( *mg.QualifiedTypeName, error )
}

type UnbindError struct { 
    path objpath.PathNode
    msg string
}

func NewUnbindError( path objpath.PathNode, msg string ) *UnbindError {
    return &UnbindError{ path: path, msg: msg }
}

func NewUnbindErrorf( 
    path objpath.PathNode, tmpl string, args ...interface{} ) *UnbindError {

    return NewUnbindError( path, fmt.Sprintf( tmpl, args... ) )
}

func ( ue *UnbindError ) Error() string {
    return formatError( ue.path, mg.FormatIdPath, ue.msg )
}

type Unbinder interface {}

type ValueUnbinder interface {
    UnbindValue( val mg.Value, path objpath.PathNode ) ( interface{}, error )
}

type ListUnbinder interface {

    StartList( path objpath.PathNode ) ( ListUnbinder, error )

    Append( val interface{}, path objpath.PathNode ) ( ListUnbinder, error )

    NextUnbinder( path objpath.PathNode ) ( Unbinder, error )

    EndList( path objpath.PathNode ) ( interface{}, error )
}

type FieldSetUnbinder interface {

    Unbinder

    UnbinderForField( 
        fld *mg.Identifier, path objpath.PathNode ) ( Unbinder, error )
}

type StructUnbinder interface {
    FieldSetUnbinder
    StartStruct( typ *mg.QualifiedTypeName, path objpath.PathNode ) error
}

func isNullValEvent( ev mg.ReactorEvent ) bool {
    if ve, ok := ev.( *mg.ValueEvent ); ok {
        if _, ok = ve.Val.( *mg.Null ); ok { return true }
    }
    return false
}

// assumed to be used downstream of a structural check reactor
type UnbindReactor struct {
    
    // could be nil either because val is still being built, or because nil was
    // actually the built value result. use hasVal to know which
    val interface{} 

    hasVal bool

    stack *stack.Stack // Unbinders
}

func newUnbindReactor() *UnbindReactor { 
    return &UnbindReactor{ stack: stack.NewStack() }
}

func ( ur *UnbindReactor ) InitializePipeline( pip *pipeline.Pipeline ) {
    mg.EnsureStructuralReactor( pip )
}

func ( ur *UnbindReactor ) HasValue() bool { return ur.hasVal }

func ( ur *UnbindReactor ) checkProcessEvent() {
    if ur.stack.IsEmpty() { panic( libError( "empty stack" ) ) }
    if ur.HasValue() { panic( libError( "already have a value" ) ) }
}

func ( ur *UnbindReactor ) UnboundValue() interface{} {
    if ur.hasVal { return ur.val }
    panic( libError( "no value is ready" ) )
}

type listUnbinderContext struct {
    lb ListUnbinder
    accumulating bool
}

func ( ur *UnbindReactor ) pushUnbinder( ub Unbinder ) { 
    var elt interface{}
    switch v := ub.( type ) {
    case ValueUnbinder: elt = v
    case ListUnbinder: elt = &listUnbinderContext{ lb: v }
    default: panic( libErrorf( "unhandled unbinder: %T", ub ) )
    }
    ur.stack.Push( elt )
}

func errorNameForStackElement( elt interface{} ) string {
    switch elt.( type ) {
    case ValueUnbinder: return "value unbind"
    case *listUnbinderContext: return "list unbind"
    }
    return fmt.Sprintf( "%T", elt ) 
}

func ( ur *UnbindReactor ) stackTypeError( 
    path objpath.PathNode, expct string, elt interface{} ) *UnbindError {

    return NewUnbindErrorf( path, "expected %s, got %s", expct, 
        errorNameForStackElement( elt ) )
}

func ( ur *UnbindReactor ) stackTopTypeError(
    path objpath.PathNode, expct string ) *UnbindError {

    return ur.stackTypeError( path, expct, ur.stack.Peek() )
}

func ( ur *UnbindReactor ) stackTypeErrorForEvent(
    path objpath.PathNode, ev mg.ReactorEvent, elt interface{} ) *UnbindError {

    return NewUnbindErrorf( path, "unexpected event %s for %s", 
        mg.EventToString( ev ), errorNameForStackElement( elt ) )
}

func ( ur *UnbindReactor ) completeWithValue( val interface{} ) error {
    ur.val, ur.hasVal = val, true 
    return nil
}

func ( ur *UnbindReactor ) processUnboundValue( 
    val interface{}, path objpath.PathNode ) ( err error ) {

    if ur.stack.IsEmpty() { return ur.completeWithValue( val ) }
    switch v := ur.stack.Peek().( type ) {
    case *listUnbinderContext: 
        v.lb, err = v.lb.Append( val, path )
        return
    }
    return ur.stackTopTypeError( path, "unbound value handler" )
}

func ( ur *UnbindReactor ) valueUnbinderProcessEvent( 
    ev mg.ReactorEvent ) error {

    path := ev.GetPath()
    vu := ur.stack.Pop().( ValueUnbinder )
    ve, ok := ev.( *mg.ValueEvent )
    if ! ok { return ur.stackTypeErrorForEvent( path, ev, vu ) }
    val, err := vu.UnbindValue( ve.Val, path )
    if err != nil { return err }
    return ur.processUnboundValue( val, path )
}

func ( ur *UnbindReactor ) listUnbinderStartValue(
    ev mg.ReactorEvent, lbCtx *listUnbinderContext ) error {
 
    if ! lbCtx.accumulating {
        return NewUnbindErrorf( ev.GetPath(), 
            "expected list start but got %s", mg.EventToString( ev ) )
    }
    if isNullValEvent( ev ) { 
        return ur.processUnboundValue( nil, ev.GetPath() ) 
    }
    ub, err := lbCtx.lb.NextUnbinder( ev.GetPath() )
    if err != nil { return err }
    ur.pushUnbinder( ub )
    return ur.ProcessEvent( ev )
}

func ( ur *UnbindReactor ) listUnbinderStartList( 
    le *mg.ListStartEvent, lbCtx *listUnbinderContext ) ( err error ) {

    if lbCtx.accumulating { return ur.listUnbinderStartValue( le, lbCtx ) }
    if lbCtx.lb, err = lbCtx.lb.StartList( le.GetPath() ); err != nil { return }
    lbCtx.accumulating = true
    return nil
}

func ( ur *UnbindReactor ) listUnbinderEndList(
    ee *mg.EndEvent, lbCtx *listUnbinderContext ) error {

    path := ee.GetPath()
    val, err := lbCtx.lb.EndList( path )
    if err != nil { return err }
    ur.stack.Pop()
    return ur.processUnboundValue( val, path )
}

func ( ur *UnbindReactor ) listUnbinderProcessEvent( 
    ev mg.ReactorEvent ) error {

    lbCtx := ur.stack.Peek().( *listUnbinderContext )
    switch v := ev.( type ) {
    case *mg.ListStartEvent: return ur.listUnbinderStartList( v, lbCtx )
    case *mg.ValueEvent: return ur.listUnbinderStartValue( v, lbCtx )
    case *mg.EndEvent: return ur.listUnbinderEndList( v, lbCtx )
    }
    panic( libErrorf( "unhandled list unbinder event: %T", ev ) )
}

func ( ur *UnbindReactor ) ProcessEvent( ev mg.ReactorEvent ) error {
    ur.checkProcessEvent()
    switch ur.stack.Peek().( type ) {
    case ValueUnbinder: return ur.valueUnbinderProcessEvent( ev )
    case *listUnbinderContext: return ur.listUnbinderProcessEvent( ev )
    }
    panic( libErrorf( "unhandled stack element: %T", ur.stack.Peek() ) )
}

type Binder struct {} 

func NewBinder() *Binder { return &Binder{} }

type bindCall struct { dest mg.ReactorEventProcessor }

func ( b *Binder ) sendBindEvent( 
    ev mg.ReactorEvent, path objpath.PathNode, bc *bindCall ) error {

    ev.SetPath( path )
    return bc.dest.ProcessEvent( ev )
}

func ( b *Binder ) callBindValue( 
    vb ValueBinding, path objpath.PathNode, bc *bindCall ) error {

    val, err := vb.BoundValue( path )
    if err != nil { return err }
    return b.sendBindEvent( mg.NewValueEvent( val ), path, bc )
}

func ( b *Binder ) callBindList( 
    lb ListBinding, path objpath.PathNode, bc *bindCall ) error {

    if err := b.sendBindEvent( mg.NewListStartEvent(), path, bc ); err != nil {
        return err
    }
    lp := objpath.StartList( path )
    for lb.HasNextBinding() {
        binding, err := lb.NextBinding( lp )
        if err != nil { return err }
        if err = b.callBind( binding, lp, bc ); err != nil { return err }
        lp = lp.Next()
    }
    return b.sendBindEvent( mg.NewEndEvent(), path, bc )
} 

func ( b *Binder ) callBind( 
    binding interface{}, path objpath.PathNode, bc *bindCall ) error {

    switch v := binding.( type ) {
    case ValueBinding: return b.callBindValue( v, path, bc )
    case ListBinding: return b.callBindList( v, path, bc )
    }
    panic( libErrorf( "unhandled binding: %T", binding ) )
}

func ( b *Binder ) Bind( 
    binding interface{}, dest mg.ReactorEventProcessor ) error {

    return b.callBind( binding, nil, &bindCall{ dest: dest } )
}

func ( b *Binder ) NewUnbindReactor( ub Unbinder ) *UnbindReactor {
    res := newUnbindReactor()
    res.pushUnbinder( ub )
    return res
}
