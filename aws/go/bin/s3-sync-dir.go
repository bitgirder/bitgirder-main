package main

import (
    "log"
    "flag"
    "fmt"
    "os"
    "aws/s3"
    "bitgirder/filerecv"
)

var remoteRoot string
var localRoot string
var bucket string
var awsAccountId string
var awsSecretKey string

func notEmpty( val, name string ) {
    if val == "" {
        fmt.Fprintf( os.Stderr, "Missing value for -%s\n", name )
        flag.PrintDefaults()
        os.Exit( -1 )
    }
}

func validateArgs() {
    notEmpty( localRoot, "local-root" )
    notEmpty( bucket, "bucket" )
    notEmpty( awsAccountId, "awsAccountId" )
    notEmpty( awsSecretKey, "awsSecretKey" )
}

func parseArgs() {
    flag.StringVar( &remoteRoot, "remote-root", "", "Remote root for sync" )
    flag.StringVar( &localRoot, "local-root", "", "Local root for sync" )
    flag.StringVar( &bucket, "bucket", "", "Remote bucket for sync" )
    flag.StringVar( 
        &awsAccountId, "aws-account-id", "", "AWS Account ID for sync" )
    flag.StringVar( 
        &awsSecretKey, "aws-secret-key", "", "AWS Secret Key for sync" )
    flag.Parse()
    validateArgs()
}

func handleEvent( ev interface{} ) {
    log.Printf( "Sync event: %#v", ev )
}

func initReceive( recv *filerecv.FileReceive, obj *s3.ListedObject ) error {
    recv.TempFile = filerecv.ExpectDottedTempOf( recv.DestFile )
    return nil
}

func main() {
    parseArgs()
    remoteRoot = s3.TrimLeadingSlash( remoteRoot )
    creds := &s3.Credentials{ 
        AccessKey: s3.AccessKey( awsAccountId ), 
        SecretKey: s3.SecretKey( awsSecretKey ),
    }
    cli := s3.NewClient( creds )
    ds := &s3.DirectorySync{
        LocalRoot: localRoot,
        RemoteRoot: &s3.ListContext{ 
            Prefix: remoteRoot, 
            Bucket: s3.Bucket( bucket ),
        },
    }
    ds.EventHandler = handleEvent
    ds.SyncPredicate = s3.RemoteFileLargerPredicate
    ds.ReceiveInitializer = initReceive
    if err := cli.SyncDirectory( ds ); err != nil { log.Fatal( err ) }
}
