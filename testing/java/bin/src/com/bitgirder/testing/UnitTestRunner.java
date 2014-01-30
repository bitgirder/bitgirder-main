package com.bitgirder.testing;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;

import com.bitgirder.application.ApplicationProcess;

import com.bitgirder.test.TestPhase;

import java.util.List;
import java.util.Arrays;
import java.util.Comparator;
import java.util.Map;

final
class UnitTestRunner
extends ApplicationProcess
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();
    
    private final UnitTestEngine engine;

    private final Map< InvocationDescriptor, InvocationInfo > events =
        Lang.newMap();

    private long startTime;

    private
    UnitTestRunner( Configurator c )
    {
        super( c );

        UnitTestEngine.Builder b = new UnitTestEngine.Builder();

        if ( c.filterPattern != null ) b.setFilterPattern( c.filterPattern );
        if ( c.poolSize != null ) b.setPoolSize( c.poolSize );

        b.setEventHandler( new EventHandlerImpl() );
        b.setClassNames( c.classNames );
        
        engine = b.build();
    }

    private
    CharSequence
    format( InvocationDescriptor id )
    {
        return
            new StringBuilder().
                append( id.getPhase().name().toLowerCase() ).
                append( " " ).
                append( id.getName() );
    }

    private
    final
    static
    class InvocationInfo
    {
        private InvocationDescriptor desc;
        private long startTime;
        private long endTime;
        private Throwable th;
    }

    private
    final
    class EventHandlerImpl
    implements InvocationEventHandler
    {
        public
        void
        invocationStarted( InvocationDescriptor id,
                           long startTime )
        {
            InvocationInfo info = new InvocationInfo();
            info.desc = id;
            info.startTime = startTime;

            state.isTrue( events.put( id, info ) == null );
            CodeLoggers.code( "Started", format( id ) );
        }

        public
        void
        invocationCompleted( InvocationDescriptor id,
                             Throwable th,
                             long endTime )
        {
            InvocationInfo info = state.get( events, id, "events" );

            info.th = th;
            info.endTime = endTime;
        }
    }

    private
    InvocationInfo[]
    validateEvents()
    {
        InvocationInfo[] res = new InvocationInfo[ events.size() ];

        int i = 0;

        for ( InvocationInfo inf : events.values() )
        {
            state.isFalse( 
                inf.startTime == 0, "start time not set for", 
                format( inf.desc ) );

            state.isFalse(
                inf.endTime == 0, "end time not set for", format( inf.desc ) );
            
            res[ i++ ] = inf;
        }

        return res;
    }

    // Lex ordering of orders ( failOrder, phaseOrder, nameOrder ), where a
    // failure is greater than success (will be reported last), where BEFORE >
    // TEST > AFTER, and names sort lexicographically
    private
    final
    static
    class InvocationInfoComparator
    implements Comparator< InvocationInfo >
    {
        private
        int
        failOrder( Throwable th )
        {
            return th == null ? -1 : 1;
        }

        public
        int
        compare( InvocationInfo i1,
                 InvocationInfo i2 )
        {
            int res = failOrder( i1.th ) - failOrder( i2.th );
            if ( res != 0 ) return res;

            res = i1.desc.getPhase().ordinal() - i2.desc.getPhase().ordinal();
            if ( res != 0 ) return res;

            return i1.desc.getName().compareTo( i2.desc.getName() );
        }
    }

    private
    String
    formatTime( long millis )
    {
        return String.format( "%1$.3fs", ( (double) millis ) / 1000.0d );
    }

    private
    void
    report( InvocationInfo inf )
    {
        String elapsedStr = formatTime( inf.endTime - inf.startTime );

        if ( inf.th == null )
        {
            CodeLoggers.code( format( inf.desc ), "succeeded in", elapsedStr );
        }
        else 
        {
            CodeLoggers.warn( 
                inf.th, format( inf.desc ), "failed in", elapsedStr );
        }
    }

    private
    void
    reportTestCounts( InvocationInfo[] arr )
    {
        int tests = 0;
        int testsFailed = 0;

        for ( InvocationInfo inf : arr )
        {
            if ( inf.desc.getPhase() == TestPhase.TEST )
            {
                if ( inf.th != null ) ++testsFailed;
                ++tests;
            }
        }
        
        System.out.printf( 
            "Ran %1$d tests; %2$d succeeded and %3$d failed\n",
            tests, ( tests - testsFailed ), testsFailed 
        );
    }

    private
    int
    getExitCode( InvocationInfo[] arr )
    {
        for ( InvocationInfo inf : arr ) if ( inf.th != null ) return -1;
        return 0;
    }
 
    private
    int
    reportResults()
    {
        long runTime = System.currentTimeMillis() - startTime;

        InvocationInfo[] arr = validateEvents();

        Arrays.sort( arr, new InvocationInfoComparator() );

        for ( InvocationInfo inf : arr ) report( inf );
        
        reportTestCounts( arr ); 
        System.out.println( "Run time was " + formatTime( runTime ) );

        return getExitCode( arr );
    }
 
    public
    int
    execute()
        throws Exception
    { 
        startTime = System.currentTimeMillis();
        engine.execute();
 
        return reportResults();
    }

    private
    final
    static
    class Configurator
    extends ApplicationProcess.Configurator
    {
        private final List< String > classNames = Lang.newList();
        private String filterPattern;
        private Integer poolSize;

        @Argument
        private
        void
        setFilterPattern( String filterPattern )
        {
            this.filterPattern = filterPattern;
        }

        @Argument
        private
        void
        setTestClass( String className )
        {
            classNames.add( className );
        }

        @Argument
        private
        void
        setPoolSize( int poolSize )
        {
            this.poolSize = poolSize;
        }
    }
}
