package mingle

import (
    "bytes"
)

const binWriterTcFail = int32( 0x64 )

type BinWriterFailureInput struct {
    Id *Identifier
    ErrMsg string
    Input []byte
}

func appendInput( data interface{}, w *BinWriter ) {
    switch v := data.( type ) {
    case *Identifier: 
        if err := w.WriteIdentifier( v ); err != nil { panic( err ) }
    case string: 
        if err := w.WriteValue( String( v ) ); err != nil { panic( err ) }
    case uint8: if err := w.w.WriteUint8( v ); err != nil { panic( err ) }
    case int32: if err := w.w.WriteInt32( v ); err != nil { panic( err ) }
    case TypeReference: 
        if err := w.WriteTypeReference( v ); err != nil { panic( err ) }
    default: panic( libErrorf( "Unrecognized input elt: %T", v ) )
    }
}

func makeBinWriterFailureInput( data ...interface{} ) []byte {
    buf := &bytes.Buffer{}
    w := NewWriter( buf )
    for _, val := range data { appendInput( val, w ) }
    return buf.Bytes()
}

var BinWriterFailureInputs = []*BinWriterFailureInput{
    &BinWriterFailureInput{
        Id: MustIdentifier( "unexpected-top-level-type-code" ),
        ErrMsg: "[offset 0]: Unrecognized value code: 0x64",
        Input: makeBinWriterFailureInput( binWriterTcFail ),
    },
    &BinWriterFailureInput{
        Id: MustIdentifier( "unexpected-symmap-val-type-code" ),
        ErrMsg: `[offset 41]: Unrecognized value code: 0x64`,
        Input: makeBinWriterFailureInput(
            tcStruct, int32( -1 ), MustTypeReference( "ns@v1/S" ),
            tcField, MustIdentifier( "f1" ), binWriterTcFail,
        ),
    },
    &BinWriterFailureInput{
        Id: MustIdentifier( "unexpected-list-val-type-code" ),
        ErrMsg: `[offset 51]: Unrecognized value code: 0x64`,
        Input: makeBinWriterFailureInput(
            tcStruct, int32( -1 ), MustTypeReference( "ns@v1/S" ),
            tcField, MustIdentifier( "f1" ),
            tcList, int32( -1 ),
            tcInt32, int32( 10 ), // an okay list val
            binWriterTcFail,
        ),
    },
}
