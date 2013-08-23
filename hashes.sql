DROP TABLE IF EXISTS annotations;
DROP TABLE IF EXISTS properties;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS hashes;

CREATE TABLE hashes (
  id SERIAL,
  bytes BYTEA,
  adler32 CHARACTER(8) NOT NULL CHECK(adler32 ~ '\A[0-9a-f]{8}\Z'),
  crc32 CHARACTER(8) NOT NULL CHECK(crc32 ~ '\A[0-9a-f]{8}\Z'),
  md5 CHARACTER(32) NOT NULL CHECK(md5 ~ '\A[0-9a-f]{32}\Z'),
  ripemd160 CHARACTER(40) NOT NULL CHECK(ripemd160 ~ '\A[0-9a-f]{40}\Z'),
  sha1 CHARACTER(40) NOT NULL CHECK(sha1 ~ '\A[0-9a-f]{40}\Z'),
  "sha2-256" CHARACTER(64) NOT NULL CHECK("sha2-256" ~ '\A[0-9a-f]{64}\Z'),
  "sha2-512" CHARACTER(128) NOT NULL CHECK("sha2-512" ~ '\A[0-9a-f]{128}\Z'),
  "sha3-256" CHARACTER(64) NOT NULL CHECK("sha3-256" ~ '\A[0-9a-f]{64}\Z'),
  ssdeep29 TEXT,
  size BIGINT NOT NULL CHECK(size >= 0),
  version TEXT NOT NULL CHECK(LENGTH(version) > 0)
);

CREATE UNIQUE INDEX ON hashes(id);
CREATE UNIQUE INDEX no_duplicate_hashes ON hashes("sha2-256", "sha3-256");

CREATE INDEX ON hashes(bytes);

CREATE INDEX ON hashes(adler32);
CREATE INDEX ON hashes(crc32);
CREATE INDEX ON hashes(md5);
CREATE INDEX ON hashes(ripemd160);
CREATE INDEX ON hashes(sha1);
CREATE INDEX ON hashes("sha2-256");
CREATE INDEX ON hashes("sha2-512");
CREATE INDEX ON hashes("sha3-256");
CREATE INDEX ON hashes(size);

DROP FUNCTION IF EXISTS binary_integrity(bytes BYTEA, mymd5 CHARACTER);
CREATE FUNCTION binary_integrity(bytes BYTEA, mymd5 CHARACTER)
RETURNS BOOLEAN
AS
$$
DECLARE
  result BOOLEAN;
BEGIN
  IF bytes IS NULL THEN
    RETURN TRUE;
  END IF;
  IF LENGTH(bytes) = 0 THEN
    RETURN TRUE;
  END IF;
  SELECT md5(bytes) = mymd5 INTO STRICT result;
  RETURN result;
END;
$$ LANGUAGE plpgsql;

ALTER TABLE hashes ADD CONSTRAINT bytes_integrity CHECK (binary_integrity(bytes, md5));

CREATE TABLE annotations (
  id SERIAL,
  type TEXT,
  annotation TEXT,
  timestamp TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
  hash_id INTEGER NOT NULL REFERENCES hashes (id)
);

CREATE INDEX ON annotations(id);
CREATE INDEX ON annotations(annotation);
CREATE INDEX ON annotations(timestamp);
CREATE INDEX ON annotations(hash_id);

CREATE TABLE properties (
  id SERIAL,
  type TEXT,
  property TEXT,
  hash_id INTEGER NOT NULL REFERENCES hashes (id)
);

CREATE UNIQUE INDEX no_duplicate_properties ON properties(hash_id, type);

CREATE INDEX ON properties(id);
CREATE INDEX ON properties(hash_id);

CREATE TABLE tags (
  tag TEXT NOT NULL,
  hash_id INTEGER NOT NULL REFERENCES hashes (id)
);

CREATE UNIQUE INDEX no_duplicate_tags ON tags(hash_id, tag);

CREATE INDEX ON tags(hash_id);
