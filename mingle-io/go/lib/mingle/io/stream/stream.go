package stream

import (
    "io"
//    "log"
    mgio "mingle/io"
)

const (
    messageVersion1 = int32( 1 )
    typeCodeHeaders = int32( 1 )
    typeCodeMessageBody = int32( 2 )
)

type BodyWriter func( w io.Writer ) error

type MessageReader func( msg *mgio.Headers, sz int64, r io.Reader ) error

func writeMessage(
    msg *mgio.Headers, sz int64, bw BodyWriter, w io.Writer ) error {
    if err := mgio.WriteInt32( messageVersion1, w ); err != nil { return err }
    if err := mgio.WriteInt32( typeCodeHeaders, w ); err != nil { return err }
    if err := mgio.WriteHeaders( msg, w ); err != nil { return err }
    if err := mgio.WriteInt32( typeCodeMessageBody, w ); err != nil { 
        return err 
    }
    if err := mgio.WriteInt64( sz, w ); err != nil { return err }
    if sz > 0 { return bw( w ) }
    return nil
}

func readMessage( rd io.Reader, mr MessageReader ) error {
    if err := mgio.ReadVersion( messageVersion1, "message", rd ); err != nil { 
        return err 
    }
    if err := mgio.ExpectTypeCode( typeCodeHeaders, rd ); err != nil { 
        return err
    }
    hdrs, err := mgio.ReadHeaders( rd )
    if err != nil { return err }
    if err = mgio.ExpectTypeCode( typeCodeMessageBody, rd ); err != nil {
        return err
    }
    var sz int64
    if sz, err = mgio.ReadInt64( rd ); err != nil { return err }
    rd = io.LimitReader( rd, sz )
    return mr( hdrs, sz, rd )
}

type Connection interface {
    WriteMessage( msg *mgio.Headers, sz int64, w BodyWriter ) error
    ReadMessage( r MessageReader ) error
    Close() error
}

type connectionImpl struct {
    rd io.ReadCloser
    wr io.WriteCloser
}

func ( c *connectionImpl ) Close() error {
    err1 := c.rd.Close()
    err2 := c.wr.Close()
    if err1 == nil { return err2 }
    return err1
}

func ( c *connectionImpl ) WriteMessage(
    msg *mgio.Headers, sz int64, w BodyWriter ) error {
    return writeMessage( msg, sz, w, c.wr )
}

func ( c *connectionImpl ) ReadMessage( r MessageReader ) error {
    return readMessage( c.rd, r )
}

func NewConnection( rd io.ReadCloser, wr io.WriteCloser ) Connection {
    return &connectionImpl{ rd, wr }
}

func OpenPipe() ( cli, srv Connection ) {
    p1Rd, p1Wr := io.Pipe()
    p2Rd, p2Wr := io.Pipe()
    return NewConnection( p1Rd, p2Wr ), NewConnection( p2Rd, p1Wr )
}

func WriteMessageBytes( 
    conn Connection, msg *mgio.Headers, body []byte ) error {
    sender := func( w io.Writer ) error {
        _, err := w.Write( body )
        return err
    }
    return conn.WriteMessage( msg, int64( len( body ) ), sender )
}

func WriteMessageNoBody( conn Connection, msg *mgio.Headers ) error {
    return WriteMessageBytes( conn, msg, []byte{} )
}

func ReadMessageBytes(
    conn Connection ) ( msg *mgio.Headers, body []byte, err error ) {
    reader := func( msgIn *mgio.Headers, sz int64, r io.Reader ) error {
        msg = msgIn
        body = make( []byte, sz )
        _, err = io.ReadFull( r, body )
        return err
    }
    err = conn.ReadMessage( reader )
    return
}
