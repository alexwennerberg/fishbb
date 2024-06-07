create table forums (
  id integer primary key,
  name text,
  description text
);

create table threads (
  id integer primary key,
  forumid integer,
  authorid integer,
  title text,
  created text
);

create table posts (
  id integer primary key,
  threadid integer,
  authorid integer,
  reports integer,
  created text,
  edited text
);

create table users (
  id integer primary key,
  username text,
  hash text,
  email text,
  role text,
  avatar blob,
  active boolean,
  emailVerified boolean,
  about text,
  website text,
  created text
);

create table auth (
  userid integer,
  hash text,
  expiry text
);

create table config (
  csrfkey text 
);


-- create table invitations (
-- );
