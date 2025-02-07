
#!/usr/bin/env ruby

require 'yajl'
require 'json'


#example pos file where issue was reported - FILE = "/var/lib/fluentd/pos/journal_pos.json"

ARGV.each do |filename|

 input = File.read(filename)

 puts "checking if #{filename} a valid json by calling yajl parser"

 @default_options ||= {:symbolize_keys => false}

 begin
   Yajl::Parser.parse(input, @default_options )
 rescue Yajl::ParseError => e
   raise e.message
 end

end

