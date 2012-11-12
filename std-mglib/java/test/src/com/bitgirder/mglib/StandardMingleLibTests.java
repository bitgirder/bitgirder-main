package com.bitgirder.mglib;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.lang.path.ObjectPath;

import com.bitgirder.mglib.v1.NativeGenHolder;
import com.bitgirder.mglib.v1.TimeUnitImpl;

import com.bitgirder.io.DataSize;
import com.bitgirder.io.DataUnit;

import com.bitgirder.concurrent.Duration;

import com.bitgirder.mingle.model.MingleTypeReference;
import com.bitgirder.mingle.model.MingleTypeCastException;
import com.bitgirder.mingle.model.AtomicTypeReference;
import com.bitgirder.mingle.model.QualifiedTypeName;
import com.bitgirder.mingle.model.MingleNamespace;
import com.bitgirder.mingle.model.MingleIdentifiedName;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleValue;
import com.bitgirder.mingle.model.MingleEnum;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleStructBuilder;
import com.bitgirder.mingle.model.MingleValidationException;

import com.bitgirder.mingle.bind.MingleBinder;
import com.bitgirder.mingle.bind.MingleBinders;
import com.bitgirder.mingle.bind.MingleBindTests;
import com.bitgirder.mingle.bind.AbstractRoundtripTest;

import com.bitgirder.test.Test;
import com.bitgirder.test.TestFailureExpector;
import com.bitgirder.test.TestRuntime;
import com.bitgirder.test.InvocationFactory;
import com.bitgirder.test.LabeledTestCall;

import java.util.List;

// Testing for DataSize is more exhaustive than that for Duration since we use
// DataSize as our proxy for testing various mechanisms in
// StandardMingleLib.AbstractMetricBinding. If all DataSize tests pass, we
// currently assume that Duration need only test basic coverage of the duration
// binding impl and anything particular to its implementation, but not of the
// overall movement of AbstractMetricBinding
@Test
final
class StandardMingleLibTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static AtomicTypeReference TYPE_DATA_SIZE =
        (AtomicTypeReference) 
            MingleTypeReference.create( "bitgirder:io@v1/DataSize" );

    private final static AtomicTypeReference TYPE_DATA_UNIT =
        (AtomicTypeReference)
            MingleTypeReference.create( "bitgirder:io@v1/DataUnit" );

    private final static AtomicTypeReference TYPE_TIME_UNIT =
        (AtomicTypeReference)
            MingleTypeReference.create( "bitgirder:concurrent@v1/TimeUnit" );

    private final static AtomicTypeReference TYPE_DURATION =
        (AtomicTypeReference)
            MingleTypeReference.create( "bitgirder:concurrent@v1/Duration" );

    private final static AtomicTypeReference TYPE_NG_HOLDER =
        (AtomicTypeReference)
            MingleTypeReference.create( "bitgirder:mglib@v1/NativeGenHolder" );
    
    private final static NativeGenHolder NG_JV_INST1;
    private final static MingleValue NG_MG_INST1;

    private final MingleBinder mb;

    private
    StandardMingleLibTests( TestRuntime rt )
    {
        this.mb = MgLibTesting.expectDefaultBinder( rt );
    }

    private
    void
    assertEqualEnum( Enum< ? > e1,
                     Enum< ? > e2 )
    {
        if ( state.sameNullity( e1, e2 ) ) state.equal( e1, e2 );
    } 

    private
    void
    assertEqual( DataSize s1,
                 DataSize s2 )
    {
        if ( state.sameNullity( s1, s2 ) )
        {
            state.equal( s1.unit(), s2.unit() );
            state.equal( s1.size(), s2.size() );
        }
    }

    // because some units (FORTNIGHT, DAYS) won't necessarily roundtrip as such,
    // we just do our test on the millis value
    private
    void
    assertEqual( Duration d1,
                 Duration d2 )
    {
        if ( state.sameNullity( d1, d2 ) )
        {
            state.equal( d1.asMillis(), d2.asMillis() );
        }
    }

    private
    void
    assertEqual( NativeGenHolder h1,
                 NativeGenHolder h2 )
    {
        if ( state.sameNullity( h1, h2 ) )
        {
            assertEqualEnum( h1.dataUnit(), h2.dataUnit() );
            assertEqual( h1.dataSize(), h2.dataSize() );
            assertEqual( h1.duration(), h2.duration() );
            state.equal( h1.mgIdent(), h2.mgIdent() );
            state.equal( h1.mgNs(), h2.mgNs() );
            state.equal( h1.mgQname(), h2.mgQname() );
            state.equal( h1.mgTypeRef(), h2.mgTypeRef() );
            state.equal( h1.mgIdentifiedName(), h2.mgIdentifiedName() );
        }
    }

    private
    void
    assertAtomic( Object jvObj1,
                  Object jvObj2,
                  AtomicTypeReference typ )
    {
        if ( typ.equals( TYPE_DATA_SIZE ) )
        {
            assertEqual( (DataSize) jvObj1, (DataSize) jvObj2 );
        }
        else if ( typ.equals( TYPE_DATA_UNIT ) )
        {
            assertEqualEnum( (DataUnit) jvObj1, (DataUnit) jvObj2 );
        }
        else if ( typ.equals( TYPE_DURATION ) )
        {
            assertEqual( (Duration) jvObj1, (Duration) jvObj2 );
        }
        else if ( typ.equals( TYPE_NG_HOLDER ) )
        {
            assertEqual( (NativeGenHolder) jvObj1, (NativeGenHolder) jvObj2 );
        }
        else state.fail( "Unhandled type:", typ );
    }

    private
    final
    class RoundtripTest
    extends AbstractRoundtripTest< RoundtripTest >
    {
        private final CharSequence lbl;

        private
        RoundtripTest( CharSequence lbl )
        {
            this.lbl = lbl;

            setBinder( mb );
        }

        @Override public final CharSequence getLabel() { return lbl; }

        protected
        void
        assertJavaValues( Object jvObj1,
                          Object jvObj2,
                          MingleTypeReference typ )
        {
            assertAtomic( jvObj1, jvObj2, (AtomicTypeReference) typ );
        }
    }

    private
    static
    MingleValue
    mgEnum( CharSequence id,
            AtomicTypeReference enTyp )
    {
        if ( enTyp == null ) return MingleModels.asMingleString( id );
        else
        {
            return 
                new MingleEnum.Builder().
                    setType( enTyp ).
                    setValue( id ).
                    build();
        }
    }

    private
    static
    MingleValue
    mgEnum( Enum< ? > e,
            AtomicTypeReference enTyp )
    {
        return mgEnum( e.name().toLowerCase(), enTyp );
    }

    private
    static
    MingleValue
    mgMetricType( Long sz,
                  Enum< ? > unit,
                  AtomicTypeReference typ,
                  AtomicTypeReference unitTyp,
                  String unitKey,
                  boolean expandEnum )
    {
        MingleStructBuilder b = MingleModels.structBuilder().setType( typ );
        
        if ( sz != null ) b.f().setInt64( unitKey, sz );

        if ( unit != null )
        {
            b.f().set( "unit", mgEnum( unit, expandEnum ? unitTyp : null ) );
        }

        return b.build();
    }

    private
    static
    MingleValue
    mgDataSize( Long sz,
                DataUnit u,
                boolean expandEnum )
    {
        return
            mgMetricType( 
                sz, u, TYPE_DATA_SIZE, TYPE_DATA_UNIT, "size", expandEnum );
    }

    private
    static
    MingleValue
    mgDuration( Long sz,
                TimeUnitImpl u,
                boolean expandEnum )
    {
        return
            mgMetricType(
                sz, u, TYPE_DURATION, TYPE_TIME_UNIT, "duration", expandEnum );
    }

    @InvocationFactory
    private
    List< RoundtripTest >
    testRoundtrip()
    {
        return Lang.asList(
 
            new RoundtripTest( "dataunit" ).
                setMgType( TYPE_DATA_UNIT ).
                setJvObj( DataUnit.MEGABYTE ).
                setMgVal( mgEnum( DataUnit.MEGABYTE, TYPE_DATA_UNIT ) ),

            new RoundtripTest( "datasize" ).
                setMgType( TYPE_DATA_SIZE ).
                setJvObj( DataSize.ofBytes( 1023 ) ).
                setMgVal( mgDataSize( 1023L, DataUnit.BYTE, true ) ),

            new RoundtripTest( "duration-with-jcur-timeunit" ).
                setMgType( TYPE_DURATION ).
                setJvObj( Duration.fromSeconds( 12 ) ).
                setMgVal( mgDuration( 12L, TimeUnitImpl.SECOND, true ) ),
            
            new RoundtripTest( "duration-with-synth-timeunit" ).
                setMgType( TYPE_DURATION ).
                setJvObj( Duration.fromDays( 3 ) ).
                setMgVal( mgDuration( 259200L, TimeUnitImpl.SECOND, true ) ),
            
            new RoundtripTest( "duration-with-fortnight" ).
                setMgType( TYPE_DURATION ).
                setJvObj( Duration.fromFortnights( 12 ) ).
                setMgVal( mgDuration( 14515200L, TimeUnitImpl.SECOND, true ) ),

            new RoundtripTest( "native-gen-holder" ).
                setMgType( "bitgirder:mglib@v1/NativeGenHolder" ).
                setJvObj( NG_JV_INST1 ).
                setMgVal( NG_MG_INST1 )
        );
    }

    private
    class AsJavaValueTest
    extends LabeledTestCall
    implements TestFailureExpector
    {
        Object jvObj;
        MingleValue mgVal;
        AtomicTypeReference mgTyp;
        CharSequence errPat;

        // default is validation exception, but only used when errPat != null
        Class< ? extends Throwable > errCls = MingleValidationException.class;

        private AsJavaValueTest( CharSequence lbl ) { super( lbl ); }

        public
        Class< ? extends Throwable >
        expectedFailureClass()
        {
            return errPat == null ? null : errCls;
        }

        public CharSequence expectedFailurePattern() { return errPat; }

        @Override
        protected
        final
        void
        call()
        {
            if ( errPat == null ) state.notNull( jvObj, "jvObj" );
            state.notNull( mgVal, "mgVal" );
            state.notNull( mgTyp, "mgTyp" );

            Object jvObj2 = MingleBinders.asJavaValue( mb, mgTyp, mgVal );
            assertAtomic( jvObj, jvObj2, mgTyp );
        }
    }

    @InvocationFactory
    private
    List< AsJavaValueTest >
    testAsJavaValue()
    {
        return Lang.< AsJavaValueTest >asList(
            
            new AsJavaValueTest( "dunit-from-string" )
            {{
                mgTyp = TYPE_DATA_UNIT;
                jvObj = DataUnit.KILOBYTE;
                mgVal = MingleModels.asMingleString( "kilobyte" );
            }},

            new AsJavaValueTest( "dunit-invalid-expanded-unit" )
            {{
                mgTyp = TYPE_DATA_UNIT;
                mgVal = mgEnum( "blah", TYPE_DATA_UNIT );
                errPat = "Invalid enum value: blah";
            }},

            new AsJavaValueTest( "dunit-invalid-string-unit" )
            {{
                mgTyp = TYPE_DATA_UNIT;
                mgVal = MingleModels.asMingleString( "blah" );
                errPat = "Invalid enum value: blah";
            }},

            new AsJavaValueTest( "dsz-from-string-no-unit" )
            {{
                jvObj = DataSize.ofBytes( 12 );
                mgVal = MingleModels.asMingleString( "12" );
                mgTyp = TYPE_DATA_SIZE;
            }},

            new AsJavaValueTest( "dsz-from-string-with-unit" )
            {{
                jvObj = DataSize.ofKilobytes( 12 );
                mgVal = MingleModels.asMingleString( "12k" );
                mgTyp = TYPE_DATA_SIZE;
            }},

            new AsJavaValueTest( "dsz-invalid-unit-abbrev" )
            {{
                mgTyp = TYPE_DATA_SIZE;
                mgVal = MingleModels.asMingleString( "12microns" );

                errPat = 
                    "\\QInvalid string for instance of type " +
                    "bitgirder:io@v1/DataSize: 12microns\\E";
            }},

            new AsJavaValueTest( "dsz-unexpanded-unit" )
            {{
                mgTyp = TYPE_DATA_SIZE;
                jvObj = DataSize.ofKilobytes( 12 );
                mgVal = mgDataSize( 12L, DataUnit.KILOBYTE, false );
            }},

            new AsJavaValueTest( "dsz-fail-no-unit" )
            {{
                mgTyp = TYPE_DATA_SIZE;
                mgVal = mgDataSize( 123L, null, true );
                errPat = "unit: value is null";
            }},

            new AsJavaValueTest( "dsz-fail-no-size" )
            {{
                mgTyp = TYPE_DATA_SIZE;
                mgVal = mgDataSize( null, DataUnit.KILOBYTE, true );
                errPat = "size: value is null";
            }},

            new AsJavaValueTest( "dsz-negative-size" )
            {{
                mgTyp = TYPE_DATA_SIZE;
                mgVal = mgDataSize( -12L, DataUnit.KILOBYTE, true );
                errPat = "size: Value is negative: -12";
            }},

            new AsJavaValueTest( "dsz-invalid-explicit-unit" )
            {{
                mgTyp = TYPE_DATA_SIZE;

                mgVal = 
                    MingleModels.structBuilder().
                        setType( TYPE_DATA_SIZE ).
                        f().setInt64( "size", 12 ).
                        f().set( "unit", mgEnum( "blah", TYPE_DATA_UNIT ) ).
                        build();
                
                errPat = "unit: Invalid enum value: blah";
            }},

            new AsJavaValueTest( "dsz-invalid-unexpanded-unit" )
            {{
                mgTyp = TYPE_DATA_SIZE;

                mgVal =
                    MingleModels.structBuilder().
                        setType( TYPE_DATA_SIZE ).
                        f().setInt64( "size", 12 ).
                        f().setString( "unit", "blah" ).
                        build();
                    
                errPat = "unit: Invalid enum value: blah";
            }},

            new AsJavaValueTest( "dsz-invalid-mingle-type" )
            {{
                mgTyp = TYPE_DATA_SIZE;
                mgVal = MingleModels.asMingleBuffer( new byte[] {} );

                errCls = MingleTypeCastException.class;

                errPat = 
                    "\\QExpected mingle value of type " +
                    "bitgirder:io@v1/DataSize " +
                    "but found mingle:core@v1/Buffer\\E";
            }},

            new AsJavaValueTest( "dur-from-millis" )
            {{
                mgTyp = TYPE_DURATION;
                mgVal = MingleModels.asMingleInt64( 12 );
                jvObj = Duration.fromMillis( 12 );
            }},

            new AsJavaValueTest( "dur-from-jcur-timestring" )
            {{
                mgTyp = TYPE_DURATION;
                mgVal = MingleModels.asMingleString( "12ms" );
                jvObj = Duration.fromMillis( 12 );
            }},

            new AsJavaValueTest( "dur-from-num-as-mgstr" )
            {{
                mgTyp = TYPE_DURATION;
                mgVal = MingleModels.asMingleString( "12" );
                jvObj = Duration.fromMillis( 12 );
            }},

            new AsJavaValueTest( "dur-from-synth-timeunit" )
            {{
                mgTyp = TYPE_DURATION;
                mgVal = MingleModels.asMingleString( "343d" );
                jvObj = Duration.fromDays( 343 );
            }},

            new AsJavaValueTest( "dur-from-fortnight" )
            {{
                mgTyp = TYPE_DURATION;
                mgVal = MingleModels.asMingleString( "3fortnights" );
                jvObj = Duration.fromFortnights( 3 );
            }},

            new AsJavaValueTest( "dur-bad-string-form" )
            {{
                mgTyp = TYPE_DURATION;
                mgVal = MingleModels.asMingleString( "23b'glock" );

                errPat = 
                    "\\QInvalid string for instance of type " +
                    "bitgirder:concurrent@v1/Duration: 23b'glock\\E";
            }}
        );
    }

    private
    MingleValue
    castNum( MingleTypeReference castTo,
             int num )
    {
        return
            MingleModels.asMingleInstance(
                castTo, 
                MingleModels.asMingleInt32( num ),
                ObjectPath.< MingleIdentifier >getRoot()
            );
    }

    private
    final
    class FromNumTest
    extends AsJavaValueTest
    {
        private
        FromNumTest( String lblPrefix,
                     MingleValue mgVal,
                     AtomicTypeReference mgTyp,
                     Object jvObj )
        {
            super( lblPrefix + mgVal.getClass().getSimpleName() );

            this.mgTyp = mgTyp;
            this.mgVal = mgVal;
            this.jvObj = jvObj;
        }
    }

    @InvocationFactory
    private
    List< AsJavaValueTest >
    testAsJavaValueFromNum()
    {
        List< AsJavaValueTest > res = Lang.newList();

        for ( AtomicTypeReference typ :
                new AtomicTypeReference[] {
                    MingleModels.TYPE_REF_MINGLE_INT64,
                    MingleModels.TYPE_REF_MINGLE_INT32,
                    MingleModels.TYPE_REF_MINGLE_DOUBLE,
                    MingleModels.TYPE_REF_MINGLE_FLOAT } )
        {
            res.add( 
                new FromNumTest( 
                    "dsz-from-", 
                    castNum( typ, 1234 ),
                    TYPE_DATA_SIZE, 
                    DataSize.ofBytes( 1234 ) 
                ) 
            );
        }

        return res;
    }

    static
    {
        NG_JV_INST1 =
            new NativeGenHolder.Builder().
                setDataUnit( DataUnit.BYTE ).
                setDataSize( DataSize.ofKilobytes( 12 ) ).
                setDuration( Duration.fromSeconds( 12 ) ).
                setMgIdent( MingleIdentifier.create( "test-ident" ) ).
                setMgNs( MingleNamespace.create( "ns1:ns2@v1" ) ).
                setMgQname( QualifiedTypeName.create( "ns1:ns2@v1/T1" ) ).
                setMgTypeRef( 
                    MingleTypeReference.create( "ns1:ns2@v1/T1/T2*+?" ) ).
                setMgIdentifiedName( 
                    MingleIdentifiedName.create( "ns1:ns2@v1/a/b/c" ) ).
                build();
        
        NG_MG_INST1 =
            MingleModels.structBuilder().
                setType( "bitgirder:mglib@v1/NativeGenHolder" ).
                f().set( "dataUnit", mgEnum( DataUnit.BYTE, TYPE_DATA_UNIT ) ).
                f().set( "dataSize",
                    mgDataSize( 12L, DataUnit.KILOBYTE, true )
                ).
                f().set( "duration",
                    mgDuration( 12L, TimeUnitImpl.SECOND, true )
                ).
                f().setString( "mg-ident", "test-ident" ).
                f().setString( "mg-ns", "ns1:ns2@v1" ).
                f().setString( "mg-qname", "ns1:ns2@v1/T1" ).
                f().setString( "mg-type-ref", "ns1:ns2@v1/T1/T2*+?" ).
                f().setString( "mg-identified-name", "ns1:ns2@v1/a/b/c" ).
                build();
    }
}
