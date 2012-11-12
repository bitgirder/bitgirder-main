package com.bitgirder.demo.mingle;

import com.bitgirder.validation.State;

import com.bitgirder.mingle.model.AtomicTypeReference;
import com.bitgirder.mingle.model.QualifiedTypeName;
import com.bitgirder.mingle.model.MingleEnum;
import com.bitgirder.mingle.model.MingleException;
import com.bitgirder.mingle.model.MingleExceptionBuilder;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleList;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleNamespace;
import com.bitgirder.mingle.model.MingleNull;
import com.bitgirder.mingle.model.MingleServiceRequest;
import com.bitgirder.mingle.model.MingleServiceResponse;
import com.bitgirder.mingle.model.MingleStruct;
import com.bitgirder.mingle.model.MingleStructBuilder;
import com.bitgirder.mingle.model.MingleSymbolMap;
import com.bitgirder.mingle.model.MingleSymbolMapBuilder;
import com.bitgirder.mingle.model.MingleTimestamp;
import com.bitgirder.mingle.model.MingleTypeName;
import com.bitgirder.mingle.model.MingleTypeReference;

import com.bitgirder.demo.Demo;
import com.bitgirder.demo.DemoRunner;

import com.bitgirder.log.CodeLoggers;

import java.nio.ByteBuffer;

// Demonstration of working with the core mingle types in java
@Demo
final
class MingleModelsDemo
implements DemoRunner.SimpleDemo
{
    private final static State state = new State();

    private static void code( Object... msg ) { CodeLoggers.code( msg ); }

    // Many declaration types, including MingleIdentifier, MingleNamespace,
    // MingleTypeName, and MingleTypeReference, can be easily create from
    // CharSequences. Most carry two variants of factory method: create() and
    // parse(). The distintion has to do with how each handles invalid input: 
    // create() fails with a runtime exception on the assumption that the
    // programmer wishes to assert that the string being passed is known to be
    // in the correct form; parse() fails with a checked
    // com.bitgirder.parser.SyntaxException, and is used in situations where the
    // input is unknown and the program is expected to handle the error in some
    // meaningful way. For an example of this type of thing in the JDK, see the
    // documentation for java.net.URI.create()
    private
    static
    void
    showDeclarationObjects()
    {
        // create identifiers from various forms and show that they are all the
        // same as java objects
        MingleIdentifier id1 = MingleIdentifier.create( "some-ident" );
        MingleIdentifier id2 = MingleIdentifier.create( "some_ident" );
        MingleIdentifier id3 = MingleIdentifier.create( "someIdent" );
        state.isTrue( id1.equals( id2 ) && id2.equals( id3 ) );
        code( "id[123] external form:", id1.getExternalForm() );

        MingleNamespace ns1 = MingleNamespace.create( "mingle:demo:ns1" );
        code( "ns1:", ns1.getExternalForm() );

        MingleTypeName mt1 = MingleTypeName.create( "Struct1" );
        MingleTypeName mt2 = MingleTypeName.create( "Struct2" );
        code( "mt1:", mt1.getExternalForm(), "; mt2:", mt2.getExternalForm() );

        // An atomic type reference (as opposed to a NullableTypeReference or a
        // QuantifiedTypeReference) is the only one mingle programmers would
        // create for the time being. It is essentially namespace + type
        // (possibly nested type) as one would expect.
        MingleTypeReference typeRef = 
            AtomicTypeReference.create( 
                QualifiedTypeName.create( ns1, mt1, mt2 ) );
        
        code( "typeRef:", typeRef.getExternalForm() );
    }

    // Various constructions for the mingle primitives
    private
    void
    showPrimitives()
    {
        code( "A MingleString:",
            MingleModels.inspect( MingleModels.asMingleString( "hello" ) ) );

        code( "A MingleIntegral:",
            MingleModels.inspect( MingleModels.asMingleIntegral( 12 ) ) );

        code( "A MingleDecimal:",
            MingleModels.inspect( MingleModels.asMingleDecimal( 923.33 ) ) );
        
        code( "A MingleBoolean:",
            MingleModels.inspect( MingleModels.asMingleBoolean( true ) ) );

        code( "A MingleBuffer:",
            MingleModels.inspect( 
                MingleModels.asMingleBuffer( ByteBuffer.allocate( 64 ) ) ) );

        code( "A MingleNull:",
            MingleModels.inspect( MingleNull.getInstance() ) );
        
        code( "A MingleTimestamp:",
            MingleModels.inspect( MingleTimestamp.now() ) );
    }

    // MingleLists are immutable lists built from mingle values (including other
    // lists if desired)
    private
    void
    showMingleList()
    {
        MingleList ml =
            new MingleList.Builder().
                add( MingleModels.asMingleString( "elt1" ) ).
                add( MingleModels.asMingleIntegral( 2 ) ).
                add( MingleNull.getInstance() ).
                add(
                    new MingleList.Builder().
                        add( MingleModels.asMingleString( "elt2" ) ).
                        add( MingleModels.asMingleString( "elt3" ) ).
                        build()
                ).
                add( MingleModels.getEmptyList() ).
                build();
        
        code( "A MingleList:", MingleModels.inspect( ml ) );
    }

    // Similarly to MingleList, a symbol map is an immutable structure built
    // using a builder. Its keys are mingle identifiers and values may be any
    // mingle value. To make java code simpler, the MingleSymbolMapBuilder class
    // exposes convenience setters which take CharSequence objects as keys that
    // will be converted internally to MingleIdentifier using
    // MingleIdentifier.create() and which take and convert some common Java
    // objects as values.
    private
    void
    showMingleSymbolMap()
    {
        MingleSymbolMap msm =
            MingleModels.symbolMapBuilder().
                setString( "key1", "hello" ).
                setIntegral( "key2", -123 ).
                setDecimal( "key3", 99.9999999 ).
                set( "key4",
                    new MingleList.Builder().
                        add( MingleModels.asMingleString( "stuff" ) ).
                        add( MingleModels.asMingleString( "more stuff" ) ).
                        build()
                ).
                set( "key5",
                    MingleModels.symbolMapBuilder().
                        setString( "sub-key1", "sub stuff" ).
                        setIntegral( "sub-key2", 999888777 ).
                        build()
                ).
                build();
        
        code( "A MingleSymbolMap:", MingleModels.inspect( msm ) );
    }

    private
    void
    showCollections()
    {
        showMingleList();
        showMingleSymbolMap();
    }

    // Creating an enum instance is as simple as passing in its type reference
    // and the constant.
    private
    void
    showEnum()
    {
        AtomicTypeReference typeRef =
            (AtomicTypeReference)
                MingleTypeReference.create( "mingle:demo:ns/AnEnum" );
        
        MingleIdentifier constant = MingleIdentifier.create( "green" );

        code( "A MingleEnum:", 
            MingleModels.inspect( MingleEnum.create( typeRef, constant ) ) );
    }

    // Manually build a struct by setting its type reference and some fields
    private
    void
    showStruct()
    {
        MingleStructBuilder b = MingleModels.structBuilder();

        b.setType( "mingle:demo:ns3/SomeStruct" );

        b.fields().
            setString( "field1", "great" ).fields().
            setIntegral( "field2", 943 ).fields().
            set( "field3", MingleModels.getEmptyList() );

        MingleStruct ms = b.build();

        code( "A MingleStruct:", MingleModels.inspect( ms ) );
    }

    // Build an exception in largely the same way as is done with a struct, but
    // using the convenience setter for the message field
    private
    void
    showException()
    {
        MingleExceptionBuilder b = MingleModels.exceptionBuilder();

        b.setType( "mingle:demo:ns3/SomeException" );

        // Shorthand for b.fields().setString( "message", "Error occurred..." )
        b.setMessage( "Error occurred doing some stuff" );

        b.fields().setIntegral( "attempts", 4 );

        MingleException me = b.build();

        code( "A MingleException:", MingleModels.inspect( me ) );
    }

    private
    void
    showStructures()
    {
        showStruct();
        showException();
    }

    // build a mingle service request
    private
    void
    showServiceRequest()
    {
        MingleServiceRequest.Builder b =
            new MingleServiceRequest.Builder().
                setNamespace( MingleNamespace.create( "mingle:demo:ns4" ) ).
                setService( MingleIdentifier.create( "cool-service" ) ).
                setOperation( MingleIdentifier.create( "do-something" ) ).
                setAuthentication( 
                    MingleModels.asMingleString( "auth-ticket" ) );
        
        b.params().setString( "user-name", "farfle" );
        b.params().setString( "user-dob", "Jan 21, 1881" );

        code( "A MingleServiceRequest:", 
            MingleModels.inspect( b.build(), true ) );
    }

    // build a success response indicative of an operation that returns an
    // integral value
    private
    void
    showSuccessResponse()
    {
        MingleServiceResponse resp =
            MingleServiceResponse.createSuccess(
                MingleModels.asMingleIntegral( 100 ) 
            );
        
        code( "A MingleServiceResponse (success response):",
            MingleModels.inspect( resp ) );
    }

    // build an exception response indicative of a failed operation
    private
    void
    showExceptionResponse()
    {
        MingleExceptionBuilder b =
            MingleModels.exceptionBuilder().
                setType( "mingle:demo:ns4/SomeAppException" );
        
        b.setMessage( "App could not complete the blah blah operation" );

        MingleException me = b.build();

        MingleServiceResponse resp = MingleServiceResponse.createFailure( me );
        
        code( "A MingleServiceResponse (failure response):",
            MingleModels.inspect( resp ) );
    }

    private
    void
    showServiceObjects()
    {
        showServiceRequest();
        showSuccessResponse();
        showExceptionResponse();
    }

    public
    void
    runDemo()
    {
        showDeclarationObjects();
        showPrimitives();
        showCollections();
        showEnum();
        showStructures();
        showServiceObjects();
    }
}
