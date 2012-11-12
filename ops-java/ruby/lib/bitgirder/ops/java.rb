require 'bitgirder/core'
require 'bitgirder/io'

module BitGirder
module Ops
module Java

include BitGirder::Core

module JavaEnvironments

    include BitGirder::Core::BitGirderMethods

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
            raise "#{ENV_JAVA_HOME} is not set"
        end
    end

    module_function :get_java_home
end

class JavaEnvironment < BitGirderClass

    include BitGirder::Io
    
    bg_attr :java_home, validation: :file_exists, required: false

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
        self.new( java_home: JavaEnvironments.get_java_home )
    end
end

class JavaRunner < BitGirderClass

    KEY_APP_RUNNER = :app_runner
    DEFAULT_APP_RUNNER = "com.bitgirder.application.ApplicationRunner"

    MINGLE_APP_RUNNER =
        "com.bitgirder.mingle.application.MingleApplicationRunner"

    bg_attr :java_env
    bg_attr :command
    bg_attr :classpath, default: proc { [] }
    bg_attr :jvm_args, default: proc { [] }
    bg_attr :sys_props, default: proc { {} }
    bg_attr :argv, :validation => :not_empty
    bg_attr :proc_env, default: proc { {} }
    bg_attr :proc_opts, default: proc { {} }

    private
    def create_jv_argv
        
        res = []
        
        unless @classpath.empty?
            res << "-classpath" << @java_env.as_classpath( @classpath )
        end

        res += @jvm_args
        @sys_props.each_pair { |k, v| res << "-D#{k}=#{v}" }
        res += argv
    end

    public
    def process_builder

        BitGirder::Io::UnixProcessBuilder.new(
            cmd: @java_env.jcmd( @command ),
            argv: create_jv_argv,
            env: @proc_env,
            opts: @proc_opts
        )
    end

    def self.create_application_runner( opts )

        not_nil( opts, :opts )

        argv = [ opts[ KEY_APP_RUNNER ] || DEFAULT_APP_RUNNER ]
        argv << has_key( opts, :app_class )
        argv += ( opts[ :argv ] || [] )

        JavaRunner.new( opts.merge( command: "java", argv: argv ) )
    end

    def self.create_mingle_app_runner( opts )
        
        not_nil( opts, :opts )

        opts = opts.merge( KEY_APP_RUNNER => MINGLE_APP_RUNNER )
        self.create_application_runner( opts )
    end
end

end
end
end
