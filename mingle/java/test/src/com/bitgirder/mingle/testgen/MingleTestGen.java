package com.bitgirder.mingle.testgen;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.mingle.QualifiedTypeName;
import com.bitgirder.mingle.MingleStruct;

import com.bitgirder.mingle.reactor.BuildReactor;
import com.bitgirder.mingle.reactor.ValueBuildFactory;

import com.bitgirder.mingle.io.MingleIo;

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
        readStructs( InputStream is )
            throws Exception
        {
            List< T > res = Lang.newList( 128 );

            while ( true )
            {
                BuildReactor br = new BuildReactor.Builder().
                    setFactory( new ValueBuildFactory() ).
                    build();

                MingleIo.feedValue( is, br );

                MingleStruct ms = (MingleStruct) br.value();

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
            try { return readStructs( is ); }
            finally { is.close(); }
        }
    }
}
