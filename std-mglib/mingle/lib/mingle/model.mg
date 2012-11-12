@version: v1;

namespace mingle:model
{
    alias MingleIdentifierPart: String~"[a-z][a-z\\d]*";

    struct MingleIdentifier { parts: MingleIdentifierPart+; }

    struct MingleNamespace
    {
        identifiers: MingleIdentifier+;
        version: MingleIdentifier;
    }

    alias MingleTypeNamePart: String~"[A-Z][a-z\\d]*";

    struct MingleTypeName { parts: MingleTypeNamePart+; }

    struct QualifiedTypeName
    {
        namespace: MingleNamespace;
        typeNames: MingleTypeName*;
    }

    struct MingleTypeReference {}

    struct MingleIdentifiedName
    {
        namespace: MingleNamespace;
        names: MingleIdentifier+;
    }
}
