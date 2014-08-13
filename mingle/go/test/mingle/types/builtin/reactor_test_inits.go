package builtin

import (
    mg "mingle"
    "mingle/parser"
    mgRct "mingle/reactor"
    "mingle/types"
    "bitgirder/objpath"
    "mingle/bind"
)

var newVcErr = mg.NewValueCastError

func addBuiltinTypeTests( b *mgRct.ReactorTestSetBuilder ) {
    dm := types.MakeV1DefMap()
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
        return parser.MustStruct( QnameIdentifier, "parts", l )
    }
    add := func( in, typ, expct interface{} ) {
        b.AddTests(
            &BuiltinTypeTest{
                In: mg.MustValue( in ),
                Type: asType( typ ),
                Expect: expct,
                Map: dm,
            },
        )
    }
    addErr := func( in, typ interface{}, err error ) {
        b.AddTests(
            &BuiltinTypeTest{
                In: mg.MustValue( in ),
                Type: asType( typ ),
                Err: err,
                Map: dm,
            },
        )
    }
    addVcErr := func( in, typ interface{}, path objpath.PathNode, msg string ) {
        addErr( in, typ, newVcErr( path, msg ) )
    }
    addBindErr := func( 
        in, typ interface{}, path objpath.PathNode, msg string ) {

        addErr( in, typ, bind.NewBindError( path, msg ) )
    }
    add( idStruct( "id1" ), TypeIdentifier, mkId( "id1" ) )
    add( idStruct( "id1", "id2" ), TypeIdentifier, mkId( "id1-id2" ) )
    add( idBytes( "id1" ), TypeIdentifier, mkId( "id1" ) )
    add(
        mkId( "id1" ).ExternalForm(),
        TypeIdentifier,
        mkId( "id1" ),
    )
    addVcErr( 
        "id$Bad", 
        TypeIdentifier, 
        nil, 
        "[<input>, line 1, col 3]: Invalid id rune: \"$\" (U+0024)",
    )
    addVcErr( 
        badBytes,
        TypeIdentifier,
        nil,
        "[offset 0]: Expected type code 0x01 but got 0x00",
    )
    addVcErr(
        idStruct( "part1", "BadPart" ),
        TypeIdentifier,
        p( "parts" ).StartList().SetIndex( 1 ),
        "Value \"BadPart\" does not satisfy restriction \"^[a-z][a-z0-9]*$\"",
    )
    add(
        parser.MustStruct( QnameNamespace,
            "version", idStruct( "v1" ),
            "parts", mg.MustList( idStruct( "ns1" ) ),
        ),
        TypeNamespace,
        mkNs( "ns1@v1" ),
    )
    add(
        parser.MustStruct( QnameNamespace,
            "version", idStruct( "v1" ),
            "parts", mg.MustList( idStruct( "ns1" ), idStruct( "ns2" ) ),
        ),
        TypeNamespace,
        mkNs( "ns1:ns2@v1" ),
    )
    add(
        parser.MustStruct( QnameNamespace,
            "version", "v1",
            "parts", mg.MustList( "ns1", "ns2" ),
        ),
        TypeNamespace,
        mkNs( "ns1:ns2@v1" ),
    )
    add(
        parser.MustStruct( QnameNamespace,
            "version", idBytes( "v1" ),
            "parts", mg.MustList( idBytes( "ns1" ), idBytes( "ns2" ) ),
        ),
        TypeNamespace,
        mkNs( "ns1:ns2@v1" ),
    )
    add(
        parser.MustStruct( QnameNamespace,
            "version", idStruct( "v1" ),
            "parts", mg.MustList( "ns1", idBytes( "ns2" ) ),
        ),
        TypeNamespace,
        mkNs( "ns1:ns2@v1" ),
    )
    add( "ns1@v1", TypeNamespace, mkNs( "ns1@v1" ) )
    add( nsBytes( "ns1@v1" ), TypeNamespace, mkNs( "ns1@v1" ) )
    addVcErr(
        parser.MustStruct( QnameNamespace,
            "version", "bad$ver",
            "parts", mg.MustList( idStruct( "ns1" ) ),
        ),
        TypeNamespace,
        p( "version" ),
        "[<input>, line 1, col 4]: Invalid id rune: \"$\" (U+0024)",
    ) 
    addVcErr(
        parser.MustStruct( QnameNamespace,
            "version", idStruct( "v1" ),
            "parts", mg.MustList( idStruct( "ns1" ), "bad$Part" ),
        ),
        TypeNamespace,
        p( "parts" ).StartList().SetIndex( 1 ),
        "[<input>, line 1, col 4]: Invalid id rune: \"$\" (U+0024)",
    ) 
    addVcErr(
        parser.MustStruct( QnameNamespace,
            "version", idStruct( "v1" ),
            "parts", mg.MustList( idStruct( "ns1" ), badBytes ),
        ),
        TypeNamespace,
        p( "parts" ).StartList().SetIndex( 1 ),
        "[offset 0]: Expected type code 0x01 but got 0x00",
    )
    addVcErr(
        badBytes,
        TypeNamespace,
        nil,
        "[offset 0]: Expected type code 0x02 but got 0x00",
    )
    addVcErr(
        "Bad@Bad",
        TypeNamespace,
        nil,
        "[<input>, line 1, col 1]: Illegal start of identifier part: \"B\" (U+0042)",
    )
    idPathStruct := func( parts ...interface{} ) *mg.Struct {
        return parser.MustStruct( QnameIdentifierPath,
            "parts", mg.MustList( parts... ),
        )
    }
    add(
        idPathStruct(
            idStruct( "p1" ),
            idStruct( "p2" ),
            int32( 1 ),
            idStruct( "p3" ),
        ),
        TypeIdentifierPath,
        p( "p1" ).
            Descend( mkId( "p2" ) ).
            StartList().SetIndex( 1 ).
            Descend( mkId( "p3" ) ),
    )
    add(
        idPathStruct(
            idStruct( "p1" ),
            "p2",
            int32( 1 ),
            uint32( 2 ),
            int64( 3 ),
            uint64( 4 ),
            idBytes( "p3" ),
        ),
        TypeIdentifierPath,
        p( "p1" ).
            Descend( mkId( "p2" ) ).
            StartList().SetIndex( 1 ).
            StartList().SetIndex( 2 ).
            StartList().SetIndex( 3 ).
            StartList().SetIndex( 4 ).
            Descend( mkId( "p3" ) ),
    )
    add(
        "p1.p2[ 3 ].p4",
        TypeIdentifierPath,
        p( "p1" ).
            Descend( mkId( "p2" ) ).
            StartList().SetIndex( 3 ).
            Descend( mkId( "p4" ) ),
    )
    addVcErr(
        "p1.bad$Id",
        TypeIdentifierPath,
        nil,
        "[<input>, line 1, col 7]: Invalid id rune: \"$\" (U+0024)",
    )
    addVcErr( 
        idPathStruct(), 
        TypeIdentifierPath,
        p( "parts" ),
        "empty list",
    )
    addBindErr(
        idPathStruct( true ),
        TypeIdentifierPath,
        p( "parts" ).StartList(),
        "unhandled value: mingle:core@v1/Boolean",
    )
    addVcErr(
        idPathStruct( "bad$Id" ),
        TypeIdentifierPath,
        p( "parts" ).StartList(),
        "[<input>, line 1, col 4]: Invalid id rune: \"$\" (U+0024)",
    )
    addVcErr(
        idPathStruct( badBytes ),
        TypeIdentifierPath,
        p( "parts" ).StartList(),
        "[offset 0]: Expected type code 0x01 but got 0x00",
    )
    addBindErr(
        idPathStruct( float32( 1 ) ),
        TypeIdentifierPath,
        p( "parts" ).StartList(),
        "unhandled value: mingle:core@v1/Float32",
    )
    addBindErr(
        idPathStruct( float64( 1 ) ),
        TypeIdentifierPath,
        p( "parts" ).StartList(),
        "unhandled value: mingle:core@v1/Float64",
    )
    addVcErr(
        idPathStruct( int32( -1 ) ),
        TypeIdentifierPath,
        p( "parts" ).StartList(),
        "value is negative",
    )
    addVcErr(
        idPathStruct( int64( -1 ) ),
        TypeIdentifierPath,
        p( "parts" ).StartList(),
        "value is negative",
    )
}

func init() {
    mgRct.AddTestInitializer( 
        reactorTestNs, 
        func( b *mgRct.ReactorTestSetBuilder ) {
            addBuiltinTypeTests( b )
        },
    )
}
