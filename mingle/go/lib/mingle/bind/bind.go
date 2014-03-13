package bind

import (
    mg "mingle"
    "fmt"
    "bitgirder/stack"
    "bitgirder/objpath"
)

func formatError( path objpath.PathNode,
                  f func( objpath.PathNode ) string,
                  msg string ) string {
    if path == nil { return msg }
    return fmt.Sprintf( "%s: %s", f( path ), msg )
}

type ValueBinding interface { BoundValue() ( mg.Value, error ) }

type ListBinding interface { 
    HasNextBinding() bool
    NextBinding() ( interface{}, error ) 
} 

type FieldSetBinding interface {
    HasNextBinding() bool
    NextFieldBinding() ( *mg.Identifier, interface{}, error )
}

type StructBinding interface {
    FieldSetBinding
    BoundType() ( *mg.QualifiedTypeName, error )
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

type Unbinder interface {
    UnboundValue( path objpath.PathNode ) ( interface{}, error )
}

type ValueUnbinder interface {
    Unbinder
    UnbindValue( val mg.Value, path objpath.PathNode ) error
}

type ListUnbinder interface {
    Unbinder
    NewElementUnbinder( path objpath.PathNode ) ( Unbinder, error )
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

func ( ur *UnbindReactor ) HasValue() bool { return ur.hasVal }

func ( ur *UnbindReactor ) UnboundValue() interface{} {
    if ur.hasVal { return ur.val }
    panic( libError( "no value is ready" ) )
}

func ( ur *UnbindReactor ) pushUnbinder( ub Unbinder ) { ur.stack.Push( ub ) }

func ( ur *UnbindReactor ) unbinderDone( 
    ub Unbinder, path objpath.PathNode ) error {

    val, err := ub.UnboundValue( path )
    if err != nil { return err }
    if ur.stack.IsEmpty() { 
        ur.val, ur.hasVal = val, true 
        return nil
    }
    panic( libError( "nested values not implemented" ) )
}

func ( ur *UnbindReactor ) processValue( ve *mg.ValueEvent ) error {
    ub := ur.stack.Pop().( Unbinder )
    vu, ok := ub.( ValueUnbinder )
    path := ve.GetPath()
    if ! ok {
        return NewUnbindErrorf( path, "expected ValueUnbinder, got %T", ub )
    }
    if err := vu.UnbindValue( ve.Val, path ); err != nil { return err }
    return ur.unbinderDone( vu, path )
}

func ( ur *UnbindReactor ) ProcessEvent( ev mg.ReactorEvent ) error {
    if ur.stack.IsEmpty() { panic( libError( "empty stack" ) ) }
    if ur.HasValue() { panic( libError( "already have a value" ) ) }
    switch v := ev.( type ) {
    case *mg.ValueEvent: return ur.processValue( v )
    }
    panic( libErrorf( "unhandled event: %T", ev ) )
}

type Binder struct {} 

func NewBinder() *Binder { return &Binder{} }

type bindCall struct {
    dest mg.ReactorEventProcessor
}

func ( b *Binder ) callBindValue( vb ValueBinding , bc *bindCall ) error {
    val, err := vb.BoundValue()
    if err != nil { return err }
    return bc.dest.ProcessEvent( mg.NewValueEvent( val ) )
}

func ( b *Binder ) callBind( binding interface{}, bc *bindCall ) error {
    switch v := binding.( type ) {
    case ValueBinding: return b.callBindValue( v, bc )
    }
    panic( libErrorf( "unhandled binding: %T", binding ) )
}

func ( b *Binder ) Bind( 
    binding interface{}, dest mg.ReactorEventProcessor ) error {

    return b.callBind( binding, &bindCall{ dest: dest } )
}

func ( b *Binder ) NewUnbindReactor( ub Unbinder ) *UnbindReactor {
    res := newUnbindReactor()
    res.pushUnbinder( ub )
    return res
}
