require 'bitgirder/core'
require 'bitgirder/io'

module BitGirder
module Ops
module Ruby

module RubyEnvVarNames
    ENV_GEM_HOME = "GEM_HOME"
    ENV_RUBYLIB = "RUBYLIB"
    ENV_PATH = "PATH"
end

module RubyIncludes

    extend BitGirder::Core::BitGirderMethods

    extend BitGirder::Io
        
    # clean up adjacent slashes and other things that can be unsightly and in
    # some cases cause more serious problems
    def self.clean_include_dir_names( dirs )
        
        dirs.map { |dir| dir.gsub( %r{/+}, "/" ) }.
             map { |dir| dir.sub( %r{/$}, "" ) }
    end
 
    def self.get_single_dir( dir, pat )

        res = 
            Dir.chdir( dir ) do |dir|
                Dir.glob( "*" ).
                    select { |f| File.directory?( f ) }.
                    grep( pat )
            end
 
        case res.size

            when 0 then nil
            when 1 then res[ 0 ]
            else raise "Multiple directories in #{dir} match #{pat}"
        end
    end
 
    def self.expect_single_dir( dir, pat )

        self.get_single_dir( dir, pat ) or
            raise "Couldn't find entry in #{dir} matching #{pat}"
    end

    def self.infer_version_str( lib_dir )
        self.expect_single_dir( lib_dir, /^\d+(?:\.\d+)*$/ )
    end

    def self.get_arch_dir( dir )

        self.get_single_dir( 
            dir, /^(?:i\d86|x86_64).*(?:darwin|linux|solaris)/ )
    end

    def self.get_ruby_include_dirs( opts )

        ruby_home = has_key( opts, :ruby_home )
        lib_dir = file_exists( "#{ruby_home}/lib/ruby" )

        ver = opts[ :version ] || self.infer_version_str( lib_dir )

        res = [ "", "site_ruby", "vendor_ruby" ].inject( [] ) do |res, dir|
            
            dir = file_exists( "#{lib_dir}/#{dir}/#{ver}" )
            res << dir

            if arch_dir = self.get_arch_dir( dir )
                res << "#{dir}/#{arch_dir}"
            end

            res
        end

        self.clean_include_dir_names( res )
    end

    def self.as_include_argv( incl_dirs )
        incl_dirs.map { |dir| [ "-I", dir ] }.flatten
    end
end

class RubyContext < BitGirderClass

    require 'set'

    bg_attr :id
    bg_attr :gem_home, validation: :file_exists
    bg_attr :ruby_home, validation: :file_exists
    bg_attr :ruby_flavor, required: false
    bg_attr :env, default: {}

    def rcmd( cmd, alts = true )
        
        not_nil( cmd, :cmd )

        cmds = [ cmd ]
        cmds << "j#{cmd}" if alts

        bin_dirs = [ @gem_home, @ruby_home ].map { |dir| "#{dir}/bin" }
        files = bin_dirs.map { |bd| cmds.map { |cmd| "#{bd}/#{cmd}" } }.flatten
        
        files.find { |f| File.exist?( f ) } or
            raise "No #{cmd} in #{bin_dirs.join( " or " )}"
    end

    def prepend_path_ruby_home
        "#@gem_home/bin:#@ruby_home/bin:#{ENV[ ENV_PATH ]}"
    end

    def get_ruby_include_dirs

        if @ruby_flavor == "jruby"
            []
        else
            RubyIncludes.get_ruby_include_dirs( ruby_home: @ruby_home )
        end
    end

    def get_env_var_rubylib
        
        res = get_ruby_include_dirs.join( ":" )

        if prev = ENV[ ENV_RUBYLIB ]
            res << ":#{prev}"
        end

        res
    end

    def proc_builder_opts( cmd = nil )
        
        opts = { opts: {}, argv: [] }

        opts[ :env ] = @env.merge( 
            ENV_GEM_HOME => @gem_home,
            ENV_PATH => prepend_path_ruby_home,
        )

        opts[ :cmd ] = rcmd( cmd ) if cmd

        opts
    end

    def self.read_env_in( s )
        
        ( s[ :env ] || [] ).inject( {} ) do |env, pair|
            
            nm, val = pair[ 0 ], pair[ 1 ]

            if env.key?( nm )
                raise "Multiple definitions of env var #{nm}"
            else
                env[ nm ] = val
            end

            env
        end
    end

    def self.from_mingle_struct( s )
        
        self.new(
            id: s.expect_string( :id ),
            ruby_home: s.expect_string( :ruby_home ),
            gem_home: s.expect_string( :gem_home ),
            ruby_flavor: s.get_string( :ruby_flavor ),
            env: read_env_in( s )
        )
    end
end

class RubyEnv < BitGirderClass

    bg_attr :rubies
    bg_attr :default_id, required: false # name of default ruby

    def initialize( *argv )
        
        super( *argv )

        if @default_id
            unless @rubies.key?( @default_id )
                raise "Default id #@default_id not a defined ruby"
            end
        end
    end

    def get_context( id = nil )
        
        if id 
            @rubies[ id ] or raise "No ruby context with id: #{id}"
        else
            if @default_id
                @rubies[ @default_id ] or 
                    raise "Default context #@default_id not set (!?)"
            else
                raise "No default ruby context set"
            end
        end
    end

    def self.read_rubies_in( s )
        
        ( s[ :rubies ] || [] ).inject( {} ) do |res, mg_ctx|
            
            ctx = RubyContext.from_mingle_struct( mg_ctx )

            if res.key?( ctx.id )
                raise "Env has multiple ruby contexts with id: #{ctx.id}"
            else
                res[ ctx.id ] = ctx
            end

            res
        end
    end

    def self.from_mingle_struct( s )
        
        self.new( 
            rubies: self.read_rubies_in( s ),
            default_id: s.get_string( :default_ruby ),
        )
    end
end

end
end
end
