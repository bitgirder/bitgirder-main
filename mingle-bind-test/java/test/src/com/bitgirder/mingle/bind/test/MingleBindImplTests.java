package com.bitgirder.mingle.bind.test;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.ObjectReceiver;

import com.bitgirder.mingle.bind.MingleBindTests;
import com.bitgirder.mingle.bind.MingleBinders;
import com.bitgirder.mingle.bind.MingleBinder;

import com.bitgirder.io.IoTestSupport;
import com.bitgirder.io.FileWrapper;

import com.bitgirder.mingle.codec.MingleCodec;

import com.bitgirder.mingle.json.JsonMingleCodecs;

import com.bitgirder.process.AbstractVoidProcess;

import com.bitgirder.test.TestRuntime;
import com.bitgirder.test.Test;

@Test
final
class MingleBindImplTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final TestRuntime rt;
    private final MingleBinder mb;

    private 
    MingleBindImplTests( TestRuntime rt ) 
    { 
        this.rt = rt; 
        this.mb = MingleBindTests.createTestBinder( rt );
    }

    @Test
    private
    final
    class MingleBinderFileRoundtripTest
    extends AbstractVoidProcess
    {
        private final MingleBindTests.Struct1 jvObj = 
            MingleBindTests.Struct1.INST1;

        private final MingleCodec codec = JsonMingleCodecs.getJsonCodec( "\n" );

        private
        MingleBinderFileRoundtripTest()
        {
            super( IoTestSupport.create( rt ) );
        }

        private
        void
        readObject( FileWrapper fw )
        {
            MingleBinders.fromFile(
                mb,
                MingleBindTests.Struct1.class,
                codec,
                fw,
                getActivityContext(),
                behavior( IoTestSupport.class ).ioProcessor(),
                new ObjectReceiver< MingleBindTests.Struct1 >() {
                    public void receive( MingleBindTests.Struct1 obj )
                    {
                        state.equal( jvObj, obj );
                        exit();
                    }
                }
            );
        }

        protected
        void
        startImpl()
            throws Exception
        {
            IoTestSupport sprt = behavior( IoTestSupport.class );

            final FileWrapper fw = sprt.createTempFile();
            
            MingleBinders.toFile(
                mb,
                jvObj,
                codec,
                fw,
                getActivityContext(),
                sprt.ioProcessor(),
                new AbstractTask() {
                    protected void runImpl() { readObject( fw ); }
                }
            );
        }
    }
}
