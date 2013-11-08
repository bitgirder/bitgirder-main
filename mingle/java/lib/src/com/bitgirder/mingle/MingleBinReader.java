package com.bitgirder.mingle;

import static com.bitgirder.mingle.MingleBinaryConstants.*;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;

import com.bitgirder.lang.path.ObjectPath;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.io.BinReader;

import java.io.IOException;
import java.io.InputStream;

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

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static byte[] RESTRICTION_TYPE_CODES = new byte[] {
        TC_NIL,
        TC_RANGE_RESTRICT,
        TC_REGEX_RESTRICT
    };

    private final static byte[] VAL_TYPE_CODES = new byte[] {
        TC_NIL,
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

    private final static byte[] TC_END_ARR = new byte[] { TC_END };

    private BinReader rd;

    private MingleBinReader( BinReader rd ) { this.rd = rd; }

    private
    static
    MingleBinaryException
    failf( String tmpl,
           Object... args )
    {
        return new MingleBinaryException( String.format( tmpl, args ) );
    }

    private
    int
    expectPosInt32( String errDesc )
        throws IOException
    {
        int i = rd.readInt();

        if ( i < 0 )
        {
            throw failf( "Value for %s is not a positive signed int32: %s",
                errDesc, Lang.toUint32String( i ) );
        }

        return i;
    }

    private byte nextTc() throws IOException { return rd.readByte(); }

    private
    byte
    acceptTc( byte tc,
              String desc,
              byte[] accpt )
        throws IOException
    {
        for ( byte a : accpt ) if ( a == tc ) { return tc; }

        if ( desc == null ) return (byte) -1;
        throw failf( "Expected %s but saw type code 0x%02x", desc, tc );
    }

    private
    byte
    nextTc( String desc,
            byte... accpt )
        throws IOException
    {
        return acceptTc( nextTc(), desc, accpt );
    }

    private
    boolean
    readJvBool( String desc )
        throws IOException
    {
        byte tc = nextTc( desc, TC_BOOL );
        return rd.readBoolean();
    }

    private
    MingleIdentifier
    processIdentifier()
        throws IOException
    {
        int sz = Lang.asOctet( rd.readByte() );

        String[] parts = new String[ sz ];
        for ( int i = 0; i < sz; ++i ) parts[ i ] = rd.readUtf8();

        return new MingleIdentifier( parts );
    }

    public
    MingleIdentifier
    readIdentifier()
        throws IOException
    {
        return (MingleIdentifier) readNext( "identifier", TC_ID );
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

    private
    MingleNamespace
    processNamespace()
        throws IOException
    {
        return new MingleNamespace( readIdentifiers(), readIdentifier() );
    }

    public
    MingleNamespace
    readNamespace()
        throws IOException
    {
        return (MingleNamespace) readNext( "namespace", TC_NS );
    }

    public
    DeclaredTypeName
    readDeclaredTypeName()
        throws IOException
    {
        return (DeclaredTypeName) readNext( "declared type name", TC_DECL_NM );
    }

    private
    DeclaredTypeName
    processDeclaredTypeName()
        throws IOException
    {
        return new DeclaredTypeName( rd.readUtf8() );
    }

    private
    QualifiedTypeName
    processQname()
        throws IOException
    {
        return new QualifiedTypeName( readNamespace(), readDeclaredTypeName() );
    }

    private
    final
    static
    class RangeVals
    {
        private boolean minClosed;
        private MingleValue min;
        private MingleValue max;
        private boolean maxClosed;

        private
        MingleValue
        convertNull( MingleValue mv )
        {
            return mv instanceof MingleNull ? null : mv;
        }

        private
        MingleRangeRestriction
        build( TypeName tn )
            throws MingleBinaryException
        {
            Class< ? extends MingleValue > typeTok = Mingle.valueClassFor( tn );

            if ( typeTok == null ) 
            {
                throw failf( "Unrecognized range target type: %s", tn );
            }

            return MingleRangeRestriction.createChecked(
                    minClosed, min, max, maxClosed, typeTok );
        }
    }

    private
    AtomicTypeReference
    processAtomicType()
        throws IOException
    {
        TypeName nm = readTypeName();

        MingleValueRestriction vr = null;
        byte tc = nextTc( "restriction", RESTRICTION_TYPE_CODES );

        if ( tc != TC_NIL ) 
        {
            Object obj = processNext( tc );
            if ( tc == TC_REGEX_RESTRICT ) vr = (MingleRegexRestriction) obj;
            else vr = ( (RangeVals) obj ).build( nm );
        } 
        
        return new AtomicTypeReference( nm, vr );
    }

    private
    RangeVals
    processRangeRestriction()
        throws IOException
    {
        RangeVals res = new RangeVals();

        res.minClosed = readJvBool( "min range inclusivity" );
        res.min = res.convertNull( readValue( "range min" ) );
        res.max = res.convertNull( readValue( "range max" ) );
        res.maxClosed = readJvBool( "max range inclusivity" );

        return res;
    }

    private
    MingleRegexRestriction
    processRegexRestriction()
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
    processListType()
        throws IOException
    {
        return new ListTypeReference( 
            readTypeReference(), 
            readJvBool( "list type empty/non-empty flag" ) 
        );
    }

    private
    NullableTypeReference
    processNullableType()
        throws IOException
    {
        return new NullableTypeReference( readTypeReference() );
    }

    private
    TypeName
    readTypeName()
        throws IOException
    {
        return (TypeName) readNext( "type name", TC_QN, TC_DECL_NM );
    }

    private
    ObjectPath< MingleIdentifier >
    processIdPath()
        throws IOException
    {
        ObjectPath< MingleIdentifier > res = ObjectPath.getRoot();

        while ( true )
        {
            byte tc = nextTc( "id path", TC_END, TC_ID, TC_ID_PATH_LIST_NODE );
    
            switch ( tc )
            {
                case TC_END: return res;
                case TC_ID: res = res.descend( readIdentifier() ); break;
    
                case TC_ID_PATH_LIST_NODE: 
                    int idx = expectPosInt32( "id path list index" );
                    res = res.startImmutableList( idx );
                    break;
            }
        }
    }

    private
    MingleTimestamp
    processTimestamp()
        throws IOException
    {
        long secs = rd.readLong();
        int nsec = rd.readInt();

        return MingleTimestamp.fromUnixNanos( secs, nsec );
    }

    private
    MingleEnum
    processEnum()
        throws IOException
    {
        return new MingleEnum( readQualifiedTypeName(), readIdentifier() );
    }

    private
    MingleSymbolMap
    processSymbolMap()
        throws IOException
    {
        Map< MingleIdentifier, MingleValue > m = Lang.newMap();

        while ( true )
        {
            byte tc = nextTc( "symbol map", TC_END, TC_MAP_PAIR );
            if ( tc == TC_END ) return new MingleSymbolMap( m );
            m.put( readIdentifier(), readValue() );
        }
    }

    // read but drop size val
    private void readSize() throws IOException { rd.readInt(); }

    private
    MingleStruct
    processStruct()
        throws IOException
    {
        readSize();
        return new MingleStruct( readQualifiedTypeName(), processSymbolMap() );
    }

    private
    MingleList
    processList()
        throws IOException
    {
        readSize();

        List< MingleValue > l = Lang.newList();

        while ( true )
        {
            byte tc = nextTc();

            if ( acceptTc( tc, null, TC_END_ARR ) == TC_END )
            {
                return MingleList.createLive( l );
            }

            acceptTc( tc, "list value", VAL_TYPE_CODES );
            l.add( (MingleValue) processNext( tc ) );
        }
    }

    private
    Object
    processNext( byte tc )
        throws IOException
    {
        switch ( tc )
        {
            case TC_ID: return processIdentifier();
            case TC_NS: return processNamespace();
            case TC_DECL_NM: return processDeclaredTypeName();
            case TC_QN: return processQname();
            case TC_ATOM_TYP: return processAtomicType();
            case TC_RANGE_RESTRICT: return processRangeRestriction();
            case TC_REGEX_RESTRICT: return processRegexRestriction();
            case TC_LIST_TYP: return processListType();
            case TC_NULLABLE_TYP: return processNullableType();
            case TC_ID_PATH: return processIdPath();
            case TC_NIL: return MingleNull.getInstance();
            case TC_BOOL: return MingleBoolean.valueOf( rd.readBoolean() ); 
            case TC_INT32: return new MingleInt32( rd.readInt() );
            case TC_UINT32: return new MingleUint32( rd.readInt() );
            case TC_INT64: return new MingleInt64( rd.readLong() );
            case TC_UINT64: return new MingleUint64( rd.readLong() );
            case TC_FLOAT32: return new MingleFloat32( rd.readFloat() );
            case TC_FLOAT64: return new MingleFloat64( rd.readDouble() );
            case TC_BUFFER: return new MingleBuffer( rd.readByteArray() );
            case TC_STRING: return new MingleString( rd.readUtf8() );
            case TC_TIMESTAMP: return processTimestamp();
            case TC_ENUM: return processEnum();
            case TC_SYM_MAP: return processSymbolMap();
            case TC_STRUCT: return processStruct();
            case TC_LIST: return processList();
            default: throw failf( "Unrecognized type code: 0x%02x", tc );
        }
    }

    private
    Object
    readNext( String desc,
              byte... accpt )
        throws IOException
    {
        return processNext( nextTc( desc, accpt ) );
    }

    private
    MingleValue
    readValue( String desc )
        throws IOException
    {
        return (MingleValue) readNext( desc, VAL_TYPE_CODES );
    }

    public
    MingleValue
    readValue()
        throws IOException
    {
        return readValue( "mingle value" );
    }

    public
    QualifiedTypeName
    readQualifiedTypeName()
        throws IOException
    {
        return (QualifiedTypeName) readNext( "qname", TC_QN );
    }

    public
    MingleTypeReference
    readTypeReference()
        throws IOException
    {
        byte tc = nextTc( "type reference", 
            TC_ATOM_TYP, TC_LIST_TYP, TC_NULLABLE_TYP);

        return (MingleTypeReference) processNext( tc );
    }

    public
    AtomicTypeReference
    readAtomicTypeReference()
        throws IOException
    {
        byte tc = nextTc( "atomic type reference", TC_ATOM_TYP );
        return (AtomicTypeReference) processNext( tc );
    }

    public
    ObjectPath< MingleIdentifier >
    readIdPath()
        throws IOException
    {
        byte tc = nextTc( "identifier path", TC_ID_PATH );
        Object res = processNext( tc );

        return Lang.< ObjectPath< MingleIdentifier > >castUnchecked( res );
    }

    public
    static
    MingleBinReader
    create( InputStream is )
    {
        inputs.notNull( is, "is" );
        return new MingleBinReader( BinReader.asReaderLe( is ) );
    }
}
