package types

import (
    mgRct "mingle/reactor"
    "mingle/parser"
    "mingle/bind"
    mg "mingle"
    "bitgirder/objpath"
)

// if err is something that should be sent to caller as a value error, a value
// error is returned; otherwise err is returned unchanged
func asValueError( ve mgRct.ReactorEvent, err error ) error {
    switch v := err.( type ) {
    case *parser.ParseError:
        err = mg.NewValueCastError( ve.GetPath(), v.Error() )
    case *mg.BinIoError: err = mg.NewValueCastError( ve.GetPath(), v.Error() )
    }
    return err
}

func idUnsafe( parts ...string ) *mg.Identifier {
    return mg.NewIdentifierUnsafe( parts )
}

func idPartsBuilder( reg *bind.Registry ) mgRct.BuilderFactory {
    res := bind.NewFunctionsBuilderFactory()
    res.ListFunc = func( 
        _ *mgRct.ListStartEvent ) ( mgRct.ListBuilder, error ) {
        
        lb := mgRct.NewFunctionsListBuilder()
        lb.Value = make( []string, 0, 2 )
        lb.NextFunc = func() mgRct.BuilderFactory { 
            return bind.NewBuilderFactory( reg )
        }
        lb.AddFunc = func(
            val interface{}, path objpath.PathNode ) error {

            lb.Value = append( lb.Value.( []string ), val.( string ) )
            return nil
        }
        return lb, nil
    }
    return res
}

func idBuilderForStruct( reg *bind.Registry ) mgRct.FieldSetBuilder {
    idBuilder := mgRct.NewFunctionsFieldSetBuilder()
    idBuilder.RegisterField( 
        idUnsafe( "parts" ),
        func( path objpath.PathNode ) ( mgRct.BuilderFactory, error ) {
            return idPartsBuilder( reg ), nil
        },
        func( val interface{}, path objpath.PathNode ) error {
            idBuilder.Value = mg.NewIdentifierUnsafe( val.( []string ) )
            return nil
        },
    )
    return idBuilder
}

func idFromBytes( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
    if b, ok := ve.Val.( mg.Buffer ); ok {
        res, err := mg.IdentifierFromBytes( []byte( b ) )
        if err != nil { err = asValueError( ve, err ) }
        return res, err, true
    }
    return nil, nil, false
}

func idFromString( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
    if s, ok := ve.Val.( mg.String ); ok {
        res, err := parser.ParseIdentifier( string( s ) )
        if err != nil { err = asValueError( ve, err ) }
        return res, err, true
    }
    return nil, nil, false
}

func newIdBuilderFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    res := bind.NewFunctionsBuilderFactory()
    res.StructFunc = func( 
        sse *mgRct.StructStartEvent ) ( mgRct.FieldSetBuilder, error ) {
        
        return idBuilderForStruct( reg ), nil
    }
    res.ValueFunc = 
        mgRct.NewBuildValueOkFunctionSequence( idFromBytes, idFromString )
    return res
}

func initBinders() {
    reg := bind.RegistryForDomain( bind.DomainDefault )
    reg.MustAddValue( mg.QnameIdentifier, newIdBuilderFactory( reg ) )
}

func init() {
    initBinders()
}
