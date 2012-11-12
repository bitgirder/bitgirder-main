package com.bitgirder.mingle.codec;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.ObjectReceiver;
import com.bitgirder.lang.Lang;

import com.bitgirder.test.TestRuntime;

import com.bitgirder.testing.Testing;

import com.bitgirder.mingle.model.MingleStruct;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleNull;
import com.bitgirder.mingle.model.MingleList;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleServiceRequest;
import com.bitgirder.mingle.model.MingleServiceResponse;
import com.bitgirder.mingle.model.ModelTestInstances;

import java.util.Map;

import java.nio.ByteBuffer;

public
final
class MingleCodecTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    public final static Object KEY_DEFAULT_CODEC_FACTORY =
        MingleCodecTests.class.getName() + ".defaultCodecFactory";

    private final static Map< MingleIdentifier, MingleStruct > STD_TEST_STRUCTS;

    private MingleCodecTests() {}

    public
    static
    Map< MingleIdentifier, MingleStruct >
    getStandardTestStructs()
    {
        return STD_TEST_STRUCTS;
    }

    public
    static
    MingleCodecFactory
    expectDefaultCodecFactory( TestRuntime rt )
    {
        inputs.notNull( rt, "rt" );

        return 
            Testing.expectObject(
                rt, 
                KEY_DEFAULT_CODEC_FACTORY, 
                MingleCodecFactory.class 
            );
    }

    public
    static
    void
    assertEqual( Object me1,
                 Object me2 )
        throws Exception
    {
        if ( state.sameNullity( me1, me2 ) )
        {
            if ( me1 instanceof MingleStruct )
            {
                ModelTestInstances.assertEqual( 
                    (MingleStruct) me1, state.cast( MingleStruct.class, me2 ) );
            }
            else if ( me1 instanceof MingleServiceRequest )
            {
                ModelTestInstances.assertEqual(
                    (MingleServiceRequest) me1,
                    state.cast( MingleServiceRequest.class, me2 ) );
            }
            else if ( me1 instanceof MingleServiceResponse )
            {
                ModelTestInstances.assertEqual(
                    (MingleServiceResponse) me2,
                    state.cast( MingleServiceResponse.class, me2 ) );
            }
            else state.fail( "Unexpected encodable:", me1 );
        }
    }

    public
    static
    void
    assertFactoryRoundtrip( CharSequence codecId,
                            TestRuntime rt,
                            ObjectReceiver< ByteBuffer > encRecv )
        throws Exception
    {
        inputs.notNull( codecId, "codecId" );
        inputs.notNull( rt, "rt" );
        inputs.notNull( encRecv, "encRecv" );

        MingleCodec codec = 
            MingleCodecTests.expectDefaultCodecFactory( rt ).
                expectCodec( codecId );

        ByteBuffer bb = 
            MingleCodecs.
                toByteBuffer( codec, ModelTestInstances.TEST_STRUCT1_INST1 );
 
        encRecv.receive( bb.slice() );

        ModelTestInstances.assertEqual(
            ModelTestInstances.TEST_STRUCT1_INST1,
            MingleCodecs.fromByteBuffer( codec, bb, MingleStruct.class )
        );
    }

    public
    static
    void
    assertFactoryRoundtrip( CharSequence codecId,
                            TestRuntime rt )
        throws Exception
    {
        assertFactoryRoundtrip( 
            codecId, rt, new ObjectReceiver< ByteBuffer >() {
                public void receive( ByteBuffer bb ) {}
            }
        );
    }

    @Testing.RuntimeInitializer
    private
    static
    void
    init( Testing.RuntimeInitializerContext ctx )
    {
        Testing.submitInitTask( 
            ctx,
            new Testing.AbstractInitTask( ctx ) {
                protected void runImpl() throws Exception 
                {
                    context().setObject(
                        KEY_DEFAULT_CODEC_FACTORY,
                        MingleCodecFactories.loadDefault()
                    );
    
                    context().complete();
                }
            }
        );
    }

    static
    {
        STD_TEST_STRUCTS =
            Lang.unmodifiableMap(
                Lang.newMap( MingleIdentifier.class, MingleStruct.class,

                    MingleIdentifier.create( "test-struct1-inst1" ),
                    ModelTestInstances.TEST_STRUCT1_INST1,
                
                    MingleIdentifier.create( "empty-struct" ),
                    MingleModels.structBuilder().
                        setType( "ns1@v1/S1" ).
                        build(),
                
                    MingleIdentifier.create( "empty-val-struct" ),
                    MingleModels.structBuilder().
                        setType( "ns1@v1/S1" ).
                        f().setBuffer( "buf1", new byte[] {} ).
                        f().setString( "str1", "" ).
                        f().set( "list1", MingleModels.getEmptyList() ).
                        f().set( "map1", MingleModels.getEmptySymbolMap() ).
                        build(),
                
                    MingleIdentifier.create( "nulls-in-list" ),
                    MingleModels.structBuilder().
                        setType( "ns1@v1/S1" ).
                        f().set( "list1",
                            new MingleList.Builder().
                                add( MingleModels.asMingleString( "s1" ) ).
                                add( MingleNull.getInstance() ).
                                add( MingleNull.getInstance() ).
                                add( MingleModels.asMingleString( "s4" ) ).
                                build()
                        ).
                        build()
            )
        );
    }
}
