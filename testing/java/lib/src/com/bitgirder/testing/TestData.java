package com.bitgirder.testing;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.io.InputStream;
import java.io.BufferedInputStream;
import java.io.IOException;
import java.io.File;
import java.io.FileInputStream;
import java.io.FileNotFoundException;

import java.util.List;

public
final
class TestData
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static String PROP_TEST_DATA_PATH = "bitgirder.testDataPath";
    private final static String PROP_TEST_BIN_PATH = "bitgirder.testBinPath";

    private TestData() {}

    private
    static
    String
    getTestPath( String pathProp )
    {
        String res = System.getProperty( pathProp );
        return res == null ? "" : res;
    }

    private
    static
    File
    expectFile( String pathProp, 
                String fileNm )
        throws FileNotFoundException
    {
        inputs.notNull( fileNm, "fileNm" );

        File res = new File( fileNm );
        if ( res.exists() ) return res;

        String[] dirs = getTestPath( pathProp ).split( File.pathSeparator );

        for ( String dir : dirs )
        {
            res = new File( dir + File.separatorChar + fileNm );
            if ( res.exists() ) return res;
        }

        throw new FileNotFoundException( 
            "Cannot find test data file: " + fileNm );
    }

    public
    static
    File
    expectDataFile( String fileNm )
        throws FileNotFoundException
    {
        return expectFile( PROP_TEST_DATA_PATH, fileNm );
    }

    public
    static
    File
    expectCommand( String fileNm )
        throws FileNotFoundException
    {
        return expectFile( PROP_TEST_BIN_PATH, fileNm );
    }

    public
    static
    InputStream
    openDataFile( String nm )
        throws IOException
    {
        return new FileInputStream( expectFile( PROP_TEST_DATA_PATH, nm ) );
    }

    public
    static
    abstract
    class TestReader< T >
    {
        private final String fileNm;

        private InputStream is;

        private boolean endCalled;

        protected
        TestReader( String fileNm )
        {
            this.fileNm = inputs.notNull( fileNm, "fileNm" );
        }

        // returns null so impls can chain end() with a null return val from
        // readNext():
        //
        //  protected T readNext() {
        //      if ( someEndCondition ) return end();
        //      return someOtherValue();
        //  }
        //
        protected 
        final 
        T
        end() 
        { 
            this.endCalled = true; 
            return null;
        }

        protected final InputStream io() { return is; }

        protected
        final
        IOException
        failf( String tmpl,
               Object... args )
        {
            return new IOException( String.format( tmpl, args ) );
        }

        protected
        abstract
        void
        readHeader()
            throws Exception;
        
        protected
        abstract
        T
        readNext()
            throws Exception;

        public
        final
        List< T >
        call()
            throws Exception
        {
            List< T > res = Lang.newList();

            is = new BufferedInputStream( openDataFile( fileNm ) );
            try 
            { 
                readHeader();

                while ( ! endCalled ) {
                    T t = readNext();
                    if ( t != null ) res.add( t );
                }

                return res;
            } 
            finally { is.close(); }
        }
    }
}
