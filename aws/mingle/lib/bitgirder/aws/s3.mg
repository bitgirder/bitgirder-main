@version: v1;

import bitgirder:io/FilePath;
import bitgirder:io/IoException;

namespace bitgirder:aws:s3
{
    alias S3Bucket: String;
    alias S3KeyPath: String;
    alias Md5: Buffer;

    alias ContentType: String;

    struct S3Location { useSsl: Boolean, default: true; }
    struct S3BucketLocation < S3Location { bucket: S3Bucket; }
    struct S3ObjectLocation < S3BucketLocation { key: S3KeyPath; }

    struct S3AmazonRequestIds
    {
        id2: String?;
        requestId: String?;
    }

    exception S3RemoteException { amazonRequestIds: S3AmazonRequestIds?; }

    exception GenericS3RemoteException < S3RemoteException
    {
        code: String?;
        message: String?;
        errorXml: Buffer?;
    }

    exception NoSuchS3ObjectException < S3RemoteException
    {
        bucket: S3Bucket;
        key: S3KeyPath;
    }

    struct S3ObjectMetaData
    {
        key: String;
        value: String;
    }

    struct S3ResponseInfo
    {
        amazonId2: String?;
        amazonRequestId: String?;
        meta: S3ObjectMetaData*;
    }

    struct S3BucketResponseInfo < S3ResponseInfo { bucket: S3Bucket; }
    struct S3ObjectResponseInfo < S3BucketResponseInfo { key: S3KeyPath; }
    struct S3PutObjectResponseInfo < S3ObjectResponseInfo { etag: String; }
    struct S3GetObjectResponseInfo < S3ObjectResponseInfo {}
    struct S3HeadObjectResponseInfo < S3ObjectResponseInfo {}
    struct S3DeleteObjectResponseInfo < S3ObjectResponseInfo {}

    struct FileUploadInfo
    {
        s3Response: S3PutObjectResponseInfo;
        location: S3ObjectLocation;
        file: FilePath;
    }

    struct FileDownloadInfo
    {
        s3Response: S3GetObjectResponseInfo;
        location: S3ObjectLocation;

        # Current impls always set this but typed to allow for it to be
        # selectively disabled on a per-download basis down the road
        md5: Md5?; 

        file: FilePath;
    }

    service S3FileService
    {
        uploadFile( path: FilePath;
                    md5: Md5?;
                    contentType: ContentType?;
                    meta: S3ObjectMetaData*;
                    location: S3ObjectLocation ): FileUploadInfo
            throws IoException,
                   S3RemoteException;
 
        downloadFile( path: FilePath;
                      location: S3ObjectLocation ): FileDownloadInfo
            throws IoException,
                   S3RemoteException;
    }
}
