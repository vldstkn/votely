box.cfg{
    listen = '0.0.0.0:3301'
}

if not box.sequence.poll_id_seq then
    box.schema.sequence.create('poll_id_seq', {if_not_exists = true})
end

if not box.space.polls then
    local polls = box.schema.create_space('polls', { if_not_exists = true })
    engine = 'memtx'
    polls:format({
        { name = 'id',         type = 'unsigned' },
        { name = 'creator_id', type = 'string'   },
        { name = 'question',   type = 'string'   },
        { name = 'is_active',  type = 'boolean'  },
        { name = 'created_at', type = 'unsigned' }
    })
    polls:create_index('primary', {
        parts = { 'id' },
        if_not_exists = true,
        sequence = 'poll_id_seq'
    })
end

if not box.space.poll_options then
    local poll_options = box.schema.create_space('poll_options', { if_not_exists = true })
    poll_options:format({
        { name = 'poll_id',     type = 'unsigned' },
        { name = 'option_id',   type = 'unsigned' },
        { name = 'option_text', type = 'string'   },
        { name = 'votes_count', type = 'unsigned' }
    })
    poll_options:create_index('primary', {
        parts = { 'poll_id', 'option_id' },
        if_not_exists = true
    })
    poll_options:create_index('poll_id', {
        parts = { 'poll_id' },
        unique = false,
        if_not_exists = true
    })
end

if not box.space.votes then
    local votes = box.schema.create_space('votes', { if_not_exists = true })
    votes:format({
        { name = 'poll_id',    type = 'unsigned' },
        { name = 'user_id',    type = 'string'   },
        { name = 'option_id',  type = 'unsigned' },
        { name = 'created_at', type = 'unsigned' }
    })
    votes:create_index('primary', {
        parts = { 'poll_id', 'user_id' },
        if_not_exists = true
    })
    votes:create_index('by_option', {
        parts = { 'poll_id', 'option_id' },
        unique = false,
        if_not_exists = true
    })
end
