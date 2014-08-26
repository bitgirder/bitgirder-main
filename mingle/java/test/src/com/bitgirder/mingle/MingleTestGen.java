package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.testing.TestData;

import java.io.InputStream;

import java.util.List;

public
final
class MingleTestGen
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static QualifiedTypeName TYPE_END =
        QualifiedTypeName.create( "mingle:testgen@v1/TestFileEnd" );

    public
    static
    abstract
    class StructFileReader< T >
    {
        private final String fname;

        protected
        StructFileReader( String fname )
        {
            this.fname = inputs.notNull( fname, "fname" );
        }

        // return null to skip this struct
        protected
        abstract
        T
        convertStruct( MingleStruct ms )
            throws Exception;

        private
        List< T >
        readStructs( MingleBinReader r )
            throws Exception
        {
            List< T > res = Lang.newList( 128 );

            while ( true )
            {
                MingleStruct ms = (MingleStruct) 
                    MingleTestMethods.readValue( r );

                if ( ms.getType().equals( TYPE_END ) ) break;
                
                T test = convertStruct( ms );
                if ( test != null ) res.add( test );
            }

            return res;
        }

        public
        final
        List< T >
        read()
            throws Exception
        {
            InputStream is = TestData.openDataFile( fname );
            try
            {
                MingleBinReader mbr = MingleBinReader.create( is );
                return readStructs( mbr );
            }
            finally { is.close(); }
        }
    }
}
