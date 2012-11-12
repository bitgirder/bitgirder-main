package com.bitgirder.mingle.codec;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.ObjectReceiver;

import com.bitgirder.process.AbstractVoidProcess;

import com.bitgirder.io.ProtocolProcessors;

import com.bitgirder.test.LabeledTestObject;
import com.bitgirder.test.TestFailureExpector;

import java.nio.ByteBuffer;

public
abstract
class AbstractCodecDetectionTest< T extends AbstractCodecDetectionTest >
extends AbstractVoidProcess
implements LabeledTestObject,
           TestFailureExpector
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final CharSequence lbl;

    private Class< ? > expctCls;
    private ByteBuffer input;
    private int copyBufSz = 20;

    private Class< ? extends Throwable > errCls;
    private CharSequence errPat;

    protected
    AbstractCodecDetectionTest( CharSequence lbl )
    {
        this.lbl = inputs.notNull( lbl, "lbl" );
    }

    public final CharSequence getLabel() { return lbl; }
    public final Object getInvocationTarget() { return this; }

    private T castThis() { return Lang.castUnchecked( this ); }

    public
    final
    T
    expectFailure( Class< ? extends Throwable > errCls,
                   CharSequence errPat )
    {
        this.errCls = inputs.notNull( errCls, "errCls" );
        this.errPat = inputs.notNull( errPat, "errPat" );

        return castThis();
    }

    public
    final
    T
    expectCodec( Class< ? > expctCls )
    {
        this.expctCls = inputs.notNull( expctCls, "expctCls" );
        return castThis();
    }

    public
    final
    T
    setInput( ByteBuffer input )
    {
        this.input = inputs.notNull( input, "input" );
        return castThis();
    }

    public
    final
    T
    setInput( byte... input )
    {
        inputs.notNull( input, "input" );

        return setInput( ByteBuffer.wrap( input ) );
    }

    public
    final
    T
    setCopyBufferSize( int copyBufSz )
    {
        this.copyBufSz = inputs.positiveI( copyBufSz, "copyBufSz" );

        return castThis();
    }

    public
    final
    Class< ? extends Throwable >
    expectedFailureClass()
    {
        return errCls;
    }

    public final CharSequence expectedFailurePattern() { return errPat; }

    protected
    abstract
    MingleCodecFactory
    codecFactory()
        throws Exception;

    private
    void
    detectionDone( MingleCodecDetection det )
        throws Exception
    {
        MingleCodec codec = det.getResult(); // may fail here
        state.isFalse( expctCls == null, "Did not expect a codec" );

        state.cast( expctCls, codec ); // if we get here then check type

        exit();
    }

    // Strictly speaking there is nothing this test needs from a
    // process/ProtocolProcessor standpoint in order to test the base
    // operation of codec detection, but since most uses of detection will
    // be taking ProtocolProcessors as input, this seems to make the most
    // sense and, in addition, gives us good coverage of
    // MingleCodecs.detectCodec()
    protected
    final
    void
    startImpl()
        throws Exception
    {
        MingleCodecs.detectCodec(
            codecFactory(),
            ProtocolProcessors.createBufferSend( input ),
            copyBufSz,
            getActivityContext(),
            new ObjectReceiver< MingleCodecDetection >() {
                public void receive( MingleCodecDetection det )
                    throws Exception
                {
                    detectionDone( det );
                }
            }
        );
    }
}
