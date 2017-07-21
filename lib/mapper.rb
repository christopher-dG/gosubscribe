class Mapper
  attr_reader :username
  attr_reader :id
  attr_reader :error

  # Supply at at least one of username or id.
  def initialize(username: '', id: '')
    begin
      if !id.to_s.empty? && !username.empty?
        @id = id.to_s
        @username = username
      elsif !id.to_s.empty?
        @id = id.to_s
        url = "#{OSU_URL}/get_user?k=#{OSU_KEY}&u=#{@id}&type=id"
        response = HTTParty.get(url).parsed_response[0]
        raise if !response.key?('username')
        @username = HTTParty.get(url).parsed_response[0]['username']
      elsif !username.empty?
        url = "#{OSU_URL}/get_user?k=#{OSU_KEY}&u=#{username}&type=string"
        response = HTTParty.get(url).parsed_response[0]
        raise if !response.key?('user_id') || !response.key?('username')
        @id = response['user_id']
        @username = response['username']
      else
        raise
      end
    rescue
      @error = true
      return
    end

    update_db
    @error = false
  end

  def to_s() "#{@username}:#{@id}:#{@error}" end

  # Insert a mapper into the database.
  def update_db
    ds = ds = DB[:mappers].select(:mapper_name).where(:mapper_id => @id)
    if ds.empty?
      DB.call(:insert_mapper, :id => @id, :name => @username)
      request_beatmaps.each do |map|
        begin
          DB.call(:insert_map, :mapper => @id, :map => map[:id], :status => map[:status])
        rescue  # Weird duplicate key issue where map has multiple statuses.
        end
      end
    else
      if ds.first[:mapper_name] != @username
        # The user has changed their name, so update the row with the new name.
        DB[:mappers].where(:mapper_id => @id).update(:mapper_name => @username)
      end
    end
  end

  # Get a list of all the mapper's beatmaps. Returns a list of hashes.
  def request_beatmaps
    # Todo: Make sure there are no duplicates here.
    url = "#{OSU_URL}/get_beatmaps?k=#{OSU_KEY}&u=#{@id}&type=id"
    maps = HTTParty.get(url).parsed_response
    return Set.new(maps.map {|m| {:id => m['beatmapset_id'], :status => m['approved']}})
  end
end
