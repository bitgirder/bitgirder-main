package builtin

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

func newDeclNmBuilderFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    return bind.CheckedStructFactory(
        reg,
        func() interface{} { return &mg.DeclaredTypeName{} },
        &bind.CheckedFieldSetter{
            Field: identifierName,
            Type: mg.TypeString,
            Assign: func( obj, val interface{} ) {
                obj.( *mg.DeclaredTypeName ).SetNameUnsafe( val.( string ) )
            },
        },
    )
}

func nsBuilderForStruct( reg *bind.Registry ) mgRct.FieldSetBuilder {
    return bind.CheckedFunctionsFieldSetBuilder(
        reg,
        &mg.Namespace{},
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

func newQnBuilderFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    return bind.CheckedStructFactory(
        reg,
        func() interface{} { return &mg.QualifiedTypeName{} },
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

func idPathPartFromValue( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
    negErr := func() error {
        return mg.NewCastError( ve.GetPath(), "value is negative" )
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

func idPathPartFailBadVal( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
    tmpl := "invalid value for identifier path part: %s"
    err := mg.NewCastErrorf( ve.GetPath(), tmpl, mgRct.TypeOfEvent( ve ) )
    return nil, err, true
}

// note that we have ValueFunc end with idPathPartFailBadVal so that we can fail
// with a CastError instead of the default error. This is to reflect the
// intent of IdentifierPart.parts being typed as Value+, but where the values
// themselves are expected to be of a finite set of types (if we had union types
// we would use that)
func idPathPartBuilderFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    res := bind.NewFunctionsBuilderFactory()
    res.StructFunc = func( 
        sse *mgRct.StructStartEvent ) ( mgRct.FieldSetBuilder, error ) {

        if qn := sse.Type; qn.Equals( mg.QnameIdentifier ) {
            if bf, ok := reg.BuilderFactoryForName( qn ); ok {
                return bf.StartStruct( sse )
            }
        }
        return nil, nil
    }
    res.ValueFunc = mgRct.NewBuildValueOkFunctionSequence(
        idFromBytes, idFromString, idPathPartFromValue, idPathPartFailBadVal )
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

func newCastErrorBuilderFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    return bind.CheckedStructFactory(
        reg,
        func() interface{} { return new( mg.CastError ) },
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
    return bind.CheckedStructFactory( reg, fact, flds... )
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
    return bind.CheckedStructFactory( reg, fact, flds... )
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

func visitNamespaceAsStruct( ns *mg.Namespace, vc bind.VisitContext ) error {
    return bind.VisitStruct( vc, mg.QnameNamespace, func() ( err error ) {
        err = bind.VisitFieldFunc( vc, identifierParts, func() error {
            return visitIdentifierList( ns.Parts, vc )
        })
        if err != nil { return }
        return bind.VisitFieldValue( vc, identifierVersion, ns.Version )
    })
}

func VisitDeclaredTypeName( 
    nm *mg.DeclaredTypeName, vc bind.VisitContext ) error {

    return bind.VisitStruct( vc, mg.QnameDeclaredTypeName, func() error {
        return bind.VisitFieldValue( vc, identifierName, nm.ExternalForm() )
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

func visitLocatableError( 
    loc objpath.PathNode, msg string, vc bind.VisitContext ) error {

    if loc != nil {
        err := bind.VisitFieldValue( vc, identifierLocation, loc )
        if err != nil { return err }
    }
    if msg == "" { return nil }
    return bind.VisitFieldValue( vc, identifierMessage, msg )
}

func VisitCastError( e *mg.CastError, vc bind.VisitContext ) error {
    return bind.VisitStruct( vc, mg.QnameCastError, func() error {
        return visitLocatableError( e.Location, e.Message, vc )
    })
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

func visitBuiltinTypeOk(
    val interface{}, vc bind.VisitContext ) ( error, bool ) {

    switch v := val.( type ) {
    case *mg.Identifier: return VisitIdentifier( v, vc ), true
    case *mg.Namespace: return VisitNamespace( v, vc ), true
    case *mg.DeclaredTypeName: return VisitDeclaredTypeName( v, vc ), true
    case *mg.QualifiedTypeName: return VisitQualifiedTypeName( v, vc ), true
    case objpath.PathNode: return VisitIdentifierPath( v, vc ), true
    case *mg.CastError: return VisitCastError( v, vc ), true
    case *mg.UnrecognizedFieldError: 
        return VisitUnrecognizedFieldError( v, vc ), true
    case *mg.MissingFieldsError: return VisitMissingFieldsError( v, vc ), true
    }
    return nil, false
}

func initBind() {
    reg := bind.RegistryForDomain( bind.DomainDefault )
    reg.MustAddValue( mg.QnameIdentifier, newIdBuilderFactory( reg ) )
    reg.MustAddValue( mg.QnameDeclaredTypeName, newDeclNmBuilderFactory( reg ) )
    reg.MustAddValue( mg.QnameNamespace, newNsBuilderFactory( reg ) )
    reg.MustAddValue( mg.QnameQualifiedTypeName, newQnBuilderFactory( reg ) )
    reg.MustAddValue( mg.QnameIdentifierPath, newIdPathBuilderFactory( reg ) )
    reg.MustAddValue( mg.QnameCastError, newCastErrorBuilderFactory( reg ) )
    reg.MustAddValue( 
        mg.QnameUnrecognizedFieldError,
        newUnrecognizedFieldErrorBuilderFactory( reg ),
    )
    reg.MustAddValue( 
        mg.QnameMissingFieldsError, newMissingFieldsErrorBuilderFactory( reg ) )
    reg.AddVisitValueOkFunc( visitBuiltinTypeOk )
}
