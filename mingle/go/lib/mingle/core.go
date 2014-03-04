package mingle

import (
    "bitgirder/objpath"
    "fmt"
    "regexp"
    "log"
    "time"
    "sort"
    "bytes"
    "unicode"
    "strings"
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
    case *Identifier: 
        if v == nil { panic( libErrorf( "nil identifier" ) ) }
        return v, nil
    }
    return nil, fmt.Errorf( "Unhandled id initializer: %v", id )
}

type Identifier struct {
    parts []string
}

// Meant for other mingle impl packages only; external callers should not count
// on the behavior of this remaining stable, or even of it continuing to exist
func NewIdentifierUnsafe( parts []string ) *Identifier {
    return &Identifier{ parts }
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

func compareIdParts( p1, p2 string ) int {
    if p1 == p2 { return 0 }
    if p1 < p2 { return -1 }
    return 1
}

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

type idSort []*Identifier
func ( s idSort ) Len() int { return len( s ) }
func ( s idSort ) Less( i, j int ) bool { return s[ i ].Compare( s[ j ] ) < 0 }
func ( s idSort ) Swap( i, j int ) { s[ j ], s[ i ] = s[ i ], s[ j ] }

func SortIds( ids []*Identifier ) []*Identifier { 
    sort.Sort( idSort( ids ) ) 
    return ids
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
    typeNameImpl()
}

type DeclaredTypeName struct { nm string }

func NewDeclaredTypeNameUnsafe( nm string ) *DeclaredTypeName {
    return &DeclaredTypeName{ nm }
}

func ( n *DeclaredTypeName ) typeNameImpl() {}

func ( n *DeclaredTypeName ) String() string { return n.nm }

func ( n *DeclaredTypeName ) ExternalForm() string { return n.String() }

func ( n *DeclaredTypeName ) Equals( other TypeName ) bool {
    if n2, ok := other.( *DeclaredTypeName ); ok { return n.nm == n2.nm }
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

func ( qn *QualifiedTypeName ) typeNameImpl() {}

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

func ( qn *QualifiedTypeName ) AsAtomicType() *AtomicTypeReference {
    return &AtomicTypeReference{ Name: qn }
}

// (atomic|list|nullable)
type TypeReference interface {
    ExternalForm() string
    Equals( t TypeReference ) bool
    String() string
    typeRefImpl()
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

func ( t *AtomicTypeReference ) typeRefImpl() {}

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

func ( t *ListTypeReference ) typeRefImpl() {}

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

func ( t *NullableTypeReference ) typeRefImpl() {}

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
    case *QualifiedTypeName, *DeclaredTypeName:
        return &AtomicTypeReference{ Name: v.( TypeName ) }
    }
    panic( libErrorf( "Unhandled type ref initializer: %T", typ ) )
}

func AtomicTypeIn( ref TypeReference ) *AtomicTypeReference {
    switch v := ref.( type ) {
    case *AtomicTypeReference: return v
    case *ListTypeReference: return AtomicTypeIn( v.ElementType )
    case *NullableTypeReference: return AtomicTypeIn( v.Type )
    }
    panic( fmt.Errorf( "No atomic type in %s (%T)", ref, ref ) )
}

type Value interface{ valImpl() }

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
    case Buffer: fmt.Fprintf( buf, "buf[%x]", []byte( v ) )
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

type ValueCastError struct {
    ve ValueErrorImpl
    msg string
}

func ( e *ValueCastError ) Message() string { return e.msg }

func ( e *ValueCastError ) Error() string { 
    return e.ve.MakeError( e.Message() ) 
}

func ( e *ValueCastError ) Location() objpath.PathNode { 
    return e.ve.Location() 
}

func NewValueCastError( path idPath, msg string ) *ValueCastError {
    res := &ValueCastError{ msg: msg, ve: ValueErrorImpl{} }
    res.ve.Path = path
    return res
}

func NewValueCastErrorf(
    path idPath, tmpl string, args ...interface{} ) *ValueCastError {
    return NewValueCastError( path, fmt.Sprintf( tmpl, args... ) )
}

func NewTypeCastError( 
    expct, act TypeReference, path objpath.PathNode ) *ValueCastError {
    return NewValueCastErrorf( 
        path,
        "Expected value of type %s but found %s",
        expct.ExternalForm(), act.ExternalForm(),
    )
}

func NewTypeCastErrorValue( 
    t TypeReference, val Value, path objpath.PathNode ) *ValueCastError {
    return NewTypeCastError( t, TypeOf( val ), path )
}

type Comparer interface {

    // Should panic (such as with type assertion panic) if val is not same type
    // as the instance or is nil
    Compare( val interface{} ) int
}

type String string
func ( s String ) valImpl() {}
func ( s String ) String() string { return string( s ) }

func ( s String ) Compare( val interface{} ) int {
    switch s2 := val.( String ); {
    case s < s2: return -1
    case s == s2: return 0
    }
    return 1
}

type Boolean bool
func ( b Boolean ) valImpl() {}
func ( b Boolean ) String() string { return fmt.Sprint( bool( b ) ) }

type Int64 int64
func ( i Int64 ) valImpl() {}
func ( i Int64 ) String() string { return fmt.Sprint( int64( i ) ) }

func ( i Int64 ) Compare( val interface{} ) int {
    switch v := val.( Int64 ); {
    case i < v: return -1
    case i > v: return 1
    }
    return 0
}

type Int32 int32
func ( i Int32 ) valImpl() {}
func ( i Int32 ) String() string { return fmt.Sprint( int32( i ) ) }

func ( i Int32 ) Compare( val interface{} ) int {
    return Int64( i ).Compare( Int64( val.( Int32 ) ) )
}

type Uint64 uint64
func ( i Uint64 ) valImpl() {}
func ( i Uint64 ) String() string { return fmt.Sprint( uint64( i ) ) }

func ( i Uint64 ) Compare( val interface{} ) int {
    switch v := val.( Uint64 ); {
    case i < v: return -1
    case i > v: return 1
    }
    return 0
}

type Uint32 uint32
func ( i Uint32 ) valImpl() {}
func ( i Uint32 ) String() string { return fmt.Sprint( uint32( i ) ) }

func ( i Uint32 ) Compare( val interface{} ) int {
    return Uint64( i ).Compare( Uint64( val.( Uint32 ) ) )
}

type Float64 float64
func ( d Float64 ) valImpl() {}
func ( d Float64 ) String() string { return fmt.Sprint( float64( d ) ) }

func ( d Float64 ) Compare ( val interface{} ) int {
    switch d2 := val.( Float64 ); {
    case d < d2: return -1
    case d == d2: return 0
    }
    return 1
}

type Float32 float32
func ( f Float32 ) valImpl() {}
func ( f Float32 ) String() string { return fmt.Sprint( float32( f ) ) }

func ( f Float32 ) Compare( val interface{} ) int {
    return Float64( f ).Compare( Float64( val.( Float32 ) ) )
}

type Buffer []byte
func ( b Buffer ) valImpl() {}

type Null struct {}
func ( n Null ) valImpl() {}
var NullVal *Null
func init() { NullVal = &Null{} }

func IsNull( val Value ) bool {
    _, isNull := val.( *Null )
    return isNull
}

type Timestamp time.Time

func ( t Timestamp ) valImpl() {}

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

var constEmptyList = &List{ []Value{} }
func EmptyList() *List { return constEmptyList }

func ( l *List ) valImpl() {}

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

func EmptySymbolMap() *SymbolMap { return MustSymbolMap() }

func ( m SymbolMap ) valImpl() {}

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
        log.Printf( "bad pairs: %v", pairs )
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
        err = NewValueCastError( acc.descend( id ), "value is null" )
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
    Type *QualifiedTypeName
    Value *Identifier
}

func ( e *Enum ) valImpl() {}

func MustEnum( typ, val string ) *Enum {
    return &Enum{ MustQualifiedTypeName( typ ), MustIdentifier( val ) }
}

type Struct struct {
    Type *QualifiedTypeName
    Fields *SymbolMap
}

func ( s *Struct ) valImpl() {}

func CreateStruct(
    typ interface{}, pairs ...interface{} ) ( *Struct, error ) {
    res := new( Struct )
    switch v := typ.( type ) {
    case *QualifiedTypeName: res.Type = v
    case string: 
        if qn, err := ParseQualifiedTypeName( v ); err == nil {
            res.Type = qn
        } else { return nil, err }
    default: return nil, libErrorf( "Not a qname: %s", typ )
    }
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
    QnameSymbolMap *QualifiedTypeName
    TypeSymbolMap *AtomicTypeReference
    QnameNull *QualifiedTypeName
    TypeNull *AtomicTypeReference
    QnameServiceRequest *QualifiedTypeName
    TypeServiceRequest *AtomicTypeReference
    QnameServiceResponse *QualifiedTypeName
    TypeServiceResponse *AtomicTypeReference
    TypeOpaqueList *ListTypeReference
    TypeNullableValue *NullableTypeReference
    IdNamespace *Identifier
    IdService *Identifier
    IdOperation *Identifier
    IdParameters *Identifier
    IdAuthentication *Identifier
    svcReqFieldOrder FieldOrder // initialized in same scope as Ids above
    IdResult *Identifier
    IdError *Identifier
    QnameTypeReference *QualifiedTypeName
    TypeTypeReference *AtomicTypeReference
    IdBuffer *Identifier
    QnameIdentifierPath *QualifiedTypeName
    TypeIdentifierPath *AtomicTypeReference
)

var coreQnameResolver map[ string ]*QualifiedTypeName
var PrimitiveTypes []*AtomicTypeReference
var NumericTypes []*AtomicTypeReference

var CoreNsV1 *Namespace

func init() {
    id := func( strs... string ) *Identifier { return &Identifier{ strs } }
    CoreNsV1 = &Namespace{
        Parts: []*Identifier{ id( "mingle" ), id( "core" ) },
        Version: id( "v1" ),
    }
    makeQn := func( s string ) *QualifiedTypeName {
        return &QualifiedTypeName{ CoreNsV1, &DeclaredTypeName{ s } }
    }
    coreQnameResolver = make( map[ string ]*QualifiedTypeName )
    f1 := func( s string ) ( *QualifiedTypeName, *AtomicTypeReference ) {
        qn := makeQn( s )
        coreQnameResolver[ qn.Name.ExternalForm() ] = qn
        return qn, &AtomicTypeReference{ Name: qn }
    }
    QnameValue, TypeValue = f1( "Value" )
    QnameBoolean, TypeBoolean = f1( "Boolean" )
    QnameBuffer, TypeBuffer = f1( "Buffer" )
    QnameString, TypeString = f1( "String" )
    QnameInt32, TypeInt32 = f1( "Int32" )
    QnameInt64, TypeInt64 = f1( "Int64" )
    QnameUint32, TypeUint32 = f1( "Uint32" )
    QnameUint64, TypeUint64 = f1( "Uint64" )
    QnameFloat32, TypeFloat32 = f1( "Float32" )
    QnameFloat64, TypeFloat64 = f1( "Float64" )
    QnameTimestamp, TypeTimestamp = f1( "Timestamp" )
    QnameSymbolMap, TypeSymbolMap = f1( "SymbolMap" )
    QnameNull, TypeNull = f1( "Null" )
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
        TypeSymbolMap,
    }
    TypeOpaqueList = &ListTypeReference{ TypeValue, true }
    TypeNullableValue = &NullableTypeReference{ TypeValue }
    NumericTypes = []*AtomicTypeReference{
        TypeInt32,
        TypeInt64,
        TypeUint32,
        TypeUint64,
        TypeFloat32,
        TypeFloat64,
    }
    QnameServiceRequest, TypeServiceRequest = f1( "Request" )
    QnameServiceResponse, TypeServiceResponse = f1( "Response" )
    IdNamespace = id( "namespace" )
    IdService = id( "service" )
    IdOperation = id( "operation" )
    IdParameters = id( "parameters" )
    IdAuthentication = id( "authentication" )
    svcReqFieldOrder = FieldOrder(
        []FieldOrderSpecification{
            { IdNamespace, true },
            { IdService, true },
            { IdOperation, true },
            { IdAuthentication, false },
            { IdParameters, false },
        },
    )
    IdResult = id( "result" )
    IdError =id( "error" )
    QnameTypeReference, TypeTypeReference = f1( "TypeReference" )
    IdBuffer = id( "buffer" )
    QnameIdentifierPath, TypeIdentifierPath = f1( "IdentifierPath" )
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
    case *Enum: return v.Type.AsAtomicType()
    case *SymbolMap: return TypeSymbolMap
    case *Struct: return v.Type.AsAtomicType()
    case *List: return TypeOpaqueList
    case *Null: return TypeNull
    }
    panic( fmt.Errorf( "Unhandled arg to typeOf (%T): %v", mgVal, mgVal ) )
}

type ValueError interface {
    Location() objpath.PathNode
    Message() string 
    error
}

type ValueErrorImpl struct {
    Path idPath
}

func ( e ValueErrorImpl ) Location() objpath.PathNode { 
    if e.Path == nil { return nil }
    return e.Path.( objpath.PathNode )
}

func ( e ValueErrorImpl ) MakeError( msg string ) string {
    if e.Path == nil { return msg }
    return fmt.Sprintf( "%s: %s", FormatIdPath( e.Path ), msg )
}

type MissingFieldsError struct {
    impl ValueErrorImpl
    flds []*Identifier // stored sorted
}

func NewMissingFieldsError( 
    path objpath.PathNode, flds []*Identifier ) *MissingFieldsError {
    flds2 := make( []*Identifier, len( flds ) )
    for i, e := 0, len( flds ); i < e; i++ { flds2[ i ] = flds[ i ] }
    SortIds( flds2 )
    return &MissingFieldsError{ impl: ValueErrorImpl{ path }, flds: flds2 }
}

func ( e *MissingFieldsError ) Message() string {
    strs := make( []string, len( e.flds ) )
    for i, fld := range e.flds { strs[ i ] = fld.ExternalForm() }
    fldsStr := strings.Join( strs, ", " )
    return fmt.Sprintf( "missing field(s): %s", fldsStr )
}

func ( e *MissingFieldsError ) Error() string {
    return e.impl.MakeError( e.Message() )
}

func ( e *MissingFieldsError ) Location() objpath.PathNode { 
    return e.impl.Location() 
}

func ( e *MissingFieldsError ) Fields() []*Identifier { return e.flds }

type UnrecognizedFieldError struct {
    impl ValueErrorImpl
    fld *Identifier
}

func NewUnrecognizedFieldError( 
    p objpath.PathNode, fld *Identifier ) *UnrecognizedFieldError {
    return &UnrecognizedFieldError{ impl: ValueErrorImpl{ p }, fld: fld }
}

func ( e *UnrecognizedFieldError ) Message() string {
    return fmt.Sprintf( "unrecognized field: %s", e.fld )
}

func ( e *UnrecognizedFieldError ) Error() string {
    return e.impl.MakeError( e.Message() )
}

func ( e *UnrecognizedFieldError ) Location() objpath.PathNode {
    return e.impl.Location()
}

type extFormer interface { ExternalForm() string }

type EndpointError struct {
    impl ValueErrorImpl
    desc string
    ef extFormer
}

func ( ee *EndpointError ) Error() string {
    msg := fmt.Sprintf( "no such %s: %s", ee.desc, ee.ef.ExternalForm() )
    return ee.impl.MakeError( msg )
}

func ( ee *EndpointError ) Location() objpath.PathNode {
    return ee.impl.Location()
}

func newEndpointError( desc string, ef extFormer, p idPath ) *EndpointError {
    return &EndpointError{ desc: desc, ef: ef, impl: ValueErrorImpl{ p } }
}

func NewEndpointErrorNamespace( 
    ns *Namespace, p objpath.PathNode ) *EndpointError {
    return newEndpointError( "namespace", ns, p )
}

func NewEndpointErrorService(
    svc *Identifier, p objpath.PathNode ) *EndpointError {
    return newEndpointError( "service", svc, p )
}

func NewEndpointErrorOperation(
    op *Identifier, p objpath.PathNode ) *EndpointError {
    return newEndpointError( "operation", op, p )
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

func ( m *mapImpl ) implGetOk( k mapImplKey ) ( interface{}, bool ) {
    res, ok := m.m[ k.ExternalForm() ]
    if ok { return res.val, ok }
    return nil, false
}

func ( m *mapImpl ) implGet( k mapImplKey ) interface{} {
    if val, ok := m.implGetOk( k ); ok { return val }
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

func ( m *IdentifierMap ) GetOk( id *Identifier ) ( interface{}, bool ) {
    return m.implGetOk( id )
}

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

func ( m *QnameMap ) GetOk( qn *QualifiedTypeName ) ( interface{}, bool ) {
    return m.implGetOk( qn )
}

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

func ( m *NamespaceMap ) GetOk( ns *Namespace ) ( interface{}, bool ) {
    return m.implGetOk( ns )
}

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

//type svcIdMapKey struct {
//    ns *Namespace
//    svc *Identifier
//}
//
//func ( k svcIdMapKey ) ExternalForm() string {
//    return k.ns.ExternalForm() + "/" + k.svc.ExternalForm()
//}
//
//type ServiceIdMap struct {
//    *mapImpl
//}
//
//func NewServiceIdMap() *ServiceIdMap { return &ServiceIdMap{ newMapImpl() } }
//
//func ( m *ServiceIdMap ) Put( 
//    ns *Namespace, svc *Identifier, val interface{} ) {
//    m.implPut( svcIdMapKey{ ns, svc }, val )
//}
//
//func ( m *ServiceIdMap ) GetOk( 
//    ns *Namespace, svc *Identifier ) ( interface{}, bool ) {
//    return m.implGetOk( svcIdMapKey{ ns, svc } )
//}
