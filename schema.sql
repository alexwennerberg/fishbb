create table if not exists boards (
  id integer primary key,
  subdomain text,
  description text
);

create table if not exists forums (
  id integer primary key,
  name text, 
  description text, 
  slug text not null unique, 
  read_permissions text not null default '',
  write_permissions text not null default 'user',
  created datetime default current_timestamp
);

create table if not exists threads (
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

create table if not exists posts (
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

create table if not exists users (
  id integer primary key,
  username text not null unique,
  hash text,
  email text not null unique,
  role text not null default 'inactive',
  email_public int not null default false,
  publicemail int not null default false,
  about text not null default 'someone',
  website text not null default '',
  created datetime default current_timestamp,
  mentions_checked datetime default '2000-01-01'
);

create table if not exists auth (
  userid integer,
  hash text,
  expiry text,
  foreign key (userid) references users(id)
);

-- sort of awkward bc it just stores a toml blob. TODO move away from toml
create table if not exists config (
  id integer primary key,
  key text unique,
  -- toml blob
  value text 
);

create index if not exists idxforums_slug on forums(slug);
create index if not exists idxposts_threadid on posts(threadid);
create index if not exists idxusers_username on users(username);

create trigger if not exists prevent_last_admin_deletion
before update of role on users 
for each row 
when OLD.role = 'admin' and (select count(*) from users where role = 'admin') = 1
begin
select raise(ABORT, 'Cannot remove the last admin'); end;

pragma journal_mode = wal;
pragma busy_timeout = 5000;
pragma synchronous = normal;
pragma cache_size = 1000000000;
pragma foreign_keys = true;
pragma temp_store = memory;
