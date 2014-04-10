package main

import (
    "log"
    "flag"
    "fmt"
    "os"
    "aws/s3"
)

var object string
var localFile string
var bucket string
var awsAccountId string
var awsSecretKey string
var amzAcl string

func notEmpty( val, name string ) {
    if val == "" {
        fmt.Fprintf( os.Stderr, "Missing value for -%s\n", name )
        flag.PrintDefaults()
        os.Exit( -1 )
    }
}

func validateArgs() {
    notEmpty( localFile, "local-file" )
    notEmpty( object, "object" )
    notEmpty( bucket, "bucket" )
    notEmpty( awsAccountId, "awsAccountId" )
    notEmpty( awsSecretKey, "awsSecretKey" )
    if ! ( amzAcl == "" || amzAcl == "public-read" ) {
        log.Fatalf( "unsupported acl: %s", amzAcl )
    }
}

func parseArgs() {
    flag.StringVar( &object, "object", "", "Store to this object" )
    flag.StringVar( &localFile, "local-file", "", "Data to store" )
    flag.StringVar( &bucket, "bucket", "", "Remote bucket" )
    flag.StringVar( &awsAccountId, "aws-account-id", "", "AWS Account ID" )
    flag.StringVar( &awsSecretKey, "aws-secret-key", "", "AWS Secret Key" )
    flag.StringVar( &amzAcl, "amz-acl", "", "Use a canned amazon acl" )
    flag.Parse()
    validateArgs()
}

func putObject() {
    log.Printf( "putting %s:%s from local file %s", bucket, object, localFile )
    creds := &s3.Credentials{ 
        AccessKey: s3.AccessKey( awsAccountId ), 
        SecretKey: s3.SecretKey( awsSecretKey ),
    }
    cli := s3.NewClient( creds )
    rt := &s3.ObjectRequestRoute{ Bucket: s3.Bucket( bucket ) }
    if key, err := s3.AsObjectKey( object ); err == nil {
        rt.Key = key
    } else {
        log.Fatalf( "bad object key format: %s", object )
    }
    req := &s3.PutObjectRequest{ Route: rt, Acl: s3.AclType( amzAcl ) }
    if err := req.SetBodyFile( localFile ); err != nil {
        log.Fatalf( "couldn't set body from file %s: %s", localFile, err )
    }
    if _, err := cli.PutObject( req ); err != nil { log.Fatal( err ) }
}

func main() {
    parseArgs()
    validateArgs()
    putObject()
}
