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
  isnew = DB[:maps].where(:mapper_id => mapper.id, :mapset_id => map['beatmapset_id']).empty?
  if isnew
    old_status = STATUS_MAP[
      DB[:maps].where(
        :mapper_id => mapper.id, :mapset_id => map['beatmapset_id']
      ).first[:status]
    ]
  end
  status = STATUS_MAP[map['approved'].to_i]
  if isnew
    puts("Notifying for new map #{map_str} by #{mapper.username}")
  else
    puts("Notifying for updated map #{map_str} by #{mapper.username}: #{old_status} -> #{status}")
  end
  ds = DB[:users].natural_join(:subscriptions).where(:mapper_id => mapper.id)
  ds.each do |sub|
    username = sub[:user_name]
    user = User.new(sub).to_discord_user

    begin
      if is_new
        user.pm("New map by #{mapper.username}: #{map_str}\nhttps://osu.ppy.sh/b/#{map['beatmap_id']}")
      else
        user.pm("#{map_str} by #{mapper.username} has been updated: #{old_status} -> #{status}\nhttps://osu.ppy.sh/b/#{map['beatmap_id']}")
      end
    rescue
      puts("Sending to #{username} failed.")
    else
      puts("Sent message to #{username}.")
    end
  end
end

if __FILE__ == $0
  # now = DateTime.now
  # puts("#{now.year}-#{now.month}-#{now.day} #{now.hour}:#{now.minute}")
  BOT = setup
  mapsets = []  # Mapsets we've already seen.
  JSON.load(HTTParty.get(SEARCH_URL).parsed_response)['beatmaps'].each do |map|
    mapper_name = map['mapper']
    mappers.include?(map['beatmapset_id']) && next
    mappers.push(map['beatmapset_id'])
    if !DB[:mappers].where(:mapper_name => mapper_name).empty?
      mapper = Mapper.new(username: mapper_name)
      if DB[:maps].where(:mapper_id => mapper.id, :mapset_id => map['beatmapset_id']).empty?
        notify(map, mapper)
        DB[:maps].insert(:mapper_id => mapper.id, :mapset_id => map['beatmapset_id'], :status => map['beatmap_status'])
      end
    end
  end
end
