@version: v1;

namespace bitgirder:io
{
    enum DataUnit { byte, kilobyte, megabyte, gigabyte, terabyte; }

    struct DataSize
    {
        unit: DataUnit;
        size: Int64~[0,);

        @constructor( String );
        @constructor( Int64~[0,) );
    }

    alias PathString: String;
    alias FilePath: PathString;
    alias DirPath: PathString;

    struct BasicFileInfo
    {
        size: DataSize;
        mtime: Timestamp;
    }

    exception IoException {}

    exception AbstractPathException < IoException
    {
        path: PathString;
    }

    exception NoSuchPathException < AbstractPathException {}
    exception PathPermissionException < AbstractPathException {}

    struct CompressionInfo
    {
        type: String;
        length: DataSize;
    }

    struct DigestInfo
    {
        digest: Buffer;
        name: String;
    }

    struct FileAttributes
    {
        dataLength: DataSize?;
        dataDigest: DigestInfo?;
        fileSize: DataSize;
        fileDigest: DigestInfo?;
        compression: CompressionInfo?;
        # cipher to come
    }
}
