package types

import (
    mgRct "mingle/reactor"
    "mingle/parser"
    mg "mingle"
)

type BindReactor interface {
    mgRct.ReactorEventProcessor
    HasValue() bool
    GetValue() interface{}
}

func asBinderError( ve mgRct.ReactorEvent, err error ) error {
    switch v := err.( type ) {
    case *parser.ParseError:
        err = mg.NewValueCastError( ve.GetPath(), v.Error() )
    case *mg.BinIoError: err = mg.NewValueCastError( ve.GetPath(), v.Error() )
    }
    return err
}

type idBinder struct {
    parts []string
    id *mg.Identifier
}

func ( ib *idBinder ) HasValue() bool { return ib.id != nil }
func ( ib *idBinder ) GetValue() interface{} { return ib.id }

func ( ib *idBinder ) setIdFromValue( ve *mgRct.ValueEvent ) ( err error ) {
    switch v := ve.Val.( type ) {
    case mg.String: ib.id, err = parser.ParseIdentifier( string( v ) )
    case mg.Buffer: ib.id, err = mg.IdentifierFromBytes( []byte( v ) )
    default: err = mgRct.NewReactorErrorf( ve.GetPath(), 
        "attempt to convert %T to identifier", ve.Val )
    }
    if err != nil { err = asBinderError( ve, err ) }
    return
}

func ( ib *idBinder ) ProcessEvent( ev mgRct.ReactorEvent ) error {
    switch v := ev.( type ) {
    case *mgRct.StructStartEvent, *mgRct.FieldStartEvent:;
    case *mgRct.ListStartEvent: ib.parts = make( []string, 0, 2 )
    case *mgRct.ValueEvent:
        if ib.parts == nil { return ib.setIdFromValue( v ) }
        ib.parts = append( ib.parts, string( v.Val.( mg.String ) ) )
    case *mgRct.EndEvent: 
        if ib.id == nil { ib.id = mg.NewIdentifierUnsafe( ib.parts ) }
    default:
        tmpl := "unexpected identifier event: %s"
        evStr := mgRct.EventToString( ev )
        return mgRct.NewReactorErrorf( ev.GetPath(), tmpl, evStr )
    }
    return nil
}

// result reactors are meant to be placed an appropriate typed cast reactor, and
// assume only valid inputs for the standard type definition of typ
func NewBindReactorForType( typ mg.TypeReference ) BindReactor {
    switch {
    case typ.Equals( mg.TypeIdentifier ): return &idBinder{}
    }
    panic( libErrorf( "don't know how to bind: %s", typ ) )
}
