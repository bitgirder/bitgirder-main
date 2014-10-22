package builtin

import (
    mgRct "mingle/reactor"
    "mingle/parser"
    "mingle/bind"
    "mingle/types"
    mg "mingle"
    "bitgirder/objpath"
//    "bitgirder/stub"
//    "log"
)

// if err is something that should be sent to caller as a value error, a value
// error is returned; otherwise err is returned unchanged
func asValueError( ve mgRct.Event, err error ) error {
    switch v := err.( type ) {
    case *parser.ParseError:
        err = mg.NewCastError( ve.GetPath(), v.Error() )
    case *mg.BinIoError: err = mg.NewCastError( ve.GetPath(), v.Error() )
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

var idSliceBuilderFactory = bind.CheckedListFieldStarter(
    func() interface{} { return make( []*mg.Identifier, 0, 2 ) },
    bind.ListElementFactoryFuncForType( mg.TypeIdentifier ),
    func( l, val interface{} ) interface{} {
        return append( l.( []*mg.Identifier ), val.( *mg.Identifier ) )
    },
)

var stringSliceBuilderFactory = bind.CheckedListFieldStarter(
    func() interface{} { return make( []string, 0, 2 ) },
    bind.ListElementFactoryFuncForType( mg.TypeString ),
    func( l, val interface{} ) interface{} {
        return append( l.( []string ), val.( string ) )
    },
)

func idBuilderForStruct( reg *bind.Registry ) mgRct.FieldSetBuilder {
    idBuilder := bind.NewFunctionsFieldSetBuilder()
    idBuilder.RegisterField( 
        identifierParts,
        func( path objpath.PathNode ) ( mgRct.BuilderFactory, error ) {
            return stringSliceBuilderFactory( reg ), nil
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

func visitIdentifierAsStruct( id *mg.Identifier, vc bind.VisitContext ) error {
    return bind.VisitStruct( vc, mg.QnameIdentifier, func() error {
        return bind.VisitFieldFunc( vc, identifierParts, func() error {
            parts := id.GetPartsUnsafe()
            l := len( parts )
            f := func( i int ) interface{} { return parts[ i ] }
            return bind.VisitListValue( vc, typeIdentifierPartsList, l, f )
        })
    })
}

func VisitIdentifier( id *mg.Identifier, vc bind.VisitContext ) error {
    es := vc.EventSender()
    switch opts := vc.BindContext.SerialOptions; opts.Format {
    case bind.SerialFormatBinary:
        return es.Value( mg.Buffer( mg.IdentifierAsBytes( id ) ) )
    case bind.SerialFormatText:
        return es.Value( mg.String( id.Format( opts.Identifiers ) ) )
    }
    return visitIdentifierAsStruct( id, vc )
}

func visitIdentifierList( ids []*mg.Identifier, vc bind.VisitContext ) error {
    lt := typeIdentifierPointerList
    switch vc.BindContext.SerialOptions.Format {
    case bind.SerialFormatText: lt = typeNonEmptyStringList
    case bind.SerialFormatBinary: lt = typeNonEmptyBufferList
    }
    return bind.VisitListFunc( vc, lt, len( ids ), func( i int ) error {
        return VisitIdentifier( ids[ i ], vc )
    })
}

func newDeclNmBuilderFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    return bind.CheckedStructFactory(
        reg,
        func() interface{} { return &mg.DeclaredTypeName{} },
        nil,
        &bind.CheckedFieldSetter{
            Field: identifierName,
            Type: mg.TypeString,
            Assign: func( obj, val interface{} ) {
                obj.( *mg.DeclaredTypeName ).SetNameUnsafe( val.( string ) )
            },
        },
    )
}

func VisitDeclaredTypeName( 
    nm *mg.DeclaredTypeName, vc bind.VisitContext ) error {

    return bind.VisitStruct( vc, mg.QnameDeclaredTypeName, func() error {
        return bind.VisitFieldValue( vc, identifierName, nm.ExternalForm() )
    })
}

func visitNamespaceAsStruct( ns *mg.Namespace, vc bind.VisitContext ) error {
    return bind.VisitStruct( vc, mg.QnameNamespace, func() ( err error ) {
        err = bind.VisitFieldFunc( vc, identifierParts, func() error {
            return visitIdentifierList( ns.Parts, vc )
        })
        if err != nil { return }
        return bind.VisitFieldValue( vc, identifierVersion, ns.Version )
    })
}

func VisitNamespace( ns *mg.Namespace, vc bind.VisitContext ) error {
    switch opts := vc.BindContext.SerialOptions; opts.Format {
    case bind.SerialFormatText:
        return vc.EventSender().Value( mg.String( ns.ExternalForm() ) )
    case bind.SerialFormatBinary:
        return vc.EventSender().Value( mg.Buffer( mg.NamespaceAsBytes( ns ) ) )
    }
    return visitNamespaceAsStruct( ns, vc )
}

func nsBuilderForStruct( reg *bind.Registry ) mgRct.FieldSetBuilder {
    return bind.CheckedFunctionsFieldSetBuilder(
        reg,
        &mg.Namespace{},
        nil,
        &bind.CheckedFieldSetter{
            Field: identifierVersion,
            Type: mg.TypeIdentifier,
            Assign: func( obj, val interface{} ) {
                obj.( *mg.Namespace ).Version = val.( *mg.Identifier )
            },
        },
        &bind.CheckedFieldSetter{
            Field: identifierParts,
            StartField: idSliceBuilderFactory,
            Assign: func( obj, val interface{} ) {
                obj.( *mg.Namespace ).Parts = val.( []*mg.Identifier )
            },
        },
    )
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

func VisitQualifiedTypeName( 
    qn *mg.QualifiedTypeName, vc bind.VisitContext ) error {

    return bind.VisitStruct( vc, mg.QnameQualifiedTypeName, func() error {
        if err := bind.VisitFieldValue( vc, identifierNamespace, qn.Namespace );
           err != nil {
            return err
        }
        return bind.VisitFieldValue( vc, identifierName, qn.Name )
    })
}

func newQnBuilderFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    return bind.CheckedStructFactory(
        reg,
        func() interface{} { return &mg.QualifiedTypeName{} },
        nil,
        &bind.CheckedFieldSetter{
            Field: identifierNamespace,
            Type: mg.TypeNamespace,
            Assign: func( obj, val interface{} ) {
                obj.( *mg.QualifiedTypeName ).Namespace = val.( *mg.Namespace )
            },
        },
        &bind.CheckedFieldSetter{
            Field: identifierName,
            Type: mg.TypeDeclaredTypeName,
            Assign: func( obj, val interface{} ) {
                obj.( *mg.QualifiedTypeName ).Name = 
                    val.( *mg.DeclaredTypeName )
            },
        },
    )
}

func VisitRangeRestriction(
    rx *mg.RangeRestriction, vc bind.VisitContext ) error {

    return bind.VisitStruct( vc, mg.QnameRangeRestriction, func() error {
        err := bind.VisitFieldValue( vc, identifierMinClosed, rx.MinClosed() )
        if err != nil { return err }
        optVis := func( val mg.Value, id *mg.Identifier ) error {
            if val == nil { return nil }
            return bind.VisitFieldValue( vc, id, val )
        }
        if err = optVis( rx.Min(), identifierMin ); err != nil { return err }
        if err = optVis( rx.Max(), identifierMax ); err != nil { return err }
        return bind.VisitFieldValue( vc, identifierMaxClosed, rx.MaxClosed() )
    })
}

func VisitRegexRestriction(
    rx *mg.RegexRestriction, vc bind.VisitContext ) error {

    return bind.VisitStruct( vc, mg.QnameRegexRestriction, func() error {
        return bind.VisitFieldValue( vc, identifierPattern, rx.Source() )
    })
}

func VisitAtomicTypeReference(
    at *mg.AtomicTypeReference, vc bind.VisitContext ) error {

    return bind.VisitStruct( vc, mg.QnameAtomicTypeReference, func() error {
        err := bind.VisitFieldValue( vc, identifierName, at.Name() )
        if err != nil { return err }
        rx := at.Restriction()
        if rx == nil { return nil }
        return bind.VisitFieldValue( vc, identifierRestriction, rx )
    })
}

type atomicBuilder struct {
    name *mg.QualifiedTypeName
    rx interface{} // will become a mg.ValueRestriction
}

type regexBuilder struct {
    pat string
}

func ( b *atomicBuilder ) buildInitial() ( *mg.AtomicTypeReference, error ) {
    var rx mg.ValueRestriction
    if b.rx != nil {
        var err error
        switch v := b.rx.( type ) {
        case *mg.RangeRestrictionBuilder: 
            v.Type = b.name
            rx, err = v.Build()
        case *regexBuilder: rx, err = mg.CreateRegexRestriction( v.pat )
        default: panic( libErrorf( "unhandled restriction: %T", b.rx ) )
        }
        if err != nil { return nil, err }
    }
    return mg.CreateAtomicTypeReference( b.name, rx )
}

func ( b *atomicBuilder ) build( 
    path objpath.PathNode ) ( interface{}, error ) {

    at, err := b.buildInitial()
    if re, ok := err.( *mg.RestrictionError ); ok {
        err = mg.NewCastError( path, re.Error() )
    }
    return at, err
}

var atomicNameFieldSetter = &bind.CheckedFieldSetter{
    Field: identifierName,
    Type: mg.TypeQualifiedTypeName,
    Assign: func( obj, fldVal interface{} ) {
        obj.( *atomicBuilder ).name = fldVal.( *mg.QualifiedTypeName )
    },
}

func setOptRangeVal( valPtr *mg.Value, val interface{} ) {
    if val == nil { return }
    *valPtr = mg.MustValue( val )
}

var rangeSetters = []*bind.CheckedFieldSetter{
    &bind.CheckedFieldSetter{
        Field: identifierMinClosed,
        Type: mg.TypeBoolean,
        Assign: func( val, obj interface{} ) {
            val.( *mg.RangeRestrictionBuilder ).MinClosed = obj.( bool )
        },
    },
    &bind.CheckedFieldSetter{
        Field: identifierMin,
        Type: mg.TypeValue,
        Assign: func( val, obj interface{} ) {
            setOptRangeVal( &( val.( *mg.RangeRestrictionBuilder ).Min ), obj )
        },
    },
    &bind.CheckedFieldSetter{
        Field: identifierMax,
        Type: mg.TypeValue,
        Assign: func( val, obj interface{} ) {
            setOptRangeVal( &( val.( *mg.RangeRestrictionBuilder ).Max ), obj )
        },
    },
    &bind.CheckedFieldSetter{
        Field: identifierMaxClosed,
        Type: mg.TypeBoolean,
        Assign: func( val, obj interface{} ) {
            val.( *mg.RangeRestrictionBuilder ).MaxClosed = obj.( bool )
        },
    },
}

func newRangeBuilderBuilder( reg *bind.Registry ) mgRct.FieldSetBuilder {
    return bind.CheckedFunctionsFieldSetBuilder(
        reg, &mg.RangeRestrictionBuilder{}, nil, rangeSetters... )
}

var regexSetters = []*bind.CheckedFieldSetter {
    &bind.CheckedFieldSetter{
        Field: identifierPattern,
        Type: mg.TypeString,
        Assign: func( val, obj interface{} ) {
            val.( *regexBuilder ).pat = obj.( string )
        },
    },
}

func newRegexBuilder( reg *bind.Registry ) mgRct.FieldSetBuilder {
    return bind.CheckedFunctionsFieldSetBuilder(
        reg, &regexBuilder{}, nil, regexSetters... )
}

func newAtomicRestrictionBuilder( reg *bind.Registry ) mgRct.BuilderFactory {
    res := bind.NewFunctionsBuilderFactory()
    res.StructFunc = func( 
        sse *mgRct.StructStartEvent ) ( mgRct.FieldSetBuilder, error ) {

        switch t := sse.Type; {
        case t.Equals( mg.QnameRangeRestriction ):
            return newRangeBuilderBuilder( reg ), nil
        case t.Equals( mg.QnameRegexRestriction ):
            return newRegexBuilder( reg ), nil
        }
        return nil, nil
    }
    return res
}

var atomicRestrictionFieldSetter = &bind.CheckedFieldSetter{
    Field: identifierRestriction,
    Type: mg.TypeValue,
    StartField: func( reg *bind.Registry ) mgRct.BuilderFactory {
        return newAtomicRestrictionBuilder( reg )
    },
    Assign: func( val, obj interface{} ) { val.( *atomicBuilder ).rx = obj },
}

func atomicBuilderForStruct( reg *bind.Registry ) mgRct.FieldSetBuilder {
    res := bind.NewFunctionsFieldSetBuilder()
    res.Value = &atomicBuilder{}
    res.FinalValue = func( path objpath.PathNode ) ( interface{}, error ) { 
        return res.Value.( *atomicBuilder ).build( path ) 
    }
    bind.AddCheckedField( res, reg, atomicNameFieldSetter )
    bind.AddCheckedField( res, reg, atomicRestrictionFieldSetter )
    return res
}

func newAtomicBuilderFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    res := bind.NewFunctionsBuilderFactory()
    setStructFunc( res, reg, atomicBuilderForStruct )
    return res
}

func VisitListTypeReference(
    lt *mg.ListTypeReference, vc bind.VisitContext ) error {

    return bind.VisitStruct( vc, mg.QnameListTypeReference, func() error {
        err := bind.VisitFieldValue( vc, identifierElementType, lt.ElementType )
        if err != nil { return err }
        return bind.VisitFieldValue( vc, identifierAllowsEmpty, lt.AllowsEmpty )
    })
}

func newListTypeBuilderFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    return bind.CheckedStructFactory(
        reg,
        func() interface{} { return &mg.ListTypeReference{} },
        nil,
        &bind.CheckedFieldSetter{
            Field: identifierElementType,
            Type: mg.TypeValue,
            Assign: func( obj, val interface{} ) {
                obj.( *mg.ListTypeReference ).ElementType = 
                    val.( mg.TypeReference )
            },
        },
        &bind.CheckedFieldSetter{
            Field: identifierAllowsEmpty,
            Type: mg.TypeBoolean,
            Assign: func( obj, val interface{} ) {
                obj.( *mg.ListTypeReference ).AllowsEmpty = val.( bool )
            },
        },
    )
}

type typeHolder struct {
    typ mg.TypeReference
}

func VisitPointerTypeReference(
    pt *mg.PointerTypeReference, vc bind.VisitContext ) error {

    return bind.VisitStruct( vc, mg.QnamePointerTypeReference, func() error {
        return bind.VisitFieldValue( vc, identifierType, pt.Type )
    })
}

func newPointerTypeBuilderFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    return bind.CheckedStructFactory(
        reg,
        func() interface{} { return &typeHolder{} },
        bind.CheckedInstanceConvertFunc(
            func( val interface{} ) interface{} {
                return mg.NewPointerTypeReference( val.( *typeHolder ).typ )
            },
        ),
        &bind.CheckedFieldSetter{
            Field: identifierType,
            Type: mg.TypeValue,
            Assign: func( obj, val interface{} ) {
                obj.( *typeHolder ).typ = val.( mg.TypeReference )
            },
        },
    )
}

func VisitNullableTypeReference(
    nt *mg.NullableTypeReference, vc bind.VisitContext ) error {

    return bind.VisitStruct( vc, mg.QnameNullableTypeReference, func() error {
        return bind.VisitFieldValue( vc, identifierType, nt.Type )
    })
}

func newNullableTypeBuilderFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    return bind.CheckedStructFactory(
        reg,
        func() interface{} { return &typeHolder{} },
        bind.CheckedInstanceConvertFunc(
            func( val interface{} ) interface{} {
                return mg.MustNullableTypeReference( val.( *typeHolder ).typ )
            },
        ),
        &bind.CheckedFieldSetter{
            Field: identifierType,
            Type: mg.TypeValue,
            Assign: func( obj, val interface{} ) {
                obj.( *typeHolder ).typ = val.( mg.TypeReference )
            },
        },
    )
}

type idPathPartsEventSendVisitor struct {
    vc bind.VisitContext
}

func ( vis idPathPartsEventSendVisitor ) Descend( elt interface{} ) error {
    return VisitIdentifier( elt.( *mg.Identifier ), vis.vc )
}

func ( vis idPathPartsEventSendVisitor ) List( idx uint64 ) error {
    return vis.vc.EventSender().Value( mg.Uint64( idx ) )
}

func visitIdPathAsStruct( p objpath.PathNode, vc bind.VisitContext ) error {
    return bind.VisitStruct( vc, mg.QnameIdentifierPath, func() error {
        return bind.VisitFieldFunc( vc, identifierParts, func() error {
            body := func() error {
                return objpath.Visit( p, idPathPartsEventSendVisitor{ vc } )
            }
            return bind.VisitList( vc, typeIdentifierPathPartsList, body )
        })
    })
}

func VisitIdentifierPath( p objpath.PathNode, vc bind.VisitContext ) error {
    if vc.BindContext.SerialOptions.Format == bind.SerialFormatText {
        return vc.EventSender().Value( mg.String( mg.FormatIdPath( p ) ) )
    }
    return visitIdPathAsStruct( p, vc )
}

func idPathPartFromUint64( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
    if i, ok := ve.Val.( mg.Uint64 ); ok { return uint64( i ), nil, true }
    return nil, nil, false
}

func idPathPartBuilderFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    res := bind.NewFunctionsBuilderFactory()
    res.StructFunc = func( 
        sse *mgRct.StructStartEvent ) ( mgRct.FieldSetBuilder, error ) {

        bf := reg.MustBuilderFactoryForType( sse.Type.AsAtomicType() )
        return bf.StartStruct( sse )
    }
    res.ValueFunc = mgRct.NewBuildValueOkFunctionSequence(
        idFromBytes, idFromString, idPathPartFromUint64 )
    return res
}

var idPathPartsStarter = bind.CheckedListFieldStarter(
    func() interface{} { return make( []interface{}, 0, 4 ) },
    idPathPartBuilderFactory,
    func( l, val interface{} ) interface{} {
        return append( l.( []interface{} ), val )
    },
)

func buildIdPath( parts []interface{} ) objpath.PathNode {
    var res objpath.PathNode
    for _, part := range parts {
        switch v := part.( type ) {
        case uint64: res = objpath.StartList( res ).SetIndex( v )
        case *mg.Identifier: res = objpath.Descend( res, v )
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
                return idPathPartsStarter( reg ), nil
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

func visitLocatableError( 
    loc objpath.PathNode, msg string, vc bind.VisitContext ) error {

    if loc != nil {
        err := bind.VisitFieldValue( vc, identifierLocation, loc )
        if err != nil { return err }
    }
    if msg == "" { return nil }
    return bind.VisitFieldValue( vc, identifierMessage, msg )
}

func createLocatableErrorSetters( 
    msg, loc bind.CheckedFieldSetValFunction ) []*bind.CheckedFieldSetter {

    return []*bind.CheckedFieldSetter{
        &bind.CheckedFieldSetter{
            Field: identifierMessage,
            Type: mg.TypeString,
            Assign: msg,
        },
        &bind.CheckedFieldSetter{
            Field: identifierLocation,
            Type: mg.TypeIdentifierPath,
            Assign: loc,
        },
    }
}

func VisitCastError( e *mg.CastError, vc bind.VisitContext ) error {
    return bind.VisitStruct( vc, mg.QnameCastError, func() error {
        return visitLocatableError( e.Location, e.Message, vc )
    })
}

func newCastErrorBuilderFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    return bind.CheckedStructFactory(
        reg,
        func() interface{} { return new( mg.CastError ) },
        nil,
        createLocatableErrorSetters(
            func( obj, val interface{} ) {
                obj.( *mg.CastError ).Message = val.( string )
            },
            func( obj, val interface{} ) {
                obj.( *mg.CastError ).Location = val.( objpath.PathNode )
            },
        )...,
    )
}

func VisitUnrecognizedFieldError( 
    e *mg.UnrecognizedFieldError, vc bind.VisitContext ) error {

    return bind.VisitStruct( vc, mg.QnameUnrecognizedFieldError, func() error {
        if err := visitLocatableError( e.Location, e.Message, vc ); err != nil {
            return err
        }
        return bind.VisitFieldValue( vc, identifierField, e.Field )
    })
}

func newUnrecognizedFieldErrorBuilderFactory( 
    reg *bind.Registry ) mgRct.BuilderFactory {

    fact := func() interface{} { return new( mg.UnrecognizedFieldError ) }
    flds := createLocatableErrorSetters(
        func( obj, val interface{} ) {
            obj.( *mg.UnrecognizedFieldError ).Message = val.( string )
        },
        func( obj, val interface{} ) {
            obj.( *mg.UnrecognizedFieldError ).Location = 
                val.( objpath.PathNode )
        },
    )
    flds = append( flds, &bind.CheckedFieldSetter{
        Field: identifierField,
        Type: mg.TypeIdentifier,
        Assign: func( obj, val interface{} ) {
            obj.( *mg.UnrecognizedFieldError ).Field = val.( *mg.Identifier )
        },
    })
    return bind.CheckedStructFactory( reg, fact, nil, flds... )
}

func VisitMissingFieldsError( 
    e *mg.MissingFieldsError, vc bind.VisitContext ) error {

    return bind.VisitStruct( vc, mg.QnameMissingFieldsError, func() error {
        if err := visitLocatableError( e.Location, e.Message, vc ); err != nil {
            return err
        }
        return bind.VisitFieldFunc( vc, identifierFields, func() error {
            return visitIdentifierList( e.Fields(), vc )
        })
    })
}

func newMissingFieldsErrorBuilderFactory( 
    reg *bind.Registry ) mgRct.BuilderFactory {

    fact := func() interface{} { return new( mg.MissingFieldsError ) }
    flds := createLocatableErrorSetters(
        func( obj, val interface{} ) {
            obj.( *mg.MissingFieldsError ).Message = val.( string )
        },
        func( obj, val interface{} ) {
            obj.( *mg.MissingFieldsError ).Location = val.( objpath.PathNode )
        },
    )
    flds = append( flds, &bind.CheckedFieldSetter{
        Field: identifierFields,
        StartField: idSliceBuilderFactory,
        Assign: func( obj, val interface{} ) {
            obj.( *mg.MissingFieldsError ).SetFields( val.( []*mg.Identifier ) )
        },
    })
    return bind.CheckedStructFactory( reg, fact, nil, flds... )
}

func VisitPrimitiveDefinition(
    def *types.PrimitiveDefinition, vc bind.VisitContext ) error {

    return bind.VisitStruct( vc, QnamePrimitiveDefinition, func() error {
        return bind.VisitFieldValue( vc, identifierName, def.Name )
    })
} 

func newPrimitiveDefFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    return bind.CheckedStructFactory(
        reg,
        func() interface{} { return &types.PrimitiveDefinition{} },
        nil,
        &bind.CheckedFieldSetter{
            Field: identifierName,
            Type: mg.TypeQualifiedTypeName,
            Assign: func( obj, val interface{} ) {
                obj.( *types.PrimitiveDefinition ).Name =
                    val.( *mg.QualifiedTypeName )
            },
        },
    )
}

func VisitFieldDefinition(
    def *types.FieldDefinition, vc bind.VisitContext ) error {

    return bind.VisitStruct( vc, QnameFieldDefinition, func() error {
        err := bind.VisitFieldValue( vc, identifierName, def.Name )
        if err != nil { return err }
        err = bind.VisitFieldValue( vc, identifierType, def.Type )
        if err != nil { return err }
        if def.Default != nil {
            err = bind.VisitFieldValue( vc, identifierDefault, def.Default )
            if err != nil { return err }
        }
        return nil
    })
}

func newFieldDefFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    return bind.CheckedStructFactory(
        reg,
        func() interface{} { return &types.FieldDefinition{} },
        nil,
        &bind.CheckedFieldSetter{
            Field: identifierName, 
            Type: mg.TypeIdentifier,
            Assign: func( obj, val interface{} ) {
                obj.( *types.FieldDefinition ).Name = val.( *mg.Identifier )
            },
        },
        &bind.CheckedFieldSetter{
            Field: identifierType,
            Type: mg.TypeValue,
            Assign: func( obj, val interface{} ) {
                obj.( *types.FieldDefinition ).Type = val.( mg.TypeReference )
            },
        },
        &bind.CheckedFieldSetter{
            Field: identifierDefault, 
            Type: mg.TypeValue,
            Assign: func( obj, val interface{} ) {
                obj.( *types.FieldDefinition ).Default = mg.MustValue( val )
            },
        },
    )
}

func VisitFieldSet( fs *types.FieldSet, vc bind.VisitContext ) error {
    return bind.VisitStruct( vc, QnameFieldSet, func() error {
        return bind.VisitFieldFunc( vc, identifierFields, func() error {        
            return bind.VisitList( vc, typeFieldDefList, func() error {
                var err error
                fs.EachDefinition( func( fd *types.FieldDefinition ) {
                    if err != nil { return }
                    err = VisitFieldDefinition( fd, vc )
                })
                return err
            })
        })
    })
}

func newFieldDefListFieldSetBuilder( reg *bind.Registry ) mgRct.ListBuilder {
    res := bind.NewFunctionsListBuilder()
    res.Value = types.NewFieldSet()
    res.NextFunc = func() mgRct.BuilderFactory {
        return reg.MustBuilderFactoryForType( TypeFieldDefinition )
    }
    res.AddFunc = func( val interface{}, path objpath.PathNode ) error {
        fs := res.Value.( *types.FieldSet )
        fd := val.( *types.FieldDefinition )
        if nm := fd.Name; fs.Get( nm ) != nil {
            return mg.NewCastErrorf( path, "field redefined: %s", nm )
        }
        fs.Add( fd )
        return nil
    }
    return res
}

func newFieldDefListFieldSetBuilderFactory( 
    reg *bind.Registry ) mgRct.BuilderFactory {

    res := bind.NewFunctionsBuilderFactory()
    res.ListFunc = func( 
        _ *mgRct.ListStartEvent ) ( mgRct.ListBuilder, error ) {

        return newFieldDefListFieldSetBuilder( reg ), nil
    }
    return res
}

func newFieldSetFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    type fsHolder struct { fs *types.FieldSet }
    return bind.CheckedStructFactory(
        reg,
        func() interface{} { return &fsHolder{} },
        bind.CheckedInstanceConvertFunc(
            func( val interface{} ) interface{} { return val.( *fsHolder ).fs },
        ),
        &bind.CheckedFieldSetter{
            Field: identifierFields,
            StartField: func( reg *bind.Registry ) mgRct.BuilderFactory {
                return newFieldDefListFieldSetBuilderFactory( reg )
            },
            Assign: func( obj, val interface{} ) {
                obj.( *fsHolder ).fs = val.( *types.FieldSet )
            },
        },
    )
}

func VisitUnionTypeDefinition( 
    utd *types.UnionTypeDefinition, vc bind.VisitContext ) error {

    return bind.VisitStruct( vc, QnameUnionTypeDefinition, func() error {
        ln := len( utd.Types )
        f := func( i int ) interface{} { return utd.Types[ i ] }
        return bind.VisitFieldFunc( vc, identifierTypes, func() error {
            return bind.VisitListValue( vc, typeUnionTypeTypesList, ln, f )
        })
    })
}

func newUnionTypeDefFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    type utBldr struct { typs []mg.TypeReference }
    return bind.CheckedStructFactory(
        reg,
        func() interface{} { return &utBldr{} },
        func( val interface{}, path objpath.PathNode ) ( interface{}, error ) {
            utb := val.( *utBldr )
            res, err := types.CreateUnionTypeDefinitionTypes( utb.typs... )
            if err == nil { return res, nil }
            return nil, mg.NewCastError( path, err.Error() )
        },
        &bind.CheckedFieldSetter{
            Field: identifierTypes,
            StartField: bind.CheckedListFieldStarter(
                func() interface{} { return make( []mg.TypeReference, 0, 4 ) },
                bind.ListElementFactoryFuncForType( mg.TypeValue ),
                func( l, val interface{} ) interface{} {
                    typs := l.( []mg.TypeReference )
                    typ := val.( mg.TypeReference )
                    return append( typs, typ )
                },
            ),
            Assign: func( obj, val interface{} ) {
                obj.( *utBldr ).typs = val.( []mg.TypeReference )
            },
        },
    )
}

func VisitUnionDefinition( 
    ud *types.UnionDefinition, vc bind.VisitContext ) error {

    return bind.VisitStruct( vc, QnameUnionDefinition, func() error {
        err := bind.VisitFieldValue( vc, identifierName, ud.Name )
        if err != nil { return err }
        err = bind.VisitFieldValue( vc, identifierUnion, ud.Union )
        return err
    })
}

func newUnionDefFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    return bind.CheckedStructFactory(
        reg,
        func() interface{} { return &types.UnionDefinition{} },
        nil,
        &bind.CheckedFieldSetter{
            Field: identifierName,
            Type: mg.TypeQualifiedTypeName,
            Assign: func( obj, val interface{} ) {
                obj.( *types.UnionDefinition ).Name =
                    val.( *mg.QualifiedTypeName )
            },
        },
        &bind.CheckedFieldSetter{
            Field: identifierUnion,
            Type: TypeUnionTypeDefinition,
            Assign: func( obj, val interface{} ) {
                obj.( *types.UnionDefinition ).Union =
                    val.( *types.UnionTypeDefinition )
            },
        },
    )
}

func VisitCallSignature( 
    cs *types.CallSignature, vc bind.VisitContext ) error {

    return bind.VisitStruct( vc, QnameCallSignature, func() error {
        err := bind.VisitFieldValue( vc, identifierFields, cs.GetFields() )
        if err != nil { return err }
        err = bind.VisitFieldValue( vc, identifierReturn, cs.Return )
        if err != nil { return err }
        if ut := cs.Throws; ut != nil {
            err = bind.VisitFieldValue( vc, identifierThrows, ut )
            if err != nil { return err }
        }
        return nil
    })
}

func newCallSigFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    return bind.CheckedStructFactory(
        reg,
        func() interface{} { return types.NewCallSignature() },
        nil,
        &bind.CheckedFieldSetter{
            Field: identifierFields,
            Type: TypeFieldSet,
            Assign: func( obj, val interface{} ) {
                obj.( *types.CallSignature ).Fields = val.( *types.FieldSet )
            },
        },
        &bind.CheckedFieldSetter{
            Field: identifierReturn,
            Type: mg.TypeValue,
            Assign: func( obj, val interface{} ) {
                obj.( *types.CallSignature ).Return = val.( mg.TypeReference )
            },
        },
        &bind.CheckedFieldSetter{
            Field: identifierThrows,
            Type: TypeUnionTypeDefinition,
            Assign: func( obj, val interface{} ) {
                obj.( *types.CallSignature ).Throws = 
                    val.( *types.UnionTypeDefinition )
            },
        },
    )
}

func VisitPrototypeDefinition(
    pd *types.PrototypeDefinition, vc bind.VisitContext ) error {

    return bind.VisitStruct( vc, QnamePrototypeDefinition, func() error {
        err := bind.VisitFieldValue( vc, identifierName, pd.Name )
        if err != nil { return err }
        err = bind.VisitFieldValue( vc, identifierSignature, pd.Signature )
        return err
    })
}

func newProtoDefFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    return bind.CheckedStructFactory(
        reg,
        func() interface{} { return &types.PrototypeDefinition{} },
        nil,
        &bind.CheckedFieldSetter{
            Field: identifierName,
            Type: mg.TypeQualifiedTypeName,
            Assign: func( obj, val interface{} ) {
                obj.( *types.PrototypeDefinition ).Name =
                    val.( *mg.QualifiedTypeName )
            },
        },
        &bind.CheckedFieldSetter{
            Field: identifierSignature,
            Type: TypeCallSignature,
            Assign: func( obj, val interface{} ) {
                obj.( *types.PrototypeDefinition ).Signature =
                    val.( *types.CallSignature )
            },
        },
    )
}

func VisitStructDefinition(
    sd *types.StructDefinition, vc bind.VisitContext ) error {

    return bind.VisitStruct( vc, QnameStructDefinition, func() error {
        err := bind.VisitFieldValue( vc, identifierName, sd.Name )
        if err != nil { return err }
        err = bind.VisitFieldValue( vc, identifierFields, sd.Fields )
        if err != nil { return err }
        if c := sd.Constructors; c != nil {
            err = bind.VisitFieldValue( vc, identifierConstructors, c )
            if err != nil { return err }
        }
        return nil
    })
}

func newStructDefFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    return bind.CheckedStructFactory(
        reg,
        func() interface{} { return types.NewStructDefinition() },
        nil,
        &bind.CheckedFieldSetter{
            Field: identifierName,
            Type: mg.TypeQualifiedTypeName,
            Assign: func( obj, val interface{} ) {
                obj.( *types.StructDefinition ).Name = 
                    val.( *mg.QualifiedTypeName )
            },
        },
        &bind.CheckedFieldSetter{
            Field: identifierFields,
            Type: TypeFieldSet,
            Assign: func( obj, val interface{} ) {
                obj.( *types.StructDefinition ).Fields = val.( *types.FieldSet )
            },
        },
        &bind.CheckedFieldSetter{
            Field: identifierConstructors,
            Type: TypeUnionTypeDefinition,
            Assign: func( obj, val interface{} ) {
                obj.( *types.StructDefinition ).Constructors =
                    val.( *types.UnionTypeDefinition )
            },
        },
    )
}

func VisitSchemaDefinition( 
    sd *types.SchemaDefinition, vc bind.VisitContext ) error {

    return bind.VisitStruct( vc, QnameSchemaDefinition, func() error {
        err := bind.VisitFieldValue( vc, identifierName, sd.Name )
        if err != nil { return err }
        return bind.VisitFieldValue( vc, identifierFields, sd.Fields )
    })
}

func newSchemaDefFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    return bind.CheckedStructFactory(
        reg,
        func() interface{} { return types.NewSchemaDefinition() },
        nil,
        &bind.CheckedFieldSetter{
            Field: identifierName,
            Type: mg.TypeQualifiedTypeName,
            Assign: func( obj, val interface{} ) {
                obj.( *types.SchemaDefinition ).Name =
                    val.( *mg.QualifiedTypeName )
            },
        },
        &bind.CheckedFieldSetter{
            Field: identifierFields,
            Type: TypeFieldSet,
            Assign: func( obj, val interface{} ) {
                obj.( *types.SchemaDefinition ).Fields = val.( *types.FieldSet )
            },
        },
    )
}

func VisitAliasedTypeDefinition(
    ad *types.AliasedTypeDefinition, vc bind.VisitContext ) error {
    
    return bind.VisitStruct( vc, QnameAliasedTypeDefinition, func() error {
        err := bind.VisitFieldValue( vc, identifierName, ad.Name )
        if err != nil { return err }
        return bind.VisitFieldValue( vc, identifierAliasedType, ad.AliasedType )
    })
}

func newAliasedTypeDefFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    return bind.CheckedStructFactory(
        reg,
        func() interface{} { return &types.AliasedTypeDefinition{} },
        nil,
        &bind.CheckedFieldSetter{
            Field: identifierName,
            Type: mg.TypeQualifiedTypeName,
            Assign: func( obj, val interface{} ) {
                obj.( *types.AliasedTypeDefinition ).Name =
                    val.( *mg.QualifiedTypeName )
            },
        },
        &bind.CheckedFieldSetter{
            Field: identifierAliasedType,
            Type: mg.TypeValue,
            Assign: func( obj, val interface{} ) {
                obj.( *types.AliasedTypeDefinition ).AliasedType =
                    val.( mg.TypeReference )
            },
        },
    )
}

func VisitEnumDefinition(
    ed *types.EnumDefinition, vc bind.VisitContext ) error {

    return bind.VisitStruct( vc, QnameEnumDefinition, func() error {
        err := bind.VisitFieldValue( vc, identifierName, ed.Name )
        if err != nil { return err }
        return bind.VisitFieldFunc( vc, identifierValues, func() error {
            ln := len( ed.Values )
            f := func( i int ) interface{} { return ed.Values[ i ] }
            return bind.VisitListValue( vc, typeIdentifierPointerList, ln, f )
        })
    })
}

func newEnumDefFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    type edBldr struct { nm *mg.QualifiedTypeName; vals []*mg.Identifier }
    return bind.CheckedStructFactory(
        reg,
        func() interface{} { return &edBldr{} },
        func( val interface{}, path objpath.PathNode ) ( interface{}, error ) {
            edb := val.( *edBldr )
            ed, err := types.CreateEnumDefinition( edb.nm, edb.vals... )
            if err == nil { return ed, nil }
            return nil, mg.NewCastError( path, err.Error() )
        },
        &bind.CheckedFieldSetter{
            Field: identifierName,
            Type: mg.TypeQualifiedTypeName,
            Assign: func( obj, val interface{} ) {
                obj.( *edBldr ).nm = val.( *mg.QualifiedTypeName )
            },
        },
        &bind.CheckedFieldSetter{
            Field: identifierValues,
            StartField: idSliceBuilderFactory,
            Assign: func( obj, val interface{} ) {
                obj.( *edBldr ).vals = val.( []*mg.Identifier )
            },
        },
    )
}

func VisitOperationDefinition(
    od *types.OperationDefinition, vc bind.VisitContext ) error {

    return bind.VisitStruct( vc, QnameOperationDefinition, func() error {
        err := bind.VisitFieldValue( vc, identifierName, od.Name )
        if err != nil { return err }
        return bind.VisitFieldValue( vc, identifierSignature, od.Signature )
    })
}

func newOpDefFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    return bind.CheckedStructFactory(
        reg,
        func() interface{} { return &types.OperationDefinition{} },
        nil,
        &bind.CheckedFieldSetter{
            Field: identifierName,
            Type: mg.TypeIdentifier,
            Assign: func( obj, val interface{} ) {
                obj.( *types.OperationDefinition ).Name = val.( *mg.Identifier )
            },
        },
        &bind.CheckedFieldSetter{
            Field: identifierSignature,
            Type: TypeCallSignature,
            Assign: func( obj, val interface{} ) {
                obj.( *types.OperationDefinition ).Signature =
                    val.( *types.CallSignature )
            },
        },
    )
}

func VisitServiceDefinition(
    sd *types.ServiceDefinition, vc bind.VisitContext ) error {

    return bind.VisitStruct( vc, QnameServiceDefinition, func() error {
        err := bind.VisitFieldValue( vc, identifierName, sd.Name )
        if err != nil { return err }
        err = bind.VisitFieldFunc( vc, identifierOperations, func() error {
            ln := len( sd.Operations )
            f := func( i int ) interface{} { return sd.Operations[ i ] }
            return bind.VisitListValue( vc, typeOpDefList, ln, f )
        })
        if err != nil { return err }
        if sec := sd.Security; sec != nil {
            err = bind.VisitFieldValue( vc, identifierSecurity, sec )
            if err != nil { return err }
        }
        return nil
    })
}

func newServiceDefFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    type svcBldr struct { 
        sd *types.ServiceDefinition
        ops []*types.OperationDefinition 
    }
    return bind.CheckedStructFactory(
        reg,
        func() interface{} { 
            return &svcBldr{ sd: types.NewServiceDefinition() }
        },
        func( val interface{}, path objpath.PathNode ) ( interface{}, error ) {
            sb := val.( *svcBldr )
            if err := sb.sd.AddOperations( sb.ops ); err != nil {
                return nil, mg.NewCastError( path, err.Error() )
            }
            return sb.sd, nil
        },
        &bind.CheckedFieldSetter{
            Field: identifierName,
            Type: mg.TypeQualifiedTypeName,
            Assign: func( obj, val interface{} ) {
                obj.( *svcBldr ).sd.Name = val.( *mg.QualifiedTypeName )
            },
        },
        &bind.CheckedFieldSetter{
            Field: identifierOperations,
            StartField: bind.CheckedListFieldStarter(
                func() interface{} { 
                    return make( []*types.OperationDefinition, 0, 4 ) 
                },
                bind.ListElementFactoryFuncForType( TypeOperationDefinition ),
                func( l, val interface{} ) interface{} {
                    ops := l.( []*types.OperationDefinition )
                    return append( ops, val.( *types.OperationDefinition ) )
                },
            ),
            Assign: func( obj, val interface{} ) {
                obj.( *svcBldr ).ops = val.( []*types.OperationDefinition )
            },
        },
        &bind.CheckedFieldSetter{
            Field: identifierSecurity,
            Type: mg.TypeQualifiedTypeName,
            Assign: func( obj, val interface{} ) {
                obj.( *svcBldr ).sd.Security = val.( *mg.QualifiedTypeName )
            },
        },
    )
}

func visitBuiltinTypeOk(
    val interface{}, vc bind.VisitContext ) ( error, bool ) {

    switch v := val.( type ) {
    case *mg.Identifier: return VisitIdentifier( v, vc ), true
    case *mg.Namespace: return VisitNamespace( v, vc ), true
    case *mg.DeclaredTypeName: return VisitDeclaredTypeName( v, vc ), true
    case *mg.QualifiedTypeName: return VisitQualifiedTypeName( v, vc ), true
    case *mg.AtomicTypeReference: return VisitAtomicTypeReference( v, vc ), true
    case *mg.ListTypeReference: return VisitListTypeReference( v, vc ), true
    case *mg.PointerTypeReference: 
        return VisitPointerTypeReference( v, vc ), true
    case *mg.NullableTypeReference:
        return VisitNullableTypeReference( v, vc ), true
    case *mg.RangeRestriction: return VisitRangeRestriction( v, vc ), true
    case *mg.RegexRestriction: return VisitRegexRestriction( v, vc ), true
    case objpath.PathNode: return VisitIdentifierPath( v, vc ), true
    case *mg.CastError: return VisitCastError( v, vc ), true
    case *mg.UnrecognizedFieldError: 
        return VisitUnrecognizedFieldError( v, vc ), true
    case *mg.MissingFieldsError: return VisitMissingFieldsError( v, vc ), true
    case *types.PrimitiveDefinition:
        return VisitPrimitiveDefinition( v, vc ), true
    case *types.FieldDefinition: return VisitFieldDefinition( v, vc ), true
    case *types.FieldSet: return VisitFieldSet( v, vc ), true
    case *types.UnionTypeDefinition:
        return VisitUnionTypeDefinition( v, vc ), true
    case *types.UnionDefinition: return VisitUnionDefinition( v, vc ), true
    case *types.CallSignature: return VisitCallSignature( v, vc ), true
    case *types.PrototypeDefinition: 
        return VisitPrototypeDefinition( v, vc ), true
    case *types.StructDefinition: return VisitStructDefinition( v, vc ), true
    case *types.SchemaDefinition: return VisitSchemaDefinition( v, vc ), true
    case *types.AliasedTypeDefinition:
        return VisitAliasedTypeDefinition( v, vc ), true
    case *types.EnumDefinition: return VisitEnumDefinition( v, vc ), true
    case *types.OperationDefinition: 
        return VisitOperationDefinition( v, vc ), true
    case *types.ServiceDefinition: return VisitServiceDefinition( v, vc ), true
    }
    return nil, false
}

func initCoreBindings( reg *bind.Registry ) {
    reg.MustAddValue( mg.QnameIdentifier, newIdBuilderFactory( reg ) )
    reg.MustAddValue( mg.QnameDeclaredTypeName, newDeclNmBuilderFactory( reg ) )
    reg.MustAddValue( mg.QnameNamespace, newNsBuilderFactory( reg ) )
    reg.MustAddValue( mg.QnameQualifiedTypeName, newQnBuilderFactory( reg ) )
    reg.MustAddValue( 
        mg.QnameAtomicTypeReference, 
        newAtomicBuilderFactory( reg ),
    )
    reg.MustAddValue(
        mg.QnameListTypeReference,
        newListTypeBuilderFactory( reg ),
    )
    reg.MustAddValue(
        mg.QnamePointerTypeReference,
        newPointerTypeBuilderFactory( reg ),
    )
    reg.MustAddValue(
        mg.QnameNullableTypeReference,
        newNullableTypeBuilderFactory( reg ),
    )
    reg.MustAddValue( mg.QnameIdentifierPath, newIdPathBuilderFactory( reg ) )
    reg.MustAddValue( mg.QnameCastError, newCastErrorBuilderFactory( reg ) )
    reg.MustAddValue( 
        mg.QnameUnrecognizedFieldError,
        newUnrecognizedFieldErrorBuilderFactory( reg ),
    )
    reg.MustAddValue( 
        mg.QnameMissingFieldsError, newMissingFieldsErrorBuilderFactory( reg ) )
}

func initTypesBindings( reg *bind.Registry ) {
    reg.MustAddValue( QnamePrimitiveDefinition, newPrimitiveDefFactory( reg ) )
    reg.MustAddValue( QnameFieldDefinition, newFieldDefFactory( reg ) )
    reg.MustAddValue( QnameFieldSet, newFieldSetFactory( reg ) )
    reg.MustAddValue( QnameUnionTypeDefinition, newUnionTypeDefFactory( reg ) )
    reg.MustAddValue( QnameUnionDefinition, newUnionDefFactory( reg ) )
    reg.MustAddValue( QnameCallSignature, newCallSigFactory( reg ) )
    reg.MustAddValue( QnamePrototypeDefinition, newProtoDefFactory( reg ) )
    reg.MustAddValue( QnameStructDefinition, newStructDefFactory( reg ) )
    reg.MustAddValue( QnameSchemaDefinition, newSchemaDefFactory( reg ) )
    reg.MustAddValue(
        QnameAliasedTypeDefinition, newAliasedTypeDefFactory( reg ) )
    reg.MustAddValue( QnameEnumDefinition, newEnumDefFactory( reg ) )
    reg.MustAddValue( QnameOperationDefinition, newOpDefFactory( reg ) )
    reg.MustAddValue( QnameServiceDefinition, newServiceDefFactory( reg ) )
}

func initBind() {
    reg := bind.RegistryForDomain( bind.DomainDefault )
    initCoreBindings( reg )
    initTypesBindings( reg )
    reg.AddVisitValueOkFunc( visitBuiltinTypeOk )
}
