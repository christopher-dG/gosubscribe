SECRETS = File.expand_path("#{File.dirname(__FILE__)}/../secrets")
TOKEN = File.open("#{SECRETS}/bot_token").read.chomp
CLIENT_ID = File.open("#{SECRETS}/client_id").read.chomp
CLIENT_SECRET = File.open("#{SECRETS}/client_secret").read.chomp
OSU_KEY = File.open("#{SECRETS}/osu_key").read.chomp
OSU_URL = 'https://osu.ppy.sh/api'
TEST = ARGV.include?('TEST') || __FILE__ != $0
DB_NAME = TEST ? 'test' : 'o_subscribe'
CHANNEL = TEST ? 'testing' : 'subscriptions'
DB = PG.connect(dbname: DB_NAME)
DB.prepare('insert_mapper', 'INSERT INTO mappers(mapper_id, mapper_name) VALUES ($1, $2)')
DB.prepare('insert_user', 'INSERT INTO users(user_disc, user_id, user_name) VALUES ($1, $2, $3)')
