package com.bitgirder.etl;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.reflect.ReflectUtils;

import com.bitgirder.concurrent.Duration;

import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.ProcessExit;
import com.bitgirder.process.Stoppable;
import com.bitgirder.process.ProcessRpcClient;

import com.bitgirder.process.management.ProcessControl;
import com.bitgirder.process.management.ProcessFactory;
import com.bitgirder.process.management.ProcessManagement;
import com.bitgirder.process.management.AbstractProcessControl;

import com.bitgirder.event.EventManager;

import com.bitgirder.mingle.model.MingleIdentifiedName;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleIdentifiedNameGenerator;

import java.util.concurrent.atomic.AtomicInteger;

public
abstract
class EtlApplicationTest
extends AbstractVoidProcess
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static EventManager evMgr = EventManager.create();

    private final static AtomicInteger appSeq = new AtomicInteger();

    private final static MingleIdentifiedNameGenerator idGen =
        MingleIdentifiedNameGenerator.forPrefix( "etl:test@v1/etlAppObj" );

    private EtlApplication app;

    protected EtlApplicationTest() { super( ProcessRpcClient.create() ); }

    @Override
    protected
    void
    childExited( AbstractProcess< ? > child,
                 ProcessExit< ? > exit )
    {
        if ( ! exit.isOk() ) fail( exit.getThrowable() );
        if ( ! hasChildren() ) exit();
    }

    private
    MingleIdentifier
    nextAppName()
    {
        return 
            MingleIdentifier.
                create( "etl-test-app" + appSeq.getAndIncrement() );
    }
 
    protected final MingleIdentifiedName nextId() { return idGen.next(); }

    private
    final
    static
    class ProcFactImpl< P extends AbstractProcess< ? > >
    implements ProcessFactory< P >
    {
        private final Class< P > cls;

        private ProcFactImpl( Class< P > cls ) { this.cls = cls; }

        public
        P
        newProcess()
            throws Exception
        {
            return ReflectUtils.newInstance( cls );
        }
    }

    protected
    final
    < P extends AbstractProcess< ? > >
    ProcessFactory< P >
    factoryFor( Class< P > cls )
    {
        inputs.notNull( cls, "cls" );
        return new ProcFactImpl< P >( cls );
    }

    protected
    final
    < P extends AbstractProcess< ? > & Stoppable >
    ProcessControl
    controlFor( Class< P > cls )
    {
        inputs.notNull( cls, "cls" );

        return
            ProcessManagement.createStoppableControl(
                factoryFor( cls ), 1, Duration.fromMinutes( 1 ) );
    }

    protected
    final
    void
    spawnApplication( EtlApplication a )
    {
        spawn( app = inputs.notNull( a, "a" ) );
    }

    protected
    final
    EtlApplication.Builder
    nextAppBuilder()
    {
        return 
            new EtlApplication.Builder().
                setApplicationName( nextAppName() ).
                setEventManager( evMgr );
    }

    protected
    abstract
    void
    startTest()
        throws Exception;

    protected final void startImpl() throws Exception { startTest(); }
}
