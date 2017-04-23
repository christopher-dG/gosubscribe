SECRETS = File.expand_path("#{File.dirname(__FILE__)}/../secrets")
TOKEN = File.open("#{SECRETS}/bot_token").read.chomp
CLIENT_ID = File.open("#{SECRETS}/client_id").read.chomp
CLIENT_SECRET = File.open("#{SECRETS}/client_secret").read.chomp
OSU_KEY = File.open("#{SECRETS}/osu_key").read.chomp
OSU_URL = 'https://osu.ppy.sh/api'
DB = 'subscriptions'
TEST = ARGV.include?('TEST')
