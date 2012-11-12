@version: v1;

import bitgirder:io/DataUnit;
import bitgirder:io/DataSize;

import bitgirder:concurrent/Duration;

import mingle:model/MingleIdentifier;
import mingle:model/MingleNamespace;
import mingle:model/MingleTypeReference;
import mingle:model/MingleIdentifiedName;
import mingle:model/QualifiedTypeName;

namespace bitgirder:mglib
{
    # Just to check that types we natively build bindings for still correctly
    # are used in codegen
    struct NativeGenHolder
    {
        dataUnit: DataUnit?;
        dataSize: DataSize?;
        duration: Duration?;
        mgIdent: MingleIdentifier?;
        mgNs: MingleNamespace?;
        mgQname: QualifiedTypeName?;
        mgTypeRef: MingleTypeReference?;
        mgIdentifiedName: MingleIdentifiedName?;
    }
}
