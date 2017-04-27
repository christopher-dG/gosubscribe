class Mapper

  attr_accessor :username
  attr_accessor :id
  attr_accessor :error

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
        response.key?('username') || raise
        @username = HTTParty.get(url).parsed_response[0]['username']
      elsif !username.empty?
        url = "#{OSU_URL}/get_user?k=#{OSU_KEY}&u=#{username}&type=string"
        response = HTTParty.get(url).parsed_response[0]
        (response.key?('user_id') && response.key?('username')) || raise
        @id = response['user_id']
        @username = response['username']
      else
        raise
      end
    rescue
      @error = true
      return
    end

    begin
      update_db
    rescue
      @error = true
    end

    @error = false
  end

  def to_s() "#{@username}:#{@id}:#{@error}" end

  # Insert a mapper into the database.
  def update_db
    ds = DB[:mappers].where(:mapper_id => @id)
    if ds.first.nil?
      DB.call(:insert_mapper, :id => @id, :name => @username)
      mapset_ids = request_beatmaps
      if !mapset_ids.empty?
        DB[:maps].multi_insert(mapset_ids.map {|m| {:mapper_id => @id, :mapset_id => m}})
      end
    else
      ds = DB[:mappers].select(:mapper_name).where(:mapper_id => @id)
      if ds.first[:mapper_name] != @username
        # The user has changed their name, so update the row with the new name.
        DB[:mappers].where(:mapper_id => @id).update(:mapper_name => @username)
      end
    end
  end

  # Get a list of all the mapper's beatmaps.
  def request_beatmaps
    url = "#{OSU_URL}/get_beatmaps?k=#{OSU_KEY}&u=#{@id}&type=id"
    return Set.new(HTTParty.get(url).parsed_response.map {|b| b['beatmapset_id']})
  end

  # Get any new maps by the mapper.
  def diff_mapper
    new_maps = DB[:maps].select(:mapset_id).where(:mapper_id => @id).all.values - request_beatmaps
    DB[:maps].multi_insert(new_maps.map {|m| {:mapper_id => @id, :mapset_id => m}})
    return new_maps
  end

end
