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

func registerBoundField0( 
    fsb *mgRct.FunctionsFieldSetBuilder,
    fld *mg.Identifier,
    typ mg.TypeReference,
    set func( fldVal, val interface{} ),
    reg *bind.Registry ) {

    fsb.RegisterField(
        fld,
        func( path objpath.PathNode ) ( mgRct.BuilderFactory, error ) {
            return reg.MustBuilderFactoryForType( typ ), nil
        },
        func( val interface{}, path objpath.PathNode ) error {
            set( val, fsb.Value )
            return nil
        },
    )
}

func builderFactFuncForType( 
    typ mg.TypeReference, reg *bind.Registry ) func() mgRct.BuilderFactory {

    return func() mgRct.BuilderFactory {
        if bf, ok := reg.BuilderFactoryForType( typ ); ok { return bf }
        return nil
    }
}

func createIdSliceBuilderFactory( reg *bind.Registry ) mgRct.BuilderFactory {
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
        identifierParts,
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
            StartField: createIdSliceBuilderFactory,
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
        StartField: createIdSliceBuilderFactory,
        Assign: func( obj, val interface{} ) {
            obj.( *mg.MissingFieldsError ).SetFields( val.( []*mg.Identifier ) )
        },
    })
    return bind.CheckedStructFactory( reg, fact, flds... )
}

func visitIdentifierAsStruct( id *mg.Identifier, es mgRct.EventSender ) error {
    if err := es.StartStruct( mg.QnameIdentifier ); err != nil { return err }
    if err := es.StartField( identifierParts ); err != nil { return err }
    if err := es.StartList( typeIdentifierPartsList ); err != nil { return err }
    for _, part := range id.GetPartsUnsafe() {
        if err := es.Value( mg.String( part ) ); err != nil { return err }
    }
    if err := es.End(); err != nil { return err } // parts
    if err := es.End(); err != nil { return err } // struct
    return nil
}

func VisitIdentifier( id *mg.Identifier, vc bind.VisitContext ) error {
    es := vc.EventSender()
    switch opts := vc.BindContext.SerialOptions; opts.Format {
    case bind.SerialFormatBinary:
        return es.Value( mg.Buffer( mg.IdentifierAsBytes( id ) ) )
    case bind.SerialFormatText:
        return es.Value( mg.String( id.Format( opts.Identifiers ) ) )
    }
    return visitIdentifierAsStruct( id, es )
}

func visitIdentifierList( 
    ids []*mg.Identifier, vc bind.VisitContext ) ( err error ) {

    lt := typeIdentifierPointerList
    switch vc.BindContext.SerialOptions.Format {
    case bind.SerialFormatText: lt = typeNonEmptyStringList
    case bind.SerialFormatBinary: lt = typeNonEmptyBufferList
    }
    es := vc.EventSender()
    if err = es.StartList( lt ); err != nil { return }
    for _, id := range ids {
        if err = VisitIdentifier( id, vc ); err != nil { return }
    }
    if err = es.End(); err != nil { return }
    return
}

func visitNamespaceAsStruct( 
    ns *mg.Namespace, vc bind.VisitContext ) ( err error ) {

    es := vc.EventSender()
    if err = es.StartStruct( mg.QnameNamespace ); err != nil { return }
    if err = es.StartField( identifierParts ); err != nil { return }
    if err = visitIdentifierList( ns.Parts, vc ); err != nil { return }
    if err = es.StartField( identifierVersion ); err != nil { return }
    if err = VisitIdentifier( ns.Version, vc ); err != nil { return }
    if err = es.End(); err != nil { return }
    return
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

type idPathPartsEventSendVisitor struct {
    vc bind.VisitContext
}

func ( vis idPathPartsEventSendVisitor ) Descend( elt interface{} ) error {
    return VisitIdentifier( elt.( *mg.Identifier ), vis.vc )
}

func ( vis idPathPartsEventSendVisitor ) List( idx uint64 ) error {
    return vis.vc.EventSender().Value( mg.Uint64( idx ) )
}

func visitIdPathAsStruct( 
    p objpath.PathNode, vc bind.VisitContext ) ( err error ) {

    es := vc.EventSender()
    if err = es.StartStruct( mg.QnameIdentifierPath ); err != nil { return }
    if err = es.StartField( identifierParts ); err != nil { return }
    if err = es.StartList( typeIdentifierPathPartsList ); err != nil { return }
    if err = objpath.Visit( p, idPathPartsEventSendVisitor{ vc } ); err != nil {
        return
    }
    if err = es.End(); err != nil { return } // parts
    if err = es.End(); err != nil { return } // struct
    return
}

func VisitIdentifierPath( p objpath.PathNode, vc bind.VisitContext ) error {
    if vc.BindContext.SerialOptions.Format == bind.SerialFormatText {
        return vc.EventSender().Value( mg.String( mg.FormatIdPath( p ) ) )
    }
    return visitIdPathAsStruct( p, vc )
}

func visitLocatableError( 
    loc objpath.PathNode, msg string, vc bind.VisitContext ) ( err error ) {

    es := vc.EventSender()
    if loc != nil {
        if err = es.StartField( identifierLocation ); err != nil { return }
        if err = VisitIdentifierPath( loc, vc ); err != nil { return }
    }
    if err = es.StartField( identifierMessage ); err != nil { return }
    if err = es.Value( mg.String( msg ) ); err != nil { return }
    return
}

func VisitCastError( 
    e *mg.CastError, vc bind.VisitContext ) ( err error ) {

    es := vc.EventSender()
    if err = es.StartStruct( mg.QnameCastError ); err != nil { return }
    if err = visitLocatableError( e.Location, e.Message, vc ); err != nil {
        return
    }
    if err = es.End(); err != nil { return }
    return
}

func VisitUnrecognizedFieldError( 
    e *mg.UnrecognizedFieldError, vc bind.VisitContext ) ( err error ) {

    es := vc.EventSender()
    if err = es.StartStruct( mg.QnameUnrecognizedFieldError ); err != nil {
        return
    }
    if err = visitLocatableError( e.Location, e.Message, vc ); err != nil {
        return
    }
    if err = es.StartField( identifierField ); err != nil { return }
    if err = VisitIdentifier( e.Field, vc ); err != nil { return }
    if err = es.End(); err != nil { return }
    return
}

func VisitMissingFieldsError( 
    e *mg.MissingFieldsError, vc bind.VisitContext ) ( err error ) {

    es := vc.EventSender()
    if err = es.StartStruct( mg.QnameMissingFieldsError ); err != nil { return }
    if err = visitLocatableError( e.Location, e.Message, vc ); err != nil {
        return
    }
    if err = es.StartField( identifierFields ); err != nil { return }
    if err = visitIdentifierList( e.Fields(), vc ); err != nil { return }
    if err = es.End(); err != nil { return }
    return
}

func visitBuiltinTypeOk(
    val interface{}, vc bind.VisitContext ) ( error, bool ) {

    switch v := val.( type ) {
    case *mg.Identifier: return VisitIdentifier( v, vc ), true
    case *mg.Namespace: return VisitNamespace( v, vc ), true
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
    reg.MustAddValue( mg.QnameNamespace, newNsBuilderFactory( reg ) )
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
