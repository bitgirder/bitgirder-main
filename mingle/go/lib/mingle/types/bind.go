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
        builderFactFuncForType( mg.TypeIdentifier, reg ),
    )
    return res
}

func nsBuilderForStruct( reg *bind.Registry ) mgRct.FieldSetBuilder {
    res := bind.NewFunctionsFieldSetBuilder()
    res.Value = new( mg.Namespace )
    res.RegisterField(
        idUnsafe( "version" ),
        func( path objpath.PathNode ) ( mgRct.BuilderFactory, error ) {
            if bf, ok := reg.BuilderFactoryForType( mg.TypeIdentifier ); ok {
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

func newIdPathBuilderFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    res := bind.NewFunctionsBuilderFactory()
    return res
}

func init() {
    reg := bind.RegistryForDomain( bind.DomainDefault )
    reg.MustAddValue( mg.QnameIdentifier, newIdBuilderFactory( reg ) )
    reg.MustAddValue( mg.QnameNamespace, newNsBuilderFactory( reg ) )
    reg.MustAddValue( mg.QnameIdentifierPath, newIdPathBuilderFactory( reg ) )
}
