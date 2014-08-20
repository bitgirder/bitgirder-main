package s3

import (
    "bitgirder/uuid"
    "bitgirder/filerecv"
    "bitgirder/iotests"
    "bitgirder/hashes"
    "bitgirder/assert"
    mg "mingle"
    "mingle/codec/json"
    "mingle/codec"
    "path/filepath"
    "testing"
    "errors"
    "crypto/rand"
    "math"
    "time"
//    "log"
    "fmt"
    "strings"
    "os"
    "io"
    "io/ioutil"
    "bytes"
    "bufio"
)

const envTestRuntime = "BITGIRDER_TEST_RUNTIME"
const credsFileName = "aws-creds.json"
const objPathPrefix = "test-object"

var testCreds *Credentials
var testBucket Bucket

const ctypeDefaultExpct = "binary/octet-stream"

// This should be used sparingly and only by tests which are making hard
// assumptions about the behavior of various code at the root of a bucket.
//
// The assumption is that this key name is the lexicographically first one in
// testBucket
const cFirstFileInBucket = "$first-file-in-bucket.txt"

var markerErr error
func init() { markerErr = errors.New( "marker-error" ) }

func readCredentials( f *os.File, creds *Credentials ) {
    r := bufio.NewReader( f )
    cdc := json.NewJsonCodec()
    if ms, err := codec.Decode( cdc, r ); err == nil {
        acc := mg.NewSymbolMapAccessor( ms.Fields, nil )
        creds.AccessKey = AccessKey( acc.MustGoStringByString( "access-key" ) )
        creds.SecretKey = SecretKey( acc.MustGoStringByString( "secret-key" ) )
    } else { panic( err ) }
}

func readAwsTestProps( creds *Credentials ) {
    if rtEnv := os.Getenv( envTestRuntime ); rtEnv != "" {
        propsFile := rtEnv + "/" + credsFileName
        if f, err := os.Open( propsFile ); err == nil {
            defer f.Close()
            readCredentials( f, creds )
            return
        } else { panic( err ) }
    }
    msgFmt := "No runtime env is set in env var %s"
    panic( fmt.Sprintf( msgFmt, envTestRuntime ) )
}

func init() {
    testBucket = Bucket( "s3test.bitgirder.com" )
    testCreds = new( Credentials )
    readAwsTestProps( testCreds )
}

func nextPathString() string {
    return fmt.Sprintf( "%s/%016x/%s", 
        objPathPrefix, time.Now().Unix(), uuid.MustType4() )
}

func nextObjectKey() ObjectKey {
    return ExpectObjectKey( nextPathString() )
}

func nextObjectRoute() *ObjectRequestRoute {
    return &ObjectRequestRoute{ Bucket: testBucket, Key: nextObjectKey() }
}

func nextBodyBuffer( sz int64 ) ( []byte, []byte ) {
    body := make( []byte, sz )
    rand.Read( body )
    return body, hashes.Md5OfBytes( body )
}

func nextBodyReader( sz int64 ) ( io.Reader, []byte ) {
    body, md5Res := nextBodyBuffer( sz )
    return bytes.NewReader( body ), md5Res
}

func nextClient() *Client { return NewClient( testCreds ) }

func nextLocalDir( t *testing.T ) string {
    locDir, err := ioutil.TempDir( "", "s3-test-dir-" )
    if err != nil { t.Fatal( err ) }
    return locDir
}

func expectMd5( in io.Reader, expct []byte ) {
    act := hashes.MustMd5OfReader( in )
    assert.Equal( expct, act )
}

func assertRespContext( ctx *ResponseContext ) {
    assert.NotEqual( "", ctx.RequestId )
    assert.NotEqual( "", ctx.RequestId2 )
}

type objectRoundtripTest struct {
    
    // included directly so methods can be called on this instance
    *testing.T

    // set before first PUT and unchanged after
    useSsl bool
    bodyBytes []byte
    bodyMd5 []byte
    route *ObjectRequestRoute
    sendMd5 bool
    useMetaHeaders bool
    cli *Client

    ContentType // optional

    // set after first PUT
    Etag
}

func ( rt *objectRoundtripTest ) init( t *testing.T ) {
    rt.T = t
    rt.cli = NewClient( testCreds )
    if rt.bodyBytes == nil { rt.bodyBytes, rt.bodyMd5 = nextBodyBuffer( 1000 ) }
    testKey := nextObjectKey()
    rt.route = &ObjectRequestRoute{ 
        Endpoint: &Endpoint{ UseSsl: rt.useSsl },
        Bucket: testBucket, 
        Key: testKey,
    }
}

func addMetaHeaders( putReq *PutObjectRequest ) {
    putReq.AddMetaHeader( "meta1", "val1-a" )
    putReq.AddMetaHeader( "meta1", "val1-b" )
    putReq.AddMetaHeader( "x-amz-meta-meta2", "val2\n" )
    putReq.AddMetaHeader( "meta3", "val3-a,val3-b" )
    putReq.AddMetaHeader( "meta4", "val4-a   \t\r\n  val4-b\n val4-c\n" )
}

func ( rt *objectRoundtripTest ) putObject() {
    putReq := &PutObjectRequest{ Route: rt.route, ContentType: rt.ContentType }
    putReq.SetBodyBytes( rt.bodyBytes )
    if rt.sendMd5 { putReq.Md5 = rt.bodyMd5 }
    if rt.useMetaHeaders { addMetaHeaders( putReq ) }
    if resp, err := rt.cli.PutObject( putReq ); err == nil {
        assertRespContext( resp.ResponseContext )
        if rt.Etag = resp.Etag; rt.Etag == "" { rt.Fatalf( "res has no Etag" ) }
    } else { rt.Fatal( err ) }
}

func ( rt *objectRoundtripTest ) assertMetaHeaders( gr *GetObjectResponse ) {
    assert.Equal( "val1-a,val1-b", gr.GetMetaHeader( "meta1" ) )
    assert.Equal( "val2", gr.GetMetaHeader( "MeTA2" ) )
    assert.Equal( "val3-a,val3-b", gr.GetMetaHeader( "meta3" ) )
    assert.Equal( "val4-a,val4-b,val4-c", gr.GetMetaHeader( "meta4" ) )
}

func ( rt *objectRoundtripTest ) assertGetResp( gr *GetObjectResponse ) {
    ctypeExpct := rt.ContentType
    if ctypeExpct == "" { ctypeExpct = ctypeDefaultExpct }
    assert.Equal( ctypeExpct, gr.ContentType )
    assert.Equal( rt.Etag, gr.Etag )
    assertRespContext( gr.ResponseContext )
    if rt.useMetaHeaders { rt.assertMetaHeaders( gr ) }
}

func ( rt *objectRoundtripTest ) createGetRequest() *GetObjectRequest {
    return &GetObjectRequest{ Route: rt.route }
}

func ( rt *objectRoundtripTest ) getObject() {
    getReq := rt.createGetRequest()
    h := func( getResp *GetObjectResponse, body io.Reader ) error {
        rt.assertGetResp( getResp )
        expectMd5( body, rt.bodyMd5 )
        return nil
    }
    if getResp, err := rt.cli.GetObject( getReq, h ); err == nil {
        rt.assertGetResp( getResp )
    } else { rt.Fatal( err ) }
}

func ( rt *objectRoundtripTest ) headObject() {
    getReq := rt.createGetRequest()
    if getResp, err := rt.cli.HeadObject( getReq ); err == nil {
        rt.assertGetResp( getResp )
    } else { rt.Fatal( err ) }
}

func ( rt *objectRoundtripTest ) deleteObject() {
    delReq := &DeleteObjectRequest{ Route: rt.route }
    if delResp, err := rt.cli.DeleteObject( delReq ); err == nil {
        assertRespContext( delResp.ResponseContext )
    } else { rt.Fatal( err ) }
}

func ( rt *objectRoundtripTest ) run() {
    rt.putObject()
    rt.getObject()
    rt.headObject()
    rt.deleteObject()
}

func getObjectRoundtripTests() []*objectRoundtripTest {
    res := make( []*objectRoundtripTest, 0, 32 )
    for _, useSsl := range []bool { true, false } {
    for _, ctype := range []string { "", "application/testing" } {
    for _, sendMd5 := range []bool { true, false } {
    for _, useMetaHeaders := range []bool { true, false } {
        rt := &objectRoundtripTest{}
        rt.useSsl = useSsl
        rt.ContentType = ContentType( ctype )
        rt.sendMd5 = sendMd5
        rt.useMetaHeaders = useMetaHeaders
        res = append( res, rt )
    }}}}
    return res
}

func TestObjectRoundtrips( t *testing.T ) {
    tests := getObjectRoundtripTests()
    for _, rt := range tests {
        rt.init( t )
        rt.run()
    }
}

func TestPutWithoutBodyFailure( t *testing.T ) {
    req := &PutObjectRequest{ Route: nextObjectRoute(), }
    if _, err := NewClient( testCreds ).PutObject( req ); err == nil {
        t.Fatalf( "Expected err" )
    } else if _, ok := err.( *NoBodyError ); ! ok { t.Fatal( err ) }
}

// Test is mainly that S3RemoteError is the error and that its xml fields were
// set. We don't exhaustively re-check all of the specific fields, leaving that
// to TestRemoteExceptionSetFieldsFromXml(), and only check that HTTP-specific
// fields are set, that xml fields were set at all (check of Code) and that the
// <Message> from the error doc takes precedence from the HTTP Status message
// from the server
func TestRemoteS3ErrorHandled( t *testing.T ) {
    body, _ := nextBodyBuffer( 1000 )
    req := &PutObjectRequest{ Route: nextObjectRoute(), }
    req.SetBodyBytes( body )
    cli := NewClient( &Credentials{ "test", "test" } )
    if _, err := cli.PutObject( req ); err == nil {
        t.Fatalf( "Expected failure" )
    } else { 
        if s3Err, ok := err.( *S3RemoteError ); ok {
            assertRespContext( s3Err.ResponseContext )
            assert.Equal( "InvalidAccessKeyId", s3Err.Code )
            assert.True( strings.HasPrefix( s3Err.Message, "The AWS Access " ) )
            assert.Equal( 403, s3Err.HttpStatusCode )
        } else { t.Fatal( err ) }
    }
}

// the xml literal below is structurally taken from an actual message, although
// the content of elements has been shortened and simplified for testing
func TestRemoteExceptionSetFieldsFromXml( t *testing.T ) {
    xml := `<?xml version="1.0" encoding="UTF-8"?> <Error><Code>TEST_CODE</Code><Message>TEST_MESSAGE</Message><RequestId>TEST_REQ_ID</RequestId><HostId>TEST_HOST_ID</HostId><AWSAccessKey>TEST_ACC_ID</AWSAccessKey></Error>`
    err := &S3RemoteError{ HttpStatusCode: 403 }
    err.setFieldsFromXml( bytes.NewBuffer( []byte( xml ) ) )
    assert.Equal( "TEST_CODE", err.Code )
    assert.Equal( "TEST_MESSAGE", err.Message )
    assert.Equal( "TEST_HOST_ID", err.HostId )
    assert.Equal( "TEST_ACC_ID", err.AWSAccessKey )
    assert.Equal( "(HTTP 403) TEST_CODE: TEST_MESSAGE", err.Error() )
}

func assertListCallCount( resSz, pgSz, calls int ) {
    if pgSz > 0 {
        expct := math.Ceil( float64( resSz ) / float64( pgSz ) )
        assert.Equal( int( expct ), calls )
    } else { assert.Equal( 1, calls ) }
}

type prefixedListingTest struct {

    // set before run()
    *testing.T
    objCount int
    pgSize int
    cli *Client
    etags []Etag
    pathPrefix string
    baseSize int64
}

func ( lpt *prefixedListingTest ) init( t *testing.T ) { 
    lpt.T = t 
    lpt.baseSize = 1000
    lpt.cli = NewClient( testCreds )
    lpt.etags = make( []Etag, 0, lpt.objCount )
    lpt.pathPrefix = nextPathString()
}

func ( t *prefixedListingTest ) keyPath( i int ) string {
    return fmt.Sprintf( "%s/%02x", t.pathPrefix, i )
}

func ( t *prefixedListingTest ) storeTestObjects() {
    for i := 0; i < t.objCount; i++ { 
        body, dig := nextBodyBuffer( t.baseSize + int64( i ) )
        putReq := &PutObjectRequest{
            Route: &ObjectRequestRoute{
                Bucket: testBucket,
                Key: ExpectObjectKey( t.keyPath( i ) ),
            },
            Md5: dig,
        }
        putReq.SetBodyBytes( body )
        if resp, err := t.cli.PutObject( putReq ); err == nil {
            t.etags = append( t.etags, resp.Etag )
        } else { t.Fatal( err ) }
    }
}

func ( t *prefixedListingTest ) assertListedObject( obj *ListedObject, i int ) {
    keyExpct := t.keyPath( i )
    assert.Equal( keyExpct, obj.Key )
    assert.Equal( t.etags[ i ], obj.Etag )
    assert.NotEqual( "", obj.LastModified )
    assert.NotEqual( "", obj.Owner.Id )
    assert.NotEqual( "", obj.Owner.DisplayName )
    assert.Equal( t.baseSize + int64( i ), obj.Size )
}

func ( t *prefixedListingTest ) listNext( 
    lsCtx *ListContext, i int ) ( *ListContext, int ) {
    if lsRes, err := t.cli.ListNext( lsCtx ); err == nil {
        assertRespContext( lsRes.ResponseContext )
        assert.NotEqual( 0, len( lsRes.Contents ) )
        for _, obj := range lsRes.Contents {
            t.assertListedObject( &obj, i )
            i++
        }
        assert.Equal( i < t.objCount, lsRes.IsTruncated )
        assert.Equal( testBucket, lsRes.Bucket )
        assert.Equal( lsCtx.Prefix, lsRes.Prefix )
        if lsCtx.MaxKeys > 0 { assert.Equal( lsCtx.MaxKeys, lsRes.MaxKeys ) }
        assert.Equal( lsCtx.Delimiter, lsRes.Delimiter )
        lsCtx = lsRes.NextListing()
    } else { t.Fatal( err ) }
    return lsCtx, i
}

func ( t *prefixedListingTest ) assertExpectedCalls( calls int ) {
    assertListCallCount( t.objCount, t.pgSize, calls )
}

func ( t *prefixedListingTest ) assertEmptyLastListing() {
    lsCtx := &ListContext{
        Bucket: testBucket,
        Prefix: t.pathPrefix,
        Marker: t.keyPath( t.objCount - 1 ),
    }
    if lsRes, err := t.cli.ListNext( lsCtx ); err == nil {
        assert.Equal( 0, len( lsRes.Contents ) )
        assert.False( lsRes.IsTruncated )
    } else { t.Fatal( err ) }
}

func ( t *prefixedListingTest ) run() {
    t.storeTestObjects()
    lsCtx := &ListContext{ 
        Bucket: testBucket,
        Prefix: t.pathPrefix,
    }
    if t.pgSize > 0 { lsCtx.MaxKeys = t.pgSize }
    calls := 0
    for i := 0; i < t.objCount; {
        lsCtx, i = t.listNext( lsCtx, i )
        calls++
    }
    t.assertExpectedCalls( calls )
    t.assertEmptyLastListing()
}

func getPrefixedListPagingTests() []*prefixedListingTest {
    return []*prefixedListingTest {

        // Try varying page sizes, include trivial (1), amz default (-1), even
        // divisor (3), non-even divisor (5), and just larger than result set
        // (7)
        &prefixedListingTest{ objCount: 6, pgSize: 1 },
        &prefixedListingTest{ objCount: 6, pgSize: 6 },
        &prefixedListingTest{ objCount: 6, pgSize: 3 },
        &prefixedListingTest{ objCount: 6, pgSize: 5 },
        &prefixedListingTest{ objCount: 6, pgSize: 7 },
        &prefixedListingTest{ objCount: 6, pgSize: -1 },
    }
}

func TestPrefixedListPaging( t *testing.T ) {
    for _, lpt := range getPrefixedListPagingTests() {
        lpt.init( t )
        lpt.run()
    }
}

type delimitedListingTest struct {
    *testing.T
    cli *Client
    prefix string
}

func ( t *delimitedListingTest ) init() {
    testPaths := []string { "a/a", "a/b", "b", "c", "d/a", "d/b", "e", "f" }
    t.cli = NewClient( testCreds )
    t.prefix = nextPathString()
    for _, path := range testPaths {
        k := ExpectObjectKey( t.prefix + "/" + path )
        req := &PutObjectRequest{ 
            Route: &ObjectRequestRoute{ Bucket: testBucket, Key: k },
        }
        req.SetBodyBytes( make( []byte, 10 ) )
        if _, err := t.cli.PutObject( req ); err != nil { t.Fatal( err ) }
    }
}

type delimitedListingRun struct {
    keys []string
    dirs []string
    calls int
    pgSz int
}

func ( t *delimitedListingTest ) listNextDelimited( 
    lsCtx *ListContext, r *delimitedListingRun ) *ListContext {
    if lsRes, err := t.cli.ListNext( lsCtx ); err == nil {
        r.calls++
        assert.Equal( "/", lsRes.Delimiter )
        r.dirs = append( r.dirs, lsRes.CommonPrefixes... )
        for _, obj := range lsRes.Contents { 
            r.keys = append( r.keys, obj.Key ) 
        }
        if lsRes.IsTruncated { 
            lsCtx = lsRes.NextListing() 
        } else { lsCtx = nil }
    } else { t.Fatal( err ) }
    return lsCtx
}

func ( t *delimitedListingTest ) assertDelimitedListResults( 
    r *delimitedListingRun ) {
    p := t.prefix // to shorten expressions below
    assert.Equal( []string { p + "/b", p + "/c", p + "/e", p + "/f" }, r.keys )
    assert.Equal( []string { p + "/a/", p + "/d/" }, r.dirs )
    assertListCallCount( 6, r.pgSz, r.calls )
}

func ( t *delimitedListingTest ) run( pgSz int ) {
    r := &delimitedListingRun{ pgSz: pgSz, keys: []string{}, dirs: []string{} }
    lsCtx := &ListContext{
        Bucket: testBucket,
        Prefix: t.prefix + "/",
        Delimiter: "/",
    }
    if pgSz > 0 { lsCtx.MaxKeys = pgSz }
    for lsCtx != nil { lsCtx = t.listNextDelimited( lsCtx, r ) }
    t.assertDelimitedListResults( r )
}

func TestDelimitedListPaging( t *testing.T ) {
    lpt := &delimitedListingTest{ T: t }
    lpt.init()
    lpt.run( -1 )
    for pgSz := 1; pgSz <= 7; pgSz ++ { lpt.run( pgSz ) }
}

func TestListAtBucketRoot( t *testing.T ) {
    lsCtx := &ListContext{ Bucket: testBucket, MaxKeys: 1 }
    if lsRes, err := NewClient( testCreds ).ListNext( lsCtx ); err == nil {
        assert.Equal( cFirstFileInBucket, lsRes.Contents[ 0 ].Key )
    } else { t.Fatal( err ) }
}

type remoteDirManager struct {
    *testing.T
    cli *Client
    objSize int64
    prefix string
    digs map[ string ][]byte
}

func ( rd *remoteDirManager ) objNameAt( i int ) string {
    return fmt.Sprintf( "file%d", i )
}

func ( rd *remoteDirManager ) keyStringAt( i int ) string {
    return fmt.Sprintf( "%s/%s", rd.prefix, rd.objNameAt( i ) )
}

func ( rd *remoteDirManager ) putObject( rel string, sz int64 ) {
    key := ExpectObjectKey( rd.prefix + "/" + rel )
    put := &PutObjectRequest{ 
        Route: &ObjectRequestRoute{ Bucket: testBucket, Key: key },
    }
    buf, dig := nextBodyBuffer( sz )
    put.SetBodyBytes( buf )
    rd.digs[ rel ] = dig
    if _, err := rd.cli.PutObject( put ); err != nil { rd.Fatal( err ) }
}

func ( rd *remoteDirManager ) putObjects( objs int ) {
    for i := 0; i < objs; i++ { rd.putObject( rd.objNameAt( i ), rd.objSize ) }
}

func ( rd *remoteDirManager ) init( objs int ) {
    rd.prefix = nextPathString()
    rd.cli = nextClient()
    if rd.objSize == 0 { rd.objSize = int64( 1000 ) }
    rd.digs = make( map[ string ][]byte )
    rd.putObjects( objs )
}

func initRemoteDir( objs int, t *testing.T ) *remoteDirManager {
    res := &remoteDirManager{ T: t }
    res.init( objs )
    return res
}

func ( rd *remoteDirManager ) listContext() *ListContext {
    return &ListContext{ Bucket: testBucket, Prefix: rd.prefix + "/" }
}

func ( rd *remoteDirManager ) assertAccumulatedKeys( acc []string ) {
    assert.Equal( len( rd.digs ), len( acc ) )
    for i, key := range acc { assert.Equal( rd.keyStringAt( i ), key ) }
}

// Ultimately does with the list nothing other than what WalkObjects() would do,
// but we just want to get basic coverage at the ScanBucket() level 
func assertScanBucketBasic( pgSz int, rd *remoteDirManager ) {
    lsCtx := rd.listContext()
    if pgSz > 0 { lsCtx.MaxKeys = pgSz }
    acc := make( []string, 0, 10 )
    calls := 0
    f := func( lsRes *ListResult ) error {
        for _, obj := range lsRes.Contents { acc = append( acc, obj.Key ) }
        calls++
        return nil
    }
    if err := rd.cli.ScanBucket( lsCtx, f ); err != nil { rd.Fatal( err ) }
    rd.assertAccumulatedKeys( acc )
    assertListCallCount( len( acc ), pgSz, calls )
}

func TestScanBucketBasic( t *testing.T ) {
    rd := initRemoteDir( 2, t )
    assertScanBucketBasic( -1, rd )
    assertScanBucketBasic( 1, rd )
    assertScanBucketBasic( 2, rd )
    assertScanBucketBasic( 10, rd )
}

type oneKeyExpector struct {
    key string
}

func ( e *oneKeyExpector ) accumulate( 
    key string, rd *remoteDirManager ) error {
    if e.key == "" { 
        e.key = key 
    } else { rd.Fatalf( "Already saw a key (%s)", key ) }
    return EndScan
} 

func ( e *oneKeyExpector ) assertFinal( rd *remoteDirManager ) {
    if e.key == "" { rd.Fatalf( "Did not see any keys" ) }
    assert.Equal( rd.keyStringAt( 0 ), e.key )
}

func TestScanBucketScanFuncStop( t *testing.T ) {
    rd := initRemoteDir( 2, t )
    e := &oneKeyExpector{}
    lsCtx := rd.listContext()
    lsCtx.MaxKeys = 1
    f := func( lsRes *ListResult ) error {
        assert.False( len( lsRes.Contents ) == 0 )
        return e.accumulate( lsRes.Contents[ 0 ].Key, rd )
    }
    if err := rd.cli.ScanBucket( lsCtx, f ); err != nil { rd.Fatal( err ) }
    e.assertFinal( rd )
}

func TestScanBucketErrorBubbleUp( t *testing.T ) {
    rd := initRemoteDir( 1, t )
    f := func( lsRes *ListResult ) error { return markerErr }
    assert.Equal( markerErr, rd.cli.ScanBucket( rd.listContext(), f ) )
}

func TestWalkObjectsBasic( t *testing.T ) {
    rd := initRemoteDir( 2, t )
    for _, pgSz := range []int { -1, 1, 2, 10 } { 
        lsCtx := rd.listContext()
        if pgSz > 0 { lsCtx.MaxKeys = pgSz }
        acc := make( []string, 0, len( rd.digs ) )
        f := func( o ListedObject ) error {
            acc = append( acc, o.Key )
            return nil
        }
        if err := rd.cli.WalkObjects( lsCtx, f ); err != nil { rd.Fatal( err ) }
        rd.assertAccumulatedKeys( acc )
    }
}

func TestWalkObjectsFuncStop( t *testing.T ) {
    rd := initRemoteDir( 2, t )
    e := &oneKeyExpector{}
    f := func( o ListedObject ) error { return e.accumulate( o.Key, rd ) }
    if err := rd.cli.WalkObjects( rd.listContext(), f ); err != nil { 
        rd.Fatal( err )
    }
    e.assertFinal( rd )
}

func TestWalkObjectsEmptyResultSet( t *testing.T ) {
    rd := initRemoteDir( 0, t )
    f := func( o ListedObject ) error {
        rd.Fatalf( "Saw key: %s", o.Key )
        return nil
    }
    if err := rd.cli.WalkObjects( rd.listContext(), f ); err != nil {
        rd.Fatal( err )
    }
}

func TestWalkObjectsErrorBubbleUp( t *testing.T ) {
    rd := initRemoteDir( 1, t )
    f := func ( o ListedObject ) error { return markerErr }
    assert.Equal( markerErr, rd.cli.WalkObjects( rd.listContext(), f ) )
}

func TestAsLeadingSlashHandling( t *testing.T ) {
    pairs := []string {
        "a", "a",
        "/a", "a",
        "%2fa", "a",
        "%2Fa", "a",
        "//a", "/a",
        "%2f%2fa", "%2fa",
        "/%2fa", "%2fa",
        "%2f/a", "/a",
    }
    for i := 0; i < len( pairs ); i += 2 {
        k, v := pairs[ i ], pairs[ i + 1 ]
        assert.Equal( v, string( ExpectObjectKey( k ) ) )
        assert.Equal( v, TrimLeadingSlash( k ) )
    }
    for _, s := range []string { "", "/", "%2f", "%2F" } {
        if k, err := AsObjectKey( s ); err == nil {
            t.Fatalf( "Expected err, got key: %s", k )
        } else if _, ok := err.( *ObjectKeyEmptyError ); ! ok { t.Fatal( err ) }
    }
}

type fileRoundtripTest struct {
    *testing.T
    size int64
    srcFile, destFile *os.File
    dig []byte
    route *ObjectRequestRoute
    cli *Client
}

func ( rt *fileRoundtripTest ) writeTestFile() {
    var buf []byte
    buf, rt.dig = nextBodyBuffer( rt.size )
    var err error
    rt.srcFile, err = ioutil.TempFile( "", "s3-test-" )
    if err != nil { rt.Fatal( err ) }
    defer rt.srcFile.Close()
    if _, err := rt.srcFile.Write( buf ); err != nil { rt.Fatal( err ) }
}

func ( rt *fileRoundtripTest ) putFile() {
    rt.writeTestFile()
    putReq := &PutObjectRequest{ Route: nextObjectRoute() }
    putReq.SetBodyFile( rt.srcFile.Name() )
    rt.route = putReq.Route
    if _, err := rt.cli.PutObject( putReq ); err != nil { rt.Fatal( err ) }
}

func ( rt *fileRoundtripTest ) getFile() {
    getReq := &GetObjectRequest{ Route: rt.route }
    var err error
    if rt.destFile, err = ioutil.TempFile( "", "s3-test-" ); err != nil {
        rt.Fatal( err )
    }
    recv := &filerecv.FileReceive{ DestFile: rt.destFile.Name() }
    h := func( gr *GetObjectResponse, in io.Reader ) error {
        _, err := filerecv.ReceiveFile( recv, in )
        return err
    }
    if _, err := rt.cli.GetObject( getReq, h ); err != nil { rt.Fatal( err ) }
}

func ( rt *fileRoundtripTest ) assertFile() {
    act := hashes.MustMd5OfReader( rt.destFile )
    assert.Equal( rt.dig, act )
}

func TestGetObjectErrorBubbleUp( t *testing.T ) {
    put := &PutObjectRequest{ Route: nextObjectRoute() }
    body, _ := nextBodyBuffer( 1000 )
    put.SetBodyBytes( body )
    cli := nextClient()
    if _, err := cli.PutObject( put ); err != nil { t.Fatal( err ) }
    get := &GetObjectRequest{ Route: put.Route }
    h := func( gr *GetObjectResponse, r io.Reader ) error { return markerErr }
    if _, err := cli.GetObject( get, h ); err != markerErr {
        t.Fatalf( "Expected marker err but got (%T) %v", err, err )
    }
}

func TestFilePutAndGet( t *testing.T ) {
    rt := &fileRoundtripTest{ T: t, size: int64( 1000 ), cli: nextClient() }
    defer iotests.WipeEntries( rt.srcFile, rt.destFile )
    rt.putFile()
    rt.getFile()
    rt.assertFile()
}

func assertSyncedFile( path string, dig []byte, rd *remoteDirManager ) {
    if f, err := os.Open( path ); err == nil {
        defer f.Close()
        act := hashes.MustMd5OfReader( f )
        assert.Equal( dig, act )
    } else { rd.Fatal( err ) }
}

func assertSyncResults( ds *DirectorySync, rd *remoteDirManager ) {
    if err := rd.cli.SyncDirectory( ds ); err != nil { rd.Fatal( err ) }
    count := 0
    f := func( path string, info os.FileInfo, err error ) error {
        if ! info.IsDir() { 
            if rel, err := filepath.Rel( ds.LocalRoot, path ); err == nil {
                if dig, ok := rd.digs[ rel ]; ! ok {
                    rd.Fatalf( "Unexpected synced file: %s", path )
                } else { assertSyncedFile( path, dig, rd ) }
                count++
            } else { rd.Fatal( err ) } 
        }
        return nil
    }
    if err := filepath.Walk( ds.LocalRoot, f ); err != nil { rd.Fatal( err ) }
    assert.Equal( len( rd.digs ), count )
}

func createBasicSync( locDir string, rd *remoteDirManager ) *DirectorySync {
    return &DirectorySync{
        LocalRoot: locDir,
        RemoteRoot: rd.listContext(),
        SyncPredicate: RemoteFileLargerPredicate,
    }
}

// Here we assert basic coverage of various event dispatches, such as
// *EvFileSyncStarting, and we won't necessarily assert all of them elsewhere.
func assertBasicSync0( locDir string, rd *remoteDirManager ) {
    ds := createBasicSync( locDir, rd )
    started := 0
    synced := 0
    ds.EventHandler = func( ev interface{} ) {
        switch ev.( type ) {
        case *EvFileSynced: synced++
        case *EvFileSyncStarting: started++
        }
    }
    assertSyncResults( ds, rd )
    assert.Equal( 2, synced )
    assert.Equal( 2, started )
}

func prepareBasicSync1( rd *remoteDirManager ) {
    rd.putObject( "new-file", 1000 )
    rd.putObject( "new-dir/new-file", 1000 )
    rd.putObject( "file0", 1001 )
}

func assertBasicSync1( locDir string, rd *remoteDirManager ) {
    ds := createBasicSync( locDir, rd )
    synced := 0
    ds.EventHandler = func( ev interface{} ) {
        if sync, ok := ev.( *EvFileSynced ); ok {
            if string( sync.Key ) == rd.keyStringAt( 1 ) {
                rd.Fatalf( "Unexpected sync of %s", sync.Key )
            }
            synced++
        }
    }
    assertSyncResults( ds, rd )
    assert.Equal( 3, synced )
}

func TestSyncDirectoryBasic( t *testing.T ) {
    rd := initRemoteDir( 2, t )
    locDir := nextLocalDir( t )
    defer iotests.WipeEntries( locDir )
    assertBasicSync0( locDir, rd )
    prepareBasicSync1( rd )
    assertBasicSync1( locDir, rd )
}

func TestSyncDirectoryDryRun( t *testing.T ) {
    rd := initRemoteDir( 1, t )
    locDir := nextLocalDir( t )
    defer iotests.WipeEntries( locDir )
    dp := createBasicSync( locDir, rd )
    dp.DryRun = true
    synced := 0
    dp.EventHandler = func( ev interface{} ) {
        if _, ok := ev.( *EvFileSyncStarting ); ok { synced++ }
    }
    if err := rd.cli.SyncDirectory( dp ); err == nil {
        if matches, err := filepath.Glob( locDir + "/*" ); err == nil {
            assert.Equal( 0, len( matches ) )
        } else { rd.Fatal( err ) }
    } else { rd.Fatal( err ) }
    assert.Equal( 1, synced )
}

// Don't actually pull it, but assert that a sync rooted at the bucket correctly
// maps root-level files into the expected local location
func TestSyncDirectoryFromBucketRoot( t *testing.T ) {
    locDir := nextLocalDir( t )
    lsCtx := &ListContext{ Bucket: testBucket }
    rs := &DirectorySync{ LocalRoot: locDir, RemoteRoot: lsCtx }
    gotFile := false
    rs.SyncPredicate = 
        func( obj *ListedObject, locFile string ) ( bool, error ) {
            if gotFile { t.Fatalf( "Listing is continuing after first file" ) }
            gotFile = locFile == locDir + "/" + cFirstFileInBucket
            return false, EndScan
        }
    if err := nextClient().SyncDirectory( rs ); err == nil {
        assert.True( gotFile )
    } else { t.Fatal( err ) }
}

// Basic coverage test that sync invokes a file receive initializer when set
func TestSyncDirectoryFileRecvInitializerCalled( t *testing.T ) {
    rd := initRemoteDir( 2, t )
    locDir := nextLocalDir( t )
    ds := createBasicSync( locDir, rd )
    calls := 0
    f := func( recv *filerecv.FileReceive, obj *ListedObject ) error {
        if f, err := ioutil.TempFile( "", "s3-test-" ); err == nil {
            defer f.Close()
            recv.TempFile = f.Name()
            calls++
        } else { return err }
        return nil
    }
    ds.ReceiveInitializer = f
    assertSyncResults( ds, rd )
    assert.Equal( 2, calls )
}

func TestSyncDirectoryFileRecvInitializerErrorBubbleUp( t *testing.T ) {
    rd := initRemoteDir( 1, t )
    locDir := nextLocalDir( t )
    ds := createBasicSync( locDir, rd )
    ds.ReceiveInitializer =
        func( recv *filerecv.FileReceive, obj *ListedObject ) error {
            return markerErr
        }
    if err := rd.cli.SyncDirectory( ds ); err != markerErr {
        t.Fatalf( "Expected marker err, got: %v", err )
    }
}
