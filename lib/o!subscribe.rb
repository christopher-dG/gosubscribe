require 'discordrb'
require 'httparty'

require_relative 'consts'

# Get usernames and user ids from a list of usernames.
# mappers: List of strings representing usernames.
# Returns: (ordered list of usernames, ordered list of user ids)
def get_mappers(mappers)
  names, ids = [], []
  mappers.each do |mapper|
    url = "#{OSU_URL}/get_user?k=#{OSU_KEY}&u=#{mapper}&type=string"
    response = HTTParty.get(url).parsed_response[0]
    response.empty? && next  # Todo: Log this error.
    names.push(response['username']) && ids.push(response['user_id'])
  end
  return names, ids
end

# For each mapper's database table, create an entry for the user.
# user: User id.
# ids: Mapper ids.
def subscribe(user, ids)
  # Todo
end

# Create a database table for the mapper.
def init_mappper(mapper_id)
  # Todo
end

# Return true if a string is empty or comprised of all spaces.
def empty?(s) s.empty? || s.each_char.all? {|c| c == ' '} end

def setup
  bot = Discordrb::Bot.new(token: TOKEN, client_id: CLIENT_ID)
  channel = TEST ? '#testing' : '#subscribe'
  bot.message(start_with: '!subscribe', in: channel) do |event|
    names, ids = get_mappers(event.text.split('"').reject {|s| empty?(s)}[1..-1])
    subscribe(event.user.discriminator, ids)
    event.respond("#{event.user.mention} has subcribed to: #{names.join(', ')}")
  end
  return bot
end

if __FILE__ == $0
  bot = setup
  bot.run(:async)

  loop do
    # Todo: Check each mapper for new maps.
  end

end
