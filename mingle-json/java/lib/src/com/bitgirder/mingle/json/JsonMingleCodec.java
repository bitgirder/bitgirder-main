package com.bitgirder.mingle.json;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.path.ObjectPath;
import com.bitgirder.lang.path.ImmutableListPath;
import com.bitgirder.lang.path.ObjectPaths;

import com.bitgirder.mingle.model.AtomicTypeReference;
import com.bitgirder.mingle.model.MingleBoolean;
import com.bitgirder.mingle.model.MingleBuffer;
import com.bitgirder.mingle.model.MingleEnum;
import com.bitgirder.mingle.model.MingleException;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleIdentifierFormat;
import com.bitgirder.mingle.model.MingleInt64;
import com.bitgirder.mingle.model.MingleInt32;
import com.bitgirder.mingle.model.MingleDouble;
import com.bitgirder.mingle.model.MingleFloat;
import com.bitgirder.mingle.model.MingleList;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleNamespace;
import com.bitgirder.mingle.model.MingleNull;
import com.bitgirder.mingle.model.MingleString;
import com.bitgirder.mingle.model.MingleStruct;
import com.bitgirder.mingle.model.MingleStructBuilder;
import com.bitgirder.mingle.model.MingleStructure;
import com.bitgirder.mingle.model.MingleStructureBuilder;
import com.bitgirder.mingle.model.MingleSymbolMap;
import com.bitgirder.mingle.model.MingleSymbolMapBuilder;
import com.bitgirder.mingle.model.MingleTimestamp;
import com.bitgirder.mingle.model.MingleTypeName;
import com.bitgirder.mingle.model.MingleTypeReference;
import com.bitgirder.mingle.model.MingleValue;

import com.bitgirder.mingle.codec.MingleCodec;
import com.bitgirder.mingle.codec.MingleEncoder;
import com.bitgirder.mingle.codec.MingleDecoder;
import com.bitgirder.mingle.codec.MingleCodecException;

import com.bitgirder.mingle.parser.MingleParsers;

import com.bitgirder.parser.SyntaxException;

import com.bitgirder.io.Base64Encoder;

import com.bitgirder.json.JsonString;
import com.bitgirder.json.JsonValue;
import com.bitgirder.json.JsonNumber;
import com.bitgirder.json.JsonObject;
import com.bitgirder.json.JsonNull;
import com.bitgirder.json.JsonArray;
import com.bitgirder.json.JsonBoolean;
import com.bitgirder.json.JsonSerializer;
import com.bitgirder.json.JsonParser;
import com.bitgirder.json.JsonParserFactory;
import com.bitgirder.json.JsonText;

import java.util.List;
import java.util.Map;

import java.nio.ByteBuffer;

import java.math.BigInteger;
import java.math.BigDecimal;

public
final
class JsonMingleCodec
implements MingleCodec
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static Base64Encoder enc = new Base64Encoder();

    private final static JsonParserFactory jpf = JsonParserFactory.create();

    private final static JsonString KEY_TYPE = JsonString.create( "$type" );

    private final static JsonString KEY_CONSTANT = 
        JsonString.create( "$constant" );

    private final static ObjectPath< JsonString > ROOT_PATH =
        ObjectPath.getRoot();

    private final MingleIdentifierFormat fmt;
    private final JsonSerializer.Options serOpts;
    private final boolean omitTypeFields;
    private final boolean expandEnums;

    private 
    JsonMingleCodec( Builder b )
    {
        this.fmt = inputs.notNull( b.fmt, "fmt" );
        this.serOpts = b.serOpts;
        this.omitTypeFields = b.omitTypeFields;
        this.expandEnums = b.expandEnums;
    }

    private
    < T >
    T
    cast( Class< T > cls,
          Object obj,
          String msg,
          ObjectPath< JsonString > loc )
        throws MingleCodecException
    {
        if ( cls.isInstance( obj ) ) return cls.cast( obj );
        else throw asCodecException( msg, loc );
    }

    private
    JsonObject.Builder
    addType( JsonObject.Builder jsonBld,
             MingleTypeReference typ )
    {
        jsonBld.addMember( KEY_TYPE, asJsonString( typ ) );
        return jsonBld;
    }

    private
    JsonNumber
    asJsonNumber( MingleInt64 mgInt )
    {
        return JsonNumber.forNumber( Long.valueOf( mgInt.longValue() ) );
    }

    private
    JsonNumber
    asJsonNumber( MingleInt32 mgInt )
    {
        return JsonNumber.forNumber( Integer.valueOf( mgInt.intValue() ) );
    }

    private
    JsonNumber
    asJsonNumber( MingleDouble mgDec )
    {
        return JsonNumber.forNumber( Double.valueOf( mgDec.doubleValue() ) );
    }

    private
    JsonNumber
    asJsonNumber( MingleFloat mgDec )
    {
        return JsonNumber.forNumber( Float.valueOf( mgDec.floatValue() ) );
    }

    private
    JsonBoolean
    asJsonBoolean( MingleBoolean mgBool )
    {
        return JsonBoolean.valueOf( mgBool.booleanValue() );
    }

    private
    JsonString
    asJsonString( MingleString mgStr )
    {
        return JsonString.create( mgStr );
    }

    private
    JsonValue
    asJsonBuffer( MingleBuffer mgBuf )
    {
        CharSequence b64 = enc.encode( mgBuf.getByteBuffer() );
        return JsonString.create( b64 );
    }

    private
    JsonValue
    asJsonValue( MingleTimestamp mgTs )
    {
        return JsonString.create( mgTs.getRfc3339String() );
    }

    private
    JsonArray
    asJsonArray( MingleList mgList )
    {
        JsonArray.Builder b = new JsonArray.Builder();

        for ( MingleValue mgVal : mgList )
        {
            JsonValue jv = asJsonValue( mgVal );
            b.add( jv );
        }

        return b.build();
    }

    private
    void
    buildFields( JsonObject.Builder jsonBld,
                 MingleSymbolMap fields )
    {
        for ( MingleIdentifier fld : fields.getFields() )
        {
            JsonString key = asJsonString( fld );
            JsonValue val = asJsonValue( fields.get( fld ) );

            jsonBld.addMember( key, val );
        }
    }

    private
    JsonObject
    asJsonObject( MingleStructure ms )
    {
        JsonObject.Builder jsonBld = new JsonObject.Builder();

        if ( ! omitTypeFields ) addType( jsonBld, ms.getType() );

        buildFields( jsonBld, ms.getFields() );

        return jsonBld.build();
    }

    private
    JsonObject
    asJsonException( MingleException me )
    {
        return asJsonObject( me );
    }

    private
    JsonObject
    asJsonObject( MingleSymbolMap msm )
    {
        JsonObject.Builder b = new JsonObject.Builder(); 
        buildFields( b, msm );

        return b.build();
    }

    // currently hardcoding that we serialize only the identifier; later callers
    // will be allowed to parameterize this behavior. Other options might be to
    // encode the fully qualified enum constant as a string, or even to convert
    // it to an object with $type and $constant fields, etc)
    private
    JsonValue
    asJsonValue( MingleEnum e )
    {
        if ( expandEnums )
        {
            return
                addType( new JsonObject.Builder(), e.getType() ).
                addMember( KEY_CONSTANT, asJsonString( e.getValue() ) ).
                build();
        }
        else return asJsonString( e.getValue() );
    }

    private
    JsonValue
    asJsonValue( MingleValue mv )
    {
        if ( mv instanceof MingleInt64 )
        {
            return asJsonNumber( (MingleInt64) mv );
        }
        else if ( mv instanceof MingleInt32 )
        {
            return asJsonNumber( (MingleInt32) mv );
        }
        else if ( mv instanceof MingleDouble )
        {
            return asJsonNumber( (MingleDouble) mv );
        }
        else if ( mv instanceof MingleFloat )
        {
            return asJsonNumber( (MingleFloat) mv );
        }
        else if ( mv instanceof MingleString )
        {
            return asJsonString( (MingleString) mv );
        }
        else if ( mv instanceof MingleBoolean )
        {
            return asJsonBoolean( (MingleBoolean) mv );
        }
        else if ( mv instanceof MingleStructure )
        {
            return asJsonObject( (MingleStructure) mv );
        }
        else if ( mv instanceof MingleNull ) return JsonNull.INSTANCE;
        else if ( mv instanceof MingleBuffer )
        {
            return asJsonBuffer( (MingleBuffer) mv );
        }
        else if ( mv instanceof MingleTimestamp )
        {
            return asJsonValue( (MingleTimestamp) mv );
        }
        else if ( mv instanceof MingleList )
        {
            return asJsonArray( (MingleList) mv );
        }
        else if ( mv instanceof MingleSymbolMap )
        {
            return asJsonObject( (MingleSymbolMap) mv );
        }
        else if ( mv instanceof MingleEnum )
        {
            return asJsonValue( (MingleEnum) mv );
        }
        else 
        {
            throw state.createFail( 
                "Conversion not yet supported for", mv.getClass() );
        }
    }

    private
    JsonString
    asJsonString( MingleNamespace ns )
    {
        return JsonString.create( ns.getExternalForm() );
    }

    private
    JsonString
    asJsonString( MingleTypeReference ref )
    {
        return JsonString.create( ref.getExternalForm() );
    }

    private
    JsonString
    asJsonString( MingleIdentifier ident )
    {
        return JsonString.create( MingleModels.format( ident, fmt ) );
    }

    public
    JsonObject
    asCodecObject( MingleStruct ms )
    {
        inputs.notNull( ms, "ms" );
        return asJsonObject( ms );
    }

    private
    MingleCodecException
    asCodecException( String msg,
                      ObjectPath< JsonString > loc,
                      Exception cause )
    {
        StringBuilder sb = new StringBuilder();
        ObjectPaths.appendFormat( loc, ObjectPaths.DOT_FORMATTER, sb );

        sb.append( ": " ).append( msg );

        return new MingleCodecException( sb.toString(), cause );
    }

    private
    MingleCodecException
    asCodecException( String msg,
                      ObjectPath< JsonString > loc )
    {
        return asCodecException( msg, loc, null );
    }

    private
    JsonValue
    expectSingleValue( List< JsonValue > vals,
                       ObjectPath< JsonString > loc )
        throws MingleCodecException
    {
        if ( vals == null || vals.size() == 0 )
        {
            throw asCodecException( "Missing value", loc );
        }
        else if ( vals.size() == 1 ) return vals.get( 0 );
        else throw asCodecException( "Multiple values", loc );
    }

    private
    JsonValue
    expectSingleValue( JsonObject jsonObj,
                       JsonString key,
                       ObjectPath< JsonString > loc )
        throws MingleCodecException
    {
        return 
            expectSingleValue( jsonObj.getValues( key ), loc.descend( key ) );
    }

    private
    JsonString
    expectSingleString( List< JsonValue > vals,
                        JsonString key,
                        ObjectPath< JsonString > loc )
        throws MingleCodecException
    {
        return cast(
            JsonString.class,
            expectSingleValue( vals, loc.descend( key ) ),
            "Expected json string",
            loc.descend( key )
        );
    }

    private
    JsonString
    expectSingleString( JsonObject jsonObj,
                        JsonString key,
                        ObjectPath< JsonString > loc )
        throws MingleCodecException
    {
        return expectSingleString( jsonObj.getValues( key ), key, loc );
    }

    private
    MingleNamespace
    parseNs( JsonString str,
             ObjectPath< JsonString > loc )
        throws MingleCodecException
    {
        try { return MingleParsers.parseNamespace( str ); }
        catch ( SyntaxException se )
        {
            throw asCodecException( "Invalid namespace syntax", loc, se ); 
        }
    }

    private
    MingleNamespace
    expectNamespace( JsonObject jsonObj,
                     JsonString key,
                     ObjectPath< JsonString > loc )
        throws MingleCodecException
    {
        return 
            parseNs( 
                expectSingleString( jsonObj, key, loc ), loc.descend( key ) );
    }

    private
    MingleIdentifier
    parseIdentifier( JsonString str,
                     ObjectPath< JsonString > loc )
        throws MingleCodecException
    {
        try { return MingleParsers.parseIdentifier( str ); }
        catch ( SyntaxException se )
        {
            throw 
                asCodecException( "Invalid identifier '" + str + "'", loc, se );
        }
    }

    private
    MingleIdentifier
    expectIdentifier( JsonObject jsonObj,
                      JsonString key,
                      ObjectPath< JsonString > loc )
        throws MingleCodecException
    {
        return 
            parseIdentifier( 
                expectSingleString( jsonObj, key, loc ),
                loc.descend( key )
            );
    }

    private
    AtomicTypeReference
    expectAtomicType( JsonString str,
                      ObjectPath< JsonString > loc )
        throws MingleCodecException
    {
        try
        {
            MingleTypeReference ref = MingleTypeReference.parse( str );
            
            if ( ! ( ref instanceof AtomicTypeReference ) ) 
            {
                throw 
                    asCodecException( 
                        "Non-atomic type reference: " + ref, loc );
            }
            else return (AtomicTypeReference) ref;
        }
        catch ( SyntaxException se )
        {
            throw asCodecException( "Invalid type reference: " + str, loc, se );
        }
    }

    private
    AtomicTypeReference
    expectAtomicType( JsonObject jsonObj,
                      JsonString key,
                      ObjectPath< JsonString > loc )
        throws MingleCodecException
    {
        return 
            expectAtomicType( 
                expectSingleString( jsonObj, key, loc ), loc.descend( key ) );
    }

    private
    AtomicTypeReference
    expectAtomicType( JsonObject obj,
                      ObjectPath< JsonString > loc )
        throws MingleCodecException
    {
        return expectAtomicType( obj, KEY_TYPE, loc );
    }

    private
    void
    setType( MingleStructureBuilder< ?, ? > b,
             JsonObject obj,
             ObjectPath< JsonString > loc )
        throws MingleCodecException
    {
        b.setType(
            expectAtomicType( 
                expectSingleString( obj, KEY_TYPE, loc ),
                loc.descend( KEY_TYPE )
            ) 
        );
    }

    private
    MingleString
    asMingleString( JsonString str )
    {
        return MingleModels.asMingleString( str );
    }

    private
    MingleBoolean
    asMingleBoolean( JsonBoolean b )
    {
        return MingleModels.asMingleBoolean( b.booleanValue() );
    }

    private
    MingleValue
    asMingleNumber( JsonNumber num )
    {
        Number jNum = num.getNumber();

        if ( jNum instanceof Byte ||
                  jNum instanceof Short ||
                  jNum instanceof Integer ||
                  jNum instanceof Long || 
                  jNum instanceof BigInteger )
        {
            return MingleModels.asMingleInt64( jNum.longValue() );
        }
        else if ( jNum instanceof Float )
        {
            return MingleModels.asMingleFloat( jNum.floatValue() );
        }
        else if ( jNum instanceof Double || jNum instanceof BigDecimal )
        {
            return MingleModels.asMingleDouble( jNum.doubleValue() );
        }
        else
        {
            throw state.createFail(
                "Conversion of number type not yet supported:", 
                jNum.getClass() );
        }
    }

    // Stub for the moment
    private
    MingleValue
    asMingleList( Iterable< JsonValue > vals,
                  ObjectPath< JsonString > loc )
        throws MingleCodecException
    {
        MingleList.Builder b = new MingleList.Builder();
        ImmutableListPath< JsonString > lp = loc.startImmutableList();

        for ( JsonValue val : vals )
        {
            MingleValue mv = asMingleValue( val, lp.next() );
            b.add( mv );
        }

        return b.build();
    }

    private
    MingleValue
    asMingleValue( List< JsonValue > vals,
                   ObjectPath< JsonString > loc )
        throws MingleCodecException
    {
        return
            vals.size() == 1 
                ? asMingleValue( vals.get( 0 ), loc ) 
                : asMingleList( vals, loc );
    }

    private
    MingleSymbolMap
    asMingleSymbolMap( JsonObject obj,
                       ObjectPath< JsonString > loc )
        throws MingleCodecException
    {
        MingleSymbolMapBuilder b = MingleModels.symbolMapBuilder();

        for ( Map.Entry< JsonString, List< JsonValue > > e : obj.entrySet() )
        {
            buildMingleField( b, e.getKey(), e.getValue(), loc );
        }

        return b.build();
    }

    // if obj has the type key then we assume it is (or is at least a malformed
    // instance of) a mingle structure; if neither is present we treat it as a
    // symbol map
    private
    MingleValue
    asMingleValue( JsonObject obj,
                   ObjectPath< JsonString > loc )
        throws MingleCodecException
    {
        if ( obj.hasMember( KEY_TYPE ) ) 
        {
            if ( obj.hasMember( KEY_CONSTANT ) ) 
            {
                return asMingleEnum( obj, loc );
            }
            else return asMingleStruct( obj, loc );
        }
        else return asMingleSymbolMap( obj, loc );
    }

    private
    MingleValue
    asMingleValue( JsonValue val,
                   ObjectPath< JsonString > loc )
        throws MingleCodecException
    {
        if ( val instanceof JsonString )
        {
            return asMingleString( (JsonString) val );
        }
        else if ( val instanceof JsonBoolean )
        {
            return asMingleBoolean( (JsonBoolean) val );
        }
        else if ( val instanceof JsonNumber )
        {
            return asMingleNumber( (JsonNumber) val );
        }
        else if ( val instanceof JsonObject )
        {
            return asMingleValue( (JsonObject) val, loc );
        }
        else if ( val instanceof JsonArray )
        {
            return asMingleList( (JsonArray) val, loc );
        }
        else if ( val instanceof JsonNull ) return MingleNull.getInstance();
        else
        {
            throw state.createFail(
                "Conversion not yet supported for instances of",
                val.getClass() );
        }
    }

    private
    MingleEnum
    asMingleEnum( JsonObject jsonObj,
                  ObjectPath< JsonString > loc )
        throws MingleCodecException
    {
        return
            MingleEnum.create(
                expectAtomicType( jsonObj, loc ),
                expectIdentifier( jsonObj, KEY_CONSTANT, loc )
            );
    }

    private
    void
    buildMingleField( MingleSymbolMapBuilder symBld,
                      JsonString key,
                      List< JsonValue > vals,
                      ObjectPath< JsonString > loc )
        throws MingleCodecException
    {
        MingleIdentifier fld = parseIdentifier( key, loc );
        MingleValue mv = asMingleValue( vals, loc.descend( key ) );
 
        symBld.set( fld, mv );
    }

    // null checking done for public inputs
    private
    < S extends MingleStructure >
    S
    asMingleStructure( JsonObject jsonObj,
                       MingleStructureBuilder< ?, S > msBld,
                       ObjectPath< JsonString > loc )
        throws MingleCodecException
    {
        inputs.notNull( jsonObj, "jsonObj" );
        inputs.notNull( loc, "loc" );

        setType( msBld, jsonObj, loc );

        for ( Map.Entry< JsonString, List< JsonValue > > e : 
                jsonObj.entrySet() )
        {
            JsonString key = e.getKey();
            List< JsonValue > vals = e.getValue();
            
            if ( ! key.equals( KEY_TYPE ) ) 
            {
                buildMingleField( msBld.fields(), key, vals, loc );
            }
        }

        return msBld.build();
    }

    public
    MingleStruct
    asMingleStruct( JsonObject jsonObj,
                    ObjectPath< JsonString > loc )
        throws MingleCodecException
    {
        return asMingleStructure( jsonObj, MingleModels.structBuilder(), loc );
    } 

    public
    MingleStruct
    asMingleStruct( JsonObject jsonObj )
        throws MingleCodecException
    {
        return asMingleStruct( jsonObj, ROOT_PATH );
    }

    private
    MingleException
    asMingleException( JsonObject jsonObj,
                       ObjectPath< JsonString > loc )
        throws MingleCodecException
    {
        return 
            asMingleStructure( jsonObj, MingleModels.exceptionBuilder(), loc );
    }

    private
    MingleException
    asMingleException( JsonObject jsonObj )
        throws MingleCodecException
    {
        return asMingleException( jsonObj, ROOT_PATH );
    }

    private
    final
    class MingleEncoderImpl
    implements MingleEncoder
    {
        private final JsonSerializer ser;

        private MingleEncoderImpl( JsonSerializer ser ) { this.ser = ser; }

        public
        boolean
        writeTo( ByteBuffer bb )
            throws Exception
        {
            return ser.writeTo( bb );
        }
    }

    public
    MingleEncoder
    createEncoder( Object me )
        throws MingleCodecException
    {
        inputs.notNull( me, "me" );

        JsonObject jsonObj;

        if ( me instanceof MingleStruct )
        {
            jsonObj = asCodecObject( (MingleStruct) me );
        }
        else throw state.createFail( "Unexpected encodable:", me );
        
        JsonSerializer ser = JsonSerializer.create( jsonObj, serOpts );

        return new MingleEncoderImpl( ser );
    }

    private
    final
    class MingleDecoderImpl< E >
    implements MingleDecoder< E >
    {
        // lazily initialized on first call to readFrom()
        private JsonParser< JsonText > p;

        private final Class< E > cls;

        private MingleDecoderImpl( Class< E > cls ) { this.cls = cls; }

        // The assumption is that this decoder is only used by callers that are
        // aware of the message boundaries and pass in endOfInput precisely when
        // the expectation is that the input buffer supplied contains no more
        // than then end of the data for this decode. For that reason the return
        // value from this method is precisely whatever is supplied as
        // endOfInput.
        public
        boolean
        readFrom( ByteBuffer buf,
                  boolean endOfInput )
            throws Exception
        {
            if ( p == null ) p = jpf.createTextParser( "<>" );

            p.update( buf, endOfInput );
            return endOfInput;
        }

        private
        JsonObject
        getJsonObject()
            throws MingleCodecException
        {
            JsonText res = p.buildResult();

            if ( res instanceof JsonObject ) return (JsonObject) res;
            else throw new MingleCodecException( "Invalid JSON input" );
        }

        private E castResult( Object o ) { return cls.cast( o ); }

        public
        E
        getResult()
            throws Exception
        {
            JsonObject jsonObj = getJsonObject();
            ObjectPath< JsonString > loc = ROOT_PATH;

            if ( MingleStruct.class.isAssignableFrom( cls ) )
            {
                return castResult( asMingleStruct( jsonObj, loc ) );
            }
            else 
            {
                throw 
                    state.createFail( 
                        "Encoder created for unhandled type:", cls );
            }
        }
    }

    public 
    < E >
    MingleDecoder< E > 
    createDecoder( Class< E > cls ) 
    { 
        inputs.notNull( cls, "cls" );

        if ( MingleStruct.class.isAssignableFrom( cls ) )
        {
            return new MingleDecoderImpl< E >( cls ); 
        }
        else throw state.createFail( "Unsupported decode target:", cls );
    }

    public static JsonMingleCodec create() { return new Builder().build(); }

    public
    final
    static
    class Builder
    {
        private MingleIdentifierFormat fmt = 
            MingleIdentifierFormat.LC_HYPHENATED;

        private JsonSerializer.Options serOpts =
            JsonSerializer.Options.getDefault();

        private boolean omitTypeFields;
        private boolean expandEnums;

        public Builder() {}

        public
        Builder
        setIdentifierFormat( MingleIdentifierFormat fmt )
        {
            this.fmt = inputs.notNull( fmt, "fmt" );
            return this;
        }

        public
        Builder
        setSerializerOptions( JsonSerializer.Options serOpts )
        {
            this.serOpts = inputs.notNull( serOpts, "serOpts" );
            return this;
        }

        public
        Builder
        setOmitTypeFields( boolean omitTypeFields )
        {
            this.omitTypeFields = omitTypeFields;
            return this;
        }

        public Builder setOmitTypeFields() { return setOmitTypeFields( true ); }

        public
        Builder
        setExpandEnums( boolean expandEnums )
        {
            this.expandEnums = expandEnums;
            return this;
        }

        public Builder setExpandEnums() { return setExpandEnums( true ); }

        public
        JsonMingleCodec
        build()
        {
            state.isFalse( 
                expandEnums && omitTypeFields,
                "Illegal combination of expandEnums and omitTypeFields"
            );

            return new JsonMingleCodec( this );
        }
    }
}
