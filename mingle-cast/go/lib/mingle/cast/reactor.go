package cast

import (
    mg "mingle"
    mgRct "mingle/reactor"
    "mingle/parser"
    "mingle/types"
    "bitgirder/objpath"
    "bitgirder/pipeline"
    "bitgirder/stack"
    "log"
    "fmt"
    "bytes"
    "strings"
    "encoding/base64"
)

func canAssignToStruct( 
    targ *types.StructDefinition, 
    def types.Definition, 
    dm types.DefinitionGetter ) bool {

    sd, ok := def.( *types.StructDefinition )
    if ! ok { return false }
    return targ.GetName().Equals( sd.GetName() ) 
}

func canAssignToSchema(
    targ *types.SchemaDefinition, 
    def types.Definition, 
    dm types.DefinitionGetter ) bool {

    if sd, ok := def.( *types.StructDefinition ); ok {
        log.Printf( "checking whether %s satisfies schema %s", 
            sd.Name, targ.Name )
        return sd.SatisfiesSchema( targ )
    }
    return false
}

func canAssignType( 
    t1, t2 *mg.QualifiedTypeName, dm types.DefinitionGetter ) bool {

    d1 := types.MustGetDefinition( t1, dm )
    d2 := types.MustGetDefinition( t2, dm )
    switch v1 := d1.( type ) {
    case *types.StructDefinition: return canAssignToStruct( v1, d2, dm )
    case *types.SchemaDefinition: return canAssignToSchema( v1, d2, dm )
    }
    return false
}

func newNullInputError( path objpath.PathNode ) *mg.InputError {
    return mg.NewInputError( path, "Value is null" )
}

func asMapStartEvent( ev mgRct.Event ) *mgRct.MapStartEvent {
    res := mgRct.NewMapStartEvent() 
    res.SetPath( ev.GetPath() )
    return res
}

func notAFieldSetTypeError( 
    p objpath.PathNode, qn *mg.QualifiedTypeName ) error {

    return mg.NewInputErrorf( p, "not a type with fields: %s", qn )
}

func fieldSetForTypeInDefMap(
    qn *mg.QualifiedTypeName, 
    dm types.DefinitionGetter, 
    path objpath.PathNode ) ( *types.FieldSet, error ) {

    if def, ok := dm.GetDefinition( qn ); ok {
        switch v := def.( type ) {
        case *types.StructDefinition: return v.Fields, nil
        case *types.SchemaDefinition: return v.Fields, nil
        default: return nil, notAFieldSetTypeError( path, qn )
        } 
    } 
    return nil, mg.NewInputErrorf( path, "unrecognized type: %s", qn )
}

type fieldTyper interface {

    // path will be positioned to the map/struct containing fld, but will not
    // itself include fld
    fieldTypeFor( 
        fld *mg.Identifier, path objpath.PathNode ) ( mg.TypeReference, error )
}

type valueFieldTyper int

func ( vt valueFieldTyper ) fieldTypeFor( 
    fld *mg.Identifier, path objpath.PathNode ) ( mg.TypeReference, error ) {
    return mg.TypeNullableValue, nil
}

type SymbolMapFieldSetGetter interface {
    GetFieldSet( path objpath.PathNode ) ( *types.FieldSet, error )
}

type TypeErrorFormatter func( 
    expct, act mg.TypeReference, path objpath.PathNode ) ( error, bool )

type CastReactor struct {

    dm types.DefinitionGetter

    stack *stack.Stack

    passFieldsByQn *mg.QnameMap

    passthroughTracker *mgRct.DepthTracker

    unionMatchFuncs *mg.QnameMap

    FieldSetFactory SymbolMapFieldSetGetter

    FormatTypeError TypeErrorFormatter

    SkipPathSetter bool
}

func ( cr *CastReactor ) dumpStack( pref string ) {
    bb := &bytes.Buffer{}
    fmt.Fprintf( bb, "%s: [", pref )
    cr.stack.VisitTop( func( v interface{} ) {
        msg := fmt.Sprintf( "%T", v )
        switch v2 := v.( type ) {
        case mg.TypeReference: msg = v2.ExternalForm()
        }
        fmt.Fprintf( bb, msg )
        fmt.Fprintf( bb, ", " )
    })
    fmt.Fprintf( bb, " ]" )
    log.Print( bb.String() )
}

func ( cr *CastReactor ) newTypeInputError(
    expct, act mg.TypeReference, path objpath.PathNode ) error {

    if f := cr.FormatTypeError; f != nil {
        if err, ok := f( expct, act, path ); ok { return err }
    }
    return mg.NewTypeInputError( expct, act, path )
}

func ( cr *CastReactor ) newTypeInputErrorValue(
    expct mg.TypeReference, val mg.Value, path objpath.PathNode ) error {

    return cr.newTypeInputError( expct, mg.TypeOf( val ), path )
}

func ( cr *CastReactor ) pushType( typ mg.TypeReference ) {
    cr.stack.Push( typ )
}

func matchIdPathPart( in types.UnionMatchInput ) ( mg.TypeReference, bool ) {
    if typ, ok := in.Union.MatchType( in.TypeIn ); ok { return typ, ok }
    def, ok := in.Definitions.GetDefinition( mg.QnameIdentifier )
    if ! ok { panic( libErrorf( "no definition for Identifier" ) ) }
    sd := def.( *types.StructDefinition )
    if typ, ok := sd.Constructors.MatchType( in.TypeIn ); ok { return typ, ok }
    if at, ok := in.TypeIn.( *mg.AtomicTypeReference ); ok {
        if mg.IsIntegerTypeName( at.Name() ) { return mg.TypeUint64, true }
    }
    return nil, false
}

// for now this is always enabled, but we may later find a need to allow callers
// to create a CastReactor with no support for builtins
func ( cr *CastReactor ) castBuiltinTypes() {
    cr.SetUnionDefinitionMatcher( mg.QnameIdentifierPathPart, matchIdPathPart )
}

func NewReactor( 
    expct mg.TypeReference, dm types.DefinitionGetter ) *CastReactor {

    res := &CastReactor{ 
        stack: stack.NewStack(), 
        dm: dm,
        passFieldsByQn: mg.NewQnameMap(),
        unionMatchFuncs: mg.NewQnameMap(),
    }
    res.pushType( expct )
    res.castBuiltinTypes()
    return res
}

func ( cr *CastReactor ) passFieldsForQn( 
    qn *mg.QualifiedTypeName ) *mg.IdentifierMap {

    if v, ok := cr.passFieldsByQn.GetOk( qn ); ok {
        return v.( *mg.IdentifierMap )
    }
    return nil
}

func ( cr *CastReactor ) AddPassthroughField( 
    qn *mg.QualifiedTypeName, fld *mg.Identifier ) {

    pf := cr.passFieldsForQn( qn )
    if pf == nil {
        pf = mg.NewIdentifierMap()
        cr.passFieldsByQn.Put( qn, pf )
    }
    pf.Put( fld, true )
}

func ( cr *CastReactor ) SetUnionDefinitionMatcher(
    qn *mg.QualifiedTypeName, mf types.UnionMatchFunction ) {

    cr.unionMatchFuncs.Put( qn, mf )
}

func ( cr *CastReactor ) InitializePipeline( pip *pipeline.Pipeline ) {
    mgRct.EnsureStructuralReactor( pip )
    if ! cr.SkipPathSetter { mgRct.EnsurePathSettingProcessor( pip ) }
}

func ( cr *CastReactor ) processPassthrough(
    ev mgRct.Event, next mgRct.EventProcessor ) error {

    if err := next.ProcessEvent( ev ); err != nil { return err }
    if err := cr.passthroughTracker.ProcessEvent( ev ); err != nil { 
        return err 
    }
    if cr.passthroughTracker.Depth() == 0 { cr.passthroughTracker = nil }
    return nil
}

type fieldCast struct {
    ft fieldTyper
    await *mg.IdentifierMap
    passFields *mg.IdentifierMap
}

func ( fc *fieldCast ) isPassthroughField( fld *mg.Identifier ) bool {
    if m := fc.passFields; m != nil {
        if _, pass := m.GetOk( fld ); pass { return true }
    }
    return false
}

func ( fc *fieldCast ) removeOptFields() {
    done := make( []*mg.Identifier, 0, fc.await.Len() )
    fc.await.EachPair( func( _ *mg.Identifier, val interface{} ) {
        fd := val.( *types.FieldDefinition)
        if _, ok := fd.Type.( *mg.NullableTypeReference ); ok {
            done = append( done, fd.Name )
        }
    })
    for _, fld := range done { fc.await.Delete( fld ) }
}

func feedDefault( 
    fld *mg.Identifier, 
    defl mg.Value, 
    p objpath.PathNode,
    next mgRct.EventProcessor ) error {

    fldPath := objpath.Descend( p, fld )
    fs := mgRct.NewFieldStartEvent( fld )
    fs.SetPath( fldPath )
    if err := next.ProcessEvent( fs ); err != nil { return err }
    ps := mgRct.NewPathSettingProcessorPath( fldPath )
    ps.SkipStructureCheck = true
    pip := mgRct.InitReactorPipeline( ps, next )
    return mgRct.VisitValue( defl, pip )
}

func processDefaults(
    fc *fieldCast,
    p objpath.PathNode, 
    next mgRct.EventProcessor ) error {

    vis := func( fld *mg.Identifier, val interface{} ) error {
        fd := val.( *types.FieldDefinition)
        if defl := fd.GetDefault(); defl != nil { 
            if err := feedDefault( fld, defl, p, next ); err != nil { 
                return err 
            }
            fc.await.Delete( fld )
        }
        return nil
    }
    return fc.await.EachPairError( vis )
}

func createMissingFieldsError( p objpath.PathNode, fc *fieldCast ) error {
    flds := make( []*mg.Identifier, 0, fc.await.Len() )
    fc.await.EachPair( func( fld *mg.Identifier, _ interface{} ) {
        flds = append( flds, fld )
    })
    return mg.NewMissingFieldsError( p, flds )
}

type listCast struct {
    sawValues bool
    lt *mg.ListTypeReference
    startPath objpath.PathNode
}

type atomicCastCall struct {
    ve *mgRct.ValueEvent
    at *mg.AtomicTypeReference
    callTyp mg.TypeReference
    cr *CastReactor
}

func ( c atomicCastCall ) val() mg.Value { return c.ve.Val }

func ( c atomicCastCall ) path() objpath.PathNode {
    if p := c.ve.GetPath(); p != nil { return p }
    return nil
}

func ( c atomicCastCall ) newTypeInputErrorValue() error {
    valTyp := mgRct.TypeOfEvent( c.ve )
    return c.cr.newTypeInputError( c.callTyp, valTyp, c.ve.GetPath() )
}

func strToBool( s mg.String, path objpath.PathNode ) ( mg.Value, error ) {
    switch lc := strings.ToLower( string( s ) ); lc { 
    case "true": return mg.Boolean( true ), nil
    case "false": return mg.Boolean( false ), nil
    }
    errTmpl :="Invalid boolean value: %s"
    errStr := mg.QuoteValue( s )
    return nil, mg.NewInputErrorf( path, errTmpl, errStr )
}

func ( c atomicCastCall ) castBoolean() ( mg.Value, error ) {
    switch v := c.val().( type ) {
    case mg.Boolean: return v, nil
    case mg.String: return strToBool( v, c.path() )
    }
    return nil, c.newTypeInputErrorValue()
}

func ( c atomicCastCall ) castBuffer() ( mg.Value, error ) {
    switch v := c.val().( type ) {
    case mg.Buffer: return v, nil
    case mg.String: 
        buf, err := base64.StdEncoding.DecodeString( string( v ) )
        if err == nil { return mg.Buffer( buf ), nil }
        msg := "Invalid base64 string: %s"
        return nil, mg.NewInputErrorf( c.path(), msg, err.Error() )
    }
    return nil, c.newTypeInputErrorValue()
}

func ( c atomicCastCall ) castString() ( mg.Value, error ) {
    switch v := c.val().( type ) {
    case mg.String: return v, nil
    case mg.Boolean, mg.Int32, mg.Int64, mg.Uint32, mg.Uint64, mg.Float32, 
         mg.Float64:
        return mg.String( v.( fmt.Stringer ).String() ), nil
    case mg.Timestamp: return mg.String( v.Rfc3339Nano() ), nil
    case mg.Buffer:
        b64 := base64.StdEncoding.EncodeToString( []byte( v ) ) 
        return mg.String( b64 ), nil
    case *mg.Enum: return mg.String( v.Value.ExternalForm() ), nil
    }
    return nil, c.newTypeInputErrorValue()
}

func isDecimalNumString( s mg.String ) bool {
    return strings.IndexAny( string( s ), "eE." ) >= 0
}

func ( c atomicCastCall ) parseNumberForCast( 
    s mg.String ) ( mg.Value, error ) {

    numTyp := c.at.Name()
    asFloat := mg.IsIntegerTypeName( numTyp ) && isDecimalNumString( s )
    parseTyp := numTyp
    if asFloat { parseTyp = mg.QnameFloat64 }
    val, err := mg.ParseNumber( string( s ), parseTyp )
    if ne, ok := err.( *mg.NumberFormatError ); ok {
        err = mg.NewInputError( c.path(), ne.Error() )
    }
    if err != nil || ( ! asFloat ) { return val, err }
    f64 := float64( val.( mg.Float64 ) )
    switch {
    case numTyp.Equals( mg.QnameInt32 ): val = mg.Int32( int32( f64 ) )
    case numTyp.Equals( mg.QnameUint32 ): val = mg.Uint32( uint32( f64 ) )
    case numTyp.Equals( mg.QnameInt64 ): val = mg.Int64( int64( f64 ) )
    case numTyp.Equals( mg.QnameUint64 ): val = mg.Uint64( uint64( f64 ) )
    }
    return val, nil
}

func ( c atomicCastCall ) castInt32() ( mg.Value, error ) {
    switch v := c.val().( type ) {
    case mg.Int32: return v, nil
    case mg.Int64: return mg.Int32( v ), nil
    case mg.Uint32: return mg.Int32( int32( v ) ), nil
    case mg.Uint64: return mg.Int32( int32( v ) ), nil
    case mg.Float32: return mg.Int32( int32( v ) ), nil
    case mg.Float64: return mg.Int32( int32( v ) ), nil
    case mg.String: return c.parseNumberForCast( v )
    }
    return nil, c.newTypeInputErrorValue()
}

func ( c atomicCastCall ) castInt64() ( mg.Value, error ) {
    switch v := c.val().( type ) {
    case mg.Int32: return mg.Int64( v ), nil
    case mg.Int64: return v, nil
    case mg.Uint32: return mg.Int64( int64( v ) ), nil
    case mg.Uint64: return mg.Int64( int64( v ) ), nil
    case mg.Float32: return mg.Int64( int64( v ) ), nil
    case mg.Float64: return mg.Int64( int64( v ) ), nil
    case mg.String: return c.parseNumberForCast( v )
    }
    return nil, c.newTypeInputErrorValue()
}

func ( c atomicCastCall ) castCheckNeg( 
    isNeg bool, in, ifOk mg.Value ) ( mg.Value, error ) {

    if isNeg { 
        return nil, mg.NewInputErrorf( c.path(), "value out of range: %s", in )
    }
    return ifOk, nil
}

func ( c atomicCastCall ) castUint32() ( mg.Value, error ) {
    switch v := c.val().( type ) {
    case mg.Int32: return c.castCheckNeg( v < 0, v, mg.Uint32( uint32( v ) ) )
    case mg.Uint32: return v, nil
    case mg.Int64: return c.castCheckNeg( v < 0, v, mg.Uint32( uint32( v ) ) )
    case mg.Uint64: return mg.Uint32( uint32( v ) ), nil
    case mg.Float32: return c.castCheckNeg( v < 0, v, mg.Uint32( uint32( v ) ) )
    case mg.Float64: return c.castCheckNeg( v < 0, v, mg.Uint32( uint32( v ) ) )
    case mg.String: return c.parseNumberForCast( v )
    }
    return nil, c.newTypeInputErrorValue()
}

func ( c atomicCastCall ) castUint64() ( mg.Value, error ) {
    switch v := c.val().( type ) {
    case mg.Int32: return c.castCheckNeg( v < 0, v, mg.Uint64( uint64( v ) ) )
    case mg.Uint32: return mg.Uint64( uint64( v ) ), nil
    case mg.Int64: return c.castCheckNeg( v < 0, v, mg.Uint64( uint64( v ) ) )
    case mg.Uint64: return v, nil
    case mg.Float32: return c.castCheckNeg( v < 0, v, mg.Uint64( uint64( v ) ) )
    case mg.Float64: return c.castCheckNeg( v < 0, v, mg.Uint64( uint64( v ) ) )
    case mg.String: return c.parseNumberForCast( v )
    }
    return nil, c.newTypeInputErrorValue()
}

func ( c atomicCastCall ) castFloat32() ( mg.Value, error ) {
    switch v := c.val().( type ) {
    case mg.Int32: return mg.Float32( float32( v ) ), nil
    case mg.Int64: return mg.Float32( float32( v ) ), nil
    case mg.Uint32: return mg.Float32( float32( v ) ), nil
    case mg.Uint64: return mg.Float32( float32( v ) ), nil
    case mg.Float32: return v, nil
    case mg.Float64: return mg.Float32( float32( v ) ), nil
    case mg.String: return c.parseNumberForCast( v )
    }
    return nil, c.newTypeInputErrorValue()
}

func ( c atomicCastCall ) castFloat64() ( mg.Value, error ) {
    switch v := c.val().( type ) {
    case mg.Int32: return mg.Float64( float64( v ) ), nil
    case mg.Int64: return mg.Float64( float64( v ) ), nil
    case mg.Uint32: return mg.Float64( float64( v ) ), nil
    case mg.Uint64: return mg.Float64( float64( v ) ), nil
    case mg.Float32: return mg.Float64( float64( v ) ), nil
    case mg.Float64: return v, nil
    case mg.String: return c.parseNumberForCast( v )
    }
    return nil, c.newTypeInputErrorValue()
}

func ( c atomicCastCall ) castTimestamp() ( mg.Value, error ) {
    switch v := c.val().( type ) {
    case mg.Timestamp: return v, nil
    case mg.String:
        tm, err := parser.ParseTimestamp( string( v ) )
        if err == nil { return tm, nil }
        msg := "Invalid timestamp: %s"
        return nil, mg.NewInputErrorf( c.path(), msg, err.Error() )
    }
    return nil, c.newTypeInputErrorValue()
}

func ( c atomicCastCall ) castSymbolMap() ( mg.Value, error ) {
    switch v := c.val().( type ) {
    case *mg.SymbolMap: return v, nil
    }
    return nil, c.newTypeInputErrorValue()
}

// switch compares based on qname not at itself since we may be dealing with
// restriction types, meaning that if at is mingle:core@v1/String~"a", it is a
// string (has qname mingle:core@v1/String) but will not equal mg.TypeString
// itself
func ( c atomicCastCall ) castAtomicUnrestricted() ( mg.Value, error ) {
    if _, ok := c.val().( *mg.Null ); ok {
        if c.at.Equals( mg.TypeNull ) { return c.val(), nil }
        return nil, newNullInputError( c.ve.GetPath() )
    }
    switch nm := c.at.Name(); {
    case nm.Equals( mg.QnameValue ): return c.val(), nil
    case nm.Equals( mg.QnameBoolean ): return c.castBoolean()
    case nm.Equals( mg.QnameBuffer ): return c.castBuffer()
    case nm.Equals( mg.QnameString ): return c.castString()
    case nm.Equals( mg.QnameInt32 ): return c.castInt32()
    case nm.Equals( mg.QnameInt64 ): return c.castInt64()
    case nm.Equals( mg.QnameUint32 ): return c.castUint32()
    case nm.Equals( mg.QnameUint64 ): return c.castUint64()
    case nm.Equals( mg.QnameFloat32 ): return c.castFloat32()
    case nm.Equals( mg.QnameFloat64 ): return c.castFloat64()
    case nm.Equals( mg.QnameTimestamp ): return c.castTimestamp()
    case nm.Equals( mg.QnameSymbolMap ): return c.castSymbolMap()
    }
    return nil, c.newTypeInputErrorValue()
}

func ( c atomicCastCall ) checkRestriction( val mg.Value ) error {
    if c.at.Restriction().AcceptsValue( val ) { return nil }
    return mg.NewInputErrorf( 
        c.ve.GetPath(), "Value %s does not satisfy restriction %s",
        mg.QuoteValue( val ), c.at.Restriction().ExternalForm() )
}

func ( c atomicCastCall ) call() ( mg.Value, error ) {
    val, err := c.castAtomicUnrestricted()
    if err == nil && c.at.Restriction() != nil { 
        err = c.checkRestriction( val )
    }
    return val, err
}

func ( cr *CastReactor ) errStackUnrecognized() error {
    return libErrorf( "unrecognized stack element: %T", cr.stack.Peek() )
}

func ( cr *CastReactor ) unionTypeDefForAtomicType( 
    at *mg.AtomicTypeReference ) *types.UnionTypeDefinition {

    if td, ok := cr.dm.GetDefinition( at.Name() ); ok {
        if utd, ok := td.( *types.UnionDefinition ); ok { return utd.Union }
    }
    return nil
}

func ( cr *CastReactor ) getStructDef( 
    nm *mg.QualifiedTypeName ) *types.StructDefinition {

    if def, ok := cr.dm.GetDefinition( nm ); ok {
        if sd, ok := def.( *types.StructDefinition ); ok { return sd }
    }
    return nil
}

// Only handles positive mismatches in which a declared type carries a type that
// corresponds to a known definition, but for which that definition makes no
// sense. We let the case when no such definition exists at all be handled
// elsewhere.
func ( cr *CastReactor ) checkWellFormed(
    ev mgRct.Event,
    typ *mg.QualifiedTypeName,
    errDesc string,
    defCheck func( def types.Definition ) bool ) error {

    if def, ok := cr.dm.GetDefinition( typ ); ok {
        if defCheck( def ) { return nil }
        tmpl := "not %s type: %s"
        return mg.NewInputErrorf( ev.GetPath(), tmpl, errDesc, typ )
    }
    return nil
}

func ( cr *CastReactor ) checkValueWellFormed( ve *mgRct.ValueEvent ) error {
    en, ok := ve.Val.( *mg.Enum )
    if ! ok { return nil }
    chk := func( def types.Definition ) bool {
        _, ok := def.( *types.EnumDefinition ); 
        return ok
    }
    return cr.checkWellFormed( ve, en.Type, "an enum", chk )
}

func ( cr *CastReactor ) constructorTypeForType(
    typ mg.TypeReference, sd *types.StructDefinition ) mg.TypeReference {

    if sd.Constructors == nil { return nil }
    res, ok := sd.Constructors.MatchType( typ )
    if ok { return res }
    return nil
}

func ( cr *CastReactor ) castStructConstructor(
    v mg.Value,
    sd *types.StructDefinition,
    path objpath.PathNode ) ( mg.Value, error, bool ) {

    if cr.constructorTypeForType( mg.TypeOf( v ), sd ) != nil { 
        return v, nil, true 
    }
    return nil, nil, false
}

func ( cr *CastReactor ) completeCastEnum(
    id *mg.Identifier, 
    ed *types.EnumDefinition, 
    path objpath.PathNode ) ( *mg.Enum, error ) {

    if res := ed.GetValue( id ); res != nil { return res, nil }
    tmpl := "illegal value for enum %s: %s"
    return nil, mg.NewInputErrorf( path, tmpl, ed.GetName(), id )
}

func ( cr *CastReactor ) castEnumFromString( 
    s string, 
    ed *types.EnumDefinition, 
    path objpath.PathNode ) ( *mg.Enum, error ) {

    id, err := parser.ParseIdentifier( s )
    if err != nil {
        tmpl := "invalid enum value %q: %s"
        return nil, mg.NewInputErrorf( path, tmpl, s, err )
    }
    return cr.completeCastEnum( id, ed, path )
}

func ( cr *CastReactor ) castEnum( 
    val mg.Value, 
    ed *types.EnumDefinition, 
    path objpath.PathNode ) ( *mg.Enum, error ) {

    switch v := val.( type ) {
    case mg.String: return cr.castEnumFromString( string( v ), ed, path )
    case *mg.Enum: 
        if v.Type.Equals( ed.GetName() ) {
            return cr.completeCastEnum( v.Value, ed, path )
        }
    }
    t := ed.GetName().AsAtomicType()
    return nil, cr.newTypeInputErrorValue( t, val, path )
}

func ( cr *CastReactor ) castAtomicForDefinition(
    v mg.Value,
    at *mg.AtomicTypeReference,
    path objpath.PathNode ) ( mg.Value, error, bool ) {

    if def, ok := cr.dm.GetDefinition( at.Name() ); ok {
        switch td := def.( type ) {
        case *types.EnumDefinition:
            res, err := cr.castEnum( v, td, path )
            return res, err, true
        case *types.StructDefinition: 
            return cr.castStructConstructor( v, td, path )
        } 
    }
    return nil, nil, false
}

func ( cr *CastReactor ) valueEventForAtomicCast( 
    ve *mgRct.ValueEvent, 
    at *mg.AtomicTypeReference, 
    callTyp mg.TypeReference ) ( *mgRct.ValueEvent, error ) {

    mv, err, ok := cr.castAtomicForDefinition( ve.Val, at, ve.GetPath() )
    if ! ok { 
        log.Printf( "at: %s, callTyp: %s", at, callTyp )
        mv, err = atomicCastCall{ ve, at, callTyp, cr }.call()
//        mv, err = castAtomicWithCallType( ve.Val, at, callTyp, ve.GetPath() ) 
    }
    if err != nil { return nil, err }
    res := mgRct.CopyEvent( ve, true ).( *mgRct.ValueEvent )
    res.Val = mv
    return res, nil
}

func ( cr *CastReactor ) processAtomicValue(
    ve *mgRct.ValueEvent,
    at *mg.AtomicTypeReference,
    callTyp mg.TypeReference,
    next mgRct.EventProcessor ) error {

    ve2, err := cr.valueEventForAtomicCast( ve, at, callTyp )
    if err != nil { return err }
    if err = next.ProcessEvent( ve2 ); err != nil { return err }
    return nil
}

func ( cr *CastReactor ) processNullableValue(
    ve *mgRct.ValueEvent,
    nt *mg.NullableTypeReference,
    callTyp mg.TypeReference,
    next mgRct.EventProcessor ) error {

    if _, ok := ve.Val.( *mg.Null ); ok { return next.ProcessEvent( ve ) }
    return cr.processValueWithType( ve, nt.Type, callTyp, next )
}

func ( cr *CastReactor ) processValueForListType(
    ve *mgRct.ValueEvent,
    typ *mg.ListTypeReference,
    callTyp mg.TypeReference,
    next mgRct.EventProcessor ) error {

    if _, ok := ve.Val.( *mg.Null ); ok {
        return newNullInputError( ve.GetPath() )
    }
    return cr.newTypeInputErrorValue( callTyp, ve.Val, ve.GetPath() )
}

func ( cr *CastReactor ) matchUnionDefType(
    ev mgRct.Event, ud *types.UnionDefinition ) ( mg.TypeReference, bool ) {

    typ := mgRct.TypeOfEvent( ev )
    ut := ud.Union
    if mf, ok := cr.unionMatchFuncs.GetOk( ud.Name ); ok {
        umf := mf.( types.UnionMatchFunction )
        return umf( types.UnionMatchInput{ typ, ut, cr.dm } )
    }
    return ut.MatchType( typ )
}

func ( cr *CastReactor ) getUnionApplication(
    ev mgRct.Event,
    at *mg.AtomicTypeReference,
    next mgRct.EventProcessor ) func() error {

    if def, ok := cr.dm.GetDefinition( at.Name() ); ok {
        if ud, ok := def.( *types.UnionDefinition ); ok {
            if mtch, ok := cr.matchUnionDefType( ev, ud ); ok {
                return func() error { 
                    cr.pushType( mtch )
                    return cr.ProcessEvent( ev, next ) 
                }
            }
        }
    }
    return nil
}

func ( cr *CastReactor ) processValueWithType(
    ve *mgRct.ValueEvent,
    typ mg.TypeReference,
    callTyp mg.TypeReference,
    next mgRct.EventProcessor ) error {

    switch v := typ.( type ) {
    case *mg.AtomicTypeReference: 
        if f := cr.getUnionApplication( ve, v, next ); f != nil { return f() }
        return cr.processAtomicValue( ve, v, callTyp, next )
    case *mg.PointerTypeReference:
        return cr.processValueWithType( ve, v.Type, callTyp, next )
    case *mg.NullableTypeReference:
        return cr.processNullableValue( ve, v, callTyp, next )
    case *mg.ListTypeReference:
        return cr.processValueForListType( ve, v, callTyp, next )
    }
    panic( libErrorf( "unhandled type: %T", typ ) )
}

func ( cr *CastReactor ) processValue( 
    ve *mgRct.ValueEvent, next mgRct.EventProcessor ) error {

    if err := cr.checkValueWellFormed( ve ); err != nil { return err }
    switch v := cr.stack.Peek().( type ) {
    case mg.TypeReference: 
        cr.stack.Pop()
        return cr.processValueWithType( ve, v, v, next )
    case *listCast:
        v.sawValues = true
        typ := v.lt.ElementType
        return cr.processValueWithType( ve, typ, typ, next )
    }
    panic( cr.errStackUnrecognized() )
}

func ( cr *CastReactor ) implMapStart(
    ev mgRct.Event, 
    ft fieldTyper, 
    fs *types.FieldSet,
    passFields *mg.IdentifierMap,
    next mgRct.EventProcessor ) error {

    fc := &fieldCast{ ft: ft, passFields: passFields }
    if fs != nil {
        fc.await = mg.NewIdentifierMap()
        fs.EachDefinition( func( fd *types.FieldDefinition) {
            fc.await.Put( fd.Name, fd )
        })
    }
    cr.stack.Push( fc )
    return next.ProcessEvent( ev )
}

type fieldSetTyper struct { 
    flds *types.FieldSet 
    dm types.DefinitionGetter
    ignoreUnrecognized bool
}

func ( ft *fieldSetTyper ) fieldTypeFor(
    fld *mg.Identifier, path objpath.PathNode ) ( mg.TypeReference, error ) {

    if fd := ft.flds.Get( fld ); fd != nil { return fd.Type, nil }
    if ft.ignoreUnrecognized { return mg.TypeValue, nil }
    return nil, mg.NewUnrecognizedFieldError( path, fld )
}

func ( cr *CastReactor ) fieldSetTyperForStruct(
    def *types.StructDefinition, 
    path objpath.PathNode ) ( *fieldSetTyper, error ) {

    return &fieldSetTyper{ flds: def.Fields, dm: cr.dm }, nil
}

func ( cr *CastReactor ) fieldSetTyperForSchema( 
    sd *types.SchemaDefinition ) *fieldSetTyper {

    return &fieldSetTyper{ 
        flds: sd.Fields, 
        dm: cr.dm, 
        ignoreUnrecognized: true,
    }
}

func ( cr *CastReactor ) fieldSetTyperFor(
    qn *mg.QualifiedTypeName, 
    path objpath.PathNode ) ( *fieldSetTyper, error ) {

    if def, ok := cr.dm.GetDefinition( qn ); ok {
        switch v := def.( type ) {
        case *types.StructDefinition: 
            return cr.fieldSetTyperForStruct( v, path )
        case *types.SchemaDefinition: return cr.fieldSetTyperForSchema( v ), nil
        default: return nil, notAFieldSetTypeError( path, qn )
        }
    }
    tmpl := "no field type info for type %s"
    return nil, mg.NewInputErrorf( path, tmpl, qn )
}

func ( cr *CastReactor ) completeStartStruct(
    ss *mgRct.StructStartEvent, next mgRct.EventProcessor ) error {

    ft, err := cr.fieldSetTyperFor( ss.Type, ss.GetPath() )
    if err != nil { return err }
    var ev mgRct.Event = ss
    fs, err := fieldSetForTypeInDefMap( ss.Type, cr.dm, ss.GetPath() )
    if err != nil { return err }
    if def, ok := cr.dm.GetDefinition( ss.Type ); ok {
        if _, ok := def.( *types.SchemaDefinition ); ok { 
            ev = asMapStartEvent( ss ) 
        }
    } 
    pf := cr.passFieldsForQn( ss.Type )
    return cr.implMapStart( ev, ft, fs, pf, next )
}

func ( cr *CastReactor ) inferStructForQname( qn *mg.QualifiedTypeName ) bool {
    if def, ok := cr.dm.GetDefinition( qn ); ok {
        if _, ok = def.( *types.StructDefinition ); ok { return true }
        if _, ok = def.( *types.SchemaDefinition ); ok { return true }
    }
    return false
}

func ( cr *CastReactor ) inferStructForMap(
    me *mgRct.MapStartEvent,
    at *mg.AtomicTypeReference,
    next mgRct.EventProcessor ) ( error, bool ) {

    if ! cr.inferStructForQname( at.Name() ) { return nil, false }

    ev := mgRct.NewStructStartEvent( at.Name() )
    ev.SetPath( me.GetPath() )

    return cr.completeStartStruct( ev, next ), true
}

func ( cr *CastReactor ) processMapStartWithAtomicType(
    me *mgRct.MapStartEvent,
    at *mg.AtomicTypeReference,
    callTyp mg.TypeReference,
    next mgRct.EventProcessor ) error {

    if at.Equals( mg.TypeSymbolMap ) || at.Equals( mg.TypeValue ) {
        var ft fieldTyper = valueFieldTyper( 1 )
        var fs *types.FieldSet
        if cr.FieldSetFactory != nil {
            var err error
            fs, err = cr.FieldSetFactory.GetFieldSet( me.GetPath() )
            if err != nil { return err }
            if fs != nil { ft = &fieldSetTyper{ flds: fs, dm: cr.dm } }
        }
        return cr.implMapStart( me, ft, fs, nil, next )
    }

    if err, ok := cr.inferStructForMap( me, at, next ); ok { return err }

    return cr.newTypeInputError( callTyp, mg.TypeSymbolMap, me.GetPath() )
}

func ( cr *CastReactor ) processMapStartWithType(
    me *mgRct.MapStartEvent, 
    typ mg.TypeReference,
    callTyp mg.TypeReference,
    next mgRct.EventProcessor ) error {

    switch v := typ.( type ) {
    case *mg.AtomicTypeReference:
        if f := cr.getUnionApplication( me, v, next ); f != nil { return f() }
        return cr.processMapStartWithAtomicType( me, v, callTyp, next )
    case *mg.PointerTypeReference:
        return cr.processMapStartWithType( me, v.Type, callTyp, next )
    case *mg.NullableTypeReference:
        return cr.processMapStartWithType( me, v.Type, callTyp, next )
    }
    return cr.newTypeInputError( callTyp, typ, me.GetPath() )
}

func ( cr *CastReactor ) processMapStart(
    me *mgRct.MapStartEvent, next mgRct.EventProcessor ) error {
    
    switch v := cr.stack.Peek().( type ) {
    case mg.TypeReference: 
        cr.stack.Pop()
        return cr.processMapStartWithType( me, v, v, next )
    case *listCast:
        v.sawValues = true
        typ := v.lt.ElementType
        return cr.processMapStartWithType( me, typ, typ, next )
    }
    panic( cr.errStackUnrecognized() )
}

func ( cr *CastReactor ) processFieldStart(
    fs *mgRct.FieldStartEvent, next mgRct.EventProcessor ) error {

    fc := cr.stack.Peek().( *fieldCast )
    if fc.await != nil { fc.await.Delete( fs.Field ) }
    
    if fc.isPassthroughField( fs.Field ) {
        cr.passthroughTracker = mgRct.NewDepthTracker()
    } else {
        typ, err := fc.ft.fieldTypeFor( fs.Field, fs.GetPath().Parent() )
        if err != nil { return err }
        cr.pushType( typ )
    }

    return next.ProcessEvent( fs )
}

func ( cr *CastReactor ) processListEnd() error {
    lc := cr.stack.Pop().( *listCast )
    if ! ( lc.sawValues || lc.lt.AllowsEmpty ) {
        return mg.NewInputError( lc.startPath, "empty list" )
    }
    return nil
}

func ( cr *CastReactor ) processFieldsEnd( 
    ee *mgRct.EndEvent, next mgRct.EventProcessor ) error {

    fc := cr.stack.Pop().( *fieldCast )
    if fc.await == nil { return nil }
    p := ee.GetPath()
    if err := processDefaults( fc, p, next ); err != nil { return err }
    fc.removeOptFields()
    if fc.await.Len() > 0 { return createMissingFieldsError( p, fc ) }
    return nil
}

func ( cr *CastReactor ) processEnd(
    ee *mgRct.EndEvent, next mgRct.EventProcessor ) error {

    switch cr.stack.Peek().( type ) {
    case *listCast: if err := cr.processListEnd(); err != nil { return err }
    case *fieldCast: 
        if err := cr.processFieldsEnd( ee, next ); err != nil { return err }
    }

    if err := next.ProcessEvent( ee ); err != nil { return err }
    return nil
}

func ( cr *CastReactor ) checkStructWellFormed(
    ss *mgRct.StructStartEvent ) error {

    return cr.checkWellFormed( ss, ss.Type, "a struct",
        func ( def types.Definition ) bool {
            _, ok := def.( *types.StructDefinition )
            return ok
        },
    )
}

func ( cr *CastReactor ) allowStructStartForType( 
    ss *mgRct.StructStartEvent, expct *mg.QualifiedTypeName ) bool {

    if _, ok := cr.dm.GetDefinition( ss.Type ); ! ok { return false }
    if sd := cr.getStructDef( expct ); sd != nil {
        if cr.constructorTypeForType( ss.Type.AsAtomicType(), sd ) != nil {
            return true
        }
    }
    return canAssignType( expct, ss.Type, cr.dm )
}

func ( cr *CastReactor ) processStructStartWithAtomicType(
    ss *mgRct.StructStartEvent,
    at *mg.AtomicTypeReference,
    callTyp mg.TypeReference,
    next mgRct.EventProcessor ) error {

    if at.Equals( mg.TypeSymbolMap ) {
        me := asMapStartEvent( ss )
        return cr.processMapStartWithAtomicType( me, at, callTyp, next )
    }

    if at.Name().Equals( ss.Type ) || at.Equals( mg.TypeValue ) ||
       cr.allowStructStartForType( ss, at.Name() ) {
        return cr.completeStartStruct( ss, next )
    }

    failTyp := mg.NewAtomicTypeReference( ss.Type, nil )
    return cr.newTypeInputError( callTyp, failTyp, ss.GetPath() )
}

func ( cr *CastReactor ) processStructStartWithType(
    ss *mgRct.StructStartEvent,
    typ mg.TypeReference,
    callTyp mg.TypeReference,
    next mgRct.EventProcessor ) error {

    switch v := typ.( type ) {
    case *mg.AtomicTypeReference:
        if f := cr.getUnionApplication( ss, v, next ); f != nil { return f() }
        return cr.processStructStartWithAtomicType( ss, v, callTyp, next )
    case *mg.PointerTypeReference:
        return cr.processStructStartWithType( ss, v.Type, callTyp, next )
    case *mg.NullableTypeReference:
        return cr.processStructStartWithType( ss, v.Type, callTyp, next )
    }
    return cr.newTypeInputError( typ, callTyp, ss.GetPath() )
}

func ( cr *CastReactor ) processStructStart(
    ss *mgRct.StructStartEvent, next mgRct.EventProcessor ) error {

    if err := cr.checkStructWellFormed( ss ); err != nil { return err }
    switch v := cr.stack.Peek().( type ) {
    case mg.TypeReference:
        cr.stack.Pop()
        return cr.processStructStartWithType( ss, v, v, next )
    case *listCast:
        v.sawValues = true
        typ := v.lt.ElementType
        return cr.processStructStartWithType( ss, typ, typ, next )
    }
    panic( cr.errStackUnrecognized() )
}

func ( cr *CastReactor ) processListStartWithAtomicType(
    le *mgRct.ListStartEvent,
    at *mg.AtomicTypeReference,
    callTyp mg.TypeReference,
    next mgRct.EventProcessor ) error {

    if at.Equals( mg.TypeValue ) {
        return cr.processListStartWithType( 
            le, mg.TypeOpaqueList, callTyp, next )
    }
    if f := cr.getUnionApplication( le, at, next ); f != nil { return f() }
    if sd := cr.getStructDef( at.Name() ); sd != nil {
        if typ := cr.constructorTypeForType( le.Type, sd ); typ != nil {
            lt := typ.( *mg.ListTypeReference )
            return cr.processListStartWithListType( le, lt, callTyp, next )
        }
    }
    return cr.newTypeInputError( callTyp, le.Type, le.GetPath() )
}

func ( cr *CastReactor ) processListStartWithListType(
    le *mgRct.ListStartEvent,
    lt *mg.ListTypeReference,
    callTyp mg.TypeReference,
    next mgRct.EventProcessor ) error {
    
    sp := objpath.CopyOf( le.GetPath() )
    cr.stack.Push( &listCast{ lt: lt, startPath: sp } )
    return next.ProcessEvent( le )
}

func ( cr *CastReactor ) processListStartWithType(
    le *mgRct.ListStartEvent,
    typ mg.TypeReference,
    callTyp mg.TypeReference,
    next mgRct.EventProcessor ) error {

    switch v := typ.( type ) {
    case *mg.AtomicTypeReference:
        return cr.processListStartWithAtomicType( le, v, callTyp, next )
    case *mg.PointerTypeReference:
        return cr.processListStartWithType( le, v.Type, callTyp, next )
    case *mg.ListTypeReference:
        return cr.processListStartWithListType( le, v, callTyp, next )
    case *mg.NullableTypeReference:
        return cr.processListStartWithType( le, v.Type, callTyp, next )
    }
    panic( libErrorf( "unhandled type: %T", typ ) )
}

func ( cr *CastReactor ) processListStart( 
    le *mgRct.ListStartEvent, next mgRct.EventProcessor ) error {

    switch v := cr.stack.Peek().( type ) {
    case mg.TypeReference:
        cr.stack.Pop()
        return cr.processListStartWithType( le, v, v, next )
    case *listCast:
        v.sawValues = true
        return cr.processListStartWithType( le, v.lt.ElementType, v.lt, next )
    }
    panic( cr.errStackUnrecognized() )
}

func ( cr *CastReactor ) ProcessEvent(
    ev mgRct.Event, next mgRct.EventProcessor ) ( err error ) {

    if cr.passthroughTracker != nil { return cr.processPassthrough( ev, next ) }
//    cr.dumpStack( "entering ProcessEvent()" )
//    defer cr.dumpStack( "after ProcessEvent()" )
    switch v := ev.( type ) {
    case *mgRct.ValueEvent: return cr.processValue( v, next )
    case *mgRct.MapStartEvent: return cr.processMapStart( v, next )
    case *mgRct.FieldStartEvent: return cr.processFieldStart( v, next )
    case *mgRct.StructStartEvent: return cr.processStructStart( v, next )
    case *mgRct.ListStartEvent: return cr.processListStart( v, next )
    case *mgRct.EndEvent: return cr.processEnd( v, next )
    }
    panic( libErrorf( "unhandled event: %T", ev ) )
}
