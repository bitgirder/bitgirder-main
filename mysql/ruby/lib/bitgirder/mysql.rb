require 'bitgirder/core'

module BitGirder
module MySql

    # Although the 3p Mysql module won't conflict with this module (cases differ
    # on the 'S' in 'sql') we nevertheless reference the external module via
    # this local instance variable to help avoid typo-related errors
    require 'mysql'
    @@mysql = Mysql # The 3p library we're working on top of

    extend BitGirder::Core::BitGirderMethods

    def self.connect_from_hash( h )

        flattened = h.values_at( :host, :user, :password, :db, :port, :socket )
        flattened << ( h[ :flag ] || 0 )

        @@mysql.connect( *flattened )
    end 
    
    def self.connect( *argv )
        
        case argv.size
            
            when 0 then raise "Need connect args"

            when 1
                case argv[ 0 ]
                    when Hash then connect_from_hash( argv[ 0 ] )
                    else @@mysql.connect( *argv ) # assume passthrough args
                end

            else raise ArgumentError, "Unexpected argv: #{argv}"
        end
    end

    def self.open( *argv )
 
        mysql = connect( *argv )

        begin
            yield( mysql )
        ensure
            mysql.close
        end
    end

    def self.flush_privileges( db )
        
        not_nil( db, :db )
        db.query( "flush privileges" )
    end

    def self.quote( str )
        @@mysql.quote( str )
    end

end
end
