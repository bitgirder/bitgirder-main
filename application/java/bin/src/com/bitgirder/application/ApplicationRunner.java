package com.bitgirder.application;

import static com.bitgirder.application.ApplicationProcess.Configurator;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.StandardThread;

import com.bitgirder.log.CodeLoggers;
import com.bitgirder.log.CodeEventSink;
import com.bitgirder.log.CodeEvent;
import com.bitgirder.log.CodeEvents;
import com.bitgirder.log.CodeEventFormatter;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;
import com.bitgirder.lang.PatternHelper;

import com.bitgirder.lang.reflect.ReflectUtils;

import java.lang.reflect.Constructor;
import java.lang.reflect.Method;

import java.io.PrintStream;

import java.util.Map;
import java.util.Collection;

import java.util.regex.Pattern;

import java.util.concurrent.TimeoutException;

final
class ApplicationRunner
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    // For now only accept an arg string starting with '--' followed by at least
    // one non-dash character
    private final static Pattern ARG_PATTERN =
        PatternHelper.compile( "^--[^\\-].*" );

    private
    final
    static
    class ConfigurationContext
    {
        private final Constructor< ? > appCons;
        private final Class configuratorCls;
        private final Configurator configurator;

        private final Map< String, ArgumentSetter > setters;

        private
        ConfigurationContext( Constructor< ? > appCons,
                              Class< ? > configuratorCls,
                              Configurator configurator,
                              Map< String, ArgumentSetter > setters )
        {
            this.appCons = appCons;
            this.configuratorCls = configuratorCls;
            this.configurator = configurator;
            this.setters = setters;
        }
    }

    private
    static
    Class< ? extends Configurator >
    extractConfiguratorClass( Constructor< ? > cons )
    {
        Class< ? >[] paramTypes = cons.getParameterTypes();

        if ( paramTypes.length == 1 )
        {
            Class< ? > cls = paramTypes[ 0 ];

            if ( Configurator.class.isAssignableFrom( cls ) )
            {
                return cls.asSubclass( Configurator.class );
            }
            else 
            {
                throw state.createFail( 
                    "Parameter to constructor", cons, "is not a descendent of",
                    Configurator.class );
            }
        }
        else
        {
            throw state.createFail( 
                "Constructor", cons, "has too many parameters" );
        }
    }

    private
    static
    Constructor< ? >
    getApplicationConstructor( Class< ? > cls )
    {
        Constructor< ? >[] constructors = cls.getDeclaredConstructors();

        if ( constructors.length == 1 ) 
        {
            Constructor< ? > res = constructors[ 0 ];
            res.setAccessible( true );

            return res;
        }
        else
        {
            throw state.createFail(
                "Application class", cls, "has more than one constructor" );
        }
    }

    private
    static
    Class< ? extends ApplicationProcess >
    getApplicationClass( String[] args )
    {
        state.isTrue( args.length > 0, "No application specified" );
        String clsName = args[ 0 ];
        
        try 
        { 
            Class< ? > appCls = Class.forName( clsName );
            return appCls.asSubclass( ApplicationProcess.class );
        }
        catch ( ClassNotFoundException cnfe )
        {
            throw state.createFail( "Can't find application class", clsName );
        }
    }

    private
    static
    abstract
    class ArgumentSetter
    {
        private final Method m;

        private ArgumentSetter( Method m ) { this.m = m; }

        abstract
        Object
        getSetterValue( String valStr )
            throws Exception;
        
        private
        final
        void
        setConfigurationValue( ConfigurationContext cCtx,
                               String valStr )
            throws Exception
        {
            Object setterVal = getSetterValue( valStr );
            ReflectUtils.invoke( m, cCtx.configurator, setterVal );
        }
    }

    private
    final
    static
    class StringSetter
    extends ArgumentSetter
    {
        private StringSetter( Method m ) { super( m ); }

        Object getSetterValue( String valStr ) { return valStr; }
    }

//    private
//    final
//    static
//    class FileWrapperSetter
//    extends ArgumentSetter
//    {
//        private FileWrapperSetter( Method m ) { super( m ); }
//
//        Object
//        getSetterValue( String valStr )
//        {
//            return new FileWrapper( valStr );
//        }
//    }
//
//    private
//    final
//    static
//    class DirWrapperSetter
//    extends ArgumentSetter
//    {
//        private DirWrapperSetter( Method m ) { super( m ); }
//
//        Object
//        getSetterValue( String valStr )
//        {
//            return new DirWrapper( valStr );
//        }
//    }

    private
    final
    static
    class EnumSetter
    extends ArgumentSetter
    {
        private final Map< String, Object > enumConstants;

        private 
        EnumSetter( Map< String, Object > enumConstants,
                    Method m )
        {
            super( m );
            this.enumConstants = enumConstants;
        }

        Object
        getSetterValue( String valStr )
        {
            Object res = enumConstants.get( valStr );

            if ( res == null )
            {
                throw inputs.createFail(
                    "Invalid enumeration value '" + valStr + "'",
                    "(must be one of:", 
                    Strings.join( ", ", enumConstants.keySet() ) + ")" );
            }
            else return res;
        }
    }

    private
    static
    ArgumentSetter
    createEnumSetter( Class< ? > cls,
                      Method m )
    {
        Object[] constantsArr = state.notNull( cls.getEnumConstants() );

        Map< String, Object > constantsMap = Lang.newMap( constantsArr.length );

        for ( Object constant : constantsArr )
        {
            Enum< ? > e = (Enum< ? >) constant;
            String valStr = e.name().toLowerCase().replace( '_', '-' );

            constantsMap.put( valStr, e );
        }

        return new EnumSetter( constantsMap, m );
    }

    private
    final
    static
    class IntegerSetter
    extends ArgumentSetter
    {
        private IntegerSetter( Method m ) { super( m ); }

        Object
        getSetterValue( String valStr )
        {
            return new Integer( valStr );
        }
    }
    
    private
    final
    static
    class LongSetter
    extends ArgumentSetter
    {
        private LongSetter( Method m ) { super( m ); }

        Object
        getSetterValue( String valStr )
        {
            return new Long( valStr );
        }
    }

    private
    final
    static
    class BooleanSetter
    extends ArgumentSetter
    {
        private BooleanSetter( Method m ) { super( m ); }

        Object
        getSetterValue( String valStr )
        {
            valStr = valStr.trim().toLowerCase();

            if ( valStr.equals( "true" ) ) return Boolean.TRUE;
            else if ( valStr.equals( "false" ) ) return Boolean.FALSE;
            else 
            {
                throw inputs.createFail( 
                    "Invalid boolean flag value:", valStr );
            }
        }
    }

    private
    static
    ArgumentSetter
    argumentSetterFor( Method m,
                       Class< ? > cls )
    {
        m.setAccessible( true );

        if ( cls.equals( String.class ) || cls.equals( CharSequence.class ) )
        {
            return new StringSetter( m );
        }
//        else if ( cls.equals( FileWrapper.class ) )
//        {
//            return new FileWrapperSetter( m );
//        }
//        else if ( cls.equals( DirWrapper.class ) )
//        {
//            return new DirWrapperSetter( m );
//        }
        else if ( Enum.class.isAssignableFrom( cls ) )
        {
            return createEnumSetter( cls, m );
        }
        else if ( cls.equals( Integer.class ) || cls.equals( Integer.TYPE ) )
        {
            return new IntegerSetter( m );
        }
        else if ( cls.equals( Long.class ) || cls.equals( Long.TYPE ) )
        {
            return new LongSetter( m );
        }
        else if ( cls.equals( Boolean.class ) || cls.equals( Boolean.TYPE ) )
        {
            return new BooleanSetter( m );
        }
        else 
        {
            throw state.createFail( 
                "Don't know how to set values of parameter type", 
                cls, "in", m );
        }
    }

    private
    static
    ArgumentSetter
    argumentSetterFor( Method m )
    {
        Class< ? >[] params = m.getParameterTypes();

        state.isTrue( 
            params.length == 1,
            "Setter method does not have exactly one parameter:", m );
        
        return argumentSetterFor( m, params[ 0 ] );
    }
 
    // Stub for later when we may allow setters to go by more than one name, or
    // just allow programmers to set a single name which bypasses the default
    // name
    private
    static
    Collection< String >
    getArgNamesForSetter( Method setter )
    {
        String name = setter.getName(); 
        state.isTrue( 
            name.startsWith( "set" ), 
            "Setter name does not start with 'set':", name );

        StringBuilder sb = new StringBuilder();

        for ( int i = 3, e = name.length(); i < e; ++i )
        {
            char ch = name.charAt( i );

            if ( Character.isUpperCase( ch ) )
            {
                if ( i > 3 ) sb.append( "-" );
                sb.append( Character.toLowerCase( ch ) );
            }
            else sb.append( ch );
        }

        return Lang.singletonList( sb.toString() );
    }

    private
    static
    Map< String, ArgumentSetter >
    buildSettersMap( Class< ? extends Configurator > configuratorCls )
    {
        Map< String, ArgumentSetter > res = Lang.newMap();

        Collection< Method > methods =
            ReflectUtils.getDeclaredAncestorMethods(
                configuratorCls, Configurator.Argument.class );

        for ( Method m : methods )
        {
            ArgumentSetter setter = argumentSetterFor( m );

            Collection< String > argNames = getArgNamesForSetter( m );

            for ( String argName : argNames ) 
            {
                Lang.putUnique( res, argName, setter );
            }
        }

        return res;
    }

    private
    static
    ConfigurationContext
    getConfigurationContext( String[] args )
        throws Exception
    {
        Class< ? extends ApplicationProcess > appCls =
            getApplicationClass( args );

        Constructor< ? > appCons = getApplicationConstructor( appCls );

        Class< ? extends Configurator > configuratorCls = 
            extractConfiguratorClass( appCons );

        Configurator configurator = ReflectUtils.newInstance( configuratorCls );

        Map< String, ArgumentSetter > setters = 
            buildSettersMap( configuratorCls );

        return 
            new ConfigurationContext( 
                appCons, configuratorCls, configurator, setters );
    }

    private
    static
    boolean
    isArgument( String arg )
    {
        return ARG_PATTERN.matcher( arg ).matches();
    }

    private
    static
    ArgumentSetter
    getArgumentSetter( String arg,
                       ConfigurationContext cCtx )
    {
        if ( arg.startsWith( "--" ) )
        {
            ArgumentSetter res = cCtx.setters.get( arg.substring( 2 ) );
            inputs.isFalse( res == null, "Unrecognized argument:", arg );

            return res;
        }
        else throw inputs.createFail( "Illegal argument:", arg );
    }

    private
    static
    void
    configureArgument( String arg,
                       String valStr,
                       ConfigurationContext cCtx )
        throws Exception
    {
        ArgumentSetter setter = getArgumentSetter( arg, cCtx );
        setter.setConfigurationValue( cCtx, valStr );
    }

    private
    static
    void
    configureApplication( ConfigurationContext cCtx,
                          String[] args,
                          int indx )
        throws Exception
    {
        while ( indx < args.length )
        {
            String arg = args[ indx++ ];
            inputs.isTrue( isArgument( arg ), "Illegal argument:", arg );
            
            state.isFalse( indx == args.length, "Need a value for", arg );
            String valStr = args[ indx++ ];

            configureArgument( arg, valStr, cCtx );
        }
    }

    private
    static
    ApplicationProcess
    createApplication( ConfigurationContext cCtx )
        throws Exception
    {
        return (ApplicationProcess)
            ReflectUtils.invoke( cCtx.appCons, cCtx.configurator );
    }

    private
    static
    void
    failApplication( Throwable th )
    {
        th.printStackTrace( System.err );
        System.exit( 1 );
    }

    public
    static
    void
    main( String[] args )
        throws Exception
    {
        try
        {
            ConfigurationContext cCtx = getConfigurationContext( args );

            configureApplication( cCtx, args, 1 );

            ApplicationProcess proc = createApplication( cCtx );
            System.exit( proc.execute() );
        }
        catch ( Throwable th ) { failApplication( th ); }
    }
}
