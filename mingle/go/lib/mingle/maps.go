package mingle

import (
    "fmt"
)

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

func ( m *IdentifierMap ) GetKeys() []*Identifier {
    res := make( []*Identifier, 0, m.Len() )
    m.EachPair( func( k *Identifier, _ interface{} ) {
        res = append( res, k )
    })
    return res
}

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

func ( m *NamespaceMap ) EachPair( f func( ns *Namespace, val interface{} ) ) {
    m.implEachPair( 
        func( k mapImplKey, v interface{} ) { f( k.( *Namespace ), v ) },
    )
}

func ( m *NamespaceMap ) EachPairError( 
    f func( ns *Namespace, val interface{} ) error ) error {

    return m.implEachPairError(
        func( k mapImplKey, val interface{} ) error {
            return f( k.( *Namespace ), val )
        },
    )
}
