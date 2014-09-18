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
    res.Put( mkId( "namespace-ns1-v1" ), mkNs( "ns1@v1" ) )
    res.Put( mkId( "namespace-ns1-ns2-v1" ), mkNs( "ns1:ns2@v1" ) )
    tp := mg.MakeTestIdPath
    res.Put( mkId( "identifier-path-f1-f2-i1-f3" ), tp( 1, 2, "1", 3 ) )
    res.Put( 
        mkId( "identifier-path-f1-f2-i1-i2-i3-i4-f3" ),
        tp( 1, 2, "1", "2", "3", "4", 3 ),
    )
    res.Put( mkId( "identifier-path-f1-i2" ), tp( 1, "2" ) )
    res.Put( mkId( "identifier-path-f1-f2-i3-f4" ), tp( 1, 2, "3", 4 ) )
    res.Put( mkId( "identifier-path-f1-i2-f3" ), tp( 1, "2", 3 ) )
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

func getBindTests() []*bind.BindTest {
    res := []*bind.BindTest{}
    idBytes := func( s string ) []byte {
        return mg.IdentifierAsBytes( mkId( s ) )
    }
    nsBytes := func( s string ) []byte {
        return mg.NamespaceAsBytes( mkNs( s ) )
    }
    badBytes := []byte{ 0, 0, 0, 0, 0, 0 }
    p := func( rootId string ) objpath.PathNode {
        return objpath.RootedAt( mkId( rootId ) )
    }
    idStruct := func( parts ...string ) *mg.Struct {
        l := mg.NewList( asType( "String+" ).( *mg.ListTypeReference ) )
        for _, part := range parts { l.AddUnsafe( mg.String( part ) ) }
        return parser.MustStruct( mg.QnameIdentifier, "parts", l )
    }
    addTest := func( t *bind.BindTest ) {
        t.Domain = bind.DomainDefault
        res = append( res, t )
    }
    addRtOpts := func( 
        in, typ interface{}, id string, opts *bind.SerialOptions ) {

        addTest(
            &bind.BindTest{
                Mingle: mg.MustValue( in ),
                Type: asType( typ ),
                BoundId: mkId( id ),
                SerialOptions: opts,
            },
        )
    }
    addRt := func( in, typ interface{}, id string ) {
        addRtOpts( in, typ, id, nil )
    }
    addIn := func( in, typ interface{}, id string ) {
        addTest(
            &bind.BindTest{
                Mingle: mg.MustValue( in ),
                Type: asType( typ ),
                BoundId: mkId( id ),
                Direction: bind.BindTestDirectionIn,
            },
        )
    }
    addOut := func( id string, expct interface{}, opts *bind.SerialOptions ) {
        addTest(
            &bind.BindTest{
                Mingle: mg.MustValue( expct ),
                BoundId: mkId( id ),
                Direction: bind.BindTestDirectionOut,
                SerialOptions: opts,
            },
        )
    }
    addInErr := func( in, typ interface{}, err error ) {
        addTest(
            &bind.BindTest{
                Mingle: mg.MustValue( in ),
                Type: asType( typ ),
                Error: err,
                Direction: bind.BindTestDirectionIn,
            },
        )
    }
    addVcErr := func( in, typ interface{}, path objpath.PathNode, msg string ) {
        addInErr( in, typ, newVcErr( path, msg ) )
    }
    addRt( idStruct( "id1" ), mg.TypeIdentifier, "identifier-id1" )
    addRt( idStruct( "id1", "id2" ), mg.TypeIdentifier, "identifier-id1-id2" )
    binOpts := bind.NewSerialOptions()
    binOpts.Format = bind.SerialFormatBinary
    txtOpts := bind.NewSerialOptions()
    txtOpts.Format = bind.SerialFormatText
    addIn( idBytes( "id1" ), mg.TypeIdentifier, "identifier-id1" )
    addOut( "identifier-id1", idBytes( "id1" ), binOpts )
    addIn( mkId( "id1" ).ExternalForm(), mg.TypeIdentifier, "identifier-id1" )
    addIdOutString := func( expct string, fmt mg.IdentifierFormat ) {
        opts := bind.NewSerialOptions()
        opts.Format = bind.SerialFormatText
        opts.Identifiers = fmt
        addOut( "identifier-id1-id2", expct, opts )
    }
    addIdOutString( "id1-id2", mg.LcHyphenated )
    addIdOutString( "id1_id2", mg.LcUnderscore )
    addIdOutString( "id1Id2", mg.LcCamelCapped )
    addVcErr(
        "id$Bad", 
        mg.TypeIdentifier, 
        nil, 
        "[<input>, line 1, col 3]: Invalid id rune: \"$\" (U+0024)",
    )
    addVcErr( 
        badBytes,
        mg.TypeIdentifier,
        nil,
        "[offset 0]: Expected type code 0x01 but got 0x00",
    )
    addVcErr(
        idStruct( "part1", "BadPart" ),
        mg.TypeIdentifier,
        p( "parts" ).StartList().SetIndex( 1 ),
        "Value \"BadPart\" does not satisfy restriction \"^[a-z][a-z0-9]*$\"",
    )
    addRt(
        parser.MustStruct( mg.QnameNamespace,
            "version", idStruct( "v1" ),
            "parts", mg.MustList( idStruct( "ns1" ) ),
        ),
        mg.TypeNamespace,
        "namespace-ns1-v1",
    )
    addRt(
        parser.MustStruct( mg.QnameNamespace,
            "version", idStruct( "v1" ),
            "parts", mg.MustList( idStruct( "ns1" ), idStruct( "ns2" ) ),
        ),
        mg.TypeNamespace,
        "namespace-ns1-ns2-v1",
    )
    addIn(
        parser.MustStruct( mg.QnameNamespace,
            "version", "v1",
            "parts", mg.MustList( "ns1", "ns2" ),
        ),
        mg.TypeNamespace,
        "namespace-ns1-ns2-v1",
    )
    addIn(
        parser.MustStruct( mg.QnameNamespace,
            "version", idBytes( "v1" ),
            "parts", mg.MustList( idBytes( "ns1" ), idBytes( "ns2" ) ),
        ),
        mg.TypeNamespace,
        "namespace-ns1-ns2-v1",
    )
    addIn(
        parser.MustStruct( mg.QnameNamespace,
            "version", idStruct( "v1" ),
            "parts", mg.MustList( "ns1", idBytes( "ns2" ) ),
        ),
        mg.TypeNamespace,
        "namespace-ns1-ns2-v1",
    )
    addIn( "ns1@v1", mg.TypeNamespace, "namespace-ns1-v1" )
    addIn( nsBytes( "ns1@v1" ), mg.TypeNamespace, "namespace-ns1-v1" )
    addOut( "namespace-ns1-v1", "ns1@v1", txtOpts )
    addOut( "namespace-ns1-v1", nsBytes( "ns1@v1" ), binOpts )
    addVcErr(
        parser.MustStruct( mg.QnameNamespace,
            "version", "bad$ver",
            "parts", mg.MustList( idStruct( "ns1" ) ),
        ),
        mg.TypeNamespace,
        p( "version" ),
        "[<input>, line 1, col 4]: Invalid id rune: \"$\" (U+0024)",
    ) 
    addVcErr(
        parser.MustStruct( mg.QnameNamespace,
            "version", idStruct( "v1" ),
            "parts", mg.MustList( idStruct( "ns1" ), "bad$Part" ),
        ),
        mg.TypeNamespace,
        p( "parts" ).StartList().SetIndex( 1 ),
        "[<input>, line 1, col 4]: Invalid id rune: \"$\" (U+0024)",
    ) 
    addVcErr(
        parser.MustStruct( mg.QnameNamespace,
            "version", idStruct( "v1" ),
            "parts", mg.MustList( idStruct( "ns1" ), badBytes ),
        ),
        mg.TypeNamespace,
        p( "parts" ).StartList().SetIndex( 1 ),
        "[offset 0]: Expected type code 0x01 but got 0x00",
    )
    addVcErr(
        badBytes,
        mg.TypeNamespace,
        nil,
        "[offset 0]: Expected type code 0x02 but got 0x00",
    )
    addVcErr(
        "Bad@Bad",
        mg.TypeNamespace,
        nil,
        "[<input>, line 1, col 1]: Illegal start of identifier part: \"B\" (U+0042)",
    )
    idPathStruct := func( parts ...interface{} ) *mg.Struct {
        return parser.MustStruct( mg.QnameIdentifierPath,
            "parts", mg.MustList( parts... ),
        )
    }
    addRt(
        idPathStruct(
            idStruct( "f1" ),
            idStruct( "f2" ),
            uint64( 1 ),
            idStruct( "f3" ),
        ),
        mg.TypeIdentifierPath,
        "identifier-path-f1-f2-i1-f3",
    )
    addIn(
        idPathStruct(
            idStruct( "f1" ),
            "f2",
            int32( 1 ),
            uint32( 2 ),
            int64( 3 ),
            uint64( 4 ),
            idBytes( "f3" ),
        ),
        mg.TypeIdentifierPath,
        "identifier-path-f1-f2-i1-i2-i3-i4-f3",
    )
    addIn( "f1[ 2 ]", mg.TypeIdentifierPath, "identifier-path-f1-i2" )
    addIn(
        "f1.f2[ 3 ].f4",
        mg.TypeIdentifierPath,
        "identifier-path-f1-f2-i3-f4",
    )
    addOut( "identifier-path-f1-i2-f3", "f1[ 2 ].f3", txtOpts )
    addOut( 
        "identifier-path-f1-i2-f3",
        idPathStruct( idBytes( "f1" ), uint64( 2 ), idBytes( "f3" ) ),
        binOpts,
    )
    addVcErr(
        "p1.bad$Id",
        mg.TypeIdentifierPath,
        nil,
        "[<input>, line 1, col 7]: Invalid id rune: \"$\" (U+0024)",
    )
    addVcErr( 
        idPathStruct(), 
        mg.TypeIdentifierPath,
        p( "parts" ),
        "empty list",
    )
    addVcErr(
        idPathStruct( true ),
        mg.TypeIdentifierPath,
        p( "parts" ).StartList(),
        "invalid value for identifier path part: mingle:core@v1/Boolean",
    )
    addVcErr(
        idPathStruct( "bad$Id" ),
        mg.TypeIdentifierPath,
        p( "parts" ).StartList(),
        "[<input>, line 1, col 4]: Invalid id rune: \"$\" (U+0024)",
    )
    addVcErr(
        idPathStruct( badBytes ),
        mg.TypeIdentifierPath,
        p( "parts" ).StartList(),
        "[offset 0]: Expected type code 0x01 but got 0x00",
    )
    addVcErr(
        idPathStruct( float32( 1 ) ),
        mg.TypeIdentifierPath,
        p( "parts" ).StartList(),
        "invalid value for identifier path part: mingle:core@v1/Float32",
    )
    addVcErr(
        idPathStruct( float64( 1 ) ),
        mg.TypeIdentifierPath,
        p( "parts" ).StartList(),
        "invalid value for identifier path part: mingle:core@v1/Float64",
    )
    addVcErr(
        idPathStruct( int32( -1 ) ),
        mg.TypeIdentifierPath,
        p( "parts" ).StartList(),
        "value is negative",
    )
    addVcErr(
        idPathStruct( int64( -1 ) ),
        mg.TypeIdentifierPath,
        p( "parts" ).StartList(),
        "value is negative",
    )
    addRt( 
        parser.MustStruct( "mingle:core@v1/CastError",
            "message", "test-message",
            "location", idPathStruct( idStruct( "f1" ), uint64( 2 ) ),
        ),
        asType( "mingle:core@v1/CastError" ),
        "cast-error-loc-f1-i2",
    )
    addIn( 
        parser.MustStruct( "mingle:core@v1/CastError",
            "message", "test-message",
            "location", "f1[ 2 ]",
        ),
        asType( "mingle:core@v1/CastError" ),
        "cast-error-loc-f1-i2",
    )
    addOut( 
        "cast-error-loc-f1-i2",
        parser.MustStruct( "mingle:core@v1/CastError",
            "message", "test-message",
            "location", "f1[ 2 ]",
        ),
        txtOpts,
    )
    addRt(
        parser.MustStruct( "mingle:core@v1/UnrecognizedFieldError",
            "message", "test-message",
            "location", idPathStruct( idStruct( "f1" ), uint64( 2 ) ),
            "field", idStruct( "bad", "field" ),
        ),
        asType( "mingle:core@v1/UnrecognizedFieldError" ),
        "unrecognized-field-error-loc-f1-i2",
    )
    addOut(
        "unrecognized-field-error-loc-f1-i2",
        parser.MustStruct( "mingle:core@v1/UnrecognizedFieldError",
            "message", "test-message",
            "location", "f1[ 2 ]",
            "field", "bad-field",
        ),
        txtOpts,
    )
    addRt(
        parser.MustStruct( "mingle:core@v1/MissingFieldsError",
            "message", "test-message",
            "location", idPathStruct( idStruct( "f1" ), uint64( 2 ) ),
            "fields", mg.MustList( 
                asType( "&mingle:core@v1/Identifier+" ),
                idStruct( "f1" ),
                idStruct( "f2" ),
                idStruct( "f3" ),
            ),
        ),
        asType( "mingle:core@v1/MissingFieldsError" ),
        "missing-fields-error-f1-f2-f3-loc-f1-i2",
    )
    addIn(
        parser.MustStruct( "mingle:core@v1/MissingFieldsError",
            "message", "test-message",
            "location", idPathStruct( idStruct( "f1" ), uint64( 2 ) ),
            "fields", mg.MustList( 
                asType( "&mingle:core@v1/Identifier+" ),
                // unsorted in the input to ensure that setter sorts
                idStruct( "f3" ),
                idStruct( "f1" ),
                idStruct( "f2" ),
            ),
        ),
        asType( "mingle:core@v1/MissingFieldsError" ),
        "missing-fields-error-f1-f2-f3-loc-f1-i2",
    )
    addIn(
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
    addOut(
        "missing-fields-error-f1-f2-f3-loc-f1-i2",
        parser.MustStruct( "mingle:core@v1/MissingFieldsError",
            "message", "test-message",
            "location", "f1[ 2 ]",
            "fields", mg.MustList( asType( "String+" ), "f1", "f2", "f3" ),
        ),
        txtOpts,
    )
    addOut(
        "missing-fields-error-f1-f2-f3-loc-f1-i2",
        parser.MustStruct( "mingle:core@v1/MissingFieldsError",
            "message", "test-message",
            "location", idPathStruct( idBytes( "f1" ), uint64( 2 ) ),
            "fields", mg.MustList( asType( "Buffer+" ), 
                idBytes( "f1" ), idBytes( "f2" ), idBytes( "f3" ),
            ),
        ),
        binOpts,
    )
    return res
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
