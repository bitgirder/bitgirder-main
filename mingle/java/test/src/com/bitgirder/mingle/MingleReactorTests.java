package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Strings;
import com.bitgirder.lang.Lang;

import com.bitgirder.test.Test;
import com.bitgirder.test.InvocationFactory;
import com.bitgirder.test.LabeledTestCall;

import com.bitgirder.io.IoTestFactory;

import java.util.List;

@Test
final
class MingleReactorTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private
    final
    static
    class ReactorSeqErrorTest
    extends LabeledTestCall
    {
        private final List< String > calls;
        private final String errMsg;
        private final String topType;

        private
        ReactorSeqErrorTest( List< String > calls,
                             String errMsg,
                             String topType )
        {
            super( Strings.join( ",", calls ) );

            this.calls = calls;
            this.errMsg = errMsg;
            this.topType = topType;
        }

        private
        void
        callReactor( ValueReactor r,
                     String call )
            throws Exception
        {
            state.fail( "Unhandled call:", call );
        }

        public
        void
        call()
            throws Exception
        {
            PipelineReactor r = 
                PipelineReactor.create( new StructureCheckReactor() );

            for ( String call : calls ) callReactor( r, call );
            state.fail( "Unimplemented" );
        }
    }

    private
    final
    static
    class ErrSeqTestReader
    extends IoTestFactory.LeTestReader< ReactorSeqErrorTest >
    {
        private ErrSeqTestReader() { super( "reactor-seq-error-tests.bin" ); } 

        protected
        void
        readHeader()
            throws Exception
        {
            expectInt32( 1 );
        }

        private
        List< String >
        readCalls()
            throws Exception
        {
            int sz = leRd().readInt();
            List< String > res = Lang.newList( sz );

            for ( int i = 0; i < sz; ++i ) res.add( leRd().readUtf8() );

            return res;
        }

        protected
        ReactorSeqErrorTest
        readNext()
            throws Exception
        {
            int tc = leRd().readByte();
            if ( tc == 0 ) return null;
            if ( tc != 1 ) throw failf( "Unexpected type code: 0x%02x", tc );

            return new ReactorSeqErrorTest(
                readCalls(),
                leRd().readUtf8(),
                leRd().readUtf8()
            );
        }
    }

    @InvocationFactory
    private
    List< ReactorSeqErrorTest >
    testReactorSeqError()
        throws Exception
    {
        return new ErrSeqTestReader().call();
    }
}
