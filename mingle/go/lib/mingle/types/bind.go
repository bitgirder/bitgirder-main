package types

import (
    mgRct "mingle/reactor"
    "mingle/parser"
    "mingle/bind"
    mg "mingle"
    "bitgirder/objpath"
//    "log"
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

func setStructFunc( 
    b *mgRct.FunctionsBuilderFactory,
    reg *bind.Registry,
    f func( *bind.Registry ) mgRct.FieldSetBuilder ) {

    b.StructFunc = func( 
        _ *mgRct.StructStartEvent ) ( mgRct.FieldSetBuilder, error ) {

        return f( reg ), nil
    }
}

func setListFunc(
    b *mgRct.FunctionsBuilderFactory, 
    valFact func() interface{},
    addVal func( val, acc interface{} ) interface{},
    nextFunc func() mgRct.BuilderFactory ) {

    b.ListFunc = func( _ *mgRct.ListStartEvent ) ( mgRct.ListBuilder, error ) {
        lb := bind.NewFunctionsListBuilder()
        lb.Value = valFact()
        lb.AddFunc = func( val interface{}, path objpath.PathNode ) error {
            lb.Value = addVal( val, lb.Value )
            return nil
        }
        lb.NextFunc = nextFunc
        return lb, nil
    }
}

func builderFactFuncForType( 
    typ mg.TypeReference, reg *bind.Registry ) func() mgRct.BuilderFactory {

    return func() mgRct.BuilderFactory {
        if bf, ok := reg.BuilderFactoryForType( typ ); ok { return bf }
        return nil
    }
}

func idPartsBuilderFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    res := bind.NewFunctionsBuilderFactory()
    setListFunc(
        res,
        func() interface{} { return make( []string, 0, 2 ) },
        func( val, acc interface{} ) interface{} {
            return append( acc.( []string ), val.( string ) )
        },
        builderFactFuncForType( mg.TypeString, reg ),
    )
    return res
}

func idBuilderForStruct( reg *bind.Registry ) mgRct.FieldSetBuilder {
    idBuilder := bind.NewFunctionsFieldSetBuilder()
    idBuilder.RegisterField( 
        idUnsafe( "parts" ),
        func( path objpath.PathNode ) ( mgRct.BuilderFactory, error ) {
            return idPartsBuilderFactory( reg ), nil
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
    setStructFunc( res, reg, idBuilderForStruct )
    res.ValueFunc = 
        mgRct.NewBuildValueOkFunctionSequence( idFromBytes, idFromString )
    return res
}

func nsPartsBuilderFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    res := bind.NewFunctionsBuilderFactory()
    setListFunc(
        res,
        func() interface{} { return make( []*mg.Identifier, 0, 2 ) },
        func( val, acc interface{} ) interface{} {
            return append( acc.( []*mg.Identifier ), val.( *mg.Identifier ) )
        },
        builderFactFuncForType( TypeIdentifier, reg ),
    )
    return res
}

func nsBuilderForStruct( reg *bind.Registry ) mgRct.FieldSetBuilder {
    res := bind.NewFunctionsFieldSetBuilder()
    res.Value = new( mg.Namespace )
    res.RegisterField(
        idUnsafe( "version" ),
        func( path objpath.PathNode ) ( mgRct.BuilderFactory, error ) {
            if bf, ok := reg.BuilderFactoryForType( TypeIdentifier ); ok {
                return bf, nil
            }
            return nil, nil
        },
        func( val interface{}, path objpath.PathNode ) error {
            res.Value.( *mg.Namespace ).Version = val.( *mg.Identifier )
            return nil
        },
    )
    res.RegisterField(
        idUnsafe( "parts" ),
        func( path objpath.PathNode ) ( mgRct.BuilderFactory, error ) {
            return nsPartsBuilderFactory( reg ), nil
        },
        func( val interface{}, path objpath.PathNode ) error {
            res.Value.( *mg.Namespace ).Parts = val.( []*mg.Identifier )
            return nil
        },
    )
    return res
}

func nsFromBytes( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
    if b, ok := ve.Val.( mg.Buffer ); ok {
        res, err := mg.NamespaceFromBytes( []byte( b ) )
        if err != nil { err = asValueError( ve, err ) }
        return res, err, true
    }
    return nil, nil, false
}

func nsFromString( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
    if s, ok := ve.Val.( mg.String ); ok {
        res, err := parser.ParseNamespace( string( s ) )
        if err != nil { err = asValueError( ve, err ) }
        return res, err, true
    }
    return nil, nil, false
}

func newNsBuilderFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    res := bind.NewFunctionsBuilderFactory()
    setStructFunc( res, reg, nsBuilderForStruct )
    res.ValueFunc = 
        mgRct.NewBuildValueOkFunctionSequence( nsFromBytes, nsFromString )
    return res
}

func idPartFromValue( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
    negErr := func() error {
        return mg.NewValueCastError( ve.GetPath(), "value is negative" )
    }
    switch v := ve.Val.( type ) {
    case mg.Int32:
        if int32( v ) < 0 { return nil, negErr(), true }
        return uint64( int32( v ) ), nil, true
    case mg.Int64:
        if int64( v ) < 0 { return nil, negErr(), true }
        return uint64( int64( v ) ), nil, true
    case mg.Uint32: return uint64( uint32( v ) ), nil, true
    case mg.Uint64: return uint64( v ), nil, true
    }
    return nil, nil, false
}

func idPathPartBuilderFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    res := bind.NewFunctionsBuilderFactory()
    res.StructFunc = func( 
        sse *mgRct.StructStartEvent ) ( mgRct.FieldSetBuilder, error ) {

        if qn := sse.Type; qn.Equals( QnameIdentifier ) {
            if bf, ok := reg.BuilderFactoryForName( qn ); ok {
                return bf.StartStruct( sse )
            }
        }
        return nil, nil
    }
    res.ValueFunc = mgRct.NewBuildValueOkFunctionSequence(
        idFromBytes, idFromString, idPartFromValue )
    return res
}

func idPathPartsBuilder( reg *bind.Registry ) mgRct.BuilderFactory {
    res := bind.NewFunctionsBuilderFactory()
    setListFunc( 
        res,
        func() interface{} { return make( []interface{}, 0, 4 ) },
        func( val, acc interface{} ) interface{} {
            return append( acc.( []interface{} ), val )
        },
        func() mgRct.BuilderFactory { return idPathPartBuilderFactory( reg ) },
    )
    return res
}

func buildIdPath( parts []interface{} ) objpath.PathNode {
    var res objpath.PathNode
    for _, part := range parts {
        switch v := part.( type ) {
        case uint64:
            if res == nil { 
                res = objpath.RootedAtList().SetIndex( v )
            } else {
                res = res.StartList().SetIndex( v )
            }
        case *mg.Identifier:
            if res == nil {
                res = objpath.RootedAt( v )
            } else {
                res = res.Descend( v )
            }
        default: panic( libErrorf( "unhandled id path part: %T", part ) )
        }
    }
    return res
}

func idPathFromString( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
    if s, ok := ve.Val.( mg.String ); ok {
        res, err := parser.ParseIdentifierPath( string( s ) )
        if err != nil { err = asValueError( ve, err ) }
        return res, err, true
    }
    return nil, nil, false
}

func newIdPathBuilderFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    res := bind.NewFunctionsBuilderFactory()
    setStructFunc( res, reg, func( reg *bind.Registry ) mgRct.FieldSetBuilder {
        res := bind.NewFunctionsFieldSetBuilder()
        res.RegisterField(
            idUnsafe( "parts" ),
            func( path objpath.PathNode ) ( mgRct.BuilderFactory, error ) {
                return idPathPartsBuilder( reg ), nil
            },
            func( val interface{}, path objpath.PathNode ) error {
                res.Value = buildIdPath( val.( []interface{} ) )
                return nil
            },
        )
        return res
    })
    res.ValueFunc = idPathFromString
    return res
}

func initBind() {
    reg := bind.RegistryForDomain( bind.DomainDefault )
    reg.MustAddValue( QnameIdentifier, newIdBuilderFactory( reg ) )
    reg.MustAddValue( QnameNamespace, newNsBuilderFactory( reg ) )
    reg.MustAddValue( QnameIdentifierPath, newIdPathBuilderFactory( reg ) )
}
