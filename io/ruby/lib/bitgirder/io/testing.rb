require 'bitgirder/core'
require 'stringio'

module BitGirder
module Io

module Testing
    
    [ :force_encoding, :encode, :encode! ].each do |meth|
        define_method( :"opt_#{meth}" ) do |s, *argv|
            RubyVersions.when_19x( s ) { |s| s.send( meth, *argv ) }
        end
    end

    def new_string_io( str = "" )
        RubyVersions.when_19x( StringIO.new( str ) ) do |io|
            io.set_encoding( "binary" )
        end
    end
    
    module_function :new_string_io

    def rand_buf( len )

        len = len.bytes if len.is_a?( DataSize )
        File.open( "/dev/random", "rb" ) { |io| Io.read_full( io, len ) }
    end

    module_function :rand_buf
end

end
end
