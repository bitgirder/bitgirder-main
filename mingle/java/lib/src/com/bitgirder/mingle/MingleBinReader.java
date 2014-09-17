package com.bitgirder.mingle;

import static com.bitgirder.mingle.MingleBinaryConstants.*;

import static com.bitgirder.log.CodeLoggers.Statics.*;

import com.bitgirder.lang.Lang;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.io.BinReader;
import com.bitgirder.io.CountingInputStream;

import java.io.IOException;
import java.io.InputStream;
import java.io.ByteArrayInputStream;

import java.util.Map;
import java.util.List;

import java.util.regex.Pattern;
import java.util.regex.PatternSyntaxException;

public
final
class MingleBinReader
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static byte[] RESTRICTION_TYPE_CODES = new byte[] {
        TC_NULL,
        TC_RANGE_RESTRICT,
        TC_REGEX_RESTRICT
    };

    private final static byte[] VAL_TYPE_CODES = new byte[] {
        TC_NULL,
        TC_BOOL,
        TC_INT32,
        TC_INT64,
        TC_UINT32,
        TC_UINT64,
        TC_FLOAT32,
        TC_FLOAT64,
        TC_STRING,
        TC_BUFFER,
        TC_TIMESTAMP,
        TC_ENUM,
        TC_SYM_MAP,
        TC_STRUCT,
        TC_LIST
    };

    private final static byte[] SCALAR_TYPE_CODES = new byte[] {
        TC_NULL,
        TC_BOOL,
        TC_INT32,
        TC_INT64,
        TC_UINT32,
        TC_UINT64,
        TC_FLOAT32,
        TC_FLOAT64,
        TC_STRING,
        TC_BUFFER,
        TC_TIMESTAMP,
        TC_ENUM
    };

    private final static byte[] RANGE_VAL_TYPE_CODES = new byte[] {
        TC_NULL,
        TC_INT32,
        TC_INT64,
        TC_UINT32,
        TC_UINT64,
        TC_FLOAT32,
        TC_FLOAT64,
        TC_STRING,
        TC_TIMESTAMP
    };

    private final static byte[] TYPE_REF_TYPE_CODES = new byte[] {
        TC_ATOM_TYP, TC_LIST_TYP, TC_NULLABLE_TYP, TC_POINTER_TYP
    };

    private final CountingInputStream cis;
    private final BinReader rd;

    private 
    MingleBinReader( CountingInputStream cis )
    {
        this.cis = cis;
        this.rd = BinReader.asReaderLe( cis );
    }

    private
    static
    MingleBinaryException
    fail( long pos,
          String msg )
    {
        String errMsg = String.format( "[offset %d]: %s", pos, msg );
        return new MingleBinaryException( errMsg );
    }

    private
    static
    MingleBinaryException
    failf( long pos,
           String tmpl,
           Object... args )
    {
        return fail( pos, String.format( tmpl, args ) );
    }

    private
    static
    MingleBinaryException
    failTypeCode( long pos,
                  String desc,
                  byte tc )
    {
        return failf( pos, "Expected %s but saw type code 0x%02x", 
            desc, tc );
    }

    public
    MingleBinaryException
    failLastTypeCode( String desc,
                      byte tc )
    {
        return failTypeCode( cis.position() - 1, desc, tc );
    }

    private
    int
    expectPosInt32( String errDesc )
        throws IOException
    {
        long pos = cis.position();
        int i = rd.readInt();

        if ( i < 0 )
        {
            throw failf( pos, "Value for %s is not a positive signed int32: %s",
                errDesc, Lang.toUint32String( i ) );
        }

        return i;
    }

    public byte nextTypeCode() throws IOException { return rd.readByte(); }

    // assumes that tc was the most recently read byte, in terms of setting the
    // error position
    private
    byte
    acceptTypeCode( byte tc,
              String desc,
              byte... accpt )
        throws IOException
    {
        for ( byte a : accpt ) if ( a == tc ) { return tc; }

        if ( desc == null ) return (byte) -1;

        long errPos = cis.position() - 1L;
        throw failTypeCode( errPos, desc, tc );
    }

    public
    byte
    nextTypeCode( String desc,
                  byte... accpt )
        throws IOException
    {
        return acceptTypeCode( nextTypeCode(), desc, accpt );
    }

    public
    MingleIdentifier
    readIdentifier()
        throws IOException
    {
        nextTypeCode( "identifier", TC_ID );

        int sz = Lang.asOctet( rd.readByte() );

        String[] parts = new String[ sz ];

        for ( int i = 0; i < sz; ++i ) 
        {
            long pos = cis.position();
            String part = rd.readUtf8();

            if ( ! MingleIdentifier.isValidPart( part ) ) {
                throw failf( pos, "invalid identifier part: %s", part );
            }

            parts[ i ] = part;
        }

        return new MingleIdentifier( parts );
    }

    private
    MingleIdentifier[]
    readIdentifiers()
        throws IOException
    {
        int sz = Lang.asOctet( rd.readByte() );

        MingleIdentifier[] res = new MingleIdentifier[ sz ];
        for ( int i = 0; i < sz; ++i ) res[ i ] = readIdentifier();

        return res;
    }

    public
    MingleNamespace
    readNamespace()
        throws IOException
    {
        nextTypeCode( "namespace", TC_NS );
        return new MingleNamespace( readIdentifiers(), readIdentifier() );
    }

    public
    DeclaredTypeName
    readDeclaredTypeName()
        throws IOException
    {
        nextTypeCode( "declared type name", TC_DECL_NM );

        long off = cis.position();

        try { return DeclaredTypeName.parse( rd.readUtf8() ); }
        catch ( MingleSyntaxException mse ) {
            throw fail( off, mse.getMessage() );
        }
    }

    private
    final
    static
    class RangeVals
    {
        private final long errPos;

        private boolean minClosed;
        private MingleValue min;
        private MingleValue max;
        private boolean maxClosed;
        private QualifiedTypeName qn;

        private RangeVals( long errPos ) { this.errPos = errPos; }

        private
        MingleValue
        convertNull( MingleValue mv )
        {
            return mv instanceof MingleNull ? null : mv;
        }

        private
        MingleRangeRestriction
        build()
            throws MingleBinaryException
        {
            state.notNull( qn, "qn" );
            Class< ? extends MingleValue > typeTok = Mingle.valueClassFor( qn );

            if ( typeTok == null ) {
                throw failf( errPos, "Unrecognized range target type: %s", qn );
            }

            return MingleRangeRestriction.createChecked(
                    minClosed, min, max, maxClosed, typeTok );
        }
    }

    private
    AtomicTypeReference
    implReadAtomicTypeReference( boolean tcChecked )
        throws IOException
    {
        if ( ! tcChecked ) nextTypeCode( "atomic type reference", TC_ATOM_TYP );

        QualifiedTypeName nm = readQualifiedTypeName();

        MingleValueRestriction vr = null;
        byte tc = nextTypeCode( "restriction", RESTRICTION_TYPE_CODES );

        switch ( tc ) {
        case TC_REGEX_RESTRICT: vr = readRegexRestriction(); break;
        case TC_RANGE_RESTRICT: vr = readRangeRestriction( nm ); break;
        } 
        
        return new AtomicTypeReference( nm, vr );
    }

    public
    AtomicTypeReference
    readAtomicTypeReference()
        throws IOException
    {
        return implReadAtomicTypeReference( false );
    }

    private
    MingleValue
    readRangeValue( String desc )
        throws IOException
    {
        return expectScalar( nextTypeCode( desc, RANGE_VAL_TYPE_CODES ) );
    }

    private
    MingleRangeRestriction
    readRangeRestriction( QualifiedTypeName nm )
        throws IOException
    {
        RangeVals res = new RangeVals( cis.position() );

        res.minClosed = rd.readBoolean();
        res.min = res.convertNull( readRangeValue( "range min" ) );
        res.max = res.convertNull( readRangeValue( "range max" ) );
        res.maxClosed = rd.readBoolean();
        res.qn = nm;

        return res.build();
    }

    private
    MingleRegexRestriction
    readRegexRestriction()
        throws IOException
    {
        try
        {
            Pattern pat = Pattern.compile( rd.readUtf8() );
            return MingleRegexRestriction.create( pat );
        }
        catch ( PatternSyntaxException pse )
        {
            throw new MingleBinaryException( "Invalid regex: " + pse, pse );
        }
    }

    private
    ListTypeReference
    implReadListTypeReference()
        throws IOException
    {
        return new ListTypeReference( readTypeReference(), rd.readBoolean() );
    }

    public
    ListTypeReference
    readListTypeReference()
        throws IOException
    {
        nextTypeCode( "list type reference", TC_LIST_TYP );
        return implReadListTypeReference();
    }

    private
    MingleTimestamp
    readTimestamp()
        throws IOException
    {
        long secs = rd.readLong();
        int nsec = rd.readInt();

        return MingleTimestamp.fromUnixNanos( secs, nsec );
    }

    private
    MingleEnum
    readEnum()
        throws IOException
    {
        return new MingleEnum( readQualifiedTypeName(), readIdentifier() );
    }

    public
    MingleValue
    expectScalar( byte tc )
        throws IOException
    {
        switch ( tc ) {
        case TC_NULL: return MingleNull.getInstance();
        case TC_BOOL: return MingleBoolean.valueOf( rd.readBoolean() ); 
        case TC_INT32: return new MingleInt32( rd.readInt() );
        case TC_UINT32: return new MingleUint32( rd.readInt() );
        case TC_INT64: return new MingleInt64( rd.readLong() );
        case TC_UINT64: return new MingleUint64( rd.readLong() );
        case TC_FLOAT32: return new MingleFloat32( rd.readFloat() );
        case TC_FLOAT64: return new MingleFloat64( rd.readDouble() );
        case TC_BUFFER: return new MingleBuffer( rd.readByteArray() );
        case TC_STRING: return new MingleString( rd.readUtf8() );
        case TC_TIMESTAMP: return readTimestamp();
        case TC_ENUM: return readEnum();
        }
        throw state.failf( "unhandled scalar type: 0x%02x", tc );
    }

    public
    MingleValue
    readScalar()
        throws IOException
    {
        byte tc = nextTypeCode( "scalar", SCALAR_TYPE_CODES );
        return expectScalar( tc );
    }

    public
    QualifiedTypeName
    readQualifiedTypeName()
        throws IOException
    {
        nextTypeCode( "qname", TC_QN );
        return new QualifiedTypeName( readNamespace(), readDeclaredTypeName() );
    }

    public
    MingleTypeReference
    readTypeReference()
        throws IOException
    {
        byte tc = nextTypeCode( "type reference", TYPE_REF_TYPE_CODES );

        switch ( tc ) {
        case TC_ATOM_TYP: return implReadAtomicTypeReference( true );
        case TC_LIST_TYP: return implReadListTypeReference();
        case TC_NULLABLE_TYP:
            return new NullableTypeReference( readTypeReference() );
        case TC_POINTER_TYP:
            return new PointerTypeReference( readTypeReference() );
        }

        throw state.failf( "unhandled tc: 0x%02x", tc );
    }

    public
    static
    MingleBinReader
    create( InputStream is )
    {
        inputs.notNull( is, "is" );
        return new MingleBinReader( new CountingInputStream( is ) );
    }

    public
    static
    MingleBinReader
    create( byte[] buf )
    {
        inputs.notNull( buf, "buf" );
        return create( new ByteArrayInputStream( buf ) );
    }

    public
    static
    MingleBinReader
    create( MingleBuffer mb )
    {
        inputs.notNull( mb, "mb" );
        return create( mb.array() );
    }
}
