package types

import (
    "io"
    "fmt"
//    "log"
    bgio "bitgirder/io"
    mg "mingle"
)

const (

    tcNull = uint8( 0x00 )
    tcDefMap = uint8( 0x01 )
    tcPrimDef = uint8( 0x02 )
    tcAliasDef = uint8( 0x03 )
    tcProtoDef = uint8( 0x04 )
    tcStructDef = uint8( 0x05 )
    tcConstructorDef = uint8( 0x06 )
    tcEnumDef = uint8( 0x07 )
    tcServiceDef = uint8( 0x08 )
)

type BinWriter struct { 
    w *bgio.BinWriter 
    mgw *mg.BinWriter
}

func NewBinWriter( wr io.Writer ) *BinWriter {
    w := bgio.NewLeWriter( wr )
    mgw := mg.AsWriter( w )
    return &BinWriter{ w: w, mgw: mgw }
}

func ( w *BinWriter ) writeLen( i int ) error {
    return w.w.WriteUint32( uint32( i ) )
}

func ( w *BinWriter ) writeOptSuperType( d Descendant ) ( err error ) {
    spr := d.GetSuperType()
    if err = w.w.WriteBool( spr != nil ); err != nil { return }
    if spr != nil {
        if err = w.mgw.WriteQualifiedTypeName( spr ); err != nil { return }
    }
    return
}

func ( w *BinWriter ) startDef( code uint8, d Definition ) ( err error ) {
    if err = w.mgw.WriteTypeCode( code ); err != nil { return }
    return w.mgw.WriteQualifiedTypeName( d.GetName() )
}

func ( w *BinWriter ) writePrimDef( pd *PrimitiveDefinition ) ( err error ) {
    return w.startDef( tcPrimDef, pd )
}

func ( w *BinWriter ) writeAliasDef( ad *AliasedTypeDefinition ) ( err error ) {
    if err = w.startDef( tcAliasDef, ad ); err != nil { return }
    return w.mgw.WriteTypeReference( ad.AliasedType )
}

func ( w *BinWriter ) writeField( fd *FieldDefinition ) ( err error ) {
    if err = w.mgw.WriteIdentifier( fd.Name ); err != nil { return }
    if err = w.mgw.WriteTypeReference( fd.Type ); err != nil { return }
    if defl := fd.Default; defl == nil {
        if err = w.mgw.WriteNull(); err != nil { return }
    } else {
        if err = w.mgw.WriteValue( defl ); err != nil { return }
    }
    return
}

func ( w *BinWriter ) writeFields( flds *FieldSet ) ( err error ) {
    if err = w.writeLen( flds.Len() ); err != nil { return }
    flds.EachDefinition( func( fd *FieldDefinition ) {
        if err == nil { err = w.writeField( fd ) }
    })
    return
}

func ( w *BinWriter ) writeCallSig( sig *CallSignature ) ( err error ) {
    if err = w.writeFields( sig.Fields ); err != nil { return }
    if err = w.mgw.WriteTypeReference( sig.Return ); err != nil { return }
    if err = w.writeLen( len( sig.Throws ) ); err != nil { return }
    for _, typ := range sig.Throws {
        if err = w.mgw.WriteTypeReference( typ ); err != nil { return }
    }
    return
}

func ( w *BinWriter ) writeProtoDef( pd *PrototypeDefinition ) ( err error ) {
    if err = w.startDef( tcProtoDef, pd ); err != nil { return }
    if err = w.writeCallSig( pd.Signature ); err != nil { return }
    return
}

func ( w *BinWriter ) writeConstructor( 
    cd *ConstructorDefinition ) ( err error ) {
    if err = w.mgw.WriteTypeCode( tcConstructorDef ); err != nil { return }
    if err = w.mgw.WriteTypeReference( cd.Type ); err != nil { return }
    return
}

func ( w *BinWriter ) writeStructDef( sd *StructDefinition ) ( err error ) {
    if err = w.startDef( tcStructDef, sd ); err != nil { return }
    if err = w.writeOptSuperType( sd ); err != nil { return }
    if err = w.writeFields( sd.Fields ); err != nil { return }
    if err = w.writeLen( len( sd.Constructors ) ); err != nil { return }
    for _, cons := range sd.Constructors {
        if err = w.writeConstructor( cons ); err != nil { return }
    }
    return
}

func ( w *BinWriter ) writeEnumDef( ed *EnumDefinition ) ( err error ) {
    if err = w.startDef( tcEnumDef, ed ); err != nil { return }
    if err = w.writeLen( len( ed.Values ) ); err != nil { return }
    for _, val := range ed.Values {
        if err = w.mgw.WriteIdentifier( val ); err != nil { return }
    }
    return
}

func ( w *BinWriter ) writeOperationDef( 
    od *OperationDefinition ) ( err error ) {
    if err = w.mgw.WriteIdentifier( od.Name ); err != nil { return }
    if err = w.writeCallSig( od.Signature ); err != nil { return }
    return
}

func ( w *BinWriter ) writeServiceDef( sd *ServiceDefinition ) ( err error ) {
    if err = w.startDef( tcServiceDef, sd ); err != nil { return }
    if err = w.writeOptSuperType( sd ); err != nil { return }
    if err = w.writeLen( len( sd.Operations ) ); err != nil { return }
    for _, od := range sd.Operations {
        if err = w.writeOperationDef( od ); err != nil { return }
    }
    sec := sd.Security
    if err = w.w.WriteBool( sec != nil ); err != nil { return }
    if sec != nil { 
        if err = w.mgw.WriteQualifiedTypeName( sec ); err != nil { return } 
    }
    return
}

func ( w *BinWriter ) WriteDefinition( d Definition ) error {
    switch v := d.( type ) {
    case *PrimitiveDefinition: return w.writePrimDef( v )
    case *AliasedTypeDefinition: return w.writeAliasDef( v )
    case *PrototypeDefinition: return w.writeProtoDef( v )
    case *StructDefinition: return w.writeStructDef( v )
    case *EnumDefinition: return w.writeEnumDef( v )
    case *ServiceDefinition: return w.writeServiceDef( v )
    default: return fmt.Errorf( "%T: Unhandled: %T", w, d )
    }
    panic( fmt.Errorf( "%T: Unhandled definition: %T", w, d ) )
}

func ( w *BinWriter ) WriteDefinitionMap( m *DefinitionMap ) ( err error ) {
    if err = w.mgw.WriteTypeCode( tcDefMap ); err != nil { return }
    m.EachDefinition( func ( d Definition ) {
        if err == nil { err = w.WriteDefinition( d ) }
    })
    if err != nil { return }
    return w.mgw.WriteTypeCode( tcNull )
}

type BinReader struct {
    mgr *mg.BinReader
}

func NewBinReader( rd io.Reader ) *BinReader {
    return &BinReader{ mgr: mg.NewReader( rd ) }
}

func ( r *BinReader ) readLen() ( int, error ) {
    i, err := r.mgr.ReadUint32()
    if err != nil { return 0, err }
    return int( i ), nil
}

func ( r *BinReader ) readOptSuperType() ( typ *mg.QualifiedTypeName, 
                                           err error ) {
    var ok bool
    if ok, err = r.mgr.ReadBool(); err != nil { return }
    if ok {
        if typ, err = r.mgr.ReadQualifiedTypeName(); err != nil { return }
    }
    return
}

func ( r *BinReader ) readPrimDef() ( pd *PrimitiveDefinition, err error ) {
    pd = &PrimitiveDefinition{}
    pd.Name, err = r.mgr.ReadQualifiedTypeName()
    return
}

func ( r *BinReader ) readAliasDef() ( ad *AliasedTypeDefinition, err error ) {
    ad = &AliasedTypeDefinition{}
    if ad.Name, err = r.mgr.ReadQualifiedTypeName(); err != nil { return }
    if ad.AliasedType, err = r.mgr.ReadTypeReference(); err != nil { return }
    return
}

func ( r *BinReader ) readField() ( fd *FieldDefinition, err error ) {
    fd = &FieldDefinition{}
    if fd.Name, err = r.mgr.ReadIdentifier(); err != nil { return }
    if fd.Type, err = r.mgr.ReadTypeReference(); err != nil { return }
    var val mg.Value
    if val, err = r.mgr.ReadValue(); err != nil { return }
    if _, isNull := val.( *mg.Null ); ! isNull { fd.Default = val }
    return
}

func ( r *BinReader ) readFields( fs *FieldSet ) ( err error ) {
    var sz int
    if sz, err = r.readLen(); err != nil { return }
    for i := 0; i < sz; i++ {
        var fld *FieldDefinition
        if fld, err = r.readField(); err != nil { return }
        if err = fs.Add( fld ); err != nil { return }
    }
    return
}

func ( r *BinReader ) readCallSig() ( sig *CallSignature, err error ) {
    sig = NewCallSignature()
    if err = r.readFields( sig.Fields ); err != nil { return }
    if sig.Return, err = r.mgr.ReadTypeReference(); err != nil { return }
    var sz int
    if sz, err = r.readLen(); err != nil { return }
    for i := 0; i < sz; i++ {
        var typ mg.TypeReference
        if typ, err = r.mgr.ReadTypeReference(); err == nil {
            sig.Throws = append( sig.Throws, typ )
        } else { return }
    }
    return
} 

func ( r *BinReader ) readProtoDef() ( pd *PrototypeDefinition, err error ) {
    pd = &PrototypeDefinition{}
    if pd.Name, err = r.mgr.ReadQualifiedTypeName(); err != nil { return }
    if pd.Signature, err = r.readCallSig(); err != nil { return }
    return
}

func ( r *BinReader ) readConstructorDef() ( cd *ConstructorDefinition, 
                                             err error ) {
    if _, err = r.mgr.ExpectTypeCode( tcConstructorDef ); err != nil { return }
    cd = &ConstructorDefinition{}
    if cd.Type, err = r.mgr.ReadTypeReference(); err != nil { return }
    return
}

func ( r *BinReader ) readStructDef() ( sd *StructDefinition, err error ) {
    sd = NewStructDefinition()
    if sd.Name, err = r.mgr.ReadQualifiedTypeName(); err != nil { return }
    if sd.SuperType, err = r.readOptSuperType(); err != nil { return }
    if err = r.readFields( sd.Fields ); err != nil { return }
    var sz int
    if sz, err = r.readLen(); err != nil { return }
    for i := 0; i < sz; i++ {
        var cons *ConstructorDefinition
        if cons, err = r.readConstructorDef(); err != nil { return }
        sd.Constructors = append( sd.Constructors, cons )
    }
    return
}

func ( r *BinReader ) readEnumDef() ( ed *EnumDefinition, err error ) {
    ed = &EnumDefinition{}
    if ed.Name, err = r.mgr.ReadQualifiedTypeName(); err != nil { return }
    var sz int
    if sz, err = r.readLen(); err != nil { return }
    ed.Values = make( []*mg.Identifier, sz )
    for i := 0; i < sz; i++ {
        if ed.Values[ i ], err = r.mgr.ReadIdentifier(); err != nil { return }
    }
    return
}

func ( r *BinReader ) readOperationDef() ( 
    od *OperationDefinition, err error ) {
    od = &OperationDefinition{}
    if od.Name, err = r.mgr.ReadIdentifier(); err != nil { return }
    if od.Signature, err = r.readCallSig(); err != nil { return }
    return
}

func ( r *BinReader ) readServiceDef() ( sd *ServiceDefinition, err error ) {
    sd = NewServiceDefinition()
    if sd.Name, err = r.mgr.ReadQualifiedTypeName(); err != nil { return }
    if sd.SuperType, err = r.readOptSuperType(); err != nil { return }
    var sz int
    if sz, err = r.readLen(); err != nil { return }
    for i := 0; i < sz; i++ {
        var od *OperationDefinition
        if od, err = r.readOperationDef(); err != nil { return }
        sd.Operations = append( sd.Operations, od )
    }
    var hasSec bool
    if hasSec, err = r.mgr.ReadBool(); err != nil { return }
    if hasSec {
        if sd.Security, err = r.mgr.ReadQualifiedTypeName(); err != nil { 
            return
        }
    }
    return
}

func ( r *BinReader ) ReadDefinitionMap() ( m *DefinitionMap, err error ) {
    if _, err = r.mgr.ExpectTypeCode( tcDefMap ); err != nil { return }
    m = NewDefinitionMap()
    for err == nil {
        var tc uint8
        var def Definition
        if tc, err = r.mgr.ReadTypeCode(); err != nil { return }
        switch tc {
        case tcNull: return
        case tcPrimDef: def, err = r.readPrimDef()
        case tcAliasDef: def, err = r.readAliasDef()
        case tcProtoDef: def, err = r.readProtoDef()
        case tcStructDef: def, err = r.readStructDef()
        case tcEnumDef: def, err = r.readEnumDef()
        case tcServiceDef: def, err = r.readServiceDef()
        default: 
            err = fmt.Errorf( "%T: Unrecognized map type code: 0x%02x", r, tc )
        }
        if err == nil { err = m.Add( def ) }
    }
    return 
}
