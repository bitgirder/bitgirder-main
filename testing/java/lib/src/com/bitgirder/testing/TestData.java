package com.bitgirder.testing;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.io.InputStream;
import java.io.IOException;
import java.io.File;
import java.io.FileInputStream;
import java.io.FileNotFoundException;

public
final
class TestData
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static String PROP_TEST_DATA_PATH = "bitgirder.testData.path";

    private TestData() {}

    public
    static
    String
    getTestPath()
    {
        String res = System.getProperty( PROP_TEST_DATA_PATH );
        return res == null ? "" : res;
    }

    public
    static
    File
    expectFile( String nm )
        throws FileNotFoundException
    {
        inputs.notNull( nm, "nm" );

        File res = new File( nm );
        if ( res.exists() ) return res;

        String[] dirs = getTestPath().split( File.pathSeparator );

        for ( String dir : dirs )
        {
            res = new File( dir + File.separatorChar + nm );
            if ( res.exists() ) return res;
        }

        throw new FileNotFoundException( "Cannot find test data file: " + nm );
    }

    public
    static
    InputStream
    openFile( String nm )
        throws IOException
    {
        return new FileInputStream( expectFile( nm ) );
    }
}
