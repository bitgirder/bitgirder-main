package mingle

import (
    "bitgirder/objpath"
    "fmt"
    "errors"
    "strings"
    "regexp"
//    "log"
    "time"
    "sort"
    "bytes"
    "strconv"
    "unicode"
    "encoding/base64"
)

type IdentifierFormat uint

const (
    _ IdentifierFormat = iota // leave zero val internally for 'unknown'
    LcUnderscore 
    LcHyphenated
    LcCamelCapped
)

func ( idFmt IdentifierFormat ) String() string {
    switch idFmt {
    case LcUnderscore: return "lc-underscore"
    case LcHyphenated: return "lc-hyphenated"
    case LcCamelCapped: return "lc-camel-capped"
    }
    panic( fmt.Errorf( "Unrecognized id format: %d", idFmt ) )
}

func IdentifierFormatString( nm string ) ( idFmt IdentifierFormat, err error ) {
    switch nm {
    case "lc-underscore": idFmt = LcUnderscore
    case "lc-hyphenated": idFmt = LcHyphenated
    case "lc-camel-capped": idFmt = LcCamelCapped
    default: err = fmt.Errorf( "Unrecognized id format: %s", nm )
    }
    return
}

func MustIdentifierFormatString( nm string ) IdentifierFormat {
    res, err := IdentifierFormatString( nm )
    if err != nil { panic( err ) }
    return res
}

var IdentifierFormats = 
    []IdentifierFormat{ LcUnderscore, LcHyphenated, LcCamelCapped }

// Marker for (Identifier|string)
type IdentifierInitializer interface{}

func asIdentifier( id IdentifierInitializer ) ( *Identifier, error ) {
    switch v := id.( type ) {
    case string: return MustIdentifier( v ), nil
    case *Identifier: return v, nil
    }
    return nil, fmt.Errorf( "Unhandled id initializer: %v", id )
}

type idPart []byte

//func ( part idPart ) compare( part2 idPart ) int {
func compareIdParts( part1, part2 idPart ) int {
    return bytes.Compare( []byte( part1 ), []byte( part2 ) )
}

type Identifier struct {
    parts []idPart
}

// Meant for other mingle impl packages only; external callers should not count
// on the behavior of this remaining stable, or even of it continuing to exist
func NewIdentifierUnsafe( parts [][]byte ) *Identifier {
    idParts := make( []idPart, len( parts ) )
    for i, p := range parts { idParts[ i ] = idPart( p ) }
    return &Identifier{ idParts }
}

func idSeparatorFor( idFmt IdentifierFormat ) (sep byte) {
    switch idFmt {
        case LcHyphenated: sep = byte( '-' )
        case LcUnderscore: sep = byte( '_' )
        default: panic( fmt.Sprintf( "Unhandled id format: %v", idFmt ) )
    }
    return
}

func ( id *Identifier ) formatToBuf( 
    buf []byte, idFmt IdentifierFormat ) []byte {
    for i, part := range id.parts {
        var cpStart int
        if idFmt == LcCamelCapped {
            if i > 0 {
                ch0 := rune( part[ 0 ] )
                buf = append( buf, byte( unicode.ToUpper( ch0 ) ) )
                cpStart = 1
            }
        } else {
            if i > 0 { buf = append( buf, idSeparatorFor( idFmt ) ) }
        }
        buf = append( buf, part[ cpStart : ]... )
    }
    return buf
}

func ( id *Identifier ) extFormToBuf( buf []byte ) []byte {
    return id.formatToBuf( buf, LcHyphenated )
}

func ( id *Identifier ) Format( idFmt IdentifierFormat ) string {
    return string( id.formatToBuf( make( []byte, 0, 32 ), idFmt ) )
}

func ( id *Identifier ) ExternalForm() string {
    return id.Format( LcHyphenated )
}

func ( id *Identifier ) String() string { return id.ExternalForm() }

func ( id *Identifier ) Compare( id2 *Identifier ) int {
    l, r, swap := id, id2, 1
    if len( id.parts ) > len( id2.parts ) { l, r, swap = id2, id, -1 }
    for i, part := range l.parts {
        if cmp := compareIdParts( part, r.parts[ i ] ); cmp != 0 { 
            return swap * cmp 
        }
    }
    return swap * ( len( l.parts ) - len( r.parts ) )
}

func ( id *Identifier ) Equals( id2 *Identifier ) bool {
    return id2 != nil && id.Compare( id2 ) == 0
}

type Namespace struct {
    Parts []*Identifier
    Version *Identifier
}

func ( ns *Namespace ) formatToBuf( buf []byte, styl IdentifierFormat ) []byte {
    for i, id := range ns.Parts {
        if i > 0 { buf = append( buf, byte( ':' ) ) }
        buf = id.formatToBuf( buf, styl )
    }
    buf = append( buf, byte( '@' ) )
    return ns.Version.formatToBuf( buf, styl )
}

func ( ns *Namespace ) extFormToBuf( buf []byte ) []byte {
    return ns.formatToBuf( buf, LcCamelCapped )
}

func ( ns *Namespace ) ExternalForm() string {
    return string( ns.extFormToBuf( make( []byte, 0, 32 ) ) )
}

func ( ns *Namespace ) String() string { return ns.ExternalForm() }

func ( ns *Namespace ) Equals( ns2 *Namespace ) bool {
    if ns2 == nil { return false }
    if ns.Version.Equals( ns2.Version ) {
        f := func( i int ) bool { 
            return ns.Parts[ i ].Equals( ns2.Parts[ i ] ) 
        }
        return equalSlices( len( ns.Parts ), len( ns2.Parts ), f )
    }
    return false
}

// Interface for (DeclaredTypeName|QualifiedTypeName) 
type TypeName interface{
    ExternalForm() string
    Equals( n TypeName ) bool
}

type DeclaredTypeName struct {
    nm []byte
}

func NewDeclaredTypeNameUnsafe( nm []byte ) *DeclaredTypeName {
    return &DeclaredTypeName{ nm }
}

func ( n *DeclaredTypeName ) String() string { return string( n.nm ) }

func ( n *DeclaredTypeName ) ExternalForm() string { return n.String() }

func ( n *DeclaredTypeName ) Equals( other TypeName ) bool {
    if n2, ok := other.( *DeclaredTypeName ); ok {
        if n2 == nil { return false }
        return bytes.Compare( n.nm, n2.nm ) == 0
    }
    return false
}

func ( n *DeclaredTypeName ) ResolveIn( ns *Namespace ) *QualifiedTypeName {
    return &QualifiedTypeName{ Namespace: ns, Name: n }
}

func equalSlices( lenL, lenR int, comp func( i int ) bool ) bool {
    if lenL == lenR {
        for i := 0; i < lenL; i++ { if ! comp( i ) { return false } }
        return true
    } 
    return false
}

type QualifiedTypeName struct {
    *Namespace
    Name *DeclaredTypeName
}

func ( qn *QualifiedTypeName ) ExternalForm() string {
    res := make( []byte, 0, 32 )
    res = append( res, []byte( qn.Namespace.ExternalForm() )... )
    res = append( res, byte( '/' ) )
    res = append( res, []byte( ( qn.Name.nm ) )... )
    return string( res )
}

func ( qn *QualifiedTypeName ) String() string { return qn.ExternalForm() }

func ( qn *QualifiedTypeName ) Equals( n2 TypeName ) bool {
    if n2 == nil { return false }
    if qn2, ok := n2.( *QualifiedTypeName ); ok {
        return qn.Namespace.Equals( qn2.Namespace ) &&
               qn.Name.Equals( qn2.Name )
    }
    return false
}

// (atomic|list|nullable)
type TypeReference interface {
    ExternalForm() string
    Equals( t TypeReference ) bool
    String() string
}

type ValueRestriction interface {

    ExternalForm() string

    // Will panic if val is nil or is not the correct type for this instance
    AcceptsValue( val Value ) bool

    equalsRestriction( vr ValueRestriction ) bool
}

func equalComparers( c1, c2 interface{} ) bool {
    if c1 == c2 { return true }
    if c1 == nil { return c2 == nil }
    if c2 == nil { return false }
    return c1.( Comparer ).Compare( c2 ) == 0
}

type RegexRestriction struct {
    src string
    exp *regexp.Regexp
}

func NewRegexRestriction( src string ) ( *RegexRestriction, error ) {
    exp, err := regexp.Compile( src )
    if err == nil { return &RegexRestriction{ src, exp }, nil }
    return nil, err
}

func ( r *RegexRestriction ) ExternalForm() string {
    return fmt.Sprintf( "%q", string( r.src ) )
}

func ( r *RegexRestriction ) equalsRestriction( vr ValueRestriction ) bool {
    if vr == nil { return false }
    if r == vr { return true }
    if r2, ok := vr.( *RegexRestriction ); ok { return r.src == r2.src }
    return false
}

var errNilVal error
func init() { errNilVal = fmt.Errorf( "Value is nil" ) }

func ( r *RegexRestriction ) AcceptsValue( val Value ) bool {
    if val == nil { panic( errNilVal ) }
    return r.exp.MatchString( string( val.( String ) ) )
}

type RangeRestriction struct {
    MinClosed bool
    Min Value 
    Max Value
    MaxClosed bool
}

func quoteRangeValue( val Value ) string {
    switch v := val.( type ) {
    case Int32, Int64, Uint32, Uint64, Float32, Float64: 
        return val.( fmt.Stringer ).String()
    case String: return fmt.Sprintf( "%q", string( v ) )
    case Timestamp: return fmt.Sprintf( "%q", v.Rfc3339Nano() )
    }
    panic( fmt.Errorf( "Unhandled range val type: %T", val ) )
}

func ( r *RangeRestriction ) ExternalForm() string {
    buf := bytes.Buffer{}
    if r.MinClosed { buf.WriteRune( '[' ) } else { buf.WriteRune( '(' ) }
    if r.Min != nil { buf.WriteString( quoteRangeValue( r.Min ) ) }
    buf.WriteRune( ',' )
    if r.Max != nil { buf.WriteString( quoteRangeValue( r.Max ) ) }
    if r.MaxClosed { buf.WriteRune( ']' ) } else { buf.WriteRune( ')' ) }
    return buf.String()
}

func ( r *RangeRestriction ) equalsRestriction( vr ValueRestriction ) bool {
    if vr == nil { return false }
    if r == vr { return true }
    if r2, ok := vr.( *RangeRestriction ); ok {
        // do cheap tests first
        return r.MinClosed == r2.MinClosed &&
               r.MaxClosed == r2.MaxClosed &&
               equalComparers( r.Min, r2.Min ) &&
               equalComparers( r.Max, r2.Max )
    }
    return false
}

func ( r *RangeRestriction ) AcceptsValue( val Value ) bool {
    if val == nil { panic( errNilVal ) }
    if r.Min != nil {
        switch i := r.Min.( Comparer ).Compare( val ); {
        case i == 0: if ! r.MinClosed { return false }
        case i > 0: return false
        }
    }
    if r.Max != nil {
        switch i := r.Max.( Comparer ).Compare( val ); {
        case i == 0: if ! r.MaxClosed { return false }
        case i < 0: return false
        }
    }
    return true
}

type AtomicTypeReference struct {
    Name TypeName
    Restriction ValueRestriction
}

func ( t *AtomicTypeReference ) ExternalForm() string {
    nm := t.Name.ExternalForm()
    if t.Restriction == nil { return nm }
    return fmt.Sprintf( "%s~%s", nm, t.Restriction.ExternalForm() )
}

func ( t *AtomicTypeReference ) String() string { return t.ExternalForm() }

func ( t *AtomicTypeReference ) Equals( t2 TypeReference ) bool {
    if t2 == nil { return false }
    if t == t2 { return true }
    if at2, ok := t2.( *AtomicTypeReference ); ok {
        if t.Name.Equals( at2.Name ) {
            if t.Restriction == nil { return at2.Restriction == nil }
            return t.Restriction.equalsRestriction( at2.Restriction )
        }
    }
    return false
}

type ListTypeReference struct {
    ElementType TypeReference
    AllowsEmpty bool
}

func ( t *ListTypeReference ) ExternalForm() string {
    var quant string
    if t.AllowsEmpty { quant = "*" } else { quant = "+" }
    return t.ElementType.ExternalForm() + quant
}

func ( t *ListTypeReference ) String() string { return t.ExternalForm() }

func ( t *ListTypeReference ) Equals( ref TypeReference ) bool {
    if ref == nil { return false }
    if t2, ok := ref.( *ListTypeReference ); ok {
        return t.AllowsEmpty == t2.AllowsEmpty &&
               t.ElementType.Equals( t2.ElementType )
    }
    return false
}

type NullableTypeReference struct {
    Type TypeReference
}

func ( t *NullableTypeReference ) ExternalForm() string {
    return t.Type.ExternalForm() + "?"
}

func ( t *NullableTypeReference ) String() string { return t.ExternalForm() }

func ( typ *NullableTypeReference ) Equals( ref TypeReference ) bool {
    if ref == nil { return false }
    if typ2, ok := ref.( *NullableTypeReference ); ok {
        return typ.Type.Equals( typ2.Type )
    }
    return false
}

func TypeNameIn( typ TypeReference ) TypeName {
    switch v := typ.( type ) {
    case *AtomicTypeReference: return v.Name
    case *ListTypeReference: return TypeNameIn( v.ElementType )
    case *NullableTypeReference: return TypeNameIn( v.Type )
    }
    panic( fmt.Errorf( "Unhandled type reference: %T", typ ) )
}

// Marker for (TypeReference|string)
type TypeReferenceInitializer interface{}

func asTypeReference( typ TypeReferenceInitializer ) TypeReference {
    switch v := typ.( type ) {
    case string: return MustTypeReference( v )
    case TypeReference: return v
    }
    panic( fmt.Sprintf( "Unhandled type ref initializer: %v", typ ) )
}

func AtomicTypeIn( ref TypeReference ) *AtomicTypeReference {
    switch v := ref.( type ) {
    case *AtomicTypeReference: return v
    case *ListTypeReference: return AtomicTypeIn( v.ElementType )
    case *NullableTypeReference: return AtomicTypeIn( v.Type )
    }
    panic( fmt.Errorf( "No atomic type in %s (%T)", ref, ref ) )
}

type Value interface{}

func appendQuotedList( buf *bytes.Buffer, l *List ) {
    buf.WriteRune( '[' )
    for i, val := range l.vals {
        appendQuotedValue( buf, val )
        if i < len( l.vals ) - 1 { buf.WriteString( ", " ) }
    }
    buf.WriteRune( ']' )
}

func appendQuotedSymbolMap( buf *bytes.Buffer, m *SymbolMap ) {
    buf.WriteRune( '{' )
    for i, fe := range m.fields {
        buf.WriteString( fe.id.Format( LcCamelCapped ) )
        buf.WriteRune( ':' )
        appendQuotedValue( buf, fe.val )
        if i < len( m.fields ) - 1 { buf.WriteString( ", " ) }
    }
    buf.WriteRune( '}' )
}

func appendQuotedStruct( buf *bytes.Buffer, ms *Struct ) {
    buf.WriteString( ms.Type.ExternalForm() )
    appendQuotedSymbolMap( buf, ms.Fields )
}

func appendQuotedValue( buf *bytes.Buffer, val Value ) {
    switch v := val.( type ) {
    case String: fmt.Fprintf( buf, "%q", string( v ) )
    case Buffer: fmt.Fprintf( buf, "%x", []byte( v ) )
    case Timestamp: fmt.Fprintf( buf, "%s", v.Rfc3339Nano() )
    case *Null: buf.WriteString( "null" )
    case Boolean, Int32, Int64, Uint32, Uint64, Float32, Float64:
        buf.WriteString( val.( fmt.Stringer ).String() )
    case *Enum: 
        fmt.Fprintf( buf, "%s.%s", 
            v.Type.ExternalForm(), v.Value.ExternalForm() )
    case *List: appendQuotedList( buf, v )
    case *SymbolMap: appendQuotedSymbolMap( buf, v )
    case *Struct: appendQuotedStruct( buf, v )
    default: fmt.Fprintf( buf, "(!%T)", val ) // seems better than a panic
    }
}

func QuoteValue( val Value ) string { 
    buf := &bytes.Buffer{}
    appendQuotedValue( buf, val )
    return buf.String()
}

type goValPath objpath.PathNode // keys are string

var inValPathRoot objpath.PathNode
func init() { inValPathRoot = objpath.RootedAt( "inVal" ) }

type ValueTypeError struct { 
    loc goValPath
    msg string 
}

var goValPathFormatter objpath.Formatter
func init() {
    goValPathFormatter = 
        objpath.DotFormatter(
            func( elt interface{}, apnd objpath.AppendFunc ) {
                apnd( elt.( string ) )
            },
        )
}

func ( e *ValueTypeError ) Error() string { 
    if e.loc == nil { return e.msg }
    locStr := objpath.Format( e.loc, goValPathFormatter )
    return locStr + ": " + e.msg 
}

type Comparer interface {

    // Should panic (such as with type assertion panic) if val is not same type
    // as the instance or is nil
    Compare( val interface{} ) int
}

type String string
func ( s String ) String() string { return string( s ) }

func ( s String ) Compare( val interface{} ) int {
    switch s2 := val.( String ); {
    case s < s2: return -1
    case s == s2: return 0
    }
    return 1
}

type Boolean bool
func ( b Boolean ) String() string { return fmt.Sprint( bool( b ) ) }

type Int64 int64
func ( i Int64 ) String() string { return fmt.Sprint( int64( i ) ) }

func ( i Int64 ) Compare( val interface{} ) int {
    switch v := val.( Int64 ); {
    case i < v: return -1
    case i > v: return 1
    }
    return 0
}

type Int32 int32
func ( i Int32 ) String() string { return fmt.Sprint( int32( i ) ) }

func ( i Int32 ) Compare( val interface{} ) int {
    return Int64( i ).Compare( Int64( val.( Int32 ) ) )
}

type Uint64 uint64
func ( i Uint64 ) String() string { return fmt.Sprint( uint64( i ) ) }

func ( i Uint64 ) Compare( val interface{} ) int {
    switch v := val.( Uint64 ); {
    case i < v: return -1
    case i > v: return 1
    }
    return 0
}

type Uint32 uint32
func ( i Uint32 ) String() string { return fmt.Sprint( uint32( i ) ) }

func ( i Uint32 ) Compare( val interface{} ) int {
    return Uint64( i ).Compare( Uint64( val.( Uint32 ) ) )
}

type Float64 float64
func ( d Float64 ) String() string { return fmt.Sprint( float64( d ) ) }

func ( d Float64 ) Compare ( val interface{} ) int {
    switch d2 := val.( Float64 ); {
    case d < d2: return -1
    case d == d2: return 0
    }
    return 1
}

type Float32 float32
func ( f Float32 ) String() string { return fmt.Sprint( float32( f ) ) }

func ( f Float32 ) Compare( val interface{} ) int {
    return Float64( f ).Compare( Float64( val.( Float32 ) ) )
}

type Buffer []byte

type Null struct {}
var NullVal *Null
func init() { NullVal = &Null{} }

func IsNull( val Value ) bool {
    _, isNull := val.( *Null )
    return isNull
}

type Timestamp time.Time

func Now() Timestamp { return Timestamp( time.Now() ) }

func ( t Timestamp ) Rfc3339Nano() string {
    return time.Time( t ).Format( time.RFC3339Nano )
}

func ( t Timestamp ) String() string { return t.Rfc3339Nano() }

func ( t Timestamp ) Compare( val interface{} ) int {
    tm1 := time.Time( t )
    switch tm2 := time.Time( val.( Timestamp ) ); {
    case tm1.Before( tm2 ): return -1
    case tm1.After( tm2 ): return 1
    }
    return 0
}

func asListValue( inVals []interface{}, path goValPath ) ( *List, error ) {
    vals := make( []Value, len( inVals ) )
    lp := path.StartList()
    for i, inVal := range inVals { 
        var err error
        if vals[ i ], err = implAsValue( inVal, lp ); err != nil { 
            return nil, err 
        }
        lp = lp.Next()
    }
    return &List{ vals }, nil
}

func asAtomicValue( 
    inVal interface{}, path goValPath ) ( val Value, err error ) {
    if inVal == nil { return NullVal, nil }
    switch v := inVal.( type ) {
    case string: val = String( v )
    case String: val = v
    case bool: val = Boolean( v )
    case Boolean: val = v
    case []byte: val = Buffer( v )
    case Buffer: val = v
    case int8: val = Int32( int32( v ) )
    case int16: val = Int32( int32( v ) )
    case int32: val = Int32( v )
    case Int32: val = v
    case int: val = Int64( int64( v ) )
    case int64: val = Int64( v )
    case Int64: val = v
    case uint64: val = Uint64( v )
    case Uint64: val = v
    case uint32: val = Uint32( v )
    case Uint32: val = v
    case float32: val = Float32( v )
    case Float32: val = v
    case float64: val = Float64( v )
    case Float64: val = v
    case Timestamp: val = v
    case time.Time: val = Timestamp( v )
    case *List: val = v
    case *SymbolMap: val = v
    case *Enum: val = v
    case *Struct: val = v
    case *Null: val = v
    default:
        msg := "Unhandled mingle value %v (%T)"
        err = &ValueTypeError{ path, fmt.Sprintf( msg, inVal, inVal ) }
    }
    return
}

func implAsValue( inVal interface{}, path goValPath ) ( Value, error ) {
    switch v := inVal.( type ) {
    case []interface{}: return asListValue( v, path )
    }
    return asAtomicValue( inVal, path )
}

func AsValue( inVal interface{} ) ( Value, error ) {
    return implAsValue( inVal, inValPathRoot )
}

func MustValue( inVal interface{} ) Value {
    val, err := AsValue( inVal )
    if err != nil { panic( err ) }
    return val
}

type List struct {
    vals []Value
}

func ( l *List ) Values() []Value { return l.vals }

func ( l *List ) Len() int { return len( l.vals ) }

func NewList( vals []Value ) *List { return &List{ vals } }

func CreateList( vals ...interface{} ) ( *List, error ) {
    sz := len( vals )
    res := &List{ vals: make( []Value, sz ) }
    for i, val := range vals {
        var err error
        if res.vals[ i ], err = AsValue( val ); err != nil { return nil, err }
    }
    return res, nil
}

func MustList( vals ...interface{} ) *List {
    res, err := CreateList( vals... )
    if err != nil { panic( err ) }
    return res
}

type fieldEntry struct {
    id *Identifier
    val Value
}

type SymbolMap struct {
    fields []fieldEntry
}

func ( m *SymbolMap ) Len() int { return len( m.fields ) }

func ( m *SymbolMap ) EachPairError ( 
    v func( *Identifier, Value ) error ) error {
    for _, e := range m.fields { 
        if err := v( e.id, e.val ); err != nil { return err }
    }
    return nil
}

func ( m *SymbolMap ) EachPair( v func( *Identifier, Value ) ) {
    m.EachPairError( func( fld *Identifier, val Value ) error {
        v( fld, val )
        return nil
    })
}

// For small maps it may be faster to scan linearly; we can sample and measure
// this down the line when optimizing that becomes necessary
func ( m *SymbolMap ) GetById( fld *Identifier ) Value {
    f := func( i int ) bool { return m.fields[ i ].id.Compare( fld ) >= 0 }
    indx := sort.Search( len( m.fields ), f )
    if indx < len( m.fields ) && m.fields[ indx ].id.Equals( fld ) {
        return m.fields[ indx ].val
    }
    return nil
}

type MapLiteralError struct {
    msg string
}

func ( e *MapLiteralError ) Error() string { return e.msg } 

func mapLiteralError( fmtStr string, args ...interface{} ) *MapLiteralError {
    return &MapLiteralError{ fmt.Sprintf( fmtStr, args... ) }
}

func makePairError( err error, indx int ) error {
    return mapLiteralError( "Error in map literal pairs at index %d: %s",
        indx, err )
}

func fieldEntryFromPair(
    pairs []interface{}, indx int ) ( *fieldEntry, error ) {
    var key *Identifier
    var val Value
    var err error
    keyIndx, valIndx := indx, indx + 1
    if key, err = asIdentifier( pairs[ keyIndx ] ); err != nil { 
        return nil, makePairError( err, keyIndx )
    }
    if val, err = AsValue( pairs[ valIndx ] ); err != nil { 
        return nil, makePairError( err, valIndx )
    }
    return &fieldEntry{ id: key, val: val }, nil
}

type fieldSorter []fieldEntry

func ( fs fieldSorter ) Len() int { return len( fs ) }

func ( fs fieldSorter ) Less( i, j int ) bool {
    return fs[ i ].id.Compare( fs[ j ].id ) < 0
}

func ( fs fieldSorter ) Swap( i, j int ) { fs[ i ], fs[ j ] = fs[ j ], fs[ i ] }

func makeSymbolMap( flds []fieldEntry ) ( *SymbolMap, error ) {
    sort.Sort( fieldSorter( flds ) )
    var curId *Identifier
    for _, e := range flds {
        if curId == nil || curId.Compare( e.id ) < 0 {
            curId = e.id 
        } else {
            return nil, mapLiteralError( "Multiple entries for key: %s", curId )
        }
    }
    return &SymbolMap{ flds }, nil
}

func CreateSymbolMap( pairs ...interface{} ) ( m *SymbolMap, err error ) {
    if pLen := len( pairs ); pLen % 2 == 1 { 
        return nil, mapLiteralError( "Invalid pairs len: %d", pLen )
    } else { 
        flds := make( []fieldEntry, 0, pLen / 2 )
        for i := 0; i < pLen; i += 2 {
            var e *fieldEntry
            if e, err = fieldEntryFromPair( pairs, i ); err == nil {
                flds = append( flds, *e )
            } else { return }
        }
        m, err = makeSymbolMap( flds )
    }
    return
}

func MustSymbolMap( pairs ...interface{} ) *SymbolMap {
    res, err := CreateSymbolMap( pairs... )
    if err != nil { panic( err ) }
    return res
}

type SymbolMapAccessor struct {
    m *SymbolMap
    path objpath.PathNode
}

func NewSymbolMapAccessor( 
    m *SymbolMap, path objpath.PathNode ) *SymbolMapAccessor {
    return &SymbolMapAccessor{ m, path }
}

func ( acc *SymbolMapAccessor ) descend( fld *Identifier ) objpath.PathNode {
    if acc.path == nil { return objpath.RootedAt( fld ) }
    return acc.path.Descend( fld )
}

func ( acc *SymbolMapAccessor ) GetValueById( id *Identifier ) ( Value, error ) {
    val, err := acc.m.GetById( id ), error( nil )
    if val == nil { 
        err = &ValidationError{ "value is null", acc.descend( id ) } 
    }
    return val, err
}

func ( acc *SymbolMapAccessor ) GetValueByString( id string ) ( Value, error ) {
    return acc.GetValueById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) MustValueById( id *Identifier ) Value {
    val, err := acc.GetValueById( id )
    if err != nil { panic( err ) }
    return val
}

func ( acc *SymbolMapAccessor ) MustValueByString( id string ) Value {
    return acc.MustValueById( MustIdentifier( id ) )
}

type Enum struct {
    Type TypeReference
    Value *Identifier
}

func MustEnum( typ, val string ) *Enum {
    return &Enum{ MustTypeReference( typ ), MustIdentifier( val ) }
}

type Struct struct {
    Type TypeReference
    Fields *SymbolMap
}

func CreateStruct(
    typ TypeReferenceInitializer, pairs ...interface{} ) ( *Struct, error ) {
    res := new( Struct )
    res.Type = asTypeReference( typ )
    if flds, err := CreateSymbolMap( pairs... ); err == nil {
        res.Fields = flds
    } else { return nil, err }
    return res, nil
}

func MustStruct( 
    typ TypeReferenceInitializer, pairs ...interface{} ) *Struct {
    res, err := CreateStruct( typ, pairs... )
    if err != nil { panic( err ) }
    return res
}

type IdentifiedName struct {
    Namespace *Namespace
    Names []*Identifier
}

func ( nm *IdentifiedName ) ExternalForm() string {
    buf := make( []byte, 0, 32 )
    buf = nm.Namespace.formatToBuf( buf, LcHyphenated )
    for _, id := range nm.Names {
        buf = append( buf, byte( '/' ) )
        buf = id.formatToBuf( buf, LcHyphenated )
    }
    return string( buf )
}

func ( nm *IdentifiedName ) String() string { return nm.ExternalForm() }

func ( nm *IdentifiedName ) Equals( nm2 *IdentifiedName ) bool {
    if nm2 == nil { return false }
    if nm.Namespace.Equals( nm2.Namespace ) {
        f := func( i int ) bool {
            return nm.Names[ i ].Equals( nm2.Names[ i ] )
        }
        return equalSlices( len( nm.Names ), len( nm2.Names ), f )
    }
    return false
}

type idPath objpath.PathNode // elts are *Identifier

var idPathRootVal idPath
func init() { idPathRootVal = objpath.RootedAt( MustIdentifier( "val" ) ) }

var idPathFormatter objpath.Formatter

func init() {
    f := func( elt interface{}, apnd objpath.AppendFunc ) {
        apnd( elt.( *Identifier ).ExternalForm() )
    }
    idPathFormatter = objpath.DotFormatter( f )
}

func FormatIdPath( p objpath.PathNode ) string {
    return objpath.Format( p, idPathFormatter )
}

var (
    QnameValue *QualifiedTypeName
    TypeValue *AtomicTypeReference
    QnameBoolean *QualifiedTypeName
    TypeBoolean *AtomicTypeReference
    QnameBuffer *QualifiedTypeName
    TypeBuffer *AtomicTypeReference
    QnameString *QualifiedTypeName
    TypeString *AtomicTypeReference
    QnameInt32 *QualifiedTypeName
    TypeInt32 *AtomicTypeReference
    QnameInt64 *QualifiedTypeName
    TypeInt64 *AtomicTypeReference
    QnameUint32 *QualifiedTypeName
    TypeUint32 *AtomicTypeReference
    QnameUint64 *QualifiedTypeName
    TypeUint64 *AtomicTypeReference
    QnameFloat32 *QualifiedTypeName
    TypeFloat32 *AtomicTypeReference
    QnameFloat64 *QualifiedTypeName
    TypeFloat64 *AtomicTypeReference
    QnameTimestamp *QualifiedTypeName
    TypeTimestamp *AtomicTypeReference
    QnameEnum *QualifiedTypeName
    TypeEnum *AtomicTypeReference
    QnameSymbolMap *QualifiedTypeName
    TypeSymbolMap *AtomicTypeReference
    QnameStruct *QualifiedTypeName
    TypeStruct *AtomicTypeReference
    QnameNull *QualifiedTypeName
    TypeNull *AtomicTypeReference
)

var coreQnameResolver map[ string ]*QualifiedTypeName
var PrimitiveTypes []*AtomicTypeReference
var NumericTypes []*AtomicTypeReference

var CoreNsV1 = MustNamespace( "mingle:core@v1" )

func init() {
    coreQnameResolver = make( map[ string ]*QualifiedTypeName )
    f := func( s string ) ( *QualifiedTypeName, *AtomicTypeReference ) {
        qn := ( &DeclaredTypeName{ []byte( s ) } ).ResolveIn( CoreNsV1 )
        coreQnameResolver[ qn.Name.ExternalForm() ] = qn
        at := &AtomicTypeReference{ Name: qn }
        return qn, at
    }
    QnameValue, TypeValue = f( "Value" )
    QnameBoolean, TypeBoolean = f( "Boolean" )
    QnameBuffer, TypeBuffer = f( "Buffer" )
    QnameString, TypeString = f( "String" )
    QnameInt32, TypeInt32 = f( "Int32" )
    QnameInt64, TypeInt64 = f( "Int64" )
    QnameUint32, TypeUint32 = f( "Uint32" )
    QnameUint64, TypeUint64 = f( "Uint64" )
    QnameFloat32, TypeFloat32 = f( "Float32" )
    QnameFloat64, TypeFloat64 = f( "Float64" )
    QnameTimestamp, TypeTimestamp = f( "Timestamp" )
    QnameEnum, TypeEnum = f( "Enum" )
    QnameSymbolMap, TypeSymbolMap = f( "SymbolMap" )
    QnameStruct, TypeStruct = f( "Struct" )
    QnameNull, TypeNull = f( "Null" )
    PrimitiveTypes = []*AtomicTypeReference{
        TypeValue,
        TypeNull,
        TypeString,
        TypeFloat64,
        TypeFloat32,
        TypeInt64,
        TypeInt32,
        TypeUint32,
        TypeUint64,
        TypeBoolean,
        TypeTimestamp,
        TypeBuffer,
    }
    NumericTypes = []*AtomicTypeReference{
        TypeInt32,
        TypeInt64,
        TypeUint32,
        TypeUint64,
        TypeFloat32,
        TypeFloat64,
    }
}

func ResolveInCore( nm *DeclaredTypeName ) ( *QualifiedTypeName, bool ) {
    qn, ok := coreQnameResolver[ nm.ExternalForm() ]
    return qn, ok
}

func IsNumericType( typ TypeReference ) bool {
    for _, num := range NumericTypes { if num.Equals( typ ) { return true } }
    return false
}

func IsIntegerType( typ TypeReference ) bool {
    return typ.Equals( TypeInt32 ) || 
           typ.Equals( TypeInt64 ) ||
           typ.Equals( TypeUint32 ) ||
           typ.Equals( TypeUint64 )
}

var typeOpaqueList *ListTypeReference
func init() { typeOpaqueList = &ListTypeReference{ TypeValue, true } }

func TypeOf( mgVal Value ) TypeReference {
    switch v := mgVal.( type ) {
    case Boolean: return TypeBoolean
    case Buffer: return TypeBuffer
    case String: return TypeString
    case Int32: return TypeInt32
    case Int64: return TypeInt64
    case Uint32: return TypeUint32
    case Uint64: return TypeUint64
    case Float32: return TypeFloat32
    case Float64: return TypeFloat64
    case Timestamp: return TypeTimestamp
    case *Enum: return v.Type
    case *SymbolMap: return TypeSymbolMap
    case *Struct: return v.Type
    case *List: return typeOpaqueList
    case *Null: return TypeNull
    }
    panic( fmt.Errorf( "Unhandled arg to typeOf (%T): %v", mgVal, mgVal ) )
}

const typCastMsgWithoutPath = "Expected value of type %s but found %s"
const typCastMsgWithPath = "%s: " + typCastMsgWithoutPath

type TypeCastError struct {
    actual, expected TypeReference
    path idPath
}

func ( e *TypeCastError ) Error() string {
    var msg string
    var off int
    args := make( []interface{}, 3 )
    args[ 1 ], args[ 2 ] = e.expected, e.actual
    if e.path == nil {
        msg, off = typCastMsgWithoutPath, 1
    } else { 
        args[ 0 ] = FormatIdPath( e.path )
        msg, off = typCastMsgWithPath, 0 
    }
    return fmt.Sprintf( msg, args[ off : ]... )
}

func asTypeCastError( 
    path idPath, t TypeReference, mgVal Value ) *TypeCastError {
    return &TypeCastError{ actual: TypeOf( mgVal ), expected: t, path: path }
}

type ValidationError struct {
    message string
    path idPath
}

func ( e *ValidationError ) Message() string { return e.message }

func ( e *ValidationError ) Error() string {
    return fmt.Sprintf( "%s: %s", FormatIdPath( e.path ), e.message )
}

func asValueCastError( 
    path idPath, 
    targType TypeReference,
    msg string, 
    args ...interface{} ) *ValidationError {
    args2 := make( []interface{}, 1, len( args ) + 1 )
    args2[ 0 ] = targType
    args2 = append( args2, args... )
    msg = "Error converting to %s: " + msg
    return &ValidationError{ message: fmt.Sprintf( msg, args2... ), path: path }
}

func strToBool( 
    s String, at *AtomicTypeReference, path idPath ) ( Value, error ) {
    switch lc := strings.ToLower( string( s ) ); lc { 
    case "true": return Boolean( true ), nil
    case "false": return Boolean( false ), nil
    }
    return nil, asValueCastError( path, at, "Invalid boolean value: %s", s )
}

func castBoolean( 
    mgVal Value, at *AtomicTypeReference, path idPath ) ( Value, error ) {
    switch v := mgVal.( type ) {
    case Boolean: return v, nil
    case String: return strToBool( v, at, path )
    }
    return nil, asTypeCastError( path, at, mgVal )
}

func castBuffer( 
    mgVal Value, at *AtomicTypeReference, path idPath ) ( Value, error ) {
    switch v := mgVal.( type ) {
    case Buffer: return v, nil
    case String: 
        buf, err := base64.StdEncoding.DecodeString( string( v ) )
        if err == nil { return Buffer( buf ), nil }
        msg := "Invalid base64 string: %s"
        return nil, asValueCastError( path, at, msg, err.Error() )
    }
    return nil, asTypeCastError( path, at, mgVal )
}

func castString( 
    mgVal Value, at *AtomicTypeReference, path idPath ) ( Value, error ) {
    switch v := mgVal.( type ) {
    case String: return mgVal, nil
    case Boolean, Int32, Int64, Uint32, Uint64, Float32, Float64:
        return String( v.( fmt.Stringer ).String() ), nil
    case Timestamp: return String( v.Rfc3339Nano() ), nil
    case Buffer:
        return String( base64.StdEncoding.EncodeToString( []byte( v ) ) ), nil
    case *Enum: return String( v.Value.ExternalForm() ), nil
    }
    return nil, asTypeCastError( path, at, mgVal )
}

func valueCastErrorForNumError(
    path idPath, at *AtomicTypeReference, err *strconv.NumError ) error {
    return asValueCastError( path, at, "%s: %s", err.Err.Error(), err.Num )
}

func parseIntInitial(
    s String,
    bitSize int,
    numType TypeReference ) ( sInt int64, uInt uint64, err error ) {
    if indx := strings.IndexAny( string( s ), "eE." ); indx >= 0 {
        var f float64
        f, err = strconv.ParseFloat( string( s ), 64 )
        if err == nil { sInt, uInt  = int64( f ), uint64( f ) }
    } else { 
        if numType == TypeUint32 || numType == TypeUint64 {
            uInt, err = strconv.ParseUint( string( s ), 10, bitSize )
            sInt = int64( uInt ) // do this even if err != nil
        } else {
            sInt, err = strconv.ParseInt( string( s ), 10, bitSize )
            uInt = uint64( sInt )
        }
    }
    return
}

func parseInt( 
    s String, 
    bitSize int, 
    numTyp TypeReference, 
    at *AtomicTypeReference, 
    path idPath ) ( Value, error ) {
    sInt, uInt, err := parseIntInitial( s, bitSize, numTyp )
    if err == nil {
        switch numTyp {
        case TypeInt32: return Int32( sInt ), nil
        case TypeInt64: return Int64( sInt ), nil
        case TypeUint32: return Uint32( uInt ), nil
        case TypeUint64: return Uint64( uInt ), nil
        default:
            msg := "Unhandled number type: %s"
            panic( asValueCastError( path, at, msg, numTyp ) )
        }
    } 
    if ne, ok := err.( *strconv.NumError ); ok {
        return nil, valueCastErrorForNumError( path, at, ne )
    }
    return nil, asValueCastError( path, at, err.Error() )
}

func castInt32( 
    mgVal Value, at *AtomicTypeReference, path idPath ) ( Value, error ) {
    switch v := mgVal.( type ) {
    case Int32: return v, nil
    case Int64: return Int32( v ), nil
    case Uint32: return Int32( int32( v ) ), nil
    case Uint64: return Int32( int32( v ) ), nil
    case Float32: return Int32( int32( v ) ), nil
    case Float64: return Int32( int32( v ) ), nil
    case String: return parseInt( v, 32, TypeInt32, at, path )
    }
    return nil, asTypeCastError( path, at, mgVal )
}

func castInt64( 
    mgVal Value, at *AtomicTypeReference, path idPath ) ( Value, error ) {
    switch v := mgVal.( type ) {
    case Int32: return Int64( v ), nil
    case Int64: return v, nil
    case Uint32: return Int64( int64( v ) ), nil
    case Uint64: return Int64( int64( v ) ), nil
    case Float32: return Int64( int64( v ) ), nil
    case Float64: return Int64( int64( v ) ), nil
    case String: return parseInt( v, 64, TypeInt64, at, path )
    }
    return nil, asTypeCastError( path, at, mgVal )
}

func castUint32(
    mgVal Value, at *AtomicTypeReference, path idPath ) ( Value, error ) {
    switch v := mgVal.( type ) {
    case Int32: return Uint32( uint32( v ) ), nil
    case Uint32: return v, nil
    case Int64: return Uint32( uint32( v ) ), nil
    case Uint64: return Uint32( uint32( v ) ), nil
    case Float32: return Uint32( uint32( v ) ), nil
    case Float64: return Uint32( uint32( v ) ), nil
    case String: return parseInt( v, 32, TypeUint32, at, path )
    }
    return nil, asTypeCastError( path, at, mgVal )
}

func castUint64(
    mgVal Value, at *AtomicTypeReference, path idPath ) ( Value, error ) {
    switch v := mgVal.( type ) {
    case Int32: return Uint64( uint64( v ) ), nil
    case Uint32: return Uint64( uint64( v ) ), nil
    case Int64: return Uint64( uint64( v ) ), nil
    case Uint64: return v, nil
    case Float32: return Uint64( uint64( v ) ), nil
    case Float64: return Uint64( uint64( v ) ), nil
    case String: return parseInt( v, 64, TypeUint64, at, path )
    }
    return nil, asTypeCastError( path, at, mgVal )
}

func parseFloat32(
    s string,
    bitSize int,
    numTyp TypeReference,
    at *AtomicTypeReference,
    path idPath ) ( Value, error ) {
    f, err := strconv.ParseFloat( string( s ), bitSize )
    if err != nil { 
        ne := err.( *strconv.NumError )
        return nil, valueCastErrorForNumError( path, at, ne )
    }
    switch numTyp {
    case TypeFloat32: return Float32( f ), nil
    case TypeFloat64: return Float64( f ), nil
    }
    panic( asValueCastError( path, at, "Unhandled num type: %s", numTyp ) )
}

func castFloat32( 
    mgVal Value, at *AtomicTypeReference, path idPath ) ( Value, error ) {
    switch v := mgVal.( type ) {
    case Int32: return Float32( float32( v ) ), nil
    case Int64: return Float32( float32( v ) ), nil
    case Uint32: return Float32( float32( v ) ), nil
    case Uint64: return Float32( float32( v ) ), nil
    case Float32: return v, nil
    case Float64: return Float32( float32( v ) ), nil
    case String: return parseFloat32( string( v ), 32, TypeFloat32, at, path )
    }
    return nil, asTypeCastError( path, at, mgVal )
}

func castFloat64( 
    mgVal Value, at *AtomicTypeReference, path idPath ) ( Value, error ) {
    switch v := mgVal.( type ) {
    case Int32: return Float64( float64( v ) ), nil
    case Int64: return Float64( float64( v ) ), nil
    case Uint32: return Float64( float64( v ) ), nil
    case Uint64: return Float64( float64( v ) ), nil
    case Float32: return Float64( float64( v ) ), nil
    case Float64: return v, nil
    case String: return parseFloat32( string( v ), 64, TypeFloat64, at, path )
    }
    return nil, asTypeCastError( path, at, mgVal )
}

func castTimestamp( 
    mgVal Value, at *AtomicTypeReference, path idPath ) ( Value, error ) {
    switch v := mgVal.( type ) {
    case Timestamp: return v, nil
    case String:
        tm, err := ParseTimestamp( string( v ) )
        if err == nil { return tm, nil }
        msg := "Invalid timestamp: %s"
        return nil, asValueCastError( path, at, msg, err.Error() )
    }
    return nil, asTypeCastError( path, at, mgVal )
}

func castEnum( 
    mgVal Value, at *AtomicTypeReference, path idPath ) ( Value, error ) {
    switch v := mgVal.( type ) {
    case *Enum: 
        if at.Equals( TypeEnum ) || v.Type.Equals( at ) { return v, nil }
    }
    return nil, asTypeCastError( path, at, mgVal )
}

func castSymbolMap( 
    mgVal Value, at *AtomicTypeReference, path idPath ) ( Value, error ) {
    switch v := mgVal.( type ) {
    case *SymbolMap: return v, nil
    }
    return nil, asTypeCastError( path, at, mgVal )
}

func castStruct( 
    mgVal Value, at *AtomicTypeReference, path idPath ) ( Value, error ) {
    switch v := mgVal.( type ) {
    case *Struct: 
        if at.Equals( TypeStruct ) || v.Type.Equals( at ) { return v, nil }
    }
    return nil, asTypeCastError( path, at, mgVal )
}

func castNull( 
    mgVal Value, at *AtomicTypeReference, path idPath ) ( Value, error ) {
    if _, ok := mgVal.( *Null ); ok { return mgVal, nil }
    return nil, asTypeCastError( path, at, mgVal )
}

func castAtomicUnrestricted(
    mgVal Value, at *AtomicTypeReference, path idPath ) ( Value, error ) {
    if _, ok := mgVal.( *Null ); ok {
        if at.Equals( TypeNull ) { return mgVal, nil }
        return nil, asValueCastError( path, at, "value is null" )
    }
    switch nm := at.Name; {
    case nm.Equals( QnameBoolean ): return castBoolean( mgVal, at, path )
    case nm.Equals( QnameBuffer ): return castBuffer( mgVal, at, path )
    case nm.Equals( QnameString ): return castString( mgVal, at, path )
    case nm.Equals( QnameInt32 ): return castInt32( mgVal, at, path )
    case nm.Equals( QnameInt64 ): return castInt64( mgVal, at, path )
    case nm.Equals( QnameUint32 ): return castUint32( mgVal, at, path )
    case nm.Equals( QnameUint64 ): return castUint64( mgVal, at, path )
    case nm.Equals( QnameFloat32 ): return castFloat32( mgVal, at, path )
    case nm.Equals( QnameFloat64 ): return castFloat64( mgVal, at, path )
    case nm.Equals( QnameTimestamp ): return castTimestamp( mgVal, at, path )
    case nm.Equals( QnameSymbolMap ): return castSymbolMap( mgVal, at, path )
    case nm.Equals( QnameEnum ): return castEnum( mgVal, at, path )
    case nm.Equals( QnameNull ): return castNull( mgVal, at, path )
    case nm.Equals( QnameValue ): return mgVal, nil
    }
    switch mgVal.( type ) {
    case *Enum: return castEnum( mgVal, at, path )
    case *Struct: return castStruct( mgVal, at, path )
    }
    return nil, asTypeCastError( path, at, mgVal )
}

func checkRestriction( val Value, at *AtomicTypeReference, path idPath ) error {
    if at.Restriction.AcceptsValue( val ) { return nil }
    return asValueCastError( 
        path, at, "Value %s does not satisfy restriction %s",
        QuoteValue( val ), at.Restriction.ExternalForm() )
}

// switch compares based on qname not at itself since we may be dealing with
// restriction types, meaning that if at is mingle:core@v1/String~"a", it is a
// string (has qname mingle:core@v1/String) but will not equal TypeString itself
func castAtomic(
    mgVal Value, 
    at *AtomicTypeReference,
    path idPath ) ( val Value, err error ) {
    if val, err = castAtomicUnrestricted( mgVal, at, path ); err == nil {
        if at.Restriction != nil { err = checkRestriction( val, at, path ) }
    }
    return
}

// Currently the list returned is a new *List representing a shallow copy of
// mgVal (assuming it is a *List) with each original element in the original
// list represented in the output by the result of casting it to the list
// element type.
//
// In cases when no actual casts need to take place (which may prove the common
// case), a space optimization will be to only lazily initiate a copy, returning
// the original list untouched if possible. That's tabled for now though.
func castList(
    mgVal Value, lt *ListTypeReference, path idPath ) ( Value, error ) {
    if ml, ok := mgVal.( *List ); ok {
        eltTyp := lt.ElementType
        vals := make( []Value, len( ml.vals ) )
        lp := path.StartList()
        for i, inVal := range ml.vals {
            if val, err := CastValue( inVal, eltTyp, lp ); err == nil {
                vals[ i ] = val
            } else { return nil, err }
            lp = lp.Next()
        }
        if len( vals ) == 0 && ( ! lt.AllowsEmpty ) {
            return nil, asValueCastError( path, lt, "list is empty" )
        }
        return &List{ vals }, nil
    }
    return nil, asTypeCastError( path, lt, mgVal )
}

func castNullable(
    mgVal Value, nt *NullableTypeReference, path idPath ) ( Value, error ) {
    if nv, ok := mgVal.( *Null ); ok { return nv, nil }
    return CastValue( mgVal, nt.Type, path )
}

func CastValue( 
    mgVal Value, typ TypeReference, path objpath.PathNode ) ( Value, error ) {
    if mgVal == nil { return nil, errors.New( "mgVal is nil" ) }
    switch v := typ.( type ) {
    case *AtomicTypeReference: return castAtomic( mgVal, v, path )
    case *ListTypeReference: return castList( mgVal, v, path )
    case *NullableTypeReference: return castNullable( mgVal, v, path )
    }
    panic( fmt.Errorf( "Unhandled target type (%T): %s", typ, typ ) )
}

type mapImplKey interface { ExternalForm() string }

type mapImplEntry struct { 
    key mapImplKey
    val interface{} 
}

type mapImpl struct {
    m map[ string ]mapImplEntry
}

func newMapImpl() *mapImpl { 
    return &mapImpl{ make( map[ string ]mapImplEntry ) }
}

func ( m *mapImpl ) Len() int { return len( m.m ) }

func ( m *mapImpl ) implGet( k mapImplKey ) interface{} {
    if e, ok := m.m[ k.ExternalForm() ]; ok { return e.val }
    return nil
}

func ( m *mapImpl ) implHasKey( k mapImplKey ) bool {
    return m.implGet( k ) != nil
}

func ( m *mapImpl ) implPut( k mapImplKey, v interface{} ) {
    m.m[ k.ExternalForm() ] = mapImplEntry{ k, v }
}

func ( m *mapImpl ) implPutSafe( k mapImplKey, v interface{} ) error {
    kStr := k.ExternalForm()
    if _, ok := m.m[ kStr ]; ok {
        tmpl := "mingle: map already contains an entry for key: %s"
        return fmt.Errorf( tmpl, kStr )
    } 
    m.implPut( k, v )
    return nil
}

func ( m *mapImpl ) implDelete( k mapImplKey ) {
    delete( m.m, k.ExternalForm() )
}

func ( m *mapImpl ) implEachPairError(
    f func( k mapImplKey, val interface{} ) error ) error {
    for _, entry := range m.m { 
        if err := f( entry.key, entry.val ); err != nil { return err }
    }
    return nil
}

func ( m *mapImpl ) implEachPair( f func( k mapImplKey, val interface{} ) ) {
    m.implEachPairError( func( k mapImplKey, val interface{} ) error {
        f( k, val )
        return nil
    })
}

type IdentifierMap struct { *mapImpl }

func NewIdentifierMap() *IdentifierMap { return &IdentifierMap{ newMapImpl() } }

func ( m *IdentifierMap ) Get( id *Identifier ) interface{} {
    return m.implGet( id )
}

func ( m *IdentifierMap ) HasKey( id *Identifier ) bool {
    return m.implHasKey( id )
}

func ( m *IdentifierMap ) Delete( id *Identifier ) { m.implDelete( id ) }

func ( m *IdentifierMap ) Put( id *Identifier, val interface{} ) {
    m.implPut( id, val )
}

func ( m *IdentifierMap ) PutSafe( id *Identifier, val interface{} ) error {
    return m.implPutSafe( id, val )
}

func ( m *IdentifierMap ) EachPairError( 
    f func( id *Identifier, val interface{} ) error ) error {
    return m.implEachPairError(
        func( k mapImplKey, val interface{} ) error {
            return f( k.( *Identifier ), val )
        },
    )
}

func ( m *IdentifierMap ) EachPair( 
    f func( id *Identifier, val interface{} ) ) {
    m.implEachPair(
        func( k mapImplKey, val interface{} ) { f( k.( *Identifier ), val ) } )
}

type QnameMap struct { *mapImpl }

func NewQnameMap() *QnameMap { return &QnameMap{ newMapImpl() } }

func ( m *QnameMap ) Get( qn *QualifiedTypeName ) interface{} {
    return m.implGet( qn )
}

func ( m *QnameMap ) HasKey( qn *QualifiedTypeName ) bool {
    return m.implHasKey( qn )
}

func ( m *QnameMap ) Put( qn *QualifiedTypeName, val interface{} ) {
    m.implPut( qn, val )
}

func ( m *QnameMap ) PutSafe( qn *QualifiedTypeName, val interface{} ) error {
    return m.implPutSafe( qn, val )
}

func ( m *QnameMap ) Delete( qn *QualifiedTypeName ) { m.implDelete( qn ) }

func ( m *QnameMap ) EachPair( 
    f func( qn *QualifiedTypeName, val interface{} ) ) {
    m.implEachPair( 
        func( k mapImplKey, v interface{} ) {
            f( k.( *QualifiedTypeName ), v )
        },
    )
}

type NamespaceMap struct { *mapImpl }

func NewNamespaceMap() *NamespaceMap { return &NamespaceMap{ newMapImpl() } }

func ( m *NamespaceMap ) Get( ns *Namespace ) interface{} {
    return m.implGet( ns )
}

func ( m *NamespaceMap ) HasKey( ns *Namespace ) bool {
    return m.implHasKey( ns )
}

func ( m *NamespaceMap ) Put( ns *Namespace, val interface{} ) {
    m.implPut( ns, val )
}

func ( m *NamespaceMap ) PutSafe( ns *Namespace, val interface{} ) error {
    return m.implPutSafe( ns, val )
}

func ( m *NamespaceMap ) Delete( ns *Namespace ) { m.implDelete( ns ) }


// -----------------------------------------------------------------------------
// Begin autogenerated accessor funcs
// -----------------------------------------------------------------------------

func ( acc *SymbolMapAccessor ) GetBooleanById(
    id *Identifier ) ( res Boolean, err error ) {
    var val Value
    val, err = acc.GetValueById( id )
    if err == nil {
        val, err = CastValue( val, TypeBoolean, acc.descend( id ) )
        if err == nil { res = val.( Boolean ) }
    }
    return 
}

func ( acc *SymbolMapAccessor ) GetBooleanByString(
    id string ) ( Boolean, error ) {
    return acc.GetBooleanById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) MustBooleanById( 
    id *Identifier ) Boolean {
    res, err := acc.GetBooleanById( id )
    if err != nil { panic( err ) }
    return res
}

func ( acc *SymbolMapAccessor ) MustBooleanByString( 
    id string ) Boolean {
    return acc.MustBooleanById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) GetGoBoolById( 
    id *Identifier ) ( res bool, err error ) {
    var val Boolean
    val, err = acc.GetBooleanById( id )
    if err == nil { res = bool( val ) }
    return
}

func ( acc *SymbolMapAccessor ) GetGoBoolByString( 
    id string ) ( bool, error ) {
    return acc.GetGoBoolById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) MustGoBoolById( 
    id *Identifier ) bool {
    s, err := acc.GetGoBoolById( id )
    if err != nil { panic( err ) }
    return s
}

func ( acc *SymbolMapAccessor ) MustGoBoolByString( 
    id string ) bool {
    return acc.MustGoBoolById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) GetBufferById(
    id *Identifier ) ( res Buffer, err error ) {
    var val Value
    val, err = acc.GetValueById( id )
    if err == nil {
        val, err = CastValue( val, TypeBuffer, acc.descend( id ) )
        if err == nil { res = val.( Buffer ) }
    }
    return 
}

func ( acc *SymbolMapAccessor ) GetBufferByString(
    id string ) ( Buffer, error ) {
    return acc.GetBufferById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) MustBufferById( 
    id *Identifier ) Buffer {
    res, err := acc.GetBufferById( id )
    if err != nil { panic( err ) }
    return res
}

func ( acc *SymbolMapAccessor ) MustBufferByString( 
    id string ) Buffer {
    return acc.MustBufferById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) GetGoBufferById( 
    id *Identifier ) ( res []byte, err error ) {
    var val Buffer
    val, err = acc.GetBufferById( id )
    if err == nil { res = []byte( val ) }
    return
}

func ( acc *SymbolMapAccessor ) GetGoBufferByString( 
    id string ) ( []byte, error ) {
    return acc.GetGoBufferById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) MustGoBufferById( 
    id *Identifier ) []byte {
    s, err := acc.GetGoBufferById( id )
    if err != nil { panic( err ) }
    return s
}

func ( acc *SymbolMapAccessor ) MustGoBufferByString( 
    id string ) []byte {
    return acc.MustGoBufferById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) GetStringById(
    id *Identifier ) ( res String, err error ) {
    var val Value
    val, err = acc.GetValueById( id )
    if err == nil {
        val, err = CastValue( val, TypeString, acc.descend( id ) )
        if err == nil { res = val.( String ) }
    }
    return 
}

func ( acc *SymbolMapAccessor ) GetStringByString(
    id string ) ( String, error ) {
    return acc.GetStringById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) MustStringById( 
    id *Identifier ) String {
    res, err := acc.GetStringById( id )
    if err != nil { panic( err ) }
    return res
}

func ( acc *SymbolMapAccessor ) MustStringByString( 
    id string ) String {
    return acc.MustStringById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) GetGoStringById( 
    id *Identifier ) ( res string, err error ) {
    var val String
    val, err = acc.GetStringById( id )
    if err == nil { res = string( val ) }
    return
}

func ( acc *SymbolMapAccessor ) GetGoStringByString( 
    id string ) ( string, error ) {
    return acc.GetGoStringById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) MustGoStringById( 
    id *Identifier ) string {
    s, err := acc.GetGoStringById( id )
    if err != nil { panic( err ) }
    return s
}

func ( acc *SymbolMapAccessor ) MustGoStringByString( 
    id string ) string {
    return acc.MustGoStringById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) GetInt32ById(
    id *Identifier ) ( res Int32, err error ) {
    var val Value
    val, err = acc.GetValueById( id )
    if err == nil {
        val, err = CastValue( val, TypeInt32, acc.descend( id ) )
        if err == nil { res = val.( Int32 ) }
    }
    return 
}

func ( acc *SymbolMapAccessor ) GetInt32ByString(
    id string ) ( Int32, error ) {
    return acc.GetInt32ById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) MustInt32ById( 
    id *Identifier ) Int32 {
    res, err := acc.GetInt32ById( id )
    if err != nil { panic( err ) }
    return res
}

func ( acc *SymbolMapAccessor ) MustInt32ByString( 
    id string ) Int32 {
    return acc.MustInt32ById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) GetGoInt32ById( 
    id *Identifier ) ( res int32, err error ) {
    var val Int32
    val, err = acc.GetInt32ById( id )
    if err == nil { res = int32( val ) }
    return
}

func ( acc *SymbolMapAccessor ) GetGoInt32ByString( 
    id string ) ( int32, error ) {
    return acc.GetGoInt32ById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) MustGoInt32ById( 
    id *Identifier ) int32 {
    s, err := acc.GetGoInt32ById( id )
    if err != nil { panic( err ) }
    return s
}

func ( acc *SymbolMapAccessor ) MustGoInt32ByString( 
    id string ) int32 {
    return acc.MustGoInt32ById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) GetInt64ById(
    id *Identifier ) ( res Int64, err error ) {
    var val Value
    val, err = acc.GetValueById( id )
    if err == nil {
        val, err = CastValue( val, TypeInt64, acc.descend( id ) )
        if err == nil { res = val.( Int64 ) }
    }
    return 
}

func ( acc *SymbolMapAccessor ) GetInt64ByString(
    id string ) ( Int64, error ) {
    return acc.GetInt64ById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) MustInt64ById( 
    id *Identifier ) Int64 {
    res, err := acc.GetInt64ById( id )
    if err != nil { panic( err ) }
    return res
}

func ( acc *SymbolMapAccessor ) MustInt64ByString( 
    id string ) Int64 {
    return acc.MustInt64ById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) GetGoInt64ById( 
    id *Identifier ) ( res int64, err error ) {
    var val Int64
    val, err = acc.GetInt64ById( id )
    if err == nil { res = int64( val ) }
    return
}

func ( acc *SymbolMapAccessor ) GetGoInt64ByString( 
    id string ) ( int64, error ) {
    return acc.GetGoInt64ById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) MustGoInt64ById( 
    id *Identifier ) int64 {
    s, err := acc.GetGoInt64ById( id )
    if err != nil { panic( err ) }
    return s
}

func ( acc *SymbolMapAccessor ) MustGoInt64ByString( 
    id string ) int64 {
    return acc.MustGoInt64ById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) GetUint32ById(
    id *Identifier ) ( res Uint32, err error ) {
    var val Value
    val, err = acc.GetValueById( id )
    if err == nil {
        val, err = CastValue( val, TypeUint32, acc.descend( id ) )
        if err == nil { res = val.( Uint32 ) }
    }
    return 
}

func ( acc *SymbolMapAccessor ) GetUint32ByString(
    id string ) ( Uint32, error ) {
    return acc.GetUint32ById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) MustUint32ById( 
    id *Identifier ) Uint32 {
    res, err := acc.GetUint32ById( id )
    if err != nil { panic( err ) }
    return res
}

func ( acc *SymbolMapAccessor ) MustUint32ByString( 
    id string ) Uint32 {
    return acc.MustUint32ById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) GetGoUint32ById( 
    id *Identifier ) ( res uint32, err error ) {
    var val Uint32
    val, err = acc.GetUint32ById( id )
    if err == nil { res = uint32( val ) }
    return
}

func ( acc *SymbolMapAccessor ) GetGoUint32ByString( 
    id string ) ( uint32, error ) {
    return acc.GetGoUint32ById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) MustGoUint32ById( 
    id *Identifier ) uint32 {
    s, err := acc.GetGoUint32ById( id )
    if err != nil { panic( err ) }
    return s
}

func ( acc *SymbolMapAccessor ) MustGoUint32ByString( 
    id string ) uint32 {
    return acc.MustGoUint32ById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) GetUint64ById(
    id *Identifier ) ( res Uint64, err error ) {
    var val Value
    val, err = acc.GetValueById( id )
    if err == nil {
        val, err = CastValue( val, TypeUint64, acc.descend( id ) )
        if err == nil { res = val.( Uint64 ) }
    }
    return 
}

func ( acc *SymbolMapAccessor ) GetUint64ByString(
    id string ) ( Uint64, error ) {
    return acc.GetUint64ById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) MustUint64ById( 
    id *Identifier ) Uint64 {
    res, err := acc.GetUint64ById( id )
    if err != nil { panic( err ) }
    return res
}

func ( acc *SymbolMapAccessor ) MustUint64ByString( 
    id string ) Uint64 {
    return acc.MustUint64ById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) GetGoUint64ById( 
    id *Identifier ) ( res uint64, err error ) {
    var val Uint64
    val, err = acc.GetUint64ById( id )
    if err == nil { res = uint64( val ) }
    return
}

func ( acc *SymbolMapAccessor ) GetGoUint64ByString( 
    id string ) ( uint64, error ) {
    return acc.GetGoUint64ById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) MustGoUint64ById( 
    id *Identifier ) uint64 {
    s, err := acc.GetGoUint64ById( id )
    if err != nil { panic( err ) }
    return s
}

func ( acc *SymbolMapAccessor ) MustGoUint64ByString( 
    id string ) uint64 {
    return acc.MustGoUint64ById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) GetFloat32ById(
    id *Identifier ) ( res Float32, err error ) {
    var val Value
    val, err = acc.GetValueById( id )
    if err == nil {
        val, err = CastValue( val, TypeFloat32, acc.descend( id ) )
        if err == nil { res = val.( Float32 ) }
    }
    return 
}

func ( acc *SymbolMapAccessor ) GetFloat32ByString(
    id string ) ( Float32, error ) {
    return acc.GetFloat32ById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) MustFloat32ById( 
    id *Identifier ) Float32 {
    res, err := acc.GetFloat32ById( id )
    if err != nil { panic( err ) }
    return res
}

func ( acc *SymbolMapAccessor ) MustFloat32ByString( 
    id string ) Float32 {
    return acc.MustFloat32ById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) GetGoFloat32ById( 
    id *Identifier ) ( res float32, err error ) {
    var val Float32
    val, err = acc.GetFloat32ById( id )
    if err == nil { res = float32( val ) }
    return
}

func ( acc *SymbolMapAccessor ) GetGoFloat32ByString( 
    id string ) ( float32, error ) {
    return acc.GetGoFloat32ById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) MustGoFloat32ById( 
    id *Identifier ) float32 {
    s, err := acc.GetGoFloat32ById( id )
    if err != nil { panic( err ) }
    return s
}

func ( acc *SymbolMapAccessor ) MustGoFloat32ByString( 
    id string ) float32 {
    return acc.MustGoFloat32ById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) GetFloat64ById(
    id *Identifier ) ( res Float64, err error ) {
    var val Value
    val, err = acc.GetValueById( id )
    if err == nil {
        val, err = CastValue( val, TypeFloat64, acc.descend( id ) )
        if err == nil { res = val.( Float64 ) }
    }
    return 
}

func ( acc *SymbolMapAccessor ) GetFloat64ByString(
    id string ) ( Float64, error ) {
    return acc.GetFloat64ById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) MustFloat64ById( 
    id *Identifier ) Float64 {
    res, err := acc.GetFloat64ById( id )
    if err != nil { panic( err ) }
    return res
}

func ( acc *SymbolMapAccessor ) MustFloat64ByString( 
    id string ) Float64 {
    return acc.MustFloat64ById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) GetGoFloat64ById( 
    id *Identifier ) ( res float64, err error ) {
    var val Float64
    val, err = acc.GetFloat64ById( id )
    if err == nil { res = float64( val ) }
    return
}

func ( acc *SymbolMapAccessor ) GetGoFloat64ByString( 
    id string ) ( float64, error ) {
    return acc.GetGoFloat64ById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) MustGoFloat64ById( 
    id *Identifier ) float64 {
    s, err := acc.GetGoFloat64ById( id )
    if err != nil { panic( err ) }
    return s
}

func ( acc *SymbolMapAccessor ) MustGoFloat64ByString( 
    id string ) float64 {
    return acc.MustGoFloat64ById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) GetTimestampById(
    id *Identifier ) ( res Timestamp, err error ) {
    var val Value
    val, err = acc.GetValueById( id )
    if err == nil {
        val, err = CastValue( val, TypeTimestamp, acc.descend( id ) )
        if err == nil { res = val.( Timestamp ) }
    }
    return 
}

func ( acc *SymbolMapAccessor ) GetTimestampByString(
    id string ) ( Timestamp, error ) {
    return acc.GetTimestampById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) MustTimestampById( 
    id *Identifier ) Timestamp {
    res, err := acc.GetTimestampById( id )
    if err != nil { panic( err ) }
    return res
}

func ( acc *SymbolMapAccessor ) MustTimestampByString( 
    id string ) Timestamp {
    return acc.MustTimestampById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) GetEnumById(
    id *Identifier ) ( res *Enum, err error ) {
    var val Value
    val, err = acc.GetValueById( id )
    if err == nil {
        val, err = CastValue( val, TypeEnum, acc.descend( id ) )
        if err == nil { res = val.( *Enum ) }
    }
    return 
}

func ( acc *SymbolMapAccessor ) GetEnumByString(
    id string ) ( *Enum, error ) {
    return acc.GetEnumById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) MustEnumById( 
    id *Identifier ) *Enum {
    res, err := acc.GetEnumById( id )
    if err != nil { panic( err ) }
    return res
}

func ( acc *SymbolMapAccessor ) MustEnumByString( 
    id string ) *Enum {
    return acc.MustEnumById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) GetSymbolMapById(
    id *Identifier ) ( res *SymbolMap, err error ) {
    var val Value
    val, err = acc.GetValueById( id )
    if err == nil {
        val, err = CastValue( val, TypeSymbolMap, acc.descend( id ) )
        if err == nil { res = val.( *SymbolMap ) }
    }
    return 
}

func ( acc *SymbolMapAccessor ) GetSymbolMapByString(
    id string ) ( *SymbolMap, error ) {
    return acc.GetSymbolMapById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) MustSymbolMapById( 
    id *Identifier ) *SymbolMap {
    res, err := acc.GetSymbolMapById( id )
    if err != nil { panic( err ) }
    return res
}

func ( acc *SymbolMapAccessor ) MustSymbolMapByString( 
    id string ) *SymbolMap {
    return acc.MustSymbolMapById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) GetStructById(
    id *Identifier ) ( res *Struct, err error ) {
    var val Value
    val, err = acc.GetValueById( id )
    if err == nil {
        val, err = CastValue( val, TypeStruct, acc.descend( id ) )
        if err == nil { res = val.( *Struct ) }
    }
    return 
}

func ( acc *SymbolMapAccessor ) GetStructByString(
    id string ) ( *Struct, error ) {
    return acc.GetStructById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) MustStructById( 
    id *Identifier ) *Struct {
    res, err := acc.GetStructById( id )
    if err != nil { panic( err ) }
    return res
}

func ( acc *SymbolMapAccessor ) MustStructByString( 
    id string ) *Struct {
    return acc.MustStructById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) GetListById(
    id *Identifier ) ( res *List, err error ) {
    var val Value
    val, err = acc.GetValueById( id )
    if err == nil {
        val, err = CastValue( val, typeOpaqueList, acc.descend( id ) )
        if err == nil { res = val.( *List ) }
    }
    return 
}

func ( acc *SymbolMapAccessor ) GetListByString(
    id string ) ( *List, error ) {
    return acc.GetListById( MustIdentifier( id ) )
}

func ( acc *SymbolMapAccessor ) MustListById( 
    id *Identifier ) *List {
    res, err := acc.GetListById( id )
    if err != nil { panic( err ) }
    return res
}

func ( acc *SymbolMapAccessor ) MustListByString( 
    id string ) *List {
    return acc.MustListById( MustIdentifier( id ) )
}

// -----------------------------------------------------------------------------
// End autogenerated accessor funcs
// -----------------------------------------------------------------------------


// # This script generates various type-specific (mingle and go) accessors. When
// # run, it will output the various functions it generates, followed by this
// # script itself, so that the functions as well as the script which generates it
// # may be kept together in the target sourcefile.
// 
// $specs = [
//     { mg_type: "Boolean" },
//     { go_type: "bool", mg_type: "Boolean" },
//     { mg_type: "Buffer" },
//     { go_type: "[]byte", go_type_name: "GoBuffer", mg_type: "Buffer" },
//     { mg_type: "String" },
//     { go_type: "string" },
//     { mg_type: "Int32" },
//     { go_type: "int32" },
//     { mg_type: "Int64" },
//     { go_type: "int64" },
//     { mg_type: "Uint32" },
//     { go_type: "uint32" },
//     { mg_type: "Uint64" },
//     { go_type: "uint64" },
//     { mg_type: "Float32" },
//     { go_type: "float32" },
//     { mg_type: "Float64" },
//     { go_type: "float64" },
//     { mg_type: "Timestamp" },
//     { mg_type: "*Enum" },
//     { mg_type: "*SymbolMap" },
//     { mg_type: "*Struct" },
//     { mg_type: "*List", cast_type: "typeOpaqueList" },
// ]
// 
// def get_mg_type_name( spec )
//     spec[ :mg_type ].sub( /^\*/, "" )
// end
// 
// def make_go_funcs( spec )
// 
//     go_typ = spec[ :go_type ]
//     go_typ_nm = (
//         spec[ :go_type_name ] or
//         ( "Go" + go_typ[ 0 ].sub( /^\*/, "" ).upcase + go_typ[ 1 .. -1 ] ) )
//     mg_typ = ( spec[ :mg_type ] || ( go_typ[ 0 ].upcase + go_typ[ 1 .. -1 ] ) )
//     mg_typ_nm = mg_typ
//     cast_fmt = ( spec[ :cast_fmt ] || "#{go_typ}( %s )" )
// 
//     <<-FUNCS
// func ( acc *SymbolMapAccessor ) Get#{go_typ_nm}ById( 
//     id *Identifier ) ( res #{go_typ}, err error ) {
//     var val #{mg_typ}
//     val, err = acc.Get#{mg_typ_nm}ById( id )
//     if err == nil { res = #{sprintf( cast_fmt, "val" )} }
//     return
// }
// 
// func ( acc *SymbolMapAccessor ) Get#{go_typ_nm}ByString( 
//     id string ) ( #{go_typ}, error ) {
//     return acc.Get#{go_typ_nm}ById( MustIdentifier( id ) )
// }
// 
// func ( acc *SymbolMapAccessor ) Must#{go_typ_nm}ById( 
//     id *Identifier ) #{go_typ} {
//     s, err := acc.Get#{go_typ_nm}ById( id )
//     if err != nil { panic( err ) }
//     return s
// }
// 
// func ( acc *SymbolMapAccessor ) Must#{go_typ_nm}ByString( 
//     id string ) #{go_typ} {
//     return acc.Must#{go_typ_nm}ById( MustIdentifier( id ) )
// }
//     FUNCS
// end
// 
// def make_mg_funcs( spec )
//     typ_nm = get_mg_type_name( spec )
//     typ = spec[ :mg_type ]
//     cast_type = ( spec[ :cast_type ] || "Type#{typ_nm}" )
// 
//     <<-FUNCS
// func ( acc *SymbolMapAccessor ) Get#{typ_nm}ById(
//     id *Identifier ) ( res #{typ}, err error ) {
//     var val Value
//     val, err = acc.GetValueById( id )
//     if err == nil {
//         val, err = CastValue( val, #{cast_type}, acc.descend( id ) )
//         if err == nil { res = val.( #{typ} ) }
//     }
//     return 
// }
// 
// func ( acc *SymbolMapAccessor ) Get#{typ_nm}ByString(
//     id string ) ( #{typ}, error ) {
//     return acc.Get#{typ_nm}ById( MustIdentifier( id ) )
// }
// 
// func ( acc *SymbolMapAccessor ) Must#{typ_nm}ById( 
//     id *Identifier ) #{typ} {
//     res, err := acc.Get#{typ_nm}ById( id )
//     if err != nil { panic( err ) }
//     return res
// }
// 
// func ( acc *SymbolMapAccessor ) Must#{typ_nm}ByString( 
//     id string ) #{typ} {
//     return acc.Must#{typ_nm}ById( MustIdentifier( id ) )
// }
//     FUNCS
// end
// 
// def make_funcs( spec )
//     if spec.key?( :go_type )
//         make_go_funcs( spec )
//     else
//         make_mg_funcs( spec )
//     end
// end
// 
// funcs = $specs.map { |spec| make_funcs( spec ) }.flatten
// 
// dashes = "// " + ( "-" * 77 )
// title = "autogenerated accessor funcs"
// puts "\n#{dashes}\n// Begin #{title}\n#{dashes}\n\n"
// puts funcs.join( "\n" )
// puts "\n#{dashes}\n// End #{title}\n#{dashes}"
// 
// puts "\n\n"
// puts File.open( __FILE__ ) { |io| io.readlines }.map { |s| "// #{s}" }
