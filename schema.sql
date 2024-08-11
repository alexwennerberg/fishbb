create table forums (
  id integer primary key,
  name text, 
  description text, 
  slug text not null unique, 
  read_permissions text not null default '',
  write_permissions text not null default 'user',
  created datetime default current_timestamp
);

-- TODO
create table forum_category (
)

create table threads (
  id integer primary key,
  forumid integer,
  authorid integer,
  title text,
  locked int not null default false,
  pinned int not null default false,
  created datetime default current_timestamp,
  foreign key (forumid) references forums(id),
  foreign key (authorid) references users(id)
);

create table posts (
  id integer primary key,
  threadid integer,
  authorid integer,
  content text,
  in_reply_to integer, 
  created datetime default current_timestamp,
  edited datetime,
  foreign key (authorid) references users(id),
  foreign key (threadid) references threads(id)
);

create table users (
  id integer primary key,
  username text,
  hash text,
  email text not null,
  role text not null default 'inactive',
  oauth text,
  emailverified int not null default false,
  about text not null default 'someone',
  website text not null default '',
  created datetime default current_timestamp
);

create table auth (
  userid integer,
  hash text,
  expiry text,
  foreign key (userid) references users(id)
);

-- sort of awkward bc it just stores a toml blob. TODO move away from toml
create table config (
  id integer primary key,
  key text unique,
  -- toml blob
  value text 
);

create index idxforums_slug on forums(slug);
create index idxposts_threadid on posts(threadid);
create index idxusers_username on users(username);

create trigger prevent_last_admin_deletion
before update of role on users 
for each row 
when NEW.role != 'admin' and (select count(*) from users where role = 'admin') = 1
begin
  select raise(ABORT, 'Cannot remove the last admin'); end;

pragma journal_mode = wal;
pragma busy_timeout = 5000;
pragma synchronous = normal;
pragma cache_size = 1000000000;
pragma foreign_keys = true;
pragma temp_store = memory;

