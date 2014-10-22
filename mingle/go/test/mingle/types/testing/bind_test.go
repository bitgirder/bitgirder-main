package testing

import ( 
    "testing"
    "bitgirder/assert"
    "bitgirder/objpath"
    mg "mingle"
    "mingle/types"
    "mingle/types/builtin"
    "mingle/parser"
    "mingle/bind"
    "sort"
    "fmt"
)

func putCoreBoundTestValues( m *mg.IdentifierMap ) {
    m.Put( mkId( "identifier-id1" ), mkId( "id1" ) )
    m.Put( mkId( "identifier-id1-id2" ), mkId( "id1-id2" ) )
    m.Put( 
        mkId( "declared-type-name-name1" ), 
        parser.MustDeclaredTypeName( "Name1" ),
    )
    m.Put( mkId( "namespace-ns1-v1" ), mkNs( "ns1@v1" ) )
    m.Put( mkId( "namespace-ns1-ns2-v1" ), mkNs( "ns1:ns2@v1" ) )
    m.Put( mkId( "qualified-type-name-ns1-v1-name1" ), mkQn( "ns1@v1/Name1" ) )
    tp := mg.MakeTestIdPath
    m.Put( mkId( "identifier-path-f1-f2-i1-f3" ), tp( 1, 2, "1", 3 ) )
    m.Put( 
        mkId( "identifier-path-f1-f2-i1-i2-i3-i4-f3" ),
        tp( 1, 2, "1", "2", "3", "4", 3 ),
    )
    m.Put( mkId( "identifier-path-f1-i2" ), tp( 1, "2" ) )
    m.Put( mkId( "identifier-path-f1-f2-i3-f4" ), tp( 1, 2, "3", 4 ) )
    m.Put( mkId( "identifier-path-f1-i2-f3" ), tp( 1, "2", 3 ) )
    m.Put( mkId( "atomic-type-name-ns1-v1-name1" ), asType( "ns1@v1/Name1" ) )
    m.Put( mkId( "int32-closed-zero-ten-open" ), asType( "Int32~[0,10)" ) )
    m.Put( mkId( "string-a-star" ), asType( `String~"a*"` ) )
    m.Put( mkId( "list-type1-allows-empty" ), asType( "ns1@v1/Name1*" ) )
    m.Put( mkId( "pointer-type1" ), asType( "&ns1@v1/Name1" ) )
    m.Put( mkId( "nullable-type1" ), asType( "&ns1@v1/Name1?" ) )
    m.Put( 
        mkId( "cast-error-loc-f1-i2" ),
        mg.NewCastError( tp( 1, "2" ), "test-message" ),
    )
    m.Put(
        mkId( "unrecognized-field-error-loc-f1-i2" ),
        &mg.UnrecognizedFieldError{
            Location: tp( 1, "2" ),
            Message: "test-message",
            Field: mkId( "bad-field" ),
        },
    )
    mfeInst1 := &mg.MissingFieldsError{
        Message: "test-message",
        Location: tp( 1, "2" ),
    }
    mfeInst1.SetFields(
        []*mg.Identifier{ mkId( "f1" ), mkId( "f2" ), mkId( "f3" ) } )
    m.Put( mkId( "missing-fields-error-f1-f2-f3-loc-f1-i2" ), mfeInst1 )
}

func putTypesBoundTestValues( m *mg.IdentifierMap ) {
    qnNs1V1Name := func( i int ) *mg.QualifiedTypeName {
        return mkQn( fmt.Sprintf( "ns1@v1/Name%d", i ) )
    }
    qnNs1V1Name1 := qnNs1V1Name( 1 )
    atomicQnNs1V1Name := func( i int ) *mg.AtomicTypeReference {
        return mg.NewAtomicTypeReference( qnNs1V1Name( i ), nil )
    }
    atomicQnNs1V1Name1 := atomicQnNs1V1Name( 1 )
    m.Put( mkId( "prim-def1" ), &types.PrimitiveDefinition{ mg.QnameInt32 } )
    m.Put(
        mkId( "field-def0" ),
        &types.FieldDefinition{ Name: mkId( "f1" ), Type: atomicQnNs1V1Name1 },
    )
    m.Put(
        mkId( "field-def1" ),
        &types.FieldDefinition{
            Name: mkId( "f1" ),
            Type: mg.TypeInt32,
            Default: mg.Int32( 1 ),
        },
    )
    m.Put( mkId( "empty-field-set" ), types.NewFieldSet() )
    fieldSet := func( sz int ) *types.FieldSet {
        flds := make( []*types.FieldDefinition, sz )
        for i := 0; i < sz; i++ {
            nm := fmt.Sprintf( "f%d", i )
            flds[ i ] = types.MakeFieldDef( nm, mg.TypeInt32, mg.Int32( i ) )
        }
        return types.MakeFieldSet( flds... )
    }
    m.Put( mkId( "field-set1" ), fieldSet( 3 ) )
    m.Put( 
        mkId( "call-sig1" ),
        types.MakeCallSig(
            []*types.FieldDefinition{
                types.MakeFieldDef( "f0", mg.TypeInt32, mg.Int32( 0 ) ),
                types.MakeFieldDef( "f1", mg.TypeInt32, mg.Int32( 1 ) ),
            },
            atomicQnNs1V1Name1.ExternalForm(),
            nil,
        ),
    )
    callSig2 := func() *types.CallSignature {
        return types.MakeCallSig(
            []*types.FieldDefinition{
                types.MakeFieldDef( "f0", mg.TypeInt32, mg.Int32( 0 ) ),
                types.MakeFieldDef( "f1", mg.TypeInt32, mg.Int32( 1 ) ),
            },
            atomicQnNs1V1Name1.ExternalForm(),
            []string{ "ns1@v1/Name0", "ns1@v1/Name1" },
        )
    } 
    m.Put( mkId( "call-sig2" ), callSig2() )
    unionTypeDef := func( sz int ) *types.UnionTypeDefinition {
        typs := make( []mg.TypeReference, sz )
        for i := 0; i < sz; i++ { typs[ i ] = atomicQnNs1V1Name( i ) }
        return types.MustUnionTypeDefinitionTypes( typs... )
    }
    m.Put( mkId( "union-type-def1" ), unionTypeDef( 2 ) )
    m.Put( 
        mkId( "union-def1" ),
        &types.UnionDefinition{ Name: qnNs1V1Name1, Union: unionTypeDef( 2 ) },
    )
    m.Put(
        mkId( "proto-def1" ),
        &types.PrototypeDefinition{ Name: qnNs1V1Name1, Signature: callSig2() },
    )
    structDef1 := types.NewStructDefinition()
    structDef1.Name = qnNs1V1Name1
    structDef1.Fields = fieldSet( 1 )
    m.Put( mkId( "struct-def1" ), structDef1 )
    structDef2 := types.NewStructDefinition()
    structDef2.Name = qnNs1V1Name1
    structDef2.Fields = fieldSet( 1 )
    structDef2.Constructors = unionTypeDef( 2 )
    m.Put( mkId( "struct-def2" ), structDef2 )
    schemaDefEmpty := types.NewSchemaDefinition()
    schemaDefEmpty.Name = qnNs1V1Name1
    m.Put( mkId( "schema-def-empty-fields" ), schemaDefEmpty )
    schemaDef1 := types.NewSchemaDefinition()
    schemaDef1.Name = qnNs1V1Name1
    schemaDef1.Fields = fieldSet( 2 )
    m.Put( mkId( "schema-def1" ), schemaDef1 )
    m.Put(
        mkId( "aliased-def1" ),
        &types.AliasedTypeDefinition{
            Name: qnNs1V1Name1,
            AliasedType: mg.TypeInt32,
        },
    )
    m.Put( 
        mkId( "enum-def1" ),
        &types.EnumDefinition{
            Name: qnNs1V1Name1,
            Values: []*mg.Identifier{ mkId( "v1" ), mkId( "v2" ) },
        },
    )
    opDef := func( nm string ) *types.OperationDefinition {
        return types.MakeOpDef( nm, callSig2() )
    }
    m.Put( mkId( "op-def1" ), opDef( "op1" ) )
    m.Put(
        mkId( "service-def1" ),
        types.MakeServiceDef( qnNs1V1Name1.ExternalForm(), "" ),
    )
    m.Put(
        mkId( "service-def2" ),
        types.MakeServiceDef( 
            qnNs1V1Name1.ExternalForm(),
            qnNs1V1Name1.ExternalForm(),
            opDef( "op1" ),
            opDef( "op2" ),
        ),
    )
}

func getBoundTestValues() *mg.IdentifierMap {
    res := mg.NewIdentifierMap()
    putCoreBoundTestValues( res )
    putTypesBoundTestValues( res )
    return res
}
 
func makeIdStruct( parts ...string ) *mg.Struct {
    l := mg.NewList( asType( "String+" ).( *mg.ListTypeReference ) )
    for _, part := range parts { l.AddUnsafe( mg.String( part ) ) }
    return parser.MustStruct( mg.QnameIdentifier, "parts", l )
}

func makeRestrictionInput( rx interface{} ) *mg.Struct {
    switch v := rx.( type ) {
    case *mg.RegexRestriction: return makeRestrictionInput( v.ExternalForm() )
    case string:
        return parser.MustStruct( mg.QnameRegexRestriction, "pattern", v ) 
    case *mg.RangeRestrictionBuilder:
        return parser.MustStruct( mg.QnameRangeRestriction,
            "min-closed", v.MinClosed,
            "min", v.Min,
            "max", v.Max,
            "max-closed", v.MaxClosed,
        )
    }
    panic( libErrorf( "unhandled restriction: %T", rx ) )
}

// ignores version
func makeNsStruct( ns *mg.Namespace ) *mg.Struct {
    nsParts := mg.NewList(
        &mg.ListTypeReference{
            AllowsEmpty: false,
            ElementType: mg.NewPointerTypeReference( mg.TypeIdentifier ),
        },
    )
    for _, nsPart := range ns.Parts {
        idPart := makeIdStruct( nsPart.GetPartsUnsafe()... )
        nsParts.AddUnsafe( idPart )
    }
    return parser.MustStruct( mg.QnameNamespace,
        "parts", nsParts,
        "version", makeIdStruct( "v1" ),
    )
}

func makeAtomicRestrictionErrorTestInput( 
    t *mg.AtomicRestrictionErrorTest ) *mg.Struct {

    return parser.MustStruct( mg.QnameAtomicTypeReference,
        "name", parser.MustStruct( mg.QnameQualifiedTypeName,
            "namespace", makeNsStruct( t.Name.Namespace ),
            "name", parser.MustStruct( mg.QnameDeclaredTypeName,
                "name", t.Name.Name.ExternalForm(),
            ),
        ),
        "restriction", makeRestrictionInput( t.Restriction ),
    )
}

func atomicRestrictionErrorTestMessageAtIndex( 
    t *mg.AtomicRestrictionErrorTest, idx int ) string {

    switch idx {
    case 24: 
        return "illegal min value of type mingle:core@v1/Int64 in range of type mingle:core@v1/Int32"
    }
    return t.Error.Error()
}

type bindTestBuilder struct {
    tests []*bind.BindTest
}

func ( b *bindTestBuilder ) addTest( t *bind.BindTest ) {
    t.Domain = bind.DomainDefault
    b.tests = append( b.tests, t )
}

func ( b *bindTestBuilder ) addRtOpts( 
    in, typ interface{}, id string, opts *bind.SerialOptions ) {

    b.addTest(
        &bind.BindTest{
            Mingle: mg.MustValue( in ),
            Type: asType( typ ),
            BoundId: mkId( id ),
            SerialOptions: opts,
        },
    )
}

func ( b *bindTestBuilder ) addRt( in, typ interface{}, id string ) {
    b.addRtOpts( in, typ, id, nil )
}
    
func ( b *bindTestBuilder ) addIn( in, typ interface{}, id string ) {
    b.addTest(
        &bind.BindTest{
            Mingle: mg.MustValue( in ),
            Type: asType( typ ),
            BoundId: mkId( id ),
            Direction: bind.BindTestDirectionIn,
        },
    )
}

func ( b *bindTestBuilder ) addOut( 
    id string, expct interface{}, opts *bind.SerialOptions ) {

    b.addTest(
        &bind.BindTest{
            Mingle: mg.MustValue( expct ),
            BoundId: mkId( id ),
            Direction: bind.BindTestDirectionOut,
            SerialOptions: opts,
        },
    )
}

func ( b *bindTestBuilder ) addInErr( in, typ interface{}, err error ) {
    b.addTest(
        &bind.BindTest{
            Mingle: mg.MustValue( in ),
            Type: asType( typ ),
            Error: err,
            Direction: bind.BindTestDirectionIn,
        },
    )
}

func ( b *bindTestBuilder ) addVcErr( 
    in, typ interface{}, path objpath.PathNode, msg string ) {

    b.addInErr( in, typ, newVcErr( path, msg ) )
}

func ( b *bindTestBuilder ) idBytes( s string ) []byte {
    return mg.IdentifierAsBytes( mkId( s ) )
}

func ( b *bindTestBuilder ) nsBytes( s string ) []byte {
    return mg.NamespaceAsBytes( mkNs( s ) )
}

func ( b *bindTestBuilder ) coreQn( nm string ) *mg.Struct {
    return parser.MustStruct( mg.QnameQualifiedTypeName,
        "namespace", parser.MustStruct( mg.QnameNamespace,
            "parts", mg.MustList( 
                makeIdStruct( "mingle" ), makeIdStruct( "core" ),
            ),
            "version", makeIdStruct( "v1" ),
        ),
        "name", parser.MustStruct( mg.QnameDeclaredTypeName, "name", nm ),
    )
}

func ( b *bindTestBuilder ) p( rootId string ) objpath.PathNode {
    return objpath.RootedAt( mkId( rootId ) )
}

func ( b *bindTestBuilder ) badBytes() []byte {
    return []byte{ 0, 0, 0, 0, 0, 0 }
}

func ( b *bindTestBuilder ) txtOpts() *bind.SerialOptions {
    res := bind.NewSerialOptions()
    res.Format = bind.SerialFormatText
    return res
}

func ( b *bindTestBuilder ) binOpts() *bind.SerialOptions {
    res := bind.NewSerialOptions()
    res.Format = bind.SerialFormatBinary
    return res
}

func ( b *bindTestBuilder ) addIdentifierTests() {
    b.addRt( makeIdStruct( "id1" ), mg.TypeIdentifier, "identifier-id1" )
    b.addRt( 
        makeIdStruct( "id1", "id2" ), 
        mg.TypeIdentifier, 
        "identifier-id1-id2",
    )
    b.addIn( b.idBytes( "id1" ), mg.TypeIdentifier, "identifier-id1" )
    b.addOut( "identifier-id1", b.idBytes( "id1" ), b.binOpts() )
    b.addIn( mkId( "id1" ).ExternalForm(), mg.TypeIdentifier, "identifier-id1" )
    addIdOutString := func( expct string, fmt mg.IdentifierFormat ) {
        opts := bind.NewSerialOptions()
        opts.Format = bind.SerialFormatText
        opts.Identifiers = fmt
        b.addOut( "identifier-id1-id2", expct, opts )
    }
    addIdOutString( "id1-id2", mg.LcHyphenated )
    addIdOutString( "id1_id2", mg.LcUnderscore )
    addIdOutString( "id1Id2", mg.LcCamelCapped )
    b.addVcErr(
        "id$Bad", 
        mg.TypeIdentifier, 
        nil, 
        "[<input>, line 1, col 3]: Invalid id rune: \"$\" (U+0024)",
    )
    b.addVcErr( 
        b.badBytes(),
        mg.TypeIdentifier,
        nil,
        "[offset 0]: Expected type code 0x01 but got 0x00",
    )
    b.addVcErr(
        makeIdStruct( "part1", "BadPart" ),
        mg.TypeIdentifier,
        b.p( "parts" ).StartList().SetIndex( 1 ),
        "Value \"BadPart\" does not satisfy restriction \"^[a-z][a-z0-9]*$\"",
    )
}

func ( b *bindTestBuilder ) declNm( i int ) *mg.Struct {
    nm := fmt.Sprintf( "Name%d", i )
    return parser.MustStruct( mg.QnameDeclaredTypeName, "name", nm )
}

func ( b *bindTestBuilder ) declNm1() *mg.Struct { return b.declNm( 1 ) }

func ( b *bindTestBuilder ) addDeclaredTypeNameTests() {
    b.addRt( b.declNm1(), mg.TypeDeclaredTypeName, "declared-type-name-name1" )
    b.addVcErr(
        parser.MustStruct( mg.QnameDeclaredTypeName, "name", "Bad$Name" ),
        mg.TypeDeclaredTypeName,
        objpath.RootedAt( mkId( "name" ) ),
        "Value \"Bad$Name\" does not satisfy restriction \"^([A-Z][a-z0-9]*)+$\"",
    )
}

func ( b *bindTestBuilder ) ns1V1() *mg.Struct {
    return parser.MustStruct( mg.QnameNamespace,
        "version", makeIdStruct( "v1" ),
        "parts", mg.MustList( makeIdStruct( "ns1" ) ),
    )
}

func ( b *bindTestBuilder ) addNamespaceTests() {
    b.addRt( b.ns1V1(), mg.TypeNamespace, "namespace-ns1-v1" )
    b.addRt(
        parser.MustStruct( mg.QnameNamespace,
            "version", makeIdStruct( "v1" ),
            "parts", mg.MustList( 
                makeIdStruct( "ns1" ), 
                makeIdStruct( "ns2" ),
            ),
        ),
        mg.TypeNamespace,
        "namespace-ns1-ns2-v1",
    )
    b.addIn(
        parser.MustStruct( mg.QnameNamespace,
            "version", "v1",
            "parts", mg.MustList( "ns1", "ns2" ),
        ),
        mg.TypeNamespace,
        "namespace-ns1-ns2-v1",
    )
    b.addIn(
        parser.MustStruct( mg.QnameNamespace,
            "version", b.idBytes( "v1" ),
            "parts", mg.MustList( b.idBytes( "ns1" ), b.idBytes( "ns2" ) ),
        ),
        mg.TypeNamespace,
        "namespace-ns1-ns2-v1",
    )
    b.addIn(
        parser.MustStruct( mg.QnameNamespace,
            "version", makeIdStruct( "v1" ),
            "parts", mg.MustList( "ns1", b.idBytes( "ns2" ) ),
        ),
        mg.TypeNamespace,
        "namespace-ns1-ns2-v1",
    )
    b.addIn( "ns1@v1", mg.TypeNamespace, "namespace-ns1-v1" )
    b.addIn( b.nsBytes( "ns1@v1" ), mg.TypeNamespace, "namespace-ns1-v1" )
    b.addOut( "namespace-ns1-v1", "ns1@v1", b.txtOpts() )
    b.addOut( "namespace-ns1-v1", b.nsBytes( "ns1@v1" ), b.binOpts() )
    b.addVcErr(
        parser.MustStruct( mg.QnameNamespace,
            "version", "bad$ver",
            "parts", mg.MustList( makeIdStruct( "ns1" ) ),
        ),
        mg.TypeNamespace,
        b.p( "version" ),
        "[<input>, line 1, col 4]: Invalid id rune: \"$\" (U+0024)",
    ) 
    b.addVcErr(
        parser.MustStruct( mg.QnameNamespace,
            "version", makeIdStruct( "v1" ),
            "parts", mg.MustList( makeIdStruct( "ns1" ), "bad$Part" ),
        ),
        mg.TypeNamespace,
        b.p( "parts" ).StartList().SetIndex( 1 ),
        "[<input>, line 1, col 4]: Invalid id rune: \"$\" (U+0024)",
    ) 
    b.addVcErr(
        parser.MustStruct( mg.QnameNamespace,
            "version", makeIdStruct( "v1" ),
            "parts", mg.MustList( makeIdStruct( "ns1" ), b.badBytes() ),
        ),
        mg.TypeNamespace,
        b.p( "parts" ).StartList().SetIndex( 1 ),
        "[offset 0]: Expected type code 0x01 but got 0x00",
    )
    b.addVcErr(
        b.badBytes(),
        mg.TypeNamespace,
        nil,
        "[offset 0]: Expected type code 0x02 but got 0x00",
    )
    b.addVcErr(
        "Bad@Bad",
        mg.TypeNamespace,
        nil,
        "[<input>, line 1, col 1]: Illegal start of identifier part: \"B\" (U+0042)",
    )
}

func ( b *bindTestBuilder ) idPathStruct( parts ...interface{} ) *mg.Struct {
    return parser.MustStruct( mg.QnameIdentifierPath,
        "parts", mg.MustList( parts... ),
    )
}

// assumes correct upstream cast tests are passing, only adds error testing for
// things not checked there
func ( b *bindTestBuilder ) addIdentifierPathTests() {
    b.addRt(
        b.idPathStruct(
            makeIdStruct( "f1" ),
            makeIdStruct( "f2" ),
            uint64( 1 ),
            makeIdStruct( "f3" ),
        ),
        mg.TypeIdentifierPath,
        "identifier-path-f1-f2-i1-f3",
    )
    b.addIn(
        b.idPathStruct(
            makeIdStruct( "f1" ),
            "f2",
            int32( 1 ),
            uint32( 2 ),
            int64( 3 ),
            uint64( 4 ),
            b.idBytes( "f3" ),
        ),
        mg.TypeIdentifierPath,
        "identifier-path-f1-f2-i1-i2-i3-i4-f3",
    )
    b.addIn( "f1[ 2 ]", mg.TypeIdentifierPath, "identifier-path-f1-i2" )
    b.addIn(
        "f1.f2[ 3 ].f4",
        mg.TypeIdentifierPath,
        "identifier-path-f1-f2-i3-f4",
    )
    b.addOut( "identifier-path-f1-i2-f3", "f1[ 2 ].f3", b.txtOpts() )
    b.addOut( 
        "identifier-path-f1-i2-f3",
        b.idPathStruct( b.idBytes( "f1" ), uint64( 2 ), b.idBytes( "f3" ) ),
        b.binOpts(),
    )
    b.addVcErr(
        "p1.bad$Id",
        mg.TypeIdentifierPath,
        nil,
        "[<input>, line 1, col 7]: Invalid id rune: \"$\" (U+0024)",
    )
    b.addVcErr(
        b.idPathStruct( "bad$Id" ),
        mg.TypeIdentifierPath,
        b.p( "parts" ).StartList(),
        "[<input>, line 1, col 4]: Invalid id rune: \"$\" (U+0024)",
    )
    b.addVcErr(
        b.idPathStruct( b.badBytes() ),
        mg.TypeIdentifierPath,
        b.p( "parts" ).StartList(),
        "[offset 0]: Expected type code 0x01 but got 0x00",
    )
}

func ( b *bindTestBuilder ) qnNs1V1Name( i int ) *mg.Struct {
    return parser.MustStruct( mg.QnameQualifiedTypeName,
        "namespace", b.ns1V1(),
        "name", b.declNm( i ),
    )
}

func ( b *bindTestBuilder ) qnNs1V1Name1() *mg.Struct {
    return b.qnNs1V1Name( 1 )
}

func ( b *bindTestBuilder ) addQualifiedTypeNameTests() {
    b.addRt( 
        b.qnNs1V1Name1(),
        mg.TypeQualifiedTypeName,
        "qualified-type-name-ns1-v1-name1",
    )
}

func ( b *bindTestBuilder ) addCoreErrorTests() {
    b.addRt( 
        parser.MustStruct( "mingle:core@v1/CastError",
            "message", "test-message",
            "location", b.idPathStruct( makeIdStruct( "f1" ), uint64( 2 ) ),
        ),
        asType( "mingle:core@v1/CastError" ),
        "cast-error-loc-f1-i2",
    )
    b.addIn( 
        parser.MustStruct( "mingle:core@v1/CastError",
            "message", "test-message",
            "location", "f1[ 2 ]",
        ),
        asType( "mingle:core@v1/CastError" ),
        "cast-error-loc-f1-i2",
    )
    b.addOut( 
        "cast-error-loc-f1-i2",
        parser.MustStruct( "mingle:core@v1/CastError",
            "message", "test-message",
            "location", "f1[ 2 ]",
        ),
        b.txtOpts(),
    )
    b.addRt(
        parser.MustStruct( "mingle:core@v1/UnrecognizedFieldError",
            "message", "test-message",
            "location", b.idPathStruct( makeIdStruct( "f1" ), uint64( 2 ) ),
            "field", makeIdStruct( "bad", "field" ),
        ),
        asType( "mingle:core@v1/UnrecognizedFieldError" ),
        "unrecognized-field-error-loc-f1-i2",
    )
    b.addOut(
        "unrecognized-field-error-loc-f1-i2",
        parser.MustStruct( "mingle:core@v1/UnrecognizedFieldError",
            "message", "test-message",
            "location", "f1[ 2 ]",
            "field", "bad-field",
        ),
        b.txtOpts(),
    )
    b.addRt(
        parser.MustStruct( "mingle:core@v1/MissingFieldsError",
            "message", "test-message",
            "location", b.idPathStruct( makeIdStruct( "f1" ), uint64( 2 ) ),
            "fields", mg.MustList( 
                asType( "&mingle:core@v1/Identifier+" ),
                makeIdStruct( "f1" ),
                makeIdStruct( "f2" ),
                makeIdStruct( "f3" ),
            ),
        ),
        asType( "mingle:core@v1/MissingFieldsError" ),
        "missing-fields-error-f1-f2-f3-loc-f1-i2",
    )
    b.addIn(
        parser.MustStruct( "mingle:core@v1/MissingFieldsError",
            "message", "test-message",
            "location", b.idPathStruct( makeIdStruct( "f1" ), uint64( 2 ) ),
            "fields", mg.MustList( 
                asType( "&mingle:core@v1/Identifier+" ),
                // unsorted in the input to ensure that setter sorts
                makeIdStruct( "f3" ),
                makeIdStruct( "f1" ),
                makeIdStruct( "f2" ),
            ),
        ),
        asType( "mingle:core@v1/MissingFieldsError" ),
        "missing-fields-error-f1-f2-f3-loc-f1-i2",
    )
    b.addIn(
        parser.MustStruct( "mingle:core@v1/MissingFieldsError",
            "message", "test-message",
            "location", "f1[ 2 ]",
            "fields", mg.MustList( asType( "String+" ), 
                "f3", "f1", "f2", // unsorted in the input
            ),
        ),
        asType( "mingle:core@v1/MissingFieldsError" ),
        "missing-fields-error-f1-f2-f3-loc-f1-i2",
    )
    b.addOut(
        "missing-fields-error-f1-f2-f3-loc-f1-i2",
        parser.MustStruct( "mingle:core@v1/MissingFieldsError",
            "message", "test-message",
            "location", "f1[ 2 ]",
            "fields", mg.MustList( asType( "String+" ), "f1", "f2", "f3" ),
        ),
        b.txtOpts(),
    )
    b.addOut(
        "missing-fields-error-f1-f2-f3-loc-f1-i2",
        parser.MustStruct( "mingle:core@v1/MissingFieldsError",
            "message", "test-message",
            "location", b.idPathStruct( b.idBytes( "f1" ), uint64( 2 ) ),
            "fields", mg.MustList( asType( "Buffer+" ), 
                b.idBytes( "f1" ), b.idBytes( "f2" ), b.idBytes( "f3" ),
            ),
        ),
        b.binOpts(),
    )
}

func ( b *bindTestBuilder ) atomicQnNs1V1Name( i int ) *mg.Struct {
    return parser.MustStruct( mg.QnameAtomicTypeReference, 
        "name", b.qnNs1V1Name( i ),
    )
}

func ( b *bindTestBuilder ) atomicQnNs1V1Name1() *mg.Struct {
    return b.atomicQnNs1V1Name( 1 )
}

func ( b *bindTestBuilder ) addAtomicTypeReferenceTests() {
    b.addRt(
        b.atomicQnNs1V1Name1(),
        mg.TypeAtomicTypeReference,
        "atomic-type-name-ns1-v1-name1",
    )
    int32Qn := b.coreQn( "Int32" )
    stringQn := b.coreQn( "String" )
    mkRng := func( lc bool, lv mg.Value, rv mg.Value, rc bool ) *mg.Struct {
        pairs := []interface{}{}
        pairs = append( pairs, "min-closed", lc, "max-closed", rc )
        if lv != nil { pairs = append( pairs, "min", lv ) }
        if rv != nil { pairs = append( pairs, "max", rv ) }
        return parser.MustStruct( mg.QnameRangeRestriction, pairs... )
    }
    b.addRt(
        parser.MustStruct( mg.QnameAtomicTypeReference,
            "name", int32Qn,
            "restriction", mkRng( true, mg.Int32( 0 ), mg.Int32( 10 ), false ),
        ),
        mg.TypeAtomicTypeReference,
        "int32-closed-zero-ten-open",
    )
    mkRegx := func( pat string ) *mg.Struct {
        return parser.MustStruct( mg.QnameRegexRestriction, "pattern", pat )
    }
    b.addRt(
        parser.MustStruct( mg.QnameAtomicTypeReference,
            "name", stringQn,
            "restriction", mkRegx( "a*" ),
        ),
        mg.TypeAtomicTypeReference,
        "string-a-star",
    )
    b.addTest(
        &bind.BindTest{
            Mingle: parser.MustStruct( mg.QnameAtomicTypeReference,
                "name", int32Qn,
                "restriction", mkRegx( "a*" ),
            ),
            Type: mg.TypeAtomicTypeReference,
            Error: newVcErr( 
                b.p( "f1" ), 
                "regex restriction cannot be applied to mingle:core@v1/Int32",
            ),
            StartPath: b.p( "f1" ),
            Direction: bind.BindTestDirectionIn,
        },
    )
    for i, aret := range mg.GetAtomicRestrictionErrorTests() {
        b.addVcErr(
            makeAtomicRestrictionErrorTestInput( aret ),
            mg.TypeAtomicTypeReference,
            nil,
            atomicRestrictionErrorTestMessageAtIndex( aret, i ),
        )
    }
}

func ( b *bindTestBuilder ) addListTypeReferenceTests() {
    b.addRt(
        parser.MustStruct( mg.QnameListTypeReference,
            "element-type", b.atomicQnNs1V1Name1(),
            "allows-empty", true,
        ),
        mg.TypeListTypeReference,
        "list-type1-allows-empty",
    )
}

func ( b *bindTestBuilder ) addPointerTypeReferenceTests() {
    b.addRt(
        parser.MustStruct( mg.QnamePointerTypeReference,
            "type", b.atomicQnNs1V1Name1(),
        ),
        mg.TypePointerTypeReference,
        "pointer-type1",
    )
}

func ( b *bindTestBuilder ) addNullableTypeReferenceTests() {
    b.addRt(
        parser.MustStruct( mg.QnameNullableTypeReference,
            "type", parser.MustStruct( mg.QnamePointerTypeReference,
                "type", b.atomicQnNs1V1Name1(),
            ),
        ),
        mg.TypeNullableTypeReference,
        "nullable-type1",
    )
}

func ( b *bindTestBuilder ) addPrimitiveDefinitionTests() {
    b.addRt(
        parser.MustStruct( builtin.QnamePrimitiveDefinition,
            "name", b.coreQn( "Int32" ),
        ),
        builtin.TypePrimitiveDefinition,
        "prim-def1",
    )
}

func ( b *bindTestBuilder ) fieldDef( i int ) *mg.Struct {
    return parser.MustStruct( builtin.QnameFieldDefinition,
        "name", makeIdStruct( fmt.Sprintf( "f%d", i ) ),
        "type", parser.MustStruct( mg.QnameAtomicTypeReference,
            "name", b.coreQn( "Int32" ),
        ),
        "default", mg.Int32( i ),
    )
}

func ( b *bindTestBuilder ) addFieldDefinitionTests() {
    b.addRt( 
        parser.MustStruct( builtin.QnameFieldDefinition,
            "name", makeIdStruct( "f1" ),
            "type", b.atomicQnNs1V1Name1(),
        ),
        builtin.TypeFieldDefinition,
        "field-def0",
    )
    b.addRt( b.fieldDef( 1 ), builtin.TypeFieldDefinition, "field-def1" )
}

func ( b *bindTestBuilder ) fieldSet( sz int ) *mg.Struct {
    fldsTyp := asType( "mingle:types@v1/FieldDefinition*" ).
        ( *mg.ListTypeReference )
    flds := mg.NewList( fldsTyp )
    for i := 0; i < sz; i++ { flds.AddUnsafe( b.fieldDef( i ) ) }
    return parser.MustStruct( builtin.QnameFieldSet, "fields", flds )
}

func ( b *bindTestBuilder ) addFieldSetTests() {
    fldDefList := asType( "mingle:types@v1/FieldDefinition*" )
    b.addRt( b.fieldSet( 0 ), builtin.TypeFieldSet, "empty-field-set" )
    b.addRt( b.fieldSet( 3 ), builtin.TypeFieldSet, "field-set1" )
    b.addInErr(
        parser.MustStruct( builtin.QnameFieldSet,
            "fields", mg.MustList( 
                fldDefList,
                b.fieldDef( 1 ),
                b.fieldDef( 1 ),
            ),
        ),
        builtin.TypeFieldSet,
        mg.NewCastError(
            objpath.RootedAt( mkId( "fields" ) ).StartList().SetIndex( 1 ),
            "field redefined: f1",
        ),
    )
}

func ( b *bindTestBuilder ) unionTypeDef( sz int ) *mg.Struct {
    l := mg.NewList( 
        asType( "mingle:core@v1/TypeReference+" ).( *mg.ListTypeReference ) )
    for i := 0; i < sz; i++ { l.AddUnsafe( b.atomicQnNs1V1Name( i ) ) }
    return parser.MustStruct( builtin.QnameUnionTypeDefinition, "types", l )
}

func ( b *bindTestBuilder ) addUnionDefinitionTests() {
    b.addRt(
        b.unionTypeDef( 2 ),
        builtin.TypeUnionTypeDefinition,
        "union-type-def1",
    )
    b.addRt(
        parser.MustStruct( builtin.QnameUnionDefinition,
            "name", b.qnNs1V1Name1(),
            "union", b.unionTypeDef( 2 ),
        ),
        builtin.TypeUnionDefinition,
        "union-def1",
    )
    b.addInErr(
        b.unionTypeDef( 0 ),
        builtin.TypeUnionTypeDefinition,
        mg.NewCastError( objpath.RootedAt( mkId( "types" ) ), "empty list" ),
    )
    b.addInErr(
        parser.MustStruct( builtin.QnameUnionTypeDefinition,
            "types", mg.MustList(
                asType( "mingle:core@v1/TypeReference+" ),
                b.atomicQnNs1V1Name1(),
                b.atomicQnNs1V1Name1(),
            ),
        ),
        builtin.TypeUnionTypeDefinition,
        mg.NewCastError( nil, "union contains one or more ambiguous types" ),
    )
}

func ( b *bindTestBuilder ) callSig2() *mg.Struct {
    return parser.MustStruct( builtin.QnameCallSignature,
        "fields", b.fieldSet( 2 ),
        "return", b.atomicQnNs1V1Name1(),
        "throws", b.unionTypeDef( 2 ),
    )
}

func ( b *bindTestBuilder ) addCallSignatureTests() {
    b.addRt(
        parser.MustStruct( builtin.QnameCallSignature,
            "fields", b.fieldSet( 2 ),
            "return", b.atomicQnNs1V1Name1(),
        ),
        builtin.TypeCallSignature,
        "call-sig1",
    )
    b.addRt( b.callSig2(), builtin.TypeCallSignature, "call-sig2" )
}

func ( b *bindTestBuilder ) addPrototypeDefinitionTests() {
    b.addRt(
        parser.MustStruct( builtin.QnamePrototypeDefinition,
            "name", b.qnNs1V1Name1(),
            "signature", b.callSig2(),
        ),
        builtin.TypePrototypeDefinition,
        "proto-def1",
    )
}

func ( b *bindTestBuilder ) addStructDefinitionTests() {
    b.addRt(
        parser.MustStruct( builtin.QnameStructDefinition,
            "name", b.qnNs1V1Name1(),
            "fields", b.fieldSet( 1 ),
        ),
        builtin.TypeStructDefinition,
        "struct-def1",
    )
    b.addRt(
        parser.MustStruct( builtin.QnameStructDefinition,
            "name", b.qnNs1V1Name1(),
            "fields", b.fieldSet( 1 ),
            "constructors", b.unionTypeDef( 2 ),
        ),
        builtin.TypeStructDefinition,
        "struct-def2",
    )
}

func ( b *bindTestBuilder ) addSchemaDefinitionTests() {
    b.addRt(
        parser.MustStruct( builtin.QnameSchemaDefinition,
            "name", b.qnNs1V1Name1(),
            "fields", b.fieldSet( 0 ),
        ),
        builtin.TypeSchemaDefinition,
        "schema-def-empty-fields",
    )
    b.addRt(
        parser.MustStruct( builtin.QnameSchemaDefinition,
            "name", b.qnNs1V1Name1(),
            "fields", b.fieldSet( 2 ),
        ),
        builtin.TypeSchemaDefinition,
        "schema-def1",
    )
}

func ( b *bindTestBuilder ) addAliasedTypeDefinition() {
    b.addRt(
        parser.MustStruct( builtin.QnameAliasedTypeDefinition,
            "name", b.qnNs1V1Name1(),
            "aliased-type", parser.MustStruct( mg.QnameAtomicTypeReference,
                "name", b.coreQn( "Int32" ),
            ),
        ),
        builtin.TypeAliasedTypeDefinition,
        "aliased-def1",
    )
}

func ( b *bindTestBuilder ) addEnumDefinition() {
    idListTyp := 
        asType( "&mingle:core@v1/Identifier+" ).( *mg.ListTypeReference )
    b.addRt(
        parser.MustStruct( builtin.QnameEnumDefinition,
            "name", b.qnNs1V1Name1(),
            "values", mg.MustList( idListTyp, 
                makeIdStruct( "v1" ), 
                makeIdStruct( "v2" ),
            ),
        ),
        builtin.TypeEnumDefinition,
        "enum-def1",
    )
    b.addInErr(
        parser.MustStruct( builtin.QnameEnumDefinition,
            "name", b.qnNs1V1Name1(),
            "values", mg.MustList( idListTyp, 
                makeIdStruct( "v1" ), 
                makeIdStruct( "v1" ),
            ),
        ),
        builtin.TypeEnumDefinition,
        mg.NewCastError( nil, "duplicate enum value(s): v1" ),
    )
}

func ( b *bindTestBuilder ) opDef( opNm string ) *mg.Struct {
    return parser.MustStruct( builtin.QnameOperationDefinition,
        "name", makeIdStruct( opNm ),
        "signature", b.callSig2(),
    )
}

func ( b *bindTestBuilder ) addOperationDefinitionTests() {
    b.addRt( b.opDef( "op1" ), builtin.TypeOperationDefinition, "op-def1" )
}

func ( b *bindTestBuilder ) addServiceDefinitionTests() {
    listTyp := asType( "&mingle:types@v1/OperationDefinition*" ).
        ( *mg.ListTypeReference )
    b.addRt(
        parser.MustStruct( builtin.QnameServiceDefinition,
            "name", b.qnNs1V1Name1(),
            "operations", mg.MustList( listTyp ),
        ),
        builtin.TypeServiceDefinition,
        "service-def1",
    )
    b.addRt(
        parser.MustStruct( builtin.QnameServiceDefinition,
            "name", b.qnNs1V1Name1(),
            "operations", mg.MustList( listTyp, 
                b.opDef( "op1" ), b.opDef( "op2" ),
            ),
            "security", b.qnNs1V1Name1(),
        ),
        builtin.TypeServiceDefinition,
        "service-def2",
    )
    b.addInErr(
        parser.MustStruct( builtin.QnameServiceDefinition,
            "name", b.qnNs1V1Name1(),
            "operations", mg.MustList( listTyp, 
                b.opDef( "op1" ), b.opDef( "op1" ),
            ),
        ),
        builtin.TypeServiceDefinition,
        mg.NewCastError( nil, "operation(s) redefined: op1" ),
    )
}

func ( b *bindTestBuilder ) addCoreBindTests() {
    b.addIdentifierTests()
    b.addDeclaredTypeNameTests()
    b.addNamespaceTests()
    b.addIdentifierPathTests()
    b.addQualifiedTypeNameTests()
    b.addCoreErrorTests()
    b.addAtomicTypeReferenceTests()
    b.addListTypeReferenceTests()
    b.addPointerTypeReferenceTests()
    b.addNullableTypeReferenceTests()
}

func ( b *bindTestBuilder ) addTypesBindTests() {
    b.addPrimitiveDefinitionTests()
    b.addFieldDefinitionTests()
    b.addFieldSetTests()
    b.addUnionDefinitionTests()
    b.addCallSignatureTests()
    b.addPrototypeDefinitionTests()
    b.addStructDefinitionTests()
    b.addSchemaDefinitionTests()
    b.addAliasedTypeDefinition()
    b.addEnumDefinition()
    b.addOperationDefinitionTests()
    b.addServiceDefinitionTests()
}

func getBindTests() []*bind.BindTest {
    b := &bindTestBuilder{ tests: make( []*bind.BindTest, 0, 128 ) }
    b.addCoreBindTests()
    b.addTypesBindTests()
    return b.tests
}

type bindTestCallInterface struct {
    boundVals *mg.IdentifierMap
}

func ( i bindTestCallInterface ) BoundValues() *mg.IdentifierMap {
    return i.boundVals
}

func ( i bindTestCallInterface ) CreateReactors( 
    t *bind.BindTest ) []interface{} {

    typ := t.Type
    if typ == nil { typ = mg.TypeNullableValue }
    cr := types.NewCastReactor( typ, builtin.BuiltinTypes() )
    builtin.CastBuiltinTypes( cr )
    return []interface{}{ cr }
}

type fldDefStructSorter []*mg.Struct

func ( s fldDefStructSorter ) Len() int { return len( s ) }

func ( s fldDefStructSorter ) Less( i, j int ) bool {
    getNm := func( i int ) *mg.Identifier {
        nm := s[ i ].Fields.Get( mkId( "name" ) ).( *mg.Struct )
        parts := nm.Fields.Get( mkId( "parts" ) ).( *mg.List )
        strs := make( []string, parts.Len() )
        for k, e := 0, parts.Len(); k < e; k++ {
            strs[ k ] = string( parts.Get( k ).( mg.String ) )
        }
        return mg.NewIdentifierUnsafe( strs )
    }
    return getNm( i ).Compare( getNm( j ) ) < 0
}

func ( s fldDefStructSorter ) Swap( i, j int ) {
    s[ i ], s[ j ] = s[ j ], s[ i ]
}

func mgFieldSetForAssert( ms *mg.Struct ) *mg.Struct {
    fldDefsIn := ms.Fields.Get( mkId( "fields" ) ).( *mg.List )
    fldDefsOut := mg.NewList( fldDefsIn.Type )
    fldDefs := make( []*mg.Struct, fldDefsIn.Len() )
    for i, e := 0, fldDefsIn.Len(); i < e; i++ {
        fldDefs[ i ] = fldDefsIn.Get( i ).( *mg.Struct )
    }
    sort.Sort( fldDefStructSorter( fldDefs ) )
    for _, fldDef := range fldDefs { fldDefsOut.AddUnsafe( fldDef ) }
    return parser.MustStruct( builtin.QnameFieldSet, "fields", fldDefsOut )
}

func mgSymbolMapForAssert( m *mg.SymbolMap ) *mg.SymbolMap {
    res := mg.NewSymbolMap()
    m.EachPair( func( fld *mg.Identifier, val mg.Value ) {
        res.Put( fld, mgValueForAssert( val ) )
    })
    return res
}

func mgStructForAssert( s *mg.Struct ) *mg.Struct {
    if s.Type.Equals( builtin.QnameFieldSet ) { 
        return mgFieldSetForAssert( s ) 
    }
    res := mg.NewStruct( s.Type )
    res.Fields = mgSymbolMapForAssert( s.Fields )
    return res
}

func mgListForAssert( l *mg.List ) *mg.List {
    l2 := mg.NewList( l.Type )
    for i, e := 0, l.Len(); i < e; i++ {
        l2.AddUnsafe( mgValueForAssert( l.Get( i ) ) )
    }
    return l2
}

func mgValueForAssert( val mg.Value ) mg.Value {
    switch v := val.( type ) {
    case *mg.Struct: return mgStructForAssert( v )
    case *mg.SymbolMap: return mgSymbolMapForAssert( v )
    case *mg.List: return mgListForAssert( v )
    }
    return val
}

func TestBind( t *testing.T ) {
    iface := bindTestCallInterface{ getBoundTestValues() }
    cc := &bind.BindTestCallControl{ Interface: iface }
    cc.MingleValueForAssert = mgValueForAssert
    bind.AssertBindTests( getBindTests(), cc, assert.NewPathAsserter( t ) )
}
