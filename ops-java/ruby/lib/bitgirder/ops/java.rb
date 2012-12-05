require 'bitgirder/core'
require 'bitgirder/io'

module BitGirder
module Ops
module Java

include BitGirder::Core

module JavaEnvironments

    extend BitGirder::Core::BitGirderMethods

    ENV_JAVA_HOME = "JAVA_HOME"

    def get_java_home

        if res = ENV[ ENV_JAVA_HOME ]
            if File.exist?( res )
                res
            else
                raise 
                    "Location specified in #{ENV_JAVA_HOME} doesn't exist: " +
                    res
            end
        else
            if jv = Io.which( "java" )
                File.dirname( File.dirname( jv ) )
            else
                raise "#{ENV_JAVA_HOME} is not set and no 'java' found on path"
            end
        end
    end

    module_function :get_java_home
end

class JavaEnvironment < BitGirderClass

    include BitGirder::Io
    
    bg_attr :java_home, :validation => :file_exists

    public
    def jcmd( cmd )
        
        not_nil( cmd, :cmd )
        file_exists( "#@java_home/bin/#{cmd}" )
    end

    public
    def as_classpath( val )
        
        if val.respond_to?( :join )
            val.join( ":" )
        else
            val.to_s
        end
    end

    def self.get_default
        self.new( :java_home => JavaEnvironments.get_java_home )
    end
end

class JavaRunner < BitGirderClass

    bg_attr :java_env
    bg_attr :command
    bg_attr :classpath, :default => proc { [] }
    bg_attr :jvm_args, :default => proc { [] }
    bg_attr :sys_props, :default => proc { {} }
    bg_attr :argv, :validation => :not_empty
    bg_attr :proc_env, :default => proc { {} }
    bg_attr :proc_opts, :default => proc { {} }

    private
    def create_jv_argv
        
        res = []
 
        unless @classpath.empty?
            res << "-classpath" << @java_env.as_classpath( @classpath )
        end

        res += @jvm_args
        @sys_props.each_pair { |k, v| res << "-D#{k}=#{v}" }
        res += argv

        res.map { |v| v.to_s }
    end

    public
    def process_builder

        BitGirder::Io::UnixProcessBuilder.new(
            :cmd => @java_env.jcmd( @command ),
            :argv => create_jv_argv,
            :env => @proc_env,
            :opts => @proc_opts
        )
    end
 
    def self.create_application_runner( opts )

        not_nil( opts, :opts )

        argv = [ has_key( opts, :main ) ]
        argv += ( opts[ :argv ] || [] )

        JavaRunner.new( opts.merge( :command => "java", :argv => argv ) )
    end

    def self.split_argv( argv )
        
        res = { argv: [], sys_props: {}, jvm_args: [] }

        argv.each do |arg|
            case arg
            when /^-X/ then res[ :jvm_args ] << arg
            when /^-D(?:([^=]+)(?:=(.*))?)?$/
                raise "Property without name" unless $1
                res[ :sys_props ][ $1 ] = $2 || "" 
            else res[ :argv ] << arg
            end
        end

        res
    end

    [ :exec, :system ].each do |meth| 
        define_method( meth ) do
            process_builder.send( meth )
        end
    end
end

end
end
end
