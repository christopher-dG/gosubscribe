config = YAML.load_file(File.expand_path("#{File.dirname(__FILE__)}/../config.yml"))
db_user = config['db_user']
db_pass = config['db_pass']

TOKEN = config['bot_token']
CLIENT_ID = config['client_id']
OSU_KEY = config['osu_key']
SEARCH_KEY = config['search_key']

OSU_URL = 'https://osu.ppy.sh/api'
SEARCH_URL = "https://osusearch.com/api/search?key=#{SEARCH_KEY}&count=500"

CHAR_LIMIT = 2000  # Discord message character limit.
TOTAL_SUB_LIMIT = 300  # Maximum subscriptions.
CMD_LIMIT = 20  # Maximum arguments per command.
DEFAULT_TOP = 3  # Default number of top mappers to display.
TOP_MAX = 15  # Maximum number of top mappers to display.

TEST = ARGV.include?('TEST') || $0 == 'irb'
DB_NAME = TEST ? 'test' : 'o_subscribe'

DB = Sequel.postgres(DB_NAME, :host => 'localhost', :user => db_user, :password => db_pass)
DB[:mappers].prepare(:insert, :insert_mapper, :mapper_name => :$name, :mapper_id => :$id)
DB[:users].prepare(:insert, :insert_user, :user_id => :$id, :user_disc => :$disc, :user_name => :$name)
DB[:maps].prepare(:insert, :insert_map, :mapper_id => :$mapper, :mapset_id => :$map, :status => :$status)
DB[:subscriptions].prepare(:insert, :subscribe, :user_id => :$user, :mapper_id => :$mapper)
