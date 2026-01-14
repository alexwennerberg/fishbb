create table if not exists board (
  id integer primary key,
  ownerid integer,
  name text,
  description text,
  foreign key (ownerid) references user(id)
);

create table if not exists forum (
  id integer primary key,
  boardid integer,
  name text,
  description text,
  slug text not null unique,
  read_permissions text not null default '',
  write_permissions text not null default 'user',
  created datetime default current_timestamp,
  foreign key (boardid) references board(id)
);

create table if not exists thread (
  id integer primary key,
  forumid integer,
  authorid integer,
  title text,
  locked int not null default false,
  pinned int not null default false,
  created datetime default current_timestamp,
  foreign key (forumid) references forum(id),
  foreign key (authorid) references user(id)
);

create table if not exists post (
  id integer primary key,
  threadid integer,
  authorid integer,
  content text,
  in_reply_to integer,
  created datetime default current_timestamp,
  edited datetime,
  foreign key (authorid) references user(id),
  foreign key (threadid) references thread(id)
);

create table if not exists user (
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
  foreign key (userid) references user(id)
);

-- sort of awkward bc it just stores a toml blob. TODO move away from toml
create table if not exists config (
  id integer primary key,
  key text unique,
  -- toml blob
  value text
);

create index if not exists idxforum_slug on forum(slug);
create index if not exists idxpost_threadid on post(threadid);
create index if not exists idxuser_username on user(username);

create trigger if not exists prevent_last_admin_deletion
before update of role on user
for each row
when OLD.role = 'admin' and (select count(*) from user where role = 'admin') = 1
begin
select raise(ABORT, 'Cannot remove the last admin'); end;

pragma journal_mode = wal;
pragma busy_timeout = 5000;
pragma synchronous = normal;
pragma cache_size = 1000000000;
pragma foreign_keys = true;
pragma temp_store = memory;
