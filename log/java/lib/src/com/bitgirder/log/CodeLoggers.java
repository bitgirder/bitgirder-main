package com.bitgirder.log;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.io.PrintStream;

public
final
class CodeLoggers
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private static CodeLogger defl = createStreamLogger( System.out );

    private CodeLoggers() {}

    public static CodeEventSink getDefaultEventSink() { return defl; }

    public
    static
    CodeLogger
    createStreamLogger( final PrintStream ps )
    {
        inputs.notNull( ps, "ps" );

        return
            new AbstractCodeLogger() {
                protected void logCodeImpl( CodeEvent ev ) {
                    ps.println( CodeEvents.format( ev ) );
                }
            };
    }

    // Note: CodeLoggers.defl is not volatile, nor are this method or
    // getDefault(Logger|EventSink) synchronized. This is to make access of defl
    // as fast as possible with the understanding that any code which calls
    // replaceDefault() is likely in a position to make its own guarantees that
    // its call to replaceDefault() will have a happens-before relationship to
    // any ensuing accesses to getDefault(Logger|EventSink) that it would care
    // to guarantee will see its new value.
    public
    static
    void
    replaceDefaultSink( final CodeEventSink sink )
    {
        inputs.notNull( sink, "sink" );

        if ( sink instanceof CodeLogger ) defl = (CodeLogger) sink;
        else
        {
            defl = 
                new AbstractCodeLogger() {
                    protected void logCodeImpl( CodeEvent ev ) {
                        sink.logCode( ev );
                    }
                };
        }
    }

    public static void code( Object... msg ) { defl.code( msg ); }

    public
    static
    void
    codef( String tmpl,
           Object... args )
    {
        code( String.format( tmpl, (Object[]) args ) );
    }

    public
    static
    void
    codef( Throwable th,
           String tmpl,
           Object... msg )
    {
        code( th, String.format( tmpl, msg ) );
    }

    public
    static
    void
    code( Throwable th,
          Object... msg )
    {
        defl.code( th, msg );
    }

    public static void warn( Object... msg ) { defl.warn( msg ); }

    public
    static
    void
    warn( Throwable th,
          Object... msg )
    {
        defl.warn( th, msg );
    }

    public
    static
    void
    warnf( String tmpl,
           Object... args )
    {
        warn( String.format( tmpl, (Object[]) args ) );
    }

    public
    static
    void
    warnf( Throwable th,
           String tmpl,
           Object... msg )
    {
        warn( th, String.format( tmpl, msg ) );
    }
}
