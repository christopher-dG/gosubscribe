config = YAML.load_file(File.expand_path("#{File.dirname(__FILE__)}/../config.yml"))

TOKEN = config['bot_token']
CLIENT_ID = config['client_id']
OSU_KEY = config['osu_key']
SEARCH_KEY = config['search_key']

OSU_URL = 'https://osu.ppy.sh/api'
SEARCH_URL = "https://osusearch.com/api/search?key=#{SEARCH_KEY}"

CHAR_LIMIT = 2000  # Discord message character limit.
TOTAL_SUB_LIMIT = 300  # Maximum subscriptions.
CMD_LIMIT = 20  # Maximum arguments per command.
DEFAULT_TOP = 3  # Default number of top mappers to display.
TOP_MAX = 15  # Maximum number of top mappers to display.

TEST = ARGV.include?('TEST')
DB_NAME = TEST ? 'test' : 'o_subscribe'
CHANNEL = TEST ? 'testing' : 'subscriptions'

DB = PG.connect(dbname: DB_NAME)
DB.prepare('insert_mapper', 'INSERT INTO mappers(mapper_id, mapper_name) VALUES ($1, $2)')
DB.prepare('insert_user', 'INSERT INTO users(user_disc, user_id, user_name) VALUES ($1, $2, $3)')
