require 'discordrb'
require 'httparty'
require 'pg'
require 'set'

require_relative 'consts'

# Get usernames and user ids from a list of usernames.
# mappers: List of strings representing usernames.
# Returns: (ordered list of usernames, ordered list of user ids)
def get_mappers(usernames)
  names, ids = [], []
  usernames.each do |username|
    begin
      mapper = request_mapper(username)
    rescue
      next
    else
      names.push(mapper['username']) && ids.push(mapper['user_id'])
    end
  end
  if names.length == ids.length
    return names, ids
  else
    return [], []
  end
end

# Get a mapper from the osu! API.
def request_mapper(username)
  url = "#{OSU_URL}/get_user?k=#{OSU_KEY}&u=#{username}&type=string"
  response = HTTParty.get(url).parsed_response[0]
  response.empty? && raise
  return response
end

# Get a list of beatmaps by a user.
# user_id: Mapper's user id.
# Returns a set of mapset ids.
def request_beatmaps(user_id)
  url = "#{OSU_URL}/get_beatmaps?k=#{OSU_KEY}&u=#{user_id}&type=id"
  response = HTTParty.get(url).parsed_response
  response.empty? && raise
  return Set.new(response.map {|b| b['beatmapset_id']})
end

# Add rows to the subscriptions table.
# user_id: User id.
# mapper_ids: Mapper ids.
def subscribe!(user_id, mapper_ids)
  mapper_ids.each do |mapper_id|
    DB.exec("SELECT * FROM maps WHERE mapper_id = #{mapper_id}") do |result|
      result.to_a.empty? && init_mappper(mapper_id)
    end
    begin
      DB.exec("INSERT INTO subscriptions(user_id, mapper_id) VALUES (#{user_id}, #{mapper_id})")
    rescue
      # Assume that it's a duplicate key.
    end
  end
end

# Remove rows from the subscriptions table.
# user_id: User id.
# mapper_ids: Mapper ids.
def unsubscribe!(user_id, mapper_ids)
  mapper_ids.each do |mapper_id|
    begin
      DB.exec("DELETE FROM subscriptions WHERE user_id = #{user_id} AND mapper_id = #{mapper_id}")
    rescue
      # Assume that it's a duplicate key.
    end
  end
end

# Add a given mapper's maps to the database table.
# mapper_id: Mapper's user id.
def init_mappper(mapper_id)
  begin
    mapset_ids = request_beatmaps(mapper_id)
  rescue  # Most probable case for this is the mapper has no maps.
    return
  end
  # We can afford to do this all in one query because we know we won't have
  # any duplicate keys.
  cmd = 'INSERT INTO maps(mapper_id, mapset_id) VALUES '
  mapset_ids.each {|mapset_id| cmd += "(#{mapper_id}, #{mapset_id}), "}
  DB.exec(cmd[0...cmd.rindex(',')])
end

# Get any new maps by a given user.
def diff_mapper(mapper_id)
  existing_mapsets = DB.exec("SELECT mapset_id FROM maps WHERE mapper_id = #{mapper_id}")
  existing_mapsets.map! {|m| m.values[0]}
  new_mapsets = []

  begin
    # Todo: use the `since` API parameter to speed this up.
    mapset_ids = request_beatmaps(mapper_id)
  rescue  # Most probable case for this is the mapper has no maps.
    return
  end

  mapset_ids.each do |mapset_id|
    if !existing_mapsets.include?(mapset_id)
      new_mapsets.push(mapset_id)
      DB.exec("INSERT INTO maps(mapper_id, mapset_id) VALUES (#{mapper_id}, #{mapset_id})")
    end
  end

  return new_mapsets
end

# Return true if a string is empty or comprised of all spaces.
def empty?(s) s.empty? || s.each_char.all? {|c| c == ' '} end

def setup
  bot = Discordrb::Bot.new(token: TOKEN, client_id: CLIENT_ID)

  bot.message(start_with: '!subscribe', in: CHANNEL) do |msg|
    usernames, user_ids = get_mappers(Set.new(msg.text.split('"').reject {|s| empty?(s)}[1..-1]))
    if !user_ids.empty?
      subscribe!(msg.user.discriminator, user_ids)
      msg.respond("#{msg.user.mention} has subcribed to: #{usernames.join(', ')}")
    else
      msg.respond("Sorry #{msg.user.mention}, something went wrong.")
    end
  end

  bot.message(start_with: '!unsubscribe', in: CHANNEL) do |msg|
    usernames, user_ids = get_mappers(Set.new(msg.text.split('"').reject {|s| empty?(s)}[1..-1]))
    if !user_ids.empty?
      unsubscribe!(msg.user.discriminator, user_ids)
      msg.respond("#{msg.user.mention} has unsubcribed from: #{usernames.join(', ')}")
    else
      msg.respond("Sorry #{msg.user.mention}, something went wrong.")
    end
  end

  return bot
end

if __FILE__ == $0
  bot = setup
  DB = PG.connect(dbname: DB_NAME)
  bot.run(:async)

  loop do
    # Todo: Check each mapper for new maps.
    sleep 3600  # Todo: Need to evaluate how much CPU power this work uses.
  end

end
