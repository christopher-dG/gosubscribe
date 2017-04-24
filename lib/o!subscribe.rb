require 'discordrb'
require 'httparty'
require 'pg'
require 'set'

require_relative 'consts'
require_relative 'user'
require_relative 'mapper'

# Return true if a string is empty or comprised of all spaces.
def empty?(s) s.empty? || s.each_char.all? {|c| c == ' '} end

def setup
  bot = Discordrb::Commands::CommandBot.new(token: TOKEN, client_id: CLIENT_ID, prefix: '!', channels: [CHANNEL])

  bot.command(
    :subscribe,
    description: 'Subscribe to a user: !subscribe "user1" "user2"'
  ) do |event|
    if event.text.split.length == 1
      error = true
    else
      string = event.text[event.text.index(' ') + 1..-1]
      string.empty? && error = true
    end
    if !error
      user = User.new(event.user)
      error = user.error
    end
    if !error
      tokens = Set.new(string.split('"').reject {|s| empty?(s)})
      mappers = tokens.map {|t| Mapper.new(username: t)}

      if !mappers.reject {|m| m.error}.empty?
        user.subscribe!(mappers)
        usernames = mappers.map {|m| m.username}
        event.respond("#{event.user.mention} has subscribed to: #{usernames.join(', ')}")
      else
        error = true
      end
    else
      event.respond("Sorry #{event.user.mention}, something went wrong.")
    end
  end

  bot.command(
    :unsubscribe,
    description: 'Unsubscribe from a user: !unsubscribe "user1" "user2"'
  ) do |event|
    if event.text.split.length == 1
      error = true
    else
      string = event.text[event.text.index(' ') + 1..-1]
      string.empty? && error = true
    end
    if !error
      user = User.new(event.user)
      error = user.error
    end
    tokens = Set.new(string.split('"').reject {|s| empty?(s)})
    mappers = tokens.map {|t| Mapper.new(username: t)}

    if !mappers.reject {|m| m.error}.empty?
      user.unsubscribe!(mappers)
      usernames = mappers.map {|m| m.username}
      event.respond("#{event.user.mention} has unsubscribed from: #{usernames.join(', ')}")
    else
      event.respond("Sorry #{event.user.mention}, something went wrong.")
    end
  end

  bot.command(:purge, description: 'Unsubscribe from all users') do |event|
    user = User.new(event.user)
    user.purge
    event.respond("#{event.user.mention} is no longer subscribed to any mappers.")
  end

  bot.command(:list, description: 'List your subscriptions') do |event|
    user = User.new(event.user)
    subscriptions = user.list
    if subscriptions.empty?
      event.respond("#{event.user.mention} is not subscribed any mappers")
    else
      event.respond("#{event.user.mention} is subscribed to: #{subscriptions.join(', ')}")
    end
  end

  return bot
end


if __FILE__ == $0
  BOT = setup
  BOT.run(:async)

  loop do
    # Todo: Check each mapper for new maps.
    sleep 3600  # Todo: Need to evaluate how much CPU power this work takes.
  end

end
