-- 002: friendships lookups for the friends feature.
--
-- The UNIQUE (requester_id, addressee_id) constraint from 001 still allows
-- A→B and B→A to coexist. One row per unordered pair keeps "are we friends?"
-- to a single lookup and makes a mutual request resolve to the same row.
CREATE UNIQUE INDEX IF NOT EXISTS uq_friendships_pair
  ON friendships (LEAST(requester_id, addressee_id), GREATEST(requester_id, addressee_id));

-- Outgoing requests / friends list for the requester side (001 only indexes
-- the addressee side).
CREATE INDEX IF NOT EXISTS idx_friendships_requester ON friendships(requester_id, status);
