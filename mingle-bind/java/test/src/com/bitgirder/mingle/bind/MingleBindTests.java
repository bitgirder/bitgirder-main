package com.bitgirder.mingle.bind;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.ObjectReceiver;

import com.bitgirder.lang.path.ObjectPath;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.io.IoUtils;

import com.bitgirder.mingle.model.AtomicTypeReference;
import com.bitgirder.mingle.model.FieldDefinition;
import com.bitgirder.mingle.model.ListTypeReference;
import com.bitgirder.mingle.model.MingleBoolean;
import com.bitgirder.mingle.model.MingleEnum;
import com.bitgirder.mingle.model.MingleException;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleList;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleNull;
import com.bitgirder.mingle.model.MingleStruct;
import com.bitgirder.mingle.model.MingleSymbolMap;
import com.bitgirder.mingle.model.MingleSymbolMapAccessor;
import com.bitgirder.mingle.model.MingleSymbolMapBuilder;
import com.bitgirder.mingle.model.MingleTypeCastException;
import com.bitgirder.mingle.model.MingleTypeReference;
import com.bitgirder.mingle.model.MingleValidationException;
import com.bitgirder.mingle.model.MingleValue;
import com.bitgirder.mingle.model.MingleTests;
import com.bitgirder.mingle.model.TypeDefinition;
import com.bitgirder.mingle.model.TypeDefinitions;
import com.bitgirder.mingle.model.TypeDefinitionLookup;
import com.bitgirder.mingle.model.NoSuchTypeDefinitionException;
import com.bitgirder.mingle.model.ModelTestInstances;
import com.bitgirder.mingle.model.NullableTypeReference;
import com.bitgirder.mingle.model.QualifiedTypeName;

import com.bitgirder.mingle.codec.MingleCodec;
import com.bitgirder.mingle.codec.MingleCodecFactory;
import com.bitgirder.mingle.codec.MingleCodecs;
import com.bitgirder.mingle.codec.MingleCodecTests;

import com.bitgirder.test.Test;
import com.bitgirder.test.TestRuntime;
import com.bitgirder.test.TestFailureExpector;
import com.bitgirder.test.InvocationFactory;
import com.bitgirder.test.LabeledTestCall;

import com.bitgirder.testing.Testing;

import java.nio.ByteBuffer;

import java.util.List;
import java.util.Iterator;

@Test
public
final
class MingleBindTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }
    
    public final static Object KEY_BINDER =
        MingleBindTests.class.getName() + ".binder";
    
    private final static Object KEY_BOOTSTRAP_TYPE_DEF_LOOKUP =
        MingleBindTests.class.getName() + ".bootstrapTypes";

    public final static StandardException STD_EXCEPTION_INST1;
    public final static MingleException STD_EXCEPTION_MG_INST1;

    private final static ObjectPath< MingleIdentifier > PATH1 =
        ObjectPath.
            getRoot( MingleIdentifier.create( "f1" ) ).
            startImmutableList().
            next().
            next().
            descend( MingleIdentifier.create( "f2" ) );

    private final static MingleList STRING_LIST1 =
        new MingleList.Builder().
            add( MingleModels.asMingleString( "s1" ) ).
            add( MingleModels.asMingleString( "s2" ) ).
            add( MingleModels.asMingleString( "s3" ) ).
            build();
    
    private final static MingleList STRING_LIST2 =
        new MingleList.Builder().
            add( MingleModels.asMingleString( "s1" ) ).
            add( MingleNull.getInstance() ).
            add( MingleModels.asMingleString( "s3" ) ).
            build();
    
    private final static MingleList EMPTY_LIST = MingleList.create();

    private final static MingleList OPAQUE_MG_LIST1;
    private final static List< ? > OPAQUE_JV_LIST1;

    private final TestRuntime rt;

    private MingleBindTests( TestRuntime rt ) { this.rt = rt; }

    public
    static
    void
    setMingleAuthentication( BoundServiceClient.AbstractCall call,
                             MingleValue mv )
    {
        inputs.notNull( call, "call" );
        inputs.notNull( mv, "mv" );

        call.setMingleAuthentication( mv );
    }

    public
    static
    QualifiedTypeName
    bindingNameForClass( MingleBinder mb,
                         Class< ? > cls )
    {
        inputs.notNull( mb, "mb" );
        inputs.notNull( cls, "cls" );

        return mb.bindingNameForClass( cls );
    }

    private
    CharSequence
    format( ObjectPath< MingleIdentifier > p )
    {
        return MingleModels.format( p );
    }

    public
    static
    enum Enum1
    {
        EN_VAL1,
        EN_VAL2,
        EN_VAL3;

        final static QualifiedTypeName QNAME =
            QualifiedTypeName.create( "mingle:bind:test@v1/Enum1" );

        public final static AtomicTypeReference TYPE =
            AtomicTypeReference.create( QNAME );

        final static MingleEnum MG_EN_VAL1 =
            MingleEnum.create( TYPE, MingleIdentifier.create( "en-val1" ) );
    }

    private
    final
    static
    class Enum1Binding
    extends AbstractEnumBinding< Enum1 >
    {
        private Enum1Binding() { super( Enum1.class ); }

        protected
        MingleEnum
        getMingleEnum( Enum1 en )
        {
            return
                MingleEnum.create(
                    Enum1.TYPE,
                    MingleIdentifier.create( en.name().toLowerCase() )
                );
        }

        protected
        Enum1
        getJavaEnum( MingleIdentifier id )
        {
            String test = id.getExternalForm().toString();

            if ( test.equals( "en-val1" ) ) return Enum1.EN_VAL1;
            else if ( test.equals( "en-val2" ) ) return Enum1.EN_VAL2;
            else if ( test.equals( "en-val3" ) ) return Enum1.EN_VAL3;
            else return null;
        }
    }

    public
    final
    static
    class Struct1
    {
        public final static Struct1 INST1;

        private final static QualifiedTypeName QNAME =
            QualifiedTypeName.create( "mingle:bind:test@v1/Struct1" );

        public final static AtomicTypeReference TYPE =
            AtomicTypeReference.create( QNAME );

        private final static MingleStruct MG_INST1;

        private final String string1;

        public
        Struct1( String string1 ) 
        { 
            this.string1 = inputs.notNull( string1, "string1" );
        }

        @Override public int hashCode() { return string1.hashCode(); }

        @Override
        public
        boolean
        equals( Object o )
        {
            return o == this ||
                   ( ( o instanceof Struct1 ) &&
                     ( (Struct1) o ).string1.equals( string1 ) );
        }

        static
        {
            INST1 = new Struct1( "hello" );

            MG_INST1 =
                MingleModels.structBuilder().
                    setType( Struct1.TYPE ).
                    f().setString( "string1", "hello" ).
                    build();
        }
    }

    private
    final
    static
    class Struct1Binding
    implements MingleBinding
    {
        public
        Object
        asJavaValue( AtomicTypeReference typ,
                     MingleValue mv,
                     MingleBinder mb,
                     ObjectPath< MingleIdentifier > path )
        {
            MingleSymbolMapAccessor acc =   
                MingleModels.expectStruct( mv, path, Struct1.TYPE );

            Struct1 res = new Struct1( acc.getString( "string1" ) );

            return res;
        }

        public
        MingleValue
        asMingleValue( Object obj,
                       MingleBinder mb,
                       ObjectPath< String > path )
        {
            Struct1 s1 = (Struct1) obj;

            return
                MingleModels.structBuilder().
                    setType( Struct1.TYPE ).
                    f().setString( "string1", s1.string1 ).
                    build();
        }
    }

    public
    static
    MingleBinder
    createTestBinder( TestRuntime rt )
    {
        TypeDefinitionLookup types = 
            Testing.expectObject(
                rt,
                KEY_BOOTSTRAP_TYPE_DEF_LOOKUP,
                TypeDefinitionLookup.class
            );

        MingleBinder.Builder b =
            new MingleBinder.Builder().
                addBinding( 
                    Struct1.QNAME, new Struct1Binding(), Struct1.class ).
                addBinding( Enum1.QNAME, new Enum1Binding(), Enum1.class ).
                setTypes( types );
        
        MingleBinders.addStandardBindings( b, types );

        return b.build();
    }

    private MingleBinder createTestBinder() { return createTestBinder( rt ); }

    private
    static
    MingleTypeReference
    typeRef( CharSequence s )
    {
        return MingleTypeReference.create( s );
    }

    private
    void
    assertEqual( Struct1 o1,
                 Struct1 o2 )
    {
        state.equalString( o1.string1, o2.string1 );
    }

    public
    static
    void
    assertEqual( BoundException e1,
                 BoundException e2 )
    {
        if ( state.sameNullity( e1.getCause(), e2.getCause() ) )
        {
            state.equal( e1.getCause(), e2.getCause() );
        }
    }

    public
    static
    void
    assertEqual( StandardException e1,
                 StandardException e2 )
    {
        if ( state.sameNullity( e1, e2 ) )
        {
            assertEqual( (BoundException) e1, (BoundException) e2 );
            state.equal( e1.getMessage(), e2.getMessage() );
        }
    }

    private
    static
    void
    assertEqual( MingleValidationException mve1,
                 MingleValidationException mve2 )
    {
        if ( state.sameNullity( mve1, mve2 ) )
        {
            state.equalString( mve1.getMessage(), mve2.getMessage() );

            state.equalString(
                MingleModels.format( mve1.getLocation() ),
                MingleModels.format( mve2.getLocation() ) );
        }
    }

    private
    boolean
    isType( MingleTypeReference mgTyp,
            CharSequence test )
    {
        return mgTyp.equals( typeRef( test ) );
    }

    private
    void
    assertEqual( Object jvObj1,
                 Object jvObj2,
                 AtomicTypeReference mgTyp )
        throws Exception
    {
        if ( isType( mgTyp, "mingle:core@v1/String" ) )
        {
            state.equalString( (CharSequence) jvObj1, (CharSequence) jvObj2 );
        }
        else if ( isType( mgTyp, "mingle:core@v1/Null" ) )
        {
            state.isTrue( jvObj1 == null || jvObj1 instanceof MingleNull );
            state.isTrue( jvObj2 == null );
        }
        else if ( mgTyp.equals( Struct1.TYPE ) )
        {
            assertEqual( (Struct1) jvObj1, (Struct1) jvObj2 );
        }
        else if ( isType( mgTyp, "mingle:core@v1/StandardException" ) )
        {
            assertEqual( 
                (StandardException) jvObj1, (StandardException) jvObj2 );
        }
        else if ( isType( mgTyp, "mingle:core@v1/ValidationException" ) )
        {
            assertEqual(
                (MingleValidationException) jvObj1,
                (MingleValidationException) jvObj2 );
        }
        else if ( isType( mgTyp, "mingle:core@v1/Value" ) )
        {
            ModelTestInstances.assertEqual(
                (MingleValue) jvObj1, (MingleValue) jvObj2 );
        }
        else if ( isType( mgTyp, "mingle:bind@v1/OpaqueAssertable" ) )
        {
            if ( ( jvObj1 instanceof MingleValue ) &&
                 ( jvObj2 instanceof MingleValue ) )
            {
                // re-entrant call to this method
                assertEqual( 
                    jvObj1, 
                    jvObj2, 
                    (AtomicTypeReference) typeRef( "mingle:core@v1/Value" )
                );
            }
            else state.equal( jvObj1, jvObj2 );
        }
        else if ( isType( mgTyp, "mingle:bind@v1/JvList1AsAtomicOpaqueVal" ) )
        {
            ModelTestInstances.assertEqual(
                OPAQUE_MG_LIST1, (MingleValue) jvObj2 );
        }
        else if ( isType( mgTyp, "mingle:bind@v1/OpaquePrimRoundtrip" ) )
        {
            // jvObj2 will already be a mingle value since it was set from
            // jvObj1; jvObj1 may or may not be a mingle primitive, but if not
            // it will be something easily and naturally converted
            ModelTestInstances.assertEqual(
                MingleModels.asMingleValue( jvObj1 ), (MingleValue) jvObj2 );
        }
        else state.equal( jvObj1, jvObj2 );
    }

    public
    static
    interface ElementAsserter
    {
        public
        void
        assertElements( Object e1,
                        Object e2,
                        MingleTypeReference t )
            throws Exception;
    }

    public
    static
    void
    assertEqualLists( Object o1,
                      Object o2,
                      ListTypeReference lt,
                      ElementAsserter a )
        throws Exception
    {
        inputs.notNull( o1, "o1" );
        inputs.notNull( o2, "o2" );
        inputs.notNull( lt, "lt" );
        inputs.notNull( a, "a" );

        List< ? > l1 = (List< ? >) o1;
        List< ? > l2 = (List< ? >) o2;

        state.equalInt( l1.size(), l2.size() );
        Iterator< ? > i1 = l1.iterator();
        Iterator< ? > i2 = l2.iterator();
        MingleTypeReference typ = lt.getElementTypeReference();

        while ( i1.hasNext() ) a.assertElements( i1.next(), i2.next(), typ );

        state.isFalse( l1.isEmpty() && ! lt.allowsEmpty() );
    }

    private
    void
    assertEqual( Object jvObj1,
                 Object jvObj2,
                 ListTypeReference lt )
        throws Exception
    {
        assertEqualLists( 
            jvObj1, 
            jvObj2, 
            lt,
            new ElementAsserter() 
            {
                public
                void
                assertElements( Object e1,
                                Object e2,
                                MingleTypeReference t )
                    throws Exception
                {
                    assertEqual( e1, e2, t );
                }
            }
        );
    }

    private
    void
    assertEqual( Object jvObj1,
                 Object jvObj2,
                 NullableTypeReference nt )
        throws Exception
    {
        if ( state.sameNullity( jvObj1, jvObj2 ) )
        {
            assertEqual( jvObj1, jvObj2, nt.getTypeReference() );
        }
    }

    private
    void
    assertEqual( Object jvObj1,
                 Object jvObj2,
                 MingleTypeReference mgTyp )
        throws Exception
    {
        if ( mgTyp instanceof AtomicTypeReference )
        {
            assertEqual( jvObj1, jvObj2, (AtomicTypeReference) mgTyp );
        }
        else if ( mgTyp instanceof ListTypeReference )
        {
            assertEqual( jvObj1, jvObj2, (ListTypeReference) mgTyp );
        }
        else if ( mgTyp instanceof NullableTypeReference )
        {
            assertEqual( jvObj1, jvObj2, (NullableTypeReference) mgTyp );
        }
        else state.fail( "Unhandled typ:", mgTyp );
    }

    private
    static
    MingleEnum
    unknownEnum()
    {
        return
            MingleEnum.create(
                (AtomicTypeReference) typeRef( "ns1@v1/En1" ),
                MingleIdentifier.create( "id1" )
            );
    }

    private
    static
    MingleStruct
    unknownStruct()
    {
        return
            MingleModels.structBuilder().setType( "ns1@v1/Struct1" ).build();
    }

    private
    final
    class RoundtripTest
    extends AbstractRoundtripTest< RoundtripTest >
    {
        private RoundtripTest() { setBinder( createTestBinder() ); }

        protected
        void
        assertJavaValues( Object jvObj1,
                          Object jvObj2,
                          MingleTypeReference mgTyp )
            throws Exception
        {
            assertEqual( jvObj1, jvObj2, mgTyp );
        }
    }

    private
    RoundtripTest
    opaque( RoundtripTest t )
    {
        return t.setMgType( "mingle:core@v1/Value" ).setUseOpaque();
    }

    @InvocationFactory
    private
    List< RoundtripTest >
    testRoundtrip()
    {
        return Lang.asList(
            
            new RoundtripTest().
                setJvObj( "hello" ).
                setMgType( "mingle:core@v1/String" ).
                setMgVal( MingleModels.asMingleString( "hello" ) ),
            
            new RoundtripTest().
                setJvObj( "hello" ).
                setMgType( "mingle:core@v1/String?" ).
                setMgVal( MingleModels.asMingleString( "hello" ) ),
            
            new RoundtripTest().
                setMgType( "mingle:core@v1/String?" ).
                setMgVal( MingleNull.getInstance() ),

            new RoundtripTest().
                setMgType( "mingle:core@v1/String" ).
                setJvPathRoot( "fluffy" ).
                setFailureClass( MingleBindingException.class ).
                setFailurePattern( "fluffy: Value is null" ),
            
            new RoundtripTest().
                setMgType( "mingle:core@v1/String" ).
                setMgPathRoot( "fld1" ).
                setFailureClass( MingleValidationException.class ).
                setFailurePattern( "Value is null" ).
                setFailureLocation( "fld1" ).
                setFromMingle(),

            new RoundtripTest().
                setJvObj( Long.valueOf( 12 ) ).
                setMgType( "mingle:core@v1/Int64" ).
                setMgVal( MingleModels.asMingleInt64( 12 ) ),
            
            new RoundtripTest().
                setMgType( "mingle:core@v1/Int64?" ).
                setMgVal( MingleNull.getInstance() ),
            
            new RoundtripTest().
                setJvObj( Integer.valueOf( 12 ) ).
                setMgType( "mingle:core@v1/Int32" ).
                setMgVal( MingleModels.asMingleInt32( 12 ) ),
            
            new RoundtripTest().
                setMgType( "mingle:core@v1/Int32?" ).
                setMgVal( MingleNull.getInstance() ),
            
            new RoundtripTest().
                setJvObj( Double.valueOf( -27.9d ) ).
                setMgType( "mingle:core@v1/Double" ).
                setMgVal( MingleModels.asMingleDouble( -27.9d ) ),
            
            new RoundtripTest().
                setMgType( "mingle:core@v1/Double?" ).
                setMgVal( MingleNull.getInstance() ),
            
            new RoundtripTest().
                setJvObj( Float.valueOf( -27.9f ) ).
                setMgType( "mingle:core@v1/Float" ).
                setMgVal( MingleModels.asMingleFloat( -27.9f ) ),
            
            new RoundtripTest().
                setMgType( "mingle:core@v1/Float?" ).
                setMgVal( MingleNull.getInstance() ),
            
            new RoundtripTest().
                setJvObj( Boolean.TRUE ).
                setMgType( "mingle:core@v1/Boolean" ).
                setMgVal( MingleBoolean.TRUE ),
            
            new RoundtripTest().
                setJvObj( ByteBuffer.wrap( new byte[] { 0x01b } ) ).
                setMgType( "mingle:core@v1/Buffer" ).
                setMgVal( MingleModels.asMingleBuffer( new byte[] { 0x01b } ) ),
            
            new RoundtripTest().
                setJvObj( ModelTestInstances.TEST_TIMESTAMP1 ).
                setMgType( "mingle:core@v1/Timestamp" ).
                setMgVal( ModelTestInstances.TEST_TIMESTAMP1 ),
            
            new RoundtripTest().
                setJvObj( MingleModels.asMingleInt64( 33 ) ).
                setMgType( "mingle:core@v1/Value" ).
                setMgVal( MingleModels.asMingleInt64( 33 ) ),

            // java-originated with MingleNull
            new RoundtripTest().
                setJvObj( MingleNull.getInstance() ).
                setMgType( "mingle:core@v1/Null" ).
                setMgVal( MingleNull.getInstance() ),

            // java-originated with implicit MingleNull (jvObj == null)
            new RoundtripTest().
                setMgType( "mingle:core@v1/Null" ).
                setMgVal( MingleNull.getInstance() ),
            
            // also test the other direction: should always go from MingleNull
            // to java null
            new RoundtripTest().
                setMgType( "mingle:core@v1/Null" ).
                setMgVal( MingleNull.getInstance() ).
                setFromMingle(),

            new RoundtripTest().
                setJvObj( Struct1.INST1 ).
                setMgType( Struct1.TYPE ).
                setMgVal( Struct1.MG_INST1 ),
            
            new RoundtripTest().
                setJvObj( Lang.< String >asList( "s1", "s2", "s3" ) ).
                setMgType( "mingle:core@v1/String*" ).
                setMgVal( STRING_LIST1 ),
            
            new RoundtripTest().
                setJvObj( Lang.< String >asList() ).
                setMgType( "mingle:core@v1/String*" ).
                setMgVal( EMPTY_LIST ),
            
            new RoundtripTest().
                setJvObj( Lang.< String >asList( "s1", "s2", "s3" ) ).
                setMgType( "mingle:core@v1/String+" ).
                setMgVal( STRING_LIST1 ),
            
            new RoundtripTest().
                setJvObj( Lang.asList( Struct1.INST1, Struct1.INST1 ) ).
                setMgType( Struct1.TYPE + "*" ).
                setMgVal( 
                    MingleList.create( Struct1.MG_INST1, Struct1.MG_INST1 ) ),
            
            new RoundtripTest().
                setJvObj( Lang.< String >asList() ).
                setMgType( "mingle:core@v1/String+" ).
                setMgVal( EMPTY_LIST ).
                setJvPathRoot( "f1" ).
                setFailureClass( MingleBindingException.class ).
                setFailurePattern( "f1: List is empty" ),
            
            new RoundtripTest().
                setJvObj( Lang.< String >asList() ).
                setMgType( "mingle:core@v1/String+" ).
                setMgVal( EMPTY_LIST ).
                setMgPathRoot( "f1" ).
                setFailureClass( MingleValidationException.class ).
                setFailurePattern( "list is empty" ).
                setFailureLocation( "f1" ).
                setFromMingle(),
            
            new RoundtripTest().
                setJvObj( Lang.< String >asList( "s1", null, "s3" ) ).
                setJvPathRoot( "f1" ).
                setMgType( "mingle:core@v1/String*" ).
                setMgVal( STRING_LIST2 ).
                setFailureClass( MingleBindingException.class ).
                setFailurePattern( "\\Qf1[ 1 ]: Value is null\\E" ),
            
            new RoundtripTest().
                setJvObj( Lang.< String >asList( "s1", null, "s3" ) ).
                setMgType( "mingle:core@v1/String*" ).
                setMgVal( STRING_LIST2 ).
                setMgPathRoot( "f1" ).
                setFailureClass( MingleValidationException.class ).
                setFailurePattern( "Value is null" ).
                setFailureLocation( "f1[ 1 ]" ).
                setFromMingle(),
            
            new RoundtripTest().
                setJvObj( Lang.< String >asList( "s1", null, "s3" ) ).
                setMgType( "mingle:core@v1/String?*" ).
                setMgVal( STRING_LIST2 ),
            
            // just testing inbound cast exceptions and caller-facing messages;
            // letting normal java casts and stack traces do the work for
            // java-side cast problems
            new RoundtripTest().
                setMgVal( MingleModels.asMingleString( "string-for-struct" ) ).
                setMgType( Struct1.TYPE ).
                setMgPathRoot( "f1" ).
                setFailureClass( MingleTypeCastException.class ).
                setFailurePattern( 
                    "Expected mingle value of type mingle:core@v1/Struct but " +
                    "found mingle:core@v1/String" ).
                setFailureLocation( "f1" ).
                setFromMingle(),
            
            new RoundtripTest().
                setMgVal(
                    MingleEnum.create(
                        Enum1.TYPE, MingleIdentifier.create( "en-val1" ) )
                ).
                setMgType( Enum1.TYPE ).
                setJvObj( Enum1.EN_VAL1 ),
            
            new RoundtripTest().
                setLabelPrefix( "Builder" ).
                setJvObj( STD_EXCEPTION_INST1 ).
                setMgType( "mingle:core@v1/StandardException" ).
                setMgVal( STD_EXCEPTION_MG_INST1 ),
            
            new RoundtripTest().
                setLabelPrefix( "FactMethod" ).
                setJvObj( 
                    StandardException.create( STD_EXCEPTION_INST1.getMessage() )
                ).
                setMgType( "mingle:core@v1/StandardException" ).
                setMgVal( STD_EXCEPTION_MG_INST1 ),
            
            new RoundtripTest().
                setLabelPrefix( "Builder" ).
                setJvObj( new StandardException.Builder().build() ).
                setMgType( "mingle:core@v1/StandardException" ).
                setMgVal(
                    MingleModels.exceptionBuilder().
                        setType( "mingle:core@v1/StandardException" ).
                        build()
                ),
            
            new RoundtripTest().
                setLabelPrefix( "FactMethod" ).
                setJvObj( StandardException.create( null ) ).
                setMgType( "mingle:core@v1/StandardException" ).
                setMgVal(
                    MingleModels.exceptionBuilder().
                        setType( "mingle:core@v1/StandardException" ).
                        build()
                ),
            
            new RoundtripTest().
                setJvObj( new MingleValidationException( "test", PATH1 ) ).
                setMgType( "mingle:core@v1/ValidationException" ).
                setMgVal(
                    MingleModels.exceptionBuilder().
                        setType( "mingle:core@v1/ValidationException" ).
                        f().setString( "message", "test" ).
                        f().setString( "location", format( PATH1 ) ).
                        build()
                ),

            opaque( new RoundtripTest() ).
                setJvObj( "hello" ).
                setAssertType( "mingle:bind@v1/OpaquePrimRoundtrip" ).
                setMgVal( MingleModels.asMingleString( "hello" ) ),
            
            opaque( new RoundtripTest() ).
                setJvObj( Struct1.INST1 ).
                setAssertType( Struct1.TYPE ).
                setMgVal( Struct1.MG_INST1 ),
            
            opaque( new RoundtripTest() ).
                setJvObj( unknownStruct() ).
                setMgVal( unknownStruct() ),
            
            opaque( new RoundtripTest() ).
                setJvObj( Enum1.EN_VAL1 ).
                setAssertType( Enum1.TYPE ).
                setMgVal( Enum1.MG_EN_VAL1 ),

            opaque( new RoundtripTest() ).
                setJvObj( unknownEnum() ).
                setMgVal( unknownEnum() ),
            
            opaque( new RoundtripTest() ).
                setJvObj( MingleModels.getEmptySymbolMap() ).
                setMgVal( MingleModels.getEmptySymbolMap() ),
            
            opaque( new RoundtripTest() ).
                setJvObj( MingleModels.asMingleString( "hello" ) ).
                setMgVal( MingleModels.asMingleString( "hello" ) ),
 
            // Typed as a list of possibly null objects (which we treat
            // opaquely)
            new RoundtripTest().
                setMgType( "mingle:core@v1/Value?*" ).
                setAssertType( typeRef( "mingle:bind@v1/OpaqueAssertable*" ) ).
                setUseOpaque().
                setMgVal( OPAQUE_MG_LIST1 ).
                setJvObj( OPAQUE_JV_LIST1 ),
 
            // Typed as just an opaque value which happens to be a list of other
            // things we can treat opaquely
            new RoundtripTest().
                setMgType( "mingle:core@v1/Value" ).
                setAssertType( 
                    typeRef( "mingle:bind@v1/JvList1AsAtomicOpaqueVal" ) ).
                setUseOpaque().
                setMgVal( OPAQUE_MG_LIST1 ).
                setJvObj( OPAQUE_JV_LIST1 )
        );
    }

    private
    abstract
    class AbstractLabeledTest
    extends LabeledTestCall
    implements TestFailureExpector
    {
        final MingleBinder mb = createTestBinder();
        
        CharSequence failExpct;

        private AbstractLabeledTest( CharSequence lbl ) { super( lbl ); }

        public CharSequence expectedFailurePattern() { return failExpct; }
    }

    // This is fundamentally a test of MingleBinding.asMingleValue(), but
    // exposed via MingleBinders.setField so as to get extra coverage of that
    // path on top of just MingleBinding.asMingleValue()
    private
    abstract
    class SetFieldTest
    extends AbstractLabeledTest
    {
        final MingleIdentifier mgFldId = MingleIdentifier.create( "mg-fld" );
        final String jvFldId = "jvFld";
        final MingleSymbolMapBuilder b = MingleModels.symbolMapBuilder();
        final ObjectPath< String > path = ObjectPath.getRoot();

        FieldDefinition fd;
        Object jvObj;
        MingleValue mgVal;

        private SetFieldTest( CharSequence lbl ) { super( lbl ); }

        final
        FieldDefinition
        fieldDef( MingleTypeReference typ )
        {
            return 
                new FieldDefinition.Builder().
                    setType( typ ).
                    setName( mgFldId ).
                    build();
        }

        public
        final
        Class< ? extends Throwable >
        expectedFailureClass()
        {
            return failExpct == null ? null : MingleBindingException.class; 
        }

        @Override
        protected
        final
        void
        call()
            throws Exception
        {
            state.notNull( fd, "fd" );

            MingleBinders.
                setField( fd, jvFldId, jvObj, b, mb, path, false );

            MingleSymbolMap m = b.build();

            ModelTestInstances.assertEqual( mgVal, m.get( mgFldId ) );
        }
    }

    // Some basic coverage tests of MingleBinders.setField
    @InvocationFactory
    private
    List< SetFieldTest >
    testSetField()
    {
        return Lang.< SetFieldTest >asList( 
            
            new SetFieldTest( "set-string-ok" ) {{
                fd = fieldDef( typeRef( "mingle:core@v1/String" ) );
                mgVal = MingleModels.asMingleString( "hello" );
                jvObj = "hello";
            }},

            new SetFieldTest( "set-null-string-fail" ) {{
                fd = fieldDef( typeRef( "mingle:core@v1/String" ) );
                failExpct = "jvFld: Value is null";
            }},

            new SetFieldTest( "null-ignored-when-default-present" ) 
            {{
                fd = 
                    new FieldDefinition.Builder().
                        setName( mgFldId ).
                        setType( typeRef( "mingle:core@v1/String" ) ).
                        setDefault( MingleModels.asMingleString( "stuff" ) ).
                        build();
                
                mgVal = MingleNull.getInstance();
                jvObj = null;
            }},

            new SetFieldTest( "set-enum-val" ) 
            {{
                fd = fieldDef( Enum1.TYPE );

                mgVal = 
                    MingleEnum.create( 
                        Enum1.TYPE, MingleIdentifier.create( "en-val1" ) );
                
                jvObj = Enum1.EN_VAL1;
            }}
        );
    }

    private
    abstract
    class AsJavaValueTest
    extends AbstractLabeledTest
    {
        final MingleIdentifier fld = MingleIdentifier.create( "mgFld" );
        final ObjectPath< MingleIdentifier > path = ObjectPath.getRoot();

        MingleTypeReference mgTyp;
        MingleTypeReference assertTyp; // null; defaults to mgTyp
        MingleValue mgVal;
        Object jvObj;

        private AsJavaValueTest( CharSequence lbl ) { super( lbl ); }

        public
        Class< ? extends Throwable >
        expectedFailureClass()
        {
            return failExpct == null ? null : MingleValidationException.class;
        }

        @Override
        public
        CharSequence
        expectedFailurePattern()
        {
            return 
                fld.getExternalForm() + ": " + super.expectedFailurePattern();
        }

        private
        FieldDefinition
        createFieldDefinition()
        {
            return
                new FieldDefinition.Builder().
                    setName( fld ).
                    setType( mgTyp ).
                    build();
        }

        Object
        asJavaValue( FieldDefinition fd,
                     MingleSymbolMap m,
                     MingleBinder mb,
                     ObjectPath< MingleIdentifier > path )
        {
            return MingleBinders.asJavaValue( fd, m, mb, path );
        }

        @Override
        protected
        void
        call()
            throws Exception
        {
            MingleSymbolMapBuilder b = MingleModels.symbolMapBuilder();
            if ( mgVal != null ) b.set( fld, mgVal );
            MingleSymbolMap m = b.build();

            FieldDefinition fd = createFieldDefinition();
            Object jvObj2;

            jvObj2 = asJavaValue( fd, m, mb, path );

            assertEqual( jvObj, jvObj2, assertTyp == null ? mgTyp : assertTyp );
        }
    }

    @InvocationFactory
    private
    List< AsJavaValueTest >
    testAsJavaValue()
    {
        return Lang.< AsJavaValueTest >asList(
            
            new AsJavaValueTest( "string-ok" ) {{
                mgTyp = typeRef( "mingle:core@v1/String" );
                mgVal = MingleModels.asMingleString( "hello" );
                jvObj = "hello";
            }},

            new AsJavaValueTest( "mingle-null-string-fail" ) {{
                mgTyp = typeRef( "mingle:core@v1/String" );
                mgVal = MingleNull.getInstance();
                failExpct = "Value is null";
            }},

            new AsJavaValueTest( "missing-string-fail" ) {{
                mgTyp = typeRef( "mingle:core@v1/String" );
                failExpct = "Value is null";
            }},

            new AsJavaValueTest( "invalid-string-fail" ) {{
                mgTyp = typeRef( "mingle:core@v1/String~\"a+\"" );
                mgVal = MingleModels.asMingleString( "bbb" );
                failExpct = "\\QValue does not match \"a+\": \"bbb\"\\E";
            }},

            new AsJavaValueTest( "enum-from-identifier-lc-hyphen" ) {{
                mgTyp = Enum1.TYPE;
                mgVal = MingleModels.asMingleString( "en-val1" );
                jvObj = Enum1.EN_VAL1;
            }},

            new AsJavaValueTest( "enum-from-identifier-lc-camel-capped" ) {{
                mgTyp = Enum1.TYPE;
                mgVal = MingleModels.asMingleString( "enVal1" );
                jvObj = Enum1.EN_VAL1;
            }},

            new AsJavaValueTest( "enum-from-identifier-lc-underscore" ) {{
                mgTyp = Enum1.TYPE;
                mgVal = MingleModels.asMingleString( "en_val1" );
                jvObj = Enum1.EN_VAL1;
            }},

            new AsJavaValueTest( "enum-from-identifier-missing" ) {{
                mgTyp = Enum1.TYPE;
                mgVal = MingleModels.asMingleString( "not-a-val" );
                failExpct = "no such enum constant";
            }},

            new AsJavaValueTest( "enum-as-string-in-enum-list" )
            {{
                mgTyp = ListTypeReference.create( Enum1.TYPE, false );

                mgVal =
                    MingleList.create(
                        MingleModels.asMingleString( "en-val1" ),
                        MingleModels.asMingleString( "en-val2" )
                    );
                
                jvObj = Lang.asList( Enum1.EN_VAL1, Enum1.EN_VAL2 );
            }},

            new AsJavaValueTest( "expanded-enum" ) 
            {{
                mgTyp = Enum1.TYPE;

                mgVal = 
                    MingleEnum.create(
                        Enum1.TYPE, MingleIdentifier.create( "en-val1" ) );

                jvObj = Enum1.EN_VAL1;
            }},

            new AsJavaValueTest( "expanded-enum-missing" ) 
            {{
                mgTyp = Enum1.TYPE;

                mgVal = 
                    MingleEnum.create(
                        Enum1.TYPE, MingleIdentifier.create( "not-a-val" ) );
                
                failExpct = "no such enum constant";
            }},

            new AsJavaValueTest( "invalid-mg-enum-val-type" )
            {{
                mgTyp = Enum1.TYPE;
                mgVal = MingleModels.asMingleBuffer( new byte[] {} );
                failExpct = "unrecognized enum value";
            }}
        );
    }

    private
    class AsJavaValueFromClassTest
    extends LabeledTestCall
    {
        Object jvObj;
        Class< ? > jvCls;
        MingleValue mgVal;

        private
        AsJavaValueFromClassTest( CharSequence lbl )
        {
            super( lbl );
        }

        protected
        void
        call()
            throws Exception
        {
            Object jvObj2 = 
                MingleBinders.asJavaValue( createTestBinder(), jvCls, mgVal );

            state.equal( jvObj, jvObj2 );
        }
    }

    @InvocationFactory
    private
    List< AsJavaValueFromClassTest >
    testAsJavaValueFromClass()
    {
        return Lang.< AsJavaValueFromClassTest >asList(
            
            new AsJavaValueFromClassTest( "success" )
            {{
                jvObj = Struct1.INST1;
                jvCls = Struct1.class;
                mgVal = Struct1.MG_INST1;
            }},

            new AsJavaValueFromClassTest( "fail-no-type" )
            {{
                jvCls = MingleBindTests.class;
                mgVal = Struct1.MG_INST1; // doesn't really matter

                expectFailure( 
                    IllegalStateException.class, 
                    "No binding keyed for class " +
                    "com.bitgirder.mingle.bind.MingleBindTests"
                );
            }},

            new AsJavaValueFromClassTest( "fail-inbound-type-cast" )
            {{
                jvCls = Struct1.class;
                mgVal = Enum1.MG_EN_VAL1;

                expectFailure( 
                    MingleTypeCastException.class, 
                    "\\QExpected mingle value of type mingle:core@v1/Struct " +
                    "but found mingle:bind:test@v1/Enum1\\E"
                );
            }}
        );
    }

    @Test
    private
    void
    testAsMingleValueFromInstanceSuccess()
        throws Exception
    {
        ModelTestInstances.assertEqual(
            Struct1.MG_INST1,
            MingleBinders.asMingleValue( createTestBinder(), Struct1.INST1 )
        );
    }

    @Test( expected = IllegalStateException.class,
           expectedPattern = 
            "\\QNo binding keyed for class java.lang.Object\\E" )
    private
    void
    testAsMingleValueUnboundInstanceFails()
        throws Exception
    {
        MingleBinders.asMingleValue( createTestBinder(), new Object() );
    }

    private
    static
    TypeDefinitionLookup
    loadBootstrapTypeDefs( MingleCodecFactory codecFact )
        throws Exception
    {
        ByteBuffer bb =
            IoUtils.toByteBuffer( 
                IoUtils.expectSingleResourceAsStream( "mingle/lib/out.mgo" ) );

        MingleStruct ms =
            MingleCodecs.fromByteBuffer(
                MingleCodecs.detectCodec( codecFact, bb ), 
                bb, 
                MingleStruct.class 
            );
        
        ObjectPath< MingleIdentifier > path = ObjectPath.getRoot();

        return TypeDefinitions.asTypeDefinitionLookup( ms, path );
    }

    private
    static
    void
    completeInitBootstrapTypes( Testing.RuntimeInitializerContext ctx,
                                final MingleCodecFactory codecFact )
    { 
        Testing.submitInitTask(
            ctx,
            new Testing.AbstractInitTask( ctx ) {
                protected void runImpl() throws Exception
                {
                    context().setObject(
                        KEY_BOOTSTRAP_TYPE_DEF_LOOKUP,
                        loadBootstrapTypeDefs( codecFact )
                    );

                    context().complete();
                }
            }
        );
    }

    @Testing.RuntimeInitializer
    private
    static
    void
    initBootstrapTypes( final Testing.RuntimeInitializerContext ctx )
    {
        Testing.awaitTestObject(
            ctx,
            MingleCodecTests.KEY_DEFAULT_CODEC_FACTORY,
            MingleCodecFactory.class,
            new ObjectReceiver< MingleCodecFactory >() {
                public void receive( MingleCodecFactory codecFact ) {
                    completeInitBootstrapTypes( ctx, codecFact );
                }
            }
        );
    }

    static
    {
        STD_EXCEPTION_INST1 =
            new StandardException.Builder().
                setMessage( "test-message" ).
                build();
        
        STD_EXCEPTION_MG_INST1 =
            MingleModels.exceptionBuilder().
                setType( "mingle:core@v1/StandardException" ).
                f().setString( "message", "test-message" ).
                build();
    }

    // Needs to come after init of things like Struct1.INST1
    static
    {
        OPAQUE_MG_LIST1 =
            new MingleList.Builder().
                add( Struct1.MG_INST1 ).
                add( unknownStruct() ).
                add( Enum1.MG_EN_VAL1 ).
                add( unknownEnum() ).
                add( MingleModels.asMingleString( "hello" ) ).
                add( MingleNull.getInstance() ).
                build();
        
        OPAQUE_JV_LIST1 =
            Lang.asList(
                Struct1.INST1,
                unknownStruct(),
                Enum1.EN_VAL1,
                unknownEnum(),
                MingleModels.asMingleString( "hello" ),
                null
            );
    }
}
