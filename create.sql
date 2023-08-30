CREATE TABLE IF NOT EXISTS segment (
   name character varying(200) NOT NULL PRIMARY KEY
);
CREATE TABLE IF NOT EXISTS users_in_segment (
   user_id integer NOT NULL,
   segment character varying(200) NOT NULL REFERENCES segment(name) ON DELETE CASCADE,
   UNIQUE (user_id, segment)
);