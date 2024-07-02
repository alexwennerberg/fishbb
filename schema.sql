create table forums (
  id integer primary key,
  name text,
  description text,
  slug text,
  created text default CURRENT_TIMESTAMP
);

create table threads (
  id integer primary key,
  forumid integer,
  authorid integer,
  title text,
  locked int not null default false,
  pinned int not null default false,
  created text default CURRENT_TIMESTAMP,
  foreign key (forumid) references forums(id),
  foreign key (authorid) references users(id)
);

create table posts (
  id integer primary key,
  threadid integer,
  authorid integer,
  content text,
  created text default CURRENT_TIMESTAMP,
  edited text,
  foreign key (authorid) references users(id),
  foreign key (threadid) references threads(id)
) strict;

create table users (
  id integer primary key,
  username text,
  hash text,
  email text,
  role text not null default 'user',
  active int not null default false,
  emailVerified int not null default false,
  about text,
  website text,
  created text default CURRENT_TIMESTAMP
) strict;

create table auth (
  userid integer,
  hash text,
  expiry text,
  foreign key (userid) references users(id)
) strict;

create table config (
  csrfkey text
) strict;

create index idxforums_slug on forums(slug);
create index idxposts_threadid on posts(threadid);

-- create table invitations ( );
-- create table reports
-- create table notifications

PRAGMA journal_mode = WAL;
PRAGMA busy_timeout = 5000;
PRAGMA synchronous = NORMAL;
PRAGMA cache_size = 1000000000;
PRAGMA foreign_keys = true;
PRAGMA temp_store = memory;
