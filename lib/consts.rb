SECRETS = File.expand_path("#{File.dirname(__FILE__)}/../secrets")
TOKEN = File.open("#{SECRETS}/bot_token").read.chomp
CLIENT_ID = File.open("#{SECRETS}/client_id").read.chomp
CLIENT_SECRET = File.open("#{SECRETS}/client_secret").read.chomp
OSU_KEY = File.open("#{SECRETS}/osu_key").read.chomp
SEARCH_KEY = File.open("#{SECRETS}/search_key").read.chomp
OSU_URL = 'https://osu.ppy.sh/api'
SEARCH_URL = "https://osusearch.com/api/search?key=#{SEARCH_KEY}"

INTERVAL = 60 * 30  # Seconds between beatmap scans.
CHAR_LIMIT = 2000
TOTAL_SUB_LIMIT = 300
CMD_LIMIT = 20

TEST = ARGV.include?('TEST')
DB_NAME = TEST ? 'test' : 'o_subscribe'
CHANNEL = TEST ? 'testing' : 'subscriptions'
DB = PG.connect(dbname: DB_NAME)
DB.prepare('insert_mapper', 'INSERT INTO mappers(mapper_id, mapper_name) VALUES ($1, $2)')
DB.prepare('insert_user', 'INSERT INTO users(user_disc, user_id, user_name) VALUES ($1, $2, $3)')

SUB_MSG = 'Subscribe to mappers. `!sub username1, username2, :userid1, :userid2`'
UNSUB_MSG = 'Unsubscribe from mappers. `!unsub username1, username2, :userid1, :userid2`'
LIST_MSG = 'List your subscriptions.'
PURGE_MSG = 'Unsubscribe from all mappers.'
COUNT_MSG = 'Get mappers\' subscriber counts. `!count username1, username2`'
