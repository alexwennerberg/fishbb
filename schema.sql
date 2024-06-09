create table forums (
  id integer primary key,
  name text,
  description text,
  slug text,
  created datetime default CURRENT_TIMESTAMP
);

create table threads (
  id integer primary key,
  forumid integer,
  authorid integer,
  title text,
  locked boolean not null default false,
  pinned boolean not null default false,
  created datetime default CURRENT_TIMESTAMP
);

create table posts (
  id integer primary key,
  threadid integer,
  authorid integer,
  content text,
  reports integer,
  created datetime default CURRENT_TIMESTAMP,
  edited datetime
);

create table users (
  id integer primary key,
  username text,
  hash text,
  email text,
  role text,
  active boolean not null default false,
  emailVerified boolean not null default false,
  about text,
  website text,
  created datetime default CURRENT_TIMESTAMP
);

create table auth (
  userid integer,
  hash text,
  expiry text
);

create table config (
  csrfkey text 
);

create index idxforums_slug on forums(slug);
create index idxposts_threadid on posts(threadid);

-- create table invitations (
-- );

-- create table reports
-- create table notifications
