# coding: utf-8
require_relative 'o!subscribe'

STATUS_MAP = {
  -2 => 'graveyard',
  -1 => 'WIP',
  0 => 'pending',
  1 => 'ranked',
  2 => 'approved',
  3 => 'qualified',
  4 => 'loved',
}

# Send messages about a new map to all subscribers of a given mapper.
# map: Map hash from the osu! API.
# mapper: Map creator.
def notify(map, mapper)
  map_str = "#{map['artist']} - #{map['title']}"
  is_new = DB[:maps].where(:mapper_id => mapper.id, :mapset_id => map['beatmapset_id']).empty?
  if !is_new
    old_status = STATUS_MAP[DB[:maps].where(:mapper_id => mapper.id, :mapset_id => map['beatmapset_id']).first[:status]]
  end
  status = STATUS_MAP[map['beatmap_status'].to_i]
  if is_new
    puts("Notifying for new map #{map_str} by #{mapper.username}")
  else
    puts("Notifying for updated map #{map_str} by #{mapper.username}: #{old_status} → #{status}")
  end
  ds = DB[:users].natural_join(:subscriptions).where(:mapper_id => mapper.id)
  ds.each do |sub|
    username = sub[:user_name]
    disc = sub[:user_disc]
    user = User.new(sub).to_discord_user
    begin
      if is_new
        user.pm("New map by #{mapper.username}: #{map_str}\nhttps://osu.ppy.sh/s/#{map['beatmapset_id']}")
      else
        user.pm("#{map_str} by #{mapper.username} has been updated: #{old_status} → #{status}\nhttps://osu.ppy.sh/s/#{map['beatmapset_id']}")
      end
    rescue  # User probably doesn't allow PMs from non-friends.
      puts("Sending to #{username}##{disc} failed.")
    else
      puts("Sent message to #{username}##{disc}")
    end
  end
end

if __FILE__ == $0
  puts("DB: #{DB_NAME}")
  puts("channel: #{CHANNEL}")
  BOT = setup
  mapsets = []  # Mapsets we've already seen.
  2.times do |i|
    result = HTTParty.get("#{SEARCH_URL}&offset=#{i}").parsed_response
    if result.start_with?('Server error')
      puts(result)
      exit
    end
    JSON.load(result)['beatmaps'].each do |map|
      mapper_name = map['mapper']
      status = map['beatmap_status'].to_i
      mapsets.include?(map['beatmapset_id']) && next
      mapsets.push(map['beatmapset_id'])
      if !DB[:mappers].where(:mapper_name => mapper_name).empty?
        mapper = Mapper.new(username: mapper_name)
        ds = DB[:maps].where(:mapper_id => mapper.id, :mapset_id => map['beatmapset_id'])
        if ds.empty? || ds.first[:status] != status
          notify(map, mapper)
          if ds.empty?
            DB[:maps].insert(:mapper_id => mapper.id, :mapset_id => map['beatmapset_id'], :status => map['beatmap_status'])
          else
            DB[:maps].where(:mapper_id => mapper.id, :mapset_id => map['beatmapset_id']).update(:status => status)
          end
        end
      end
    end
  end
end
