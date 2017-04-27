require_relative 'o!subscribe'


# Send messages about a new map to all subscribers of a given mapper.
# map: Map hash from the osu! API.
# mapper: Map creator.
def notify(map, mapper)
  puts("Notifying for new map by #{mapper.username}")
  ds = DB[:users].natural_join(:subscriptions).where(:mapper_id => mapper.id)
  ds.each do |data|
    username = data[:user_name]
    user = User.new(data).to_discord_user
    begin
      status = map['approved'] == '1' ? 'ranked' : 'qualified'
      user.pm("New #{status} map by #{mapper.username}: https://osu.ppy.sh/b/#{map['beatmap_id']}")
    rescue
      puts("Sending to #{username} for #{mapper.username}'s map failed.")
    else
      puts("Sent message to #{username} for #{mapper.username}'s map.")
    end
  end
end

if __FILE__ == $0
  now = DateTime.now
  puts("#{now.year}-#{now.month}-#{now.day} #{now.hour}:#{now.minute}")
  BOT = setup
  date = now - 1
  date = "#{date.year}-#{date.month}-#{date.day} #{date.hour}:#{date.minute}"
  url = "#{OSU_URL}/get_beatmaps?k=#{OSU_KEY}&since=#{date}"
  mappers = []
  HTTParty.get(url).parsed_response.each do |map|
    mapper_name = map['creator']
    mappers.include?(mapper_name) && next
    mappers.push(mapper_name)
    puts("Mapper: #{mapper_name}.")
    if !DB[:mappers].where(:mapper_name => mapper_name).empty?
      mapper = Mapper.new(username: mapper_name)
      if DB[:maps].where(:mapper_id => mapper.id, :mapset_id => map['beatmapset_id']).empty?
        notify(map, mapper)
        DB[:maps].insert(:mapper_id => mapper.id, :mapset_id => map['beatmapset_id'])
      end
    end
  end
end
