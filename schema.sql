create table forums (
  id integer primary key,
  name text, description text, slug text,
  created text default current_timestamp
);

create table threads (
  id integer primary key,
  forumid integer,
  authorid integer,
  title text,
  locked int not null default false,
  pinned int not null default false,
  created text default current_timestamp,
  foreign key (forumid) references forums(id),
  foreign key (authorid) references users(id)
);

create table posts (
  id integer primary key,
  threadid integer,
  authorid integer,
  content text,
  created text default current_timestamp,
  edited text,
  foreign key (authorid) references users(id),
  foreign key (threadid) references threads(id)
) strict;

create table users (
  id integer primary key,
  username text,
  hash text,
  email text not null,
  role text not null default 'user',
  active int not null default false,
  emailverified int not null default false,
  about text not null default 'someone',
  website text not null default '',
  created text default current_timestamp
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

pragma journal_mode = wal;
pragma busy_timeout = 5000;
pragma synchronous = normal;
pragma cache_size = 1000000000;
pragma foreign_keys = true;
pragma temp_store = memory;
