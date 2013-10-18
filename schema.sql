--
-- PostgreSQL database dump
--

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


SET search_path = public, pg_catalog;

--
-- Name: binary_integrity(bytea, integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION binary_integrity(bytz bytea, hid integer) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
DECLARE
  result BOOLEAN;
BEGIN
  SELECT md5(bytz) = hashes.md5 INTO STRICT result FROM hashes WHERE hashes.id = hid;
  RETURN result;
END;
$$;


SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: annotations; Type: TABLE; Schema: public; Owner: -; Tablespace: 
--

CREATE TABLE annotations (
    id integer NOT NULL,
    type text,
    annotation text,
    "timestamp" timestamp without time zone DEFAULT now(),
    hash_id integer NOT NULL
);


--
-- Name: annotations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE annotations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: annotations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE annotations_id_seq OWNED BY annotations.id;


--
-- Name: contents; Type: TABLE; Schema: public; Owner: -; Tablespace: 
--

CREATE TABLE contents (
    hash_id integer NOT NULL,
    bytes bytea,
    CONSTRAINT bytes_integrity CHECK (binary_integrity(bytes, hash_id))
);


--
-- Name: hashes; Type: TABLE; Schema: public; Owner: -; Tablespace: 
--

CREATE TABLE hashes (
    id integer NOT NULL,
    adler32 character(8) NOT NULL,
    crc32 character(8) NOT NULL,
    md5 character(32) NOT NULL,
    ripemd160 character(40) NOT NULL,
    sha1 character(40) NOT NULL,
    "sha2-256" character(64) NOT NULL,
    "sha2-512" character(128) NOT NULL,
    "sha3-256" character(64) NOT NULL,
    ssdeep29 text,
    size bigint NOT NULL,
    version text NOT NULL,
    CONSTRAINT hashes_adler32_check CHECK ((adler32 ~ '\A[0-9a-f]{8}\Z'::text)),
    CONSTRAINT hashes_crc32_check CHECK ((crc32 ~ '\A[0-9a-f]{8}\Z'::text)),
    CONSTRAINT hashes_md5_check CHECK ((md5 ~ '\A[0-9a-f]{32}\Z'::text)),
    CONSTRAINT hashes_ripemd160_check CHECK ((ripemd160 ~ '\A[0-9a-f]{40}\Z'::text)),
    CONSTRAINT hashes_sha1_check CHECK ((sha1 ~ '\A[0-9a-f]{40}\Z'::text)),
    CONSTRAINT "hashes_sha2-256_check" CHECK (("sha2-256" ~ '\A[0-9a-f]{64}\Z'::text)),
    CONSTRAINT "hashes_sha2-512_check" CHECK (("sha2-512" ~ '\A[0-9a-f]{128}\Z'::text)),
    CONSTRAINT "hashes_sha3-256_check" CHECK (("sha3-256" ~ '\A[0-9a-f]{64}\Z'::text)),
    CONSTRAINT hashes_size_check CHECK ((size >= 0)),
    CONSTRAINT hashes_version_check CHECK ((length(version) > 0))
);


--
-- Name: hashes_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE hashes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: hashes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE hashes_id_seq OWNED BY hashes.id;


--
-- Name: properties; Type: TABLE; Schema: public; Owner: -; Tablespace: 
--

CREATE TABLE properties (
    id integer NOT NULL,
    type text,
    property text,
    hash_id integer NOT NULL
);


--
-- Name: properties_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE properties_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: properties_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE properties_id_seq OWNED BY properties.id;


--
-- Name: tags; Type: TABLE; Schema: public; Owner: -; Tablespace: 
--

CREATE TABLE tags (
    tag text NOT NULL,
    hash_id integer NOT NULL
);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY annotations ALTER COLUMN id SET DEFAULT nextval('annotations_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY hashes ALTER COLUMN id SET DEFAULT nextval('hashes_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY properties ALTER COLUMN id SET DEFAULT nextval('properties_id_seq'::regclass);


--
-- Name: annotations_annotation_idx; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX annotations_annotation_idx ON annotations USING btree (annotation);


--
-- Name: annotations_hash_id_idx; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX annotations_hash_id_idx ON annotations USING btree (hash_id);


--
-- Name: annotations_id_idx; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX annotations_id_idx ON annotations USING btree (id);


--
-- Name: annotations_timestamp_idx; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX annotations_timestamp_idx ON annotations USING btree ("timestamp");


--
-- Name: contents_bytes_idx; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE UNIQUE INDEX contents_bytes_idx ON contents USING btree (bytes);


--
-- Name: contents_hash_id_idx; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE UNIQUE INDEX contents_hash_id_idx ON contents USING btree (hash_id);


--
-- Name: hashes_adler32_idx; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX hashes_adler32_idx ON hashes USING btree (adler32);


--
-- Name: hashes_crc32_idx; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX hashes_crc32_idx ON hashes USING btree (crc32);


--
-- Name: hashes_id_idx; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE UNIQUE INDEX hashes_id_idx ON hashes USING btree (id);


--
-- Name: hashes_md5_idx; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX hashes_md5_idx ON hashes USING btree (md5);


--
-- Name: hashes_ripemd160_idx; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX hashes_ripemd160_idx ON hashes USING btree (ripemd160);


--
-- Name: hashes_sha1_idx; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX hashes_sha1_idx ON hashes USING btree (sha1);


--
-- Name: hashes_sha2-256_idx; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX "hashes_sha2-256_idx" ON hashes USING btree ("sha2-256");


--
-- Name: hashes_sha2-512_idx; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX "hashes_sha2-512_idx" ON hashes USING btree ("sha2-512");


--
-- Name: hashes_sha3-256_idx; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX "hashes_sha3-256_idx" ON hashes USING btree ("sha3-256");


--
-- Name: hashes_size_idx; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX hashes_size_idx ON hashes USING btree (size);


--
-- Name: no_duplicate_hashes; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE UNIQUE INDEX no_duplicate_hashes ON hashes USING btree ("sha2-256", "sha3-256");


--
-- Name: no_duplicate_properties; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE UNIQUE INDEX no_duplicate_properties ON properties USING btree (hash_id, type);


--
-- Name: no_duplicate_tags; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE UNIQUE INDEX no_duplicate_tags ON tags USING btree (hash_id, tag);


--
-- Name: properties_hash_id_idx; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX properties_hash_id_idx ON properties USING btree (hash_id);


--
-- Name: properties_id_idx; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX properties_id_idx ON properties USING btree (id);


--
-- Name: tags_hash_id_idx; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX tags_hash_id_idx ON tags USING btree (hash_id);


--
-- Name: annotations_hash_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY annotations
    ADD CONSTRAINT annotations_hash_id_fkey FOREIGN KEY (hash_id) REFERENCES hashes(id);


--
-- Name: content_hash_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY contents
    ADD CONSTRAINT content_hash_id_fkey FOREIGN KEY (hash_id) REFERENCES hashes(id);


--
-- Name: properties_hash_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY properties
    ADD CONSTRAINT properties_hash_id_fkey FOREIGN KEY (hash_id) REFERENCES hashes(id);


--
-- Name: tags_hash_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY tags
    ADD CONSTRAINT tags_hash_id_fkey FOREIGN KEY (hash_id) REFERENCES hashes(id);


--
-- Name: public; Type: ACL; Schema: -; Owner: -
--

REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON SCHEMA public FROM postgres;
GRANT ALL ON SCHEMA public TO postgres;
GRANT ALL ON SCHEMA public TO PUBLIC;


--
-- PostgreSQL database dump complete
--

