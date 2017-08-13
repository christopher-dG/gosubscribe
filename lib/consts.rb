config = YAML.load_file("#{ENV['APP']}/config.yml")

TOKEN = config['bot_token']
CLIENT_ID = config['client_id']
OSU_KEY = config['osu_key']
SEARCH_KEY = config['search_key']

OSU_URL = 'https://osu.ppy.sh'
SEARCH_URL = "https://osusearch.com/api/search?key=#{SEARCH_KEY}&count=500"
SERVER_INVITE = "https://discord.gg/qaUhTKJ"

CHAR_LIMIT = 2000  # Discord message character limit.
TOTAL_SUB_LIMIT = 300  # Maximum subscriptions.
CMD_LIMIT = 20  # Maximum arguments per command.
DEFAULT_TOP = 3  # Default number of top mappers to display.
TOP_MAX = 15  # Maximum number of top mappers to display.

DB_NAME = 'o_subscribe'
conn = "postgres://#{config['db_user']}:#{config['db_pass']}@#{config['db_endpoint']}/#{DB_NAME}"
DB = Sequel.connect(conn)
DB[:mappers].prepare(:insert, :insert_mapper, :mapper_name => :$name, :mapper_id => :$id)
DB[:users].prepare(:insert, :insert_user, :user_id => :$id, :user_disc => :$disc, :user_name => :$name)
DB[:maps].prepare(:insert, :insert_map, :mapper_id => :$mapper, :mapset_id => :$map, :status => :$status)
DB[:subscriptions].prepare(:insert, :subscribe, :user_id => :$user, :mapper_id => :$mapper)
