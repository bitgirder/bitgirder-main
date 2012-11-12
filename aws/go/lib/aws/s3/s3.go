package s3

import (
    "bitgirder/hashes"
    "bitgirder/filerecv"
    "bytes"
    "crypto/hmac"
    "crypto/sha1"
    "encoding/base64"
    "encoding/xml"
    "errors"
    "path/filepath"
    "fmt"
    "hash"
    "io"
    "log"
    "net/http"
    "net/url"
    "os"
    "sort"
    "strconv"
    "strings"
    "time"
)

const hdrAuth = "Authorization"
const hdrDate = "Date"
const hdrEtag = "Etag"
const hdrReqId = "X-Amz-Request-Id"
const hdrReqId2 = "X-Amz-Id-2"
const hdrMd5 = "Content-Md5"
const hdrCtype = "Content-Type"

const metaHdrPrefix = "x-amz-meta-"
const amzHdrPrefix = "x-amz-"

const ctypeS3Err = "application/xml"
const ctypeListRes = "application/xml"

const keyPrefix = "prefix"
const keyMarker = "marker"
const keyDelim = "delimiter"
const keyMaxKeys = "max-keys"

var AwsSslEndpoint *Endpoint

func init() {
    AwsSslEndpoint = &Endpoint{ UseSsl: true }
}

func makeHeader() http.Header {
    return http.Header( make( map[ string ][]string ) )
}

type AccessKeyId string
type SecretKey string

type Credentials struct {
    AccessKeyId
    SecretKey
}

type Etag string

type EtagMissingError struct {}

func ( e *EtagMissingError ) Error() string { return "Missing header 'Etag'" }

type ContentType string

type ObjectKey string

type ObjectKeyEmptyError struct {}

func ( e *ObjectKeyEmptyError ) Error() string { return "Key is empty or '/'" }

func TrimLeadingSlash( s string ) string {
    switch l := len( s ); {
        case l > 0 && s[ 0 ] == '/': s = s[ 1 : ]
        case l > 2 && 
             s[ 0 ] == '%' && s[ 1 ] == '2' && 
             ( s[ 2 ] == 'f' || s[ 2 ] == 'F' ): s = s[ 3 : ]
    }
    return s
}

func AsObjectKey( s string ) ( ObjectKey, error ) {
    s = TrimLeadingSlash( s )
    if len( s ) == 0 { return "", &ObjectKeyEmptyError{} }
    return ObjectKey( s ), nil
}

func ExpectObjectKey( s string ) ObjectKey { 
    res, err := AsObjectKey( s )
    if err != nil { panic( err ) }
    return res
}

type Bucket string

type ImplementationError struct { msg string }

func ( e *ImplementationError ) Error() string { return e.msg }

func implError( fmtStr string, args ...interface{} ) *ImplementationError {
    return &ImplementationError{ fmt.Sprintf( fmtStr, args... ) }
}

type Endpoint struct {
    UseSsl bool
}

func ( e *Endpoint ) beginUrl() *bytes.Buffer {
    buf := &bytes.Buffer{}
    buf.WriteString( "http" )
    if e.UseSsl { buf.WriteString( "s" ) }
    buf.WriteString( "://s3.amazonaws.com" )
    return buf
}

type ObjectRequestRoute struct {
    *Endpoint
    Bucket
    Key ObjectKey 
}

func ( r *ObjectRequestRoute ) resourceToSign( httpReq *http.Request ) string {
    return httpReq.URL.RequestURI()
}

func ( r *ObjectRequestRoute ) getUrl( ep *Endpoint ) string {
    buf := ep.beginUrl()
    buf.WriteString( "/" )
    buf.WriteString( url.QueryEscape( string( r.Bucket ) ) )
    buf.WriteString( "/" )
    buf.WriteString( url.QueryEscape( string( r.Key ) ) )
    return buf.String()
}

type ResponseContext struct {
    RequestId string
    RequestId2 string
}

func responseContextFor( resp *http.Response ) *ResponseContext {
    return &ResponseContext{
        RequestId: resp.Header.Get( hdrReqId ),
        RequestId2: resp.Header.Get( hdrReqId2 ),
    }
}

type GetObjectRequest struct {
    Route *ObjectRequestRoute
}

type GetObjectResponse struct {
    *ResponseContext
    Etag
    ContentType
    metaHeader http.Header
}

func ( gr *GetObjectResponse ) GetMetaHeader( key string ) string {
    if gr.metaHeader == nil { return "" }
    if headerHasPrefix( key, metaHdrPrefix ) {
        key = key[ len( metaHdrPrefix ) : ]
    }
    return gr.metaHeader.Get( key )
}

type PutObjectRequest struct {
    Route *ObjectRequestRoute
    Body io.Reader
    ContentLength int64
    ContentType
    Md5 []byte
    metaHeaders http.Header
}

func ( req *PutObjectRequest ) SetBodyBytes( body []byte ) {
    req.Body = bytes.NewReader( body )
    req.ContentLength = int64( len( body ) )
}

func ( req *PutObjectRequest ) SetBodyFile( nm string ) error {
    if f, err := os.Open( nm ); err == nil {
        if stat, err := f.Stat(); err == nil {
            req.Body, req.ContentLength = f, stat.Size()
        } else { return err }
    } else { return err }
    return nil
}

func headerHasPrefix( key, prefix string ) bool {
    prefLen := len( prefix )
    if len( key ) < prefLen { return false }
    return strings.ToLower( key[ : prefLen ] ) == prefix
} 

func asMetaHeaderKey( key string ) string {
    if ! headerHasPrefix( key, metaHdrPrefix ) { key = metaHdrPrefix + key }
    return key
}

func ( r *PutObjectRequest ) AddMetaHeader( key, value string ) {
    key = asMetaHeaderKey( key )
    if r.metaHeaders == nil {
        r.metaHeaders = makeHeader()
    }
    r.metaHeaders.Add( key, value )
}

func ( r *PutObjectRequest ) setOptMd5( httpReq *http.Request ) {
    if r.Md5 != nil {
        md5Str := base64.StdEncoding.EncodeToString( r.Md5 )
        httpReq.Header.Set( hdrMd5, md5Str )
    }
}

func ( r *PutObjectRequest ) setOptCtype( httpReq *http.Request ) {
    if r.ContentType != "" {
        httpReq.Header.Set( hdrCtype, string( r.ContentType ) )
    }
}

func ( r *PutObjectRequest ) setMetaHeaders( httpReq *http.Request ) {
    if r.metaHeaders != nil {
        for k, v := range r.metaHeaders { httpReq.Header[ k ] = v }
    }
}

type PutObjectResponse struct {
    *ResponseContext
    Etag
}

type DeleteObjectRequest struct {
    Route *ObjectRequestRoute
}

type DeleteObjectResponse struct {
    *ResponseContext
}

type NoBodyError struct {}

func ( e *NoBodyError ) Error() string { return "Request is missing body" }

type Client struct {
    *Credentials
    cli *http.Client
    defaultEndpoint *Endpoint // hardcoded now during construction
}

func NewClient( creds *Credentials ) *Client {
    return &Client{ 
        Credentials: creds, 
        cli: &http.Client{},
        defaultEndpoint: AwsSslEndpoint,
    }
}

func ( c *Client ) asEndpoint( ep *Endpoint ) *Endpoint {
    if ep == nil { ep = c.defaultEndpoint }
    return ep
}

// Even if xml doc contains a <RequestId> element, we ignore it here in favor of
// the id(s) set in the ResponseContext associated with whatever response led to
// the creation of this instance
type S3RemoteError struct {
    *ResponseContext `xml:-`
    Code string 
    HttpStatusCode int
    Message string 
    HostId string 
    AWSAccessKeyId string 
}

const (
    tmplRemoteErrWithoutCode = "(HTTP %d): %s"
    tmplRemoteErrWithCode = "(HTTP %d) %s: %s"
)

func ( e *S3RemoteError ) Error() string { 
    args := make( []interface{}, 1, 3 )
    args[ 0 ] = e.HttpStatusCode
    var tmpl string
    if e.Code == "" { 
        tmpl = tmplRemoteErrWithoutCode 
    } else {
        tmpl = tmplRemoteErrWithCode
        args = append( args, e.Code )
    }
    args = append( args, e.Message )
    return fmt.Sprintf( tmpl, args... )
}

func ( e *S3RemoteError ) setFieldsFromXml( r io.Reader ) error {
    dec := xml.NewDecoder( r )
    return dec.Decode( e )
}

func ( c *Client ) asS3Error( httpResp *http.Response ) error {
    res := &S3RemoteError{ 
        ResponseContext: responseContextFor( httpResp ),
        HttpStatusCode: httpResp.StatusCode,
    }
    if httpResp.StatusCode >= 400 && 
       httpResp.Header.Get( hdrCtype ) == ctypeS3Err {
        if err := res.setFieldsFromXml( httpResp.Body ); err != nil {
            log.Printf( 
                "Warning: error processing xml error body (returning generic " +
                "S3 error to caller): %s", err )
        }
    } else { res.Message = httpResp.Status }
    return res
}

func ( c *Client ) httpDo( 
    httpReq *http.Request, expctCode int ) ( *http.Response, error ) {
    var httpResp *http.Response
    var err error
    httpResp, err = c.cli.Do( httpReq )
    if err == nil {
        if httpResp.StatusCode != expctCode { err = c.asS3Error( httpResp ) }
    }
    return httpResp, err
}

func writeUnfoldedHeaderValue( w *bytes.Buffer, val string ) {
    for i, line := range strings.Split( val, "\n" ) {
        if line = strings.TrimSpace( line ); line != "" {
            if i > 0 { w.WriteString( "," ) }
            w.WriteString( strings.TrimSpace( line ) )
        }
    }
}

func canonicalizeHeaderValues( vals []string ) string {
    buf := bytes.Buffer{}
    for i, val := range vals {
        writeUnfoldedHeaderValue( &buf, val )
        if i < len( vals ) - 1 { buf.WriteString( "," ) }
    }
    return buf.String()
}

// Side effect: canonicalizes x-amz- headers in place as found in req. We have
// to do this to ensure that S3 receives the same form that we send. Since we
// try to adhere to S3 stated strict requirements (not clear that S3 would care
// if we loosened them a bit, but no need to risk it), and since net/http.Header
// doesn't compress contiguous whitespace as the S3 docs request, we need to do
// it ourselves.
func addCanonicalizedHeaders( req *http.Request, str *bytes.Buffer ) {
    strs := make( []string, 0, len( req.Header ) )
    for k, v := range req.Header {
        if headerHasPrefix( k, amzHdrPrefix ) {
            buf := &bytes.Buffer{}
            buf.WriteString( strings.ToLower( k ) )
            buf.WriteString( ":" )
            canon := canonicalizeHeaderValues( v )
            buf.WriteString( canon )
            strs = append( strs, buf.String() )
            req.Header.Set( k, canon )
        }
    }
    if len( strs ) > 0 {
        sort.Strings( strs )
        str.WriteString( strings.Join( strs, "\n" ) )
        str.WriteString( "\n" )
    }
}

// Side effect: sets Date header
func writeSigDate( str *bytes.Buffer, req *http.Request ) {
    dtStr := time.Now().Format( time.RFC1123 )
    req.Header.Set( hdrDate, dtStr )
    str.WriteString( dtStr )
    str.WriteString( "\n" )
}

func ( c *Client ) completeRequestSign( buf []byte, req *http.Request ) {
    key := []byte( c.Credentials.SecretKey )
    h := hmac.New( func() hash.Hash { return sha1.New() }, key )
    sig := hashes.HashOfBytes( h, buf )
    b64 := base64.StdEncoding.EncodeToString( sig )
    authStr := "AWS " + string( c.Credentials.AccessKeyId ) + ":" + b64
    req.Header.Set( hdrAuth, authStr )
}

// Side effect: sets Date header and canonicalizes x-amz- headers in req
// in-place (see addCanonicalizedHeaders())
func ( c *Client ) signRequest( req *http.Request, canonRsrc string ) {
    str := bytes.Buffer{}
    str.WriteString( req.Method )
    str.WriteString( "\n" )
    if s := req.Header.Get( hdrMd5 ); s != "" { str.WriteString( s ) }
    str.WriteString( "\n" )
    if s := req.Header.Get( hdrCtype ); s != "" { str.WriteString( s ) }
    str.WriteString( "\n" )
    writeSigDate( &str, req )
    addCanonicalizedHeaders( req, &str )
    str.WriteString( canonRsrc )
//    log.Printf( "Signing %s", str.String() )
    c.completeRequestSign( str.Bytes(), req )
}

func asObjectHttpRequest( 
    method string, 
    r *ObjectRequestRoute, 
    body io.Reader,
    c *Client ) ( *http.Request, error ) {
    ep := c.asEndpoint( r.Endpoint )
    return http.NewRequest( method, r.getUrl( ep ), body )
}

func expectEtagIn( httpResp *http.Response ) ( Etag, error ) {
    var etag string
    if etag = httpResp.Header.Get( hdrEtag ); etag == "" {
        return "", &EtagMissingError{}
    } 
    return Etag( etag ), nil
}

func createPutObjectResponse( 
    httpResp *http.Response ) ( *PutObjectResponse, error ) {
    resp := new( PutObjectResponse )
    resp.ResponseContext = responseContextFor( httpResp )
    if etag := httpResp.Header.Get( hdrEtag ); etag == "" {
        return nil, &EtagMissingError{}
    } else { resp.Etag = Etag( etag ) }
    return resp, nil
}

func ( c *Client ) PutObject( 
    req *PutObjectRequest ) ( *PutObjectResponse, error ) {
    if req.Body == nil { return nil, &NoBodyError{} }
    httpReq, err := asObjectHttpRequest( "PUT", req.Route, req.Body, c )
    if err != nil { return nil, err }
    httpReq.ContentLength = req.ContentLength
    req.setOptMd5( httpReq )
    req.setOptCtype( httpReq )
    req.setMetaHeaders( httpReq )
    c.signRequest( httpReq, req.Route.resourceToSign( httpReq ) )
    httpResp, err := c.httpDo( httpReq, 200 )
    if err != nil { return nil, err }
    defer httpResp.Body.Close()
    return createPutObjectResponse( httpResp )
}

// As of this writing, the inner loop below is not necessary since S3 returns a
// single ws-folded header val for each meta header. To protect against the
// (unlikely) chance that this changes later, we nevertheless assume that values
// could come separately and recollect them in the inner loop
func metaHeaderFor( httpResp *http.Response ) ( hdr http.Header ) {
    for k, vals := range httpResp.Header {
        if headerHasPrefix( k, metaHdrPrefix ) {
            if hdr == nil { hdr = makeHeader() }
            k = k[ len( metaHdrPrefix ) : ]
            for _, val := range vals { hdr.Add( k, val ) }
        }
    }
    return
}

func createGetObjectResponse( 
    httpResp *http.Response ) ( resp *GetObjectResponse, err error ) {
    resp = new( GetObjectResponse )
    resp.ResponseContext = responseContextFor( httpResp )
    resp.Etag, err = expectEtagIn( httpResp )
    resp.ContentType = ContentType( httpResp.Header.Get( hdrCtype ) )
    resp.metaHeader = metaHeaderFor( httpResp )
    return
}

// Note on defer statement below: we are trying to be aggressive and close the
// Body even if createGetObjectResponse() panics for some reason. That is why
// inside the defer block the check is 'getResp == nil' instead of simply 'err
// != nil'
//
// Assuming that this method does not panic, it returns non-nil for httpResp and
// getResp simultaneously if and only if err == nil
func execGetObjectRequest(
    req *GetObjectRequest, 
    method string, 
    c *Client ) ( httpResp *http.Response, 
                  getResp *GetObjectResponse, 
                  err error ) {
    var httpReq *http.Request
    httpReq, err = asObjectHttpRequest( method, req.Route, nil, c )
    if err != nil { return }
    c.signRequest( httpReq, req.Route.resourceToSign( httpReq ) )
    httpResp, err = c.httpDo( httpReq, 200 )
    if err != nil { return }
    defer func() { 
        if getResp == nil { 
            httpResp.Body.Close() 
            httpResp = nil
        }
    }()
    getResp, err = createGetObjectResponse( httpResp )
    return 
}

type GetObjectBodyHandler func( *GetObjectResponse, io.Reader ) error

func ( c *Client ) GetObject( 
    req *GetObjectRequest, 
    h GetObjectBodyHandler ) ( *GetObjectResponse, error ) {
    httpResp, getResp, err := execGetObjectRequest( req, "GET", c )
    if err != nil { return nil, err }
    defer httpResp.Body.Close()
    if err := h( getResp, httpResp.Body ); err != nil { return nil, err }
    return getResp, nil
}

func ( c *Client ) HeadObject( 
    req *GetObjectRequest ) ( *GetObjectResponse, error ) {
    httpResp, getResp, err := execGetObjectRequest( req, "HEAD", c )
    if err != nil { return nil, err }
    httpResp.Body.Close()
    return getResp, nil
}

func ( c *Client ) DeleteObject(
    req *DeleteObjectRequest ) ( *DeleteObjectResponse, error ) {
    httpReq, err := asObjectHttpRequest( "DELETE", req.Route, nil, c )
    if err != nil { return nil, err }
    c.signRequest( httpReq, req.Route.resourceToSign( httpReq ) )
    httpResp, err := c.httpDo( httpReq, 204 )
    if err != nil { return nil, err }
    httpResp.Body.Close()
    return &DeleteObjectResponse{ responseContextFor( httpResp ) }, nil
}

type ListContext struct {
    *Endpoint
    Bucket
    Prefix string
    Marker string
    Delimiter string
    MaxKeys int
}

func ( ls *ListContext ) validate() error {
    if ls.Bucket == "" { return errors.New( "Listing has no Bucket" ) }
    return nil
}

func appendOptQuery( buf *bytes.Buffer, isFirst *bool, k, v string ) {
    if len( v ) > 0 {
        if *isFirst { 
            buf.WriteByte( '?' ) 
            *isFirst = false
        } else { buf.WriteByte( '&' ) }
        buf.WriteString( url.QueryEscape( k ) )
        buf.WriteByte( '=' )
        buf.WriteString( url.QueryEscape( v ) )
    }
}

func createListHttpRequest( 
    lsCtx *ListContext, c *Client ) ( *http.Request, error ) {
    buf := c.asEndpoint( lsCtx.Endpoint ).beginUrl()
    buf.WriteString( "/" )
    buf.WriteString( url.QueryEscape( string( lsCtx.Bucket ) ) )
    isFirst := true
    appendOptQuery( buf, &isFirst, keyPrefix, lsCtx.Prefix )
    appendOptQuery( buf, &isFirst, keyDelim, lsCtx.Delimiter )
    appendOptQuery( buf, &isFirst, keyMarker, lsCtx.Marker )
    if mx := lsCtx.MaxKeys; mx > 0 {
        appendOptQuery( buf, &isFirst, keyMaxKeys, strconv.Itoa( mx ) )
    }
    return http.NewRequest( "GET", buf.String(), nil )
}

func ( lsCtx *ListContext ) resourceToSign( httpReq *http.Request ) string {
    str := httpReq.URL.RequestURI()
    if i := strings.IndexRune( str, '?' ); i >= 0 { str = str[ : i ] }
    return str
}

type ObjectOwner struct {
    Id string `xml:"ID"`
    DisplayName string
}

type ListedObject struct {
    Etag Etag `xml:"ETag"`
    Key string
    LastModified string
    Size int64
    Owner ObjectOwner
}

type ListResult struct {
    *ResponseContext `xml:-`
    Bucket Bucket `xml:"Name"`
    Prefix string
    Marker string
    IsTruncated bool
    MaxKeys int
    Delimiter string

    // May be empty; never nil on an instance returned by this library
    Contents []ListedObject `xml:"Contents"`
    CommonPrefixes []string `xml:"CommonPrefixes>Prefix"`
}

func ( r *ListResult ) InferMarker() string {
    var lastPref, lastKey string
    if l := len( r.Contents ); l > 0 { lastKey = r.Contents[ l - 1 ].Key }
    if l := len( r.CommonPrefixes ); l > 0 { 
        lastPref = r.CommonPrefixes[ l - 1 ]
    }
    res := lastPref
    if lastKey > lastPref { res = lastKey }
    return res
}

func ( r *ListResult ) NextListing() *ListContext {
    return &ListContext{
        Bucket: r.Bucket,
        Prefix: r.Prefix,
        Marker: r.InferMarker(),
        MaxKeys: r.MaxKeys,
        Delimiter: r.Delimiter,
    }
}

func createListResult( httpResp *http.Response ) ( *ListResult, error ) {
    if ctype := httpResp.Header.Get( hdrCtype ); ctype != ctypeListRes {
        return nil, implError( "Unexpected list result ctype: %s", ctype )
    }
    res := &ListResult{ ResponseContext: responseContextFor( httpResp ) }
    dec := xml.NewDecoder( httpResp.Body );
    if err := dec.Decode( res ); err != nil { return nil, err }
    // ensure that Contents and CommonPrefixes are never nil
    if res.Contents == nil { res.Contents = make( []ListedObject, 0 ) }
    if res.CommonPrefixes == nil { res.CommonPrefixes = make( []string, 0 ) }
    return res, nil
}

func ( c *Client ) ListNext( lsCtx *ListContext ) ( *ListResult, error ) {
    if err := lsCtx.validate(); err != nil { return nil, err }
    httpReq, err := createListHttpRequest( lsCtx, c )
    if err != nil { return nil, err }
    c.signRequest( httpReq, lsCtx.resourceToSign( httpReq ) )
    httpResp, err := c.httpDo( httpReq, 200 )
    if err != nil { return nil, err }
    defer httpResp.Body.Close()
    return createListResult( httpResp )
}

var EndScan error
func init() { EndScan = errors.New( "End Scan (sentinel)" ) }

type ScanBucketFunc func( *ListResult ) error

func ( c *Client ) ScanBucket( lsCtx *ListContext, f ScanBucketFunc ) error {
    for lsCtx != nil {
        if lsRes, err := c.ListNext( lsCtx ); err == nil {
            lsCtx = nil
            if err := f( lsRes ); err == nil && lsRes.IsTruncated {
                lsCtx = lsRes.NextListing()
            } else {
                if err == EndScan { return nil } else { return err }
            }
        } else { return err }
    }
    return nil
}

type WalkObjectsFunc func( obj ListedObject ) error

func ( c *Client ) WalkObjects( lsCtx *ListContext, f WalkObjectsFunc ) error {
    wrapper := func( lsRes *ListResult ) error {
        for _, obj := range lsRes.Contents { 
            if err := f( obj ); err != nil { return err }
        }
        return nil
    }
    return c.ScanBucket( lsCtx, wrapper )
}

type EvFileSynced struct {
    Key ObjectKey
    DestFile string
}

type EvFileSyncStarting struct {
    Key ObjectKey
    DestFile string
}

type SyncPredicate func( obj *ListedObject, locFile string ) ( bool, error )

func RemoteFileLargerPredicate( 
    obj *ListedObject, locFile string ) ( bool, error ) {
    info, err := os.Stat( locFile )
    if os.IsNotExist( err ) { return true, nil }
    if err != nil { return false, err }
    return info.Size() < obj.Size, nil
}

type ReceiveInitializer func( 
    recv *filerecv.FileReceive, obj *ListedObject ) error
    

type DirectorySync struct {
    LocalRoot string
    RemoteRoot *ListContext
    DryRun bool
    SyncPredicate SyncPredicate
    EventHandler func( ev interface{} )
    ReceiveInitializer ReceiveInitializer
}

func ( ds *DirectorySync ) validate() error {
    if ds.LocalRoot == "" { 
        return errors.New( "Directory sync is missing a local root" )
    }
    if ds.RemoteRoot == nil {
        return errors.New( "Directory sync is missing a remote root" )
    }
    if ds.SyncPredicate == nil {
        return errors.New( "Directory sync is missing sync predicate" )
    }
    return nil
}

func ( ds *DirectorySync ) sendEvent( ev interface{} ) {
    if ds.EventHandler != nil { ds.EventHandler( ev ) }
}

func ( ds *DirectorySync ) fileReceiveFor( 
    locFile string, obj *ListedObject ) ( *filerecv.FileReceive, error ) {
    recv := &filerecv.FileReceive{ DestFile: locFile }
    if ds.ReceiveInitializer == nil { return recv, nil }
    return recv, ds.ReceiveInitializer( recv, obj )
}

func ( ds *DirectorySync ) getFile( 
    obj *ListedObject, locFile string, lsRes *ListResult, c *Client ) error {
    get := &GetObjectRequest{
        Route: &ObjectRequestRoute{
            Bucket: lsRes.Bucket, 
            Key: ExpectObjectKey( obj.Key ),
        },
    }
    f := func( gr *GetObjectResponse, r io.Reader ) error {
        recv, err := ds.fileReceiveFor( locFile, obj )
        if err != nil { return err }
        if _, err := filerecv.ReceiveFile( recv, r ); err != nil { return err }
        ds.sendEvent( &EvFileSynced{ get.Route.Key, recv.DestFile } )
        return nil
    }
    if _, err := c.GetObject( get, f ); err != nil { return err }
    return nil
}

func ensureParentOf( f string ) error {
    parent := filepath.Dir( f )
    return os.MkdirAll( parent, os.ModeDir | 0777 )
}

func ( ds *DirectorySync ) syncObject( 
    obj *ListedObject, lsRes *ListResult, c *Client ) error {
    rel, err := filepath.Rel( lsRes.Prefix, obj.Key )
    if err != nil { return err }
    dest := filepath.Join( ds.LocalRoot, rel )
    shouldSync, err := ds.SyncPredicate( obj, dest )
    if err != nil { return err }
    if shouldSync {
        ds.sendEvent( &EvFileSyncStarting{ ExpectObjectKey( obj.Key ), dest } )
        if ! ds.DryRun {
            if err := ensureParentOf( dest ); err == nil {
                if err := ds.getFile( obj, dest, lsRes, c ); err != nil {
                    return err
                }
            } else { return err }
        }
    }
    return nil
}

func ( c *Client ) SyncDirectory( ds *DirectorySync ) error {
    if err := ds.validate(); err != nil { return err }
    f := func( lsRes *ListResult ) error {
        for _, obj := range lsRes.Contents {
            if err := ds.syncObject( &obj, lsRes, c ); err != nil { return err }
        }
        return nil
    }
    return c.ScanBucket( ds.RemoteRoot, f )
}
