package com.bitgirder.mingle.codec;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.reflect.ReflectUtils;

import com.bitgirder.io.IoUtils;

import java.io.BufferedReader;
import java.io.InputStreamReader;

import java.util.List;

import java.net.URL;

public
final
class MingleCodecFactories
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static String RSRC_NAME = "codec-factory-init.txt";

    private MingleCodecFactories() {}

    private
    static
    MingleCodecFactoryInitializerException
    factoryInitException( URL src,
                          int lineNo,
                          Throwable cause,
                          String msg )
    {
        StringBuilder sb = new StringBuilder();

        if ( src != null )
        {
            sb.append( "[loading from " ).append( src );
            if ( lineNo > 0 ) sb.append( ", line " ).append( lineNo );
            sb.append( "]: " );
        }

        sb.append( msg );

        MingleCodecFactoryInitializerException res =
            new MingleCodecFactoryInitializerException( msg );
        
        if ( cause != null ) res.initCause( cause );

        return res;
    }

    private
    static
    List< URL >
    getResources()
        throws MingleCodecFactoryInitializerException
    {
        try { return IoUtils.getResources( RSRC_NAME ); }
        catch ( Throwable th )
        {
            throw 
                factoryInitException( 
                    null, -1, th, "Couldn't get factory initializers" );
        }
    }

    private
    static
    Class< ? >
    getInitClass( String line,
                  URL u,
                  int lineNo )
        throws MingleCodecFactoryInitializerException
    {
        try { return Class.forName( line ); }
        catch ( ClassNotFoundException cnfe )
        {
            throw
                factoryInitException(
                    u, 
                    lineNo, 
                    cnfe,
                    "Couldn't load factory class " + line + " (see cause)"
                );
        }
    }

    private
    static
    void
    initCodecs( String line,
                MingleCodecFactory.Builder b,
                URL u,
                int lineNo )
        throws MingleCodecFactoryInitializerException
    {
        Class< ? > cls = getInitClass( line, u, lineNo );

        try
        {
            MingleCodecFactoryInitializer init =
                (MingleCodecFactoryInitializer) 
                    ReflectUtils.newInstance( cls );
            
            init.initialize( b );
        }
        catch ( Throwable th )
        {
            throw 
                factoryInitException( 
                    u, lineNo, th, "Codec factory init failed (see cause)"
                );
        }
    }

    private
    static
    void
    initCodecs( URL u,
                BufferedReader br,
                MingleCodecFactory.Builder b )
        throws Exception
    {
        String line = null;

        for ( int lineNo = 1; ( line = br.readLine() ) != null; ++lineNo )
        {
            initCodecs( line, b, u, lineNo );
        }
    }

    private
    static
    void
    initCodecs( URL u,
                MingleCodecFactory.Builder b )
        throws MingleCodecFactoryInitializerException
    {
        try
        {
            BufferedReader br = 
                new BufferedReader( 
                    new InputStreamReader( u.openStream(), "utf-8" ) );
            
            try { initCodecs( u, br, b ); }
            finally { IoUtils.closeQuietly( br, u.toString() ); }
        }
        catch ( MingleCodecFactoryInitializerException mcfe ) { throw mcfe; }
        catch ( Throwable th ) 
        { 
            throw 
                factoryInitException( 
                    u, -1, null, "Couldn't load factories from" );
        }
    }

    public
    static
    MingleCodecFactory
    loadDefault()
        throws MingleCodecFactoryInitializerException
    {
        MingleCodecFactory.Builder b = new MingleCodecFactory.Builder();

        for ( URL u : getResources() ) initCodecs( u, b );

        return b.build();
    }
}
