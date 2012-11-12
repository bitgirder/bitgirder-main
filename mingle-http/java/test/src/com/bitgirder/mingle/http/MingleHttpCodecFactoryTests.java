package com.bitgirder.mingle.http;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.http.HttpRequestMessage;

import com.bitgirder.mingle.codec.MingleCodecException;
import com.bitgirder.mingle.codec.MingleCodec;

import com.bitgirder.test.Test;
import com.bitgirder.test.LabeledTestCall;
import com.bitgirder.test.InvocationFactory;

import java.util.List;

@Test
final
class MingleHttpCodecFactoryTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private
    static
    abstract
    class StubCodecContext
    implements MingleHttpCodecContext
    {
        public
        MingleCodec
        codec()
        {
            throw new UnsupportedOperationException( "Unimplemented" );
        }

        public
        CharSequence
        contentType()
        {
            throw new UnsupportedOperationException( "Unimplemented" );
        }
    }

    private final static class CodecContext1 extends StubCodecContext {}
    private final static class CodecContext2 extends StubCodecContext {}
    private final static class CodecContext3 extends StubCodecContext {}

    private
    final
    static
    MingleHttpCodecFactory
    asFact( final MingleHttpCodecContext ctx )
    {
        return
            new MingleHttpCodecFactory() 
            {
                public 
                MingleHttpCodecContext
                codecContextFor( HttpRequestMessage req )
                    throws MingleCodecException
                {
                    String ctype = req.h().expectContentTypeString().toString();

                    if ( ctype.equals( "application/fail" ) )
                    {
                        throw new MingleCodecException( "marker" );
                    }
                    else return ctx;
                }
            };
    }

    private
    final
    class DefaultCodecFactoryTest
    extends LabeledTestCall
    {
        private CharSequence ctype;
        private Class< ? > expctCls;

        private DefaultCodecFactoryTest( CharSequence lbl ) { super( lbl ); }

        private
        DefaultCodecFactoryTest
        expectContext( Class< ? > expctCls )
        {
            this.expctCls = expctCls;
            return this;
        }

        private
        DefaultCodecFactoryTest
        setContentType( CharSequence ctype )
        {
            this.ctype = ctype;
            return this;
        }

        private
        MingleHttpCodecFactory
        factory()
        {
            return
                new MingleHttpCodecFactorySelector.Builder().
                    selectSubType( "codec1", asFact( new CodecContext1() ) ).
                    selectFullType( 
                        "application/codec2", asFact( new CodecContext2() ) ).
                    selectRegex( 
                        "(application|text)/codecOne", 
                        asFact( new CodecContext1() ) ).
                    selectRegex( ".*ambig.*", asFact( new CodecContext3() ) ).
                    selectRegex( ".*biguo.*", asFact( new CodecContext3() ) ).
                    selectFullType(
                        "application/fail", asFact( new CodecContext3() ) ).
                        
                    build();
        }

        protected
        void
        call()
            throws Exception
        {
            HttpRequestMessage.Builder b =
                new HttpRequestMessage.Builder().
                    setMethod( "GET" ).
                    setRequestUri( "/" );
            
            if ( ctype != null ) b.h().setContentType( ctype );

            MingleHttpCodecContext ctx = factory().codecContextFor( b.build() );

            state.cast( expctCls, ctx );
        }
    }

    @InvocationFactory
    private
    List< DefaultCodecFactoryTest >
    testDefaultCodecFactory()
    {
        return Lang.asList(
            
            new DefaultCodecFactoryTest( "app-codec1" ).
                expectContext( CodecContext1.class ).
                setContentType( "application/codec1" ),
            
            new DefaultCodecFactoryTest( "text-codec1" ).
                expectContext( CodecContext1.class ).
                setContentType( "text/codec1" ),
            
            new DefaultCodecFactoryTest( "app-codecOne" ).
                expectContext( CodecContext1.class ).
                setContentType( "application/codecOne" ),
            
            new DefaultCodecFactoryTest( "app-codec2" ).
                expectContext( CodecContext2.class ).
                setContentType( "application/codec2" ),

            new DefaultCodecFactoryTest( "app-codec2-with-charset" ).
                expectContext( CodecContext2.class ).
                setContentType( "application/codec2;charset=utf-8" ),
            
            (DefaultCodecFactoryTest)
            new DefaultCodecFactoryTest( "missing-ctype" ).
                expectFailure( 
                    MingleCodecException.class,
                    "Missing content type"
                ),
 
            (DefaultCodecFactoryTest)
            new DefaultCodecFactoryTest( "unknown-codec" ).
                setContentType( "application/codec3" ).
                expectFailure(
                    MingleCodecException.class,
                    "Unrecognized content type: application/codec3"
                ),
 
            (DefaultCodecFactoryTest)
            new DefaultCodecFactoryTest( "ambiguous-codec" ).
                setContentType( "application/ambiguous" ).
                expectFailure(
                    IllegalStateException.class,
                    "\\QMultiple selector patterns matched content type " +
                    "'application/ambiguous': [.*ambig.*, .*biguo.*]\\E"
                ),
            
            (DefaultCodecFactoryTest)
            new DefaultCodecFactoryTest( "selected-fact-failure-is-thrown" ).
                setContentType( "application/fail" ).
                expectFailure( MingleCodecException.class, "^marker$" )
        );
    }
}
