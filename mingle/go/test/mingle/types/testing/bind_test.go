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
)

func getBoundTestValues() *mg.IdentifierMap {
    res := mg.NewIdentifierMap()
    res.Put( mkId( "identifier-id1" ), mkId( "id1" ) )
    res.Put( mkId( "identifier-id1-id2" ), mkId( "id1-id2" ) )
    res.Put( 
        mkId( "declared-type-name-name1" ), 
        parser.MustDeclaredTypeName( "Name1" ),
    )
    res.Put( mkId( "namespace-ns1-v1" ), mkNs( "ns1@v1" ) )
    res.Put( mkId( "namespace-ns1-ns2-v1" ), mkNs( "ns1:ns2@v1" ) )
    res.Put(
        mkId( "qualified-type-name-ns1-v1-name1" ),
        mkQn( "ns1@v1/Name1" ),
    )
    tp := mg.MakeTestIdPath
    res.Put( mkId( "identifier-path-f1-f2-i1-f3" ), tp( 1, 2, "1", 3 ) )
    res.Put( 
        mkId( "identifier-path-f1-f2-i1-i2-i3-i4-f3" ),
        tp( 1, 2, "1", "2", "3", "4", 3 ),
    )
    res.Put( mkId( "identifier-path-f1-i2" ), tp( 1, "2" ) )
    res.Put( mkId( "identifier-path-f1-f2-i3-f4" ), tp( 1, 2, "3", 4 ) )
    res.Put( mkId( "identifier-path-f1-i2-f3" ), tp( 1, "2", 3 ) )
    res.Put( mkId( "atomic-type-name-ns1-v1-name1" ), asType( "ns1@v1/Name1" ) )
    res.Put( mkId( "int32-closed-zero-ten-open" ), asType( "Int32~[0,10)" ) )
    res.Put( mkId( "string-a-star" ), asType( `String~"a*"` ) )
    res.Put( mkId( "list-type1-allows-empty" ), asType( "ns1@v1/Name1*" ) )
    res.Put( mkId( "pointer-type1" ), asType( "&ns1@v1/Name1" ) )
    res.Put( mkId( "nullable-type1" ), asType( "&ns1@v1/Name1?" ) )
    res.Put( 
        mkId( "cast-error-loc-f1-i2" ),
        mg.NewCastError( tp( 1, "2" ), "test-message" ),
    )
    res.Put(
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
    res.Put( mkId( "missing-fields-error-f1-f2-f3-loc-f1-i2" ), mfeInst1 )
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

func ( b *bindTestBuilder ) declNm1() *mg.Struct {
    return parser.MustStruct( mg.QnameDeclaredTypeName, "name", "Name1" )
}

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
        b.idPathStruct(), 
        mg.TypeIdentifierPath,
        b.p( "parts" ),
        "empty list",
    )
    b.addVcErr(
        b.idPathStruct( true ),
        mg.TypeIdentifierPath,
        b.p( "parts" ).StartList(),
        "invalid value for identifier path part: mingle:core@v1/Boolean",
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
    b.addVcErr(
        b.idPathStruct( float32( 1 ) ),
        mg.TypeIdentifierPath,
        b.p( "parts" ).StartList(),
        "invalid value for identifier path part: mingle:core@v1/Float32",
    )
    b.addVcErr(
        b.idPathStruct( float64( 1 ) ),
        mg.TypeIdentifierPath,
        b.p( "parts" ).StartList(),
        "invalid value for identifier path part: mingle:core@v1/Float64",
    )
    b.addVcErr(
        b.idPathStruct( int32( -1 ) ),
        mg.TypeIdentifierPath,
        b.p( "parts" ).StartList(),
        "value is negative",
    )
    b.addVcErr(
        b.idPathStruct( int64( -1 ) ),
        mg.TypeIdentifierPath,
        b.p( "parts" ).StartList(),
        "value is negative",
    )
}

func ( b *bindTestBuilder ) qnNs1V1Name1() *mg.Struct {
    return parser.MustStruct( mg.QnameQualifiedTypeName,
        "namespace", b.ns1V1(),
        "name", b.declNm1(),
    )
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

func ( b *bindTestBuilder ) atomicQnNs1V1Name1() *mg.Struct {
    return parser.MustStruct( mg.QnameAtomicTypeReference, 
        "name", b.qnNs1V1Name1(),
    )
}

func ( b *bindTestBuilder ) addAtomicTypeReferenceTests() {
    b.addRt(
        b.atomicQnNs1V1Name1(),
        mg.TypeAtomicTypeReference,
        "atomic-type-name-ns1-v1-name1",
    )
    coreQn := func( nm string ) *mg.Struct {
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
    int32Qn := coreQn( "Int32" )
    stringQn := coreQn( "String" )
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

func getBindTests() []*bind.BindTest {
    b := &bindTestBuilder{ tests: make( []*bind.BindTest, 0, 128 ) }
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
    return []interface{}{ cr }
}

func TestBind( t *testing.T ) {
    iface := bindTestCallInterface{ getBoundTestValues() }
    bind.AssertBindTests( getBindTests(), iface, assert.NewPathAsserter( t ) )
}
