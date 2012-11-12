@version: v1;

namespace bitgirder:mingle:http
{
    struct HttpTestServerConfig {}

    alias ServerName: String;
    alias Host: String;
    alias Port: Int32~(0,);
    alias Uri: String;

    struct HttpTestServerInfo
    {
        name: ServerName;
        host: Host;
        port: Port;
        uri: Uri;
        isSsl: Boolean;
    }

    struct HttpTestServersInfo
    {
        servers: HttpTestServerInfo*;
    }
}
