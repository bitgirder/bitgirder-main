package com.bitgirder.mingle;

import static com.bitgirder.mingle.MingleBinaryConstants.*;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.ObjectReceiver;

import com.bitgirder.lang.path.ListPath;
import com.bitgirder.lang.path.DictionaryPath;

import com.bitgirder.io.BinWriter;

import java.util.List;
import java.util.Map;

import java.io.OutputStream;
import java.io.IOException;

public
final
class MingleBinWriter
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final BinWriter w;

    private MingleBinWriter( BinWriter w ) { this.w = w; }

    public
    void
    writeTypeCode( byte tc )
        throws IOException
    {
        w.writeByte( tc );
    }

    private
    void
    writeUint8( int i )
        throws IOException
    {
        w.writeByte( Lang.fromOctet( i ) );
    }

    public
    void
    writeIdentifier( MingleIdentifier id )
        throws IOException
    {
        inputs.notNull( id, "id" );

        writeTypeCode( TC_ID );

        String[] parts = id.getPartsArray();

        writeUint8( parts.length );
        for ( String part : parts ) w.writeUtf8( part );
    }

    private
    void
    writeIdentifiers( List< MingleIdentifier > ids )
        throws IOException
    {
        writeUint8( ids.size() );
        for ( MingleIdentifier id : ids ) writeIdentifier( id );
    }

    public
    void
    writeDeclaredTypeName( DeclaredTypeName nm )
        throws IOException
    {
        inputs.notNull( nm, "nm" );

        writeTypeCode( TC_DECL_NM );
        w.writeUtf8( nm.getExternalForm() );
    }

    private
    void
    writeTypeName( TypeName nm )
        throws IOException
    {
        if ( nm instanceof QualifiedTypeName ) {
            writeQualifiedTypeName( (QualifiedTypeName) nm );
        } else if ( nm instanceof DeclaredTypeName ) {
            writeDeclaredTypeName( (DeclaredTypeName) nm );
        } else {
            state.failf( "unhandled type name: %s", nm.getClass() );
        }
    }

    public
    void
    writeNamespace( MingleNamespace ns )
        throws IOException
    {
        writeTypeCode( TC_NS );
        
        writeIdentifiers( ns.getParts() );
        writeIdentifier( ns.getVersion() );
    }

    public
    void
    writeQualifiedTypeName( QualifiedTypeName qn )
        throws IOException
    {
        inputs.notNull( qn, "qn" );

        writeTypeCode( TC_QN );
        writeNamespace( qn.getNamespace() );
        writeDeclaredTypeName( qn.getName() );
    }

    public
    void
    writeScalar( MingleValue mv )
        throws IOException
    {
        inputs.notNull( mv, "mv" );

        if ( mv instanceof MingleNull ) writeTypeCode( TC_NULL );
        else if ( mv instanceof MingleBoolean ) {
            writeTypeCode( TC_BOOL );
            w.writeBoolean( ( (MingleBoolean) mv ).booleanValue() );
        } else if ( mv instanceof MingleString ) {
            writeTypeCode( TC_STRING );
            w.writeUtf8( ( (MingleString) mv ).toString() );
        } else if ( mv instanceof MingleBuffer ) {
            writeTypeCode( TC_BUFFER );
            w.writeByteArray( ( (MingleBuffer) mv ).array() );
        } else if ( mv instanceof MingleInt32 ) {
            writeTypeCode( TC_INT32 );
            w.writeInt( ( (MingleInt32) mv ).intValue() );
        } else if ( mv instanceof MingleUint32 ) {
            writeTypeCode( TC_UINT32 );
            w.writeInt( ( (MingleUint32) mv ).intValue() );
        } else if ( mv instanceof MingleInt64 ) {
            writeTypeCode( TC_INT64 );
            w.writeLong( ( (MingleInt64) mv ).longValue() );
        } else if ( mv instanceof MingleUint64 ) {
            writeTypeCode( TC_UINT64 );
            w.writeLong( ( (MingleUint64) mv ).longValue() );
        } else if ( mv instanceof MingleFloat32 ) {
            writeTypeCode( TC_FLOAT32 );
            w.writeFloat( ( (MingleFloat32) mv ).floatValue() );
        } else if ( mv instanceof MingleFloat64 ) {
            writeTypeCode( TC_FLOAT64 );
            w.writeDouble( ( (MingleFloat64) mv ).doubleValue() );
        } else if ( mv instanceof MingleTimestamp ) {
            writeTypeCode( TC_TIMESTAMP );
            MingleTimestamp ts = (MingleTimestamp) mv;
            w.writeLong( ts.seconds() );
            w.writeInt( ts.nanos() );
        } else if ( mv instanceof MingleEnum ) {
            writeTypeCode( TC_ENUM );
            MingleEnum me = (MingleEnum) mv;
            writeQualifiedTypeName( me.getType() );
            writeIdentifier( me.getValue() );
        } else {
            state.failf( "not a scalar: %s", mv.getClass() );
        }
    }

    private
    void
    writeRangeValue( MingleValue mv )
        throws IOException
    {
        writeScalar( mv == null ? MingleNull.getInstance() : mv );
    }

    private
    void
    writeRestriction( MingleRangeRestriction r )
        throws IOException
    {
        writeTypeCode( TC_RANGE_RESTRICT );
        w.writeBoolean( r.minClosed );
        writeRangeValue( r.min );
        writeRangeValue( r.max );
        w.writeBoolean( r.maxClosed );
    }

    private
    void
    writeRestriction( MingleRegexRestriction r )
        throws IOException
    {
        writeTypeCode( TC_REGEX_RESTRICT );
        w.writeUtf8( r.pattern().pattern() );
    }

    private
    void
    writeAtomicTypeReference( AtomicTypeReference typ )
        throws IOException
    {
        writeTypeCode( TC_ATOM_TYP );
        writeTypeName( typ.getName() );

        MingleValueRestriction r = typ.getRestriction();

        if ( r == null ) writeTypeCode( TC_NULL ); 
        else if ( r instanceof MingleRangeRestriction ) {
            writeRestriction( (MingleRangeRestriction) r );
        } else if ( r instanceof MingleRegexRestriction ) {
            writeRestriction( (MingleRegexRestriction) r );
        } else state.failf( "unhandled restriction: %s", r.getClass() );
    }

    private
    void
    writeListTypeReference( ListTypeReference typ )
        throws IOException
    {
        writeTypeCode( TC_LIST_TYP );
        writeTypeReference( typ.getElementType() );
        w.writeBoolean( typ.allowsEmpty() );
    }

    private
    void
    writeNullableTypeReference( NullableTypeReference typ )
        throws IOException
    {
        writeTypeCode( TC_NULLABLE_TYP );
        writeTypeReference( typ.getValueType() );
    }

    private
    void
    writePointerTypeReference( PointerTypeReference typ )
        throws IOException
    {
        writeTypeCode( TC_POINTER_TYP );
        writeTypeReference( typ.getType() );
    }

    public
    void
    writeTypeReference( MingleTypeReference typ )
        throws IOException
    {
        inputs.notNull( typ, "typ" );
        
        if ( typ instanceof AtomicTypeReference ) {
            writeAtomicTypeReference( (AtomicTypeReference) typ );
        } else if ( typ instanceof ListTypeReference ) {
            writeListTypeReference( (ListTypeReference) typ );
        } else if ( typ instanceof NullableTypeReference ) {
            writeNullableTypeReference( (NullableTypeReference) typ );
        } else if ( typ instanceof PointerTypeReference ) {
            writePointerTypeReference( (PointerTypeReference) typ );
        } else {
            state.failf( "unhandled type reference: %s", typ.getClass() );
        }
    }

    public
    static
    MingleBinWriter
    create( OutputStream os )
    {
        inputs.notNull( os, "os" );
        return new MingleBinWriter( BinWriter.asWriterLe( os ) );
    }
}
