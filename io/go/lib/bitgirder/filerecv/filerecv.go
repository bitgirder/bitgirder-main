package filerecv

import (
    "os"
    "io"
    "errors"
    "log"
    "path/filepath"
)

var emptyPathErr error
func init() { emptyPathErr = errors.New( "Empty path" ) }

func DottedTempOf( p string ) ( string, error ) {
    if p = filepath.Clean( p ); p == "." || p == "/" { return "", emptyPathErr }
    dir, base := filepath.Dir( p ), filepath.Base( p )
    return filepath.Join( dir, "." + base ), nil
}

func ExpectDottedTempOf( p string ) string {
    res, err := DottedTempOf( p )
    if err != nil { panic( err ) }
    return res
}

type FileReceive struct {
    TempFile string
    DestFile string
    PreserveOnFail bool
}

func ( recv *FileReceive ) validate() error {
    if recv.DestFile == "" {
        return errors.New( "FileReceive has no destination" )
    }
    return nil
}

func ( recv *FileReceive ) receiver() ( *os.File, error ) {
    fName := recv.TempFile
    if fName == "" { fName = recv.DestFile }
    return os.Create( fName )
}

// see note at ReceiveFile() on error handling here
func cleanupFile( f string ) {
    if err := os.Remove( f ); err != nil {
        log.Printf( "Could not remove %s during cleanup of failed receive: %s",
            f, err )
    }
}

func receiveFailed( recv *FileReceive ) {
    if ! recv.PreserveOnFail { 
        if recv.TempFile != "" && recv.TempFile != recv.DestFile {
            cleanupFile( recv.TempFile )
        }
        cleanupFile( recv.DestFile )
    }
}

// Return values are those from io.Copy but with some possible other error
// conditions, such as might arise from renaming the tempfile. If error X causes
// us to abort the receive and error Y is encountered during cleanup, X will be
// returned and Y will be logged. This last behavior could be further
// parameterized later by adding an optional EventListener/Reactor field to
// FileReceive which has some sort of callback method which could receive Y
//
// If rename from temp --> dest fails, that error is returned and both files are
// left in whatever state they were in after the failed Rename() call
//
// If recv.PreserveOnFail is true, this function will leave recv.TempFile
// and recv.DestFile in whatever state they were in (including not existing) at
// the time of failure; if recv.PreserveOnFail is false, a best effort attempt
// will be made to ensure that neither file exists before this function returns
// its error
func ReceiveFile( 
    recv *FileReceive, src io.Reader ) ( written int64, err error ) {
    if err = recv.validate(); err != nil { return }
    var f *os.File
    if f, err = recv.receiver(); err != nil { return }
    defer f.Close()
    if written, err = io.Copy( f, src ); err == nil {
        f.Close() // ensure closed and flushed before rename
        if recv.TempFile != "" && recv.TempFile != recv.DestFile {
            err = os.Rename( recv.TempFile, recv.DestFile )
        }
    } else { receiveFailed( recv ) }
    return
}
