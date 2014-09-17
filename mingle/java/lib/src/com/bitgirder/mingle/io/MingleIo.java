package com.bitgirder.mingle.io;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.mingle.MingleValue;
import com.bitgirder.mingle.MingleBinReader;
import com.bitgirder.mingle.MingleBinWriter;
import com.bitgirder.mingle.MingleIdentifier;
import com.bitgirder.mingle.QualifiedTypeName;
import com.bitgirder.mingle.ListTypeReference;

import static com.bitgirder.mingle.MingleBinaryConstants.*;

import com.bitgirder.mingle.reactor.MingleReactor;
import com.bitgirder.mingle.reactor.MingleReactorEvent;

import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;

public
final
class MingleIo
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private MingleIo() {}

    private
    final
    static
    class Feed
    {
        private final MingleReactor rct;
        private final MingleBinReader rd;

        private final MingleReactorEvent ev = new MingleReactorEvent();

        private 
        Feed( MingleReactor rct,
              MingleBinReader rd ) 
        { 
            this.rct = rct; 
            this.rd = rd;
        }

        private void feedNext() throws Exception { rct.processEvent( ev ); }
    
        private
        void
        feedScalar( byte tc )
            throws Exception
        {
            ev.setValue( rd.expectScalar( tc ) );
            feedNext();
        }
    
        private
        void
        feedList()
            throws Exception
        {
            ev.setStartList( rd.readListTypeReference() );
            feedNext();
    
            while ( true )
            {
                byte tc = rd.nextTypeCode();
    
                if ( tc == TC_END ) {
                    ev.setEnd();
                    feedNext();
                    return;
                }
    
                feedValue( tc );
            }
        }
    
        private
        void
        feedSymbolMapPairs()
            throws Exception
        {
            while ( true )
            {
                byte tc = rd.nextTypeCode( "symbol map", TC_END, TC_FIELD );
                if ( tc == TC_END ) break;
    
                ev.setStartField( rd.readIdentifier() );
                feedNext();
    
                feedValue( rd.nextTypeCode() );
            }
    
            ev.setEnd();
            feedNext();
        }
    
        private
        void
        feedSymbolMap()
            throws Exception
        {
            ev.setStartMap();
            feedNext();
    
            feedSymbolMapPairs();
        } 
    
        private
        void
        feedStruct()
            throws Exception
        {
            ev.setStartStruct( rd.readQualifiedTypeName() );
            feedNext();
    
            feedSymbolMapPairs();
        }
    
        private
        void
        feedValue( byte tc )
            throws Exception
        {
            switch ( tc ) {
            case TC_NULL:
            case TC_BOOL:
            case TC_BUFFER:
            case TC_STRING:
            case TC_INT32:
            case TC_UINT32:
            case TC_INT64:
            case TC_UINT64:
            case TC_FLOAT32:
            case TC_FLOAT64:
            case TC_TIMESTAMP:
            case TC_ENUM:
                feedScalar( tc );
                break;
            case TC_LIST: feedList(); break;
            case TC_SYM_MAP: feedSymbolMap(); break;
            case TC_STRUCT: feedStruct(); break;
            default: throw rd.failLastTypeCode( "mingle value", tc );
            }
        }
    }

    public
    static
    void
    feedValue( InputStream is,
               MingleReactor rct )
        throws Exception
    {
        inputs.notNull( is, "is" );
        inputs.notNull( rct, "rct" );

        MingleBinReader rd = MingleBinReader.create( is );

        Feed f = new Feed( rct, rd );
        f.feedValue( f.rd.nextTypeCode() );
    }

    private
    final
    static
    class WriteReactor
    implements MingleReactor
    {
        private final MingleBinWriter w;

        private WriteReactor( MingleBinWriter w ) { this.w = w; }

        private 
        void 
        writeEnd() 
            throws IOException 
        { 
            w.writeTypeCode( TC_END ); 
        }

        private
        void
        writeListStart( ListTypeReference lt )
            throws IOException
        {
            w.writeTypeCode( TC_LIST );
            w.writeTypeReference( lt );
        }
        
        private
        void
        writeFieldStart( MingleIdentifier fld )
            throws IOException
        {
            w.writeTypeCode( TC_FIELD );
            w.writeIdentifier( fld );
        }

        private
        void
        writeStructStart( QualifiedTypeName qn )
            throws IOException
        {
            w.writeTypeCode( TC_STRUCT );
            w.writeQualifiedTypeName( qn );
        }

        public
        void
        processEvent( MingleReactorEvent ev )
            throws Exception
        {
            switch ( ev.type() ) {
            case VALUE: w.writeScalar( ev.value() ); return;
            case LIST_START: writeListStart( ev.listType() ); return;
            case MAP_START: w.writeTypeCode( TC_SYM_MAP ); return;
            case STRUCT_START: writeStructStart( ev.structType() ); return;
            case FIELD_START: writeFieldStart( ev.field() ); return;
            case END: writeEnd(); return;
            default: state.failf( "unhandled type: %s", ev.type() );
            }
        }
    }

    public
    static
    MingleReactor
    createWriteReactor( OutputStream os )
    {
        inputs.notNull( os, "os" );

        MingleBinWriter w = MingleBinWriter.create( os );
        return new WriteReactor( w );
    }
}
