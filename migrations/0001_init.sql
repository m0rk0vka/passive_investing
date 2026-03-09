-- 0001_init.sql
-- Minimal schema for Telegram uploads + raw storage + import batches

begin;

-- updated_at helper trigger (optional, but handy)
create or replace function set_updated_at()
returns trigger as $$
begin
  new.updated_at = now();
  return new;
end;
$$ language plpgsql;

-- --- Telegram entities

create table if not exists tg_user (
  id                bigserial primary key,
  telegram_user_id  bigint not null unique,
  created_at        timestamptz not null default now()
);

create table if not exists tg_chat (
  id                bigserial primary key,
  telegram_chat_id  bigint not null unique,
  created_at        timestamptz not null default now()
);

-- --- Uploads (fact of receiving a document)

do $$
begin
  if not exists (select 1 from pg_type where typname = 'upload_status') then
    create type upload_status as enum ('RECEIVED','PROCESSING','DONE','FAILED');
  end if;
end $$;

create table if not exists upload (
  id                      bigserial primary key,

  tg_user_id              bigint not null references tg_user(id),
  tg_chat_id              bigint not null references tg_chat(id),

  telegram_message_id     bigint,
  telegram_file_id        text,
  telegram_file_unique_id text,

  original_filename       text,
  mime_type               text,
  file_size               bigint,

  status                  upload_status not null default 'RECEIVED',
  error_message           text,

  created_at              timestamptz not null default now(),
  updated_at              timestamptz not null default now()
);

create index if not exists upload_status_created_idx
  on upload(status, created_at);

drop trigger if exists upload_set_updated_at on upload;
create trigger upload_set_updated_at
before update on upload
for each row execute function set_updated_at();

-- --- Raw file storage (where the XLSX ended up)

create table if not exists raw_file (
  id            bigserial primary key,
  upload_id     bigint not null unique references upload(id),

  sha256        text not null,
  storage_kind  text not null,   -- 'local' or 's3'
  storage_key   text not null,   -- path or object key
  stored_at     timestamptz not null default now(),

  unique(sha256)
);

-- --- Logical accounts (IIS/Broker/etc) - can be used later

do $$
begin
  if not exists (select 1 from pg_type where typname = 'account_type') then
    create type account_type as enum ('IIS','BROKER','SAVINGS','PDS','OTHER');
  end if;
end $$;

create table if not exists account (
  id          bigserial primary key,
  name        text not null,
  institution text not null,
  type        account_type not null,
  currency    text not null default 'RUB',
  created_at  timestamptz not null default now()
);

-- --- Import batches (processing result for an upload)

do $$
begin
  if not exists (select 1 from pg_type where typname = 'import_status') then
    create type import_status as enum ('NEW','PARSED','NORMALIZED','DONE','FAILED');
  end if;
end $$;

create table if not exists import_batch (
  id           bigserial primary key,
  upload_id    bigint not null unique references upload(id),

  account_id   bigint references account(id),

  period_start date,
  period_end   date,

  status       import_status not null default 'NEW',
  error_message text,

  created_at   timestamptz not null default now(),
  updated_at   timestamptz not null default now()
);

drop trigger if exists import_batch_set_updated_at on import_batch;
create trigger import_batch_set_updated_at
before update on import_batch
for each row execute function set_updated_at();

-- --- Snapshots: monthly portfolio state

-- valuation_snapshot: общая сумма портфеля за месяц
create table if not exists valuation_snapshot (
  id              bigserial primary key,
  account_id      bigint not null references account(id),
  period          text not null,  -- "2025-10" format (YYYY-MM)
  
  total_value     numeric(20,2) not null,
  cash_balance    numeric(20,2) not null default 0,
  securities_value numeric(20,2) not null default 0,
  
  currency        text not null default 'RUB',
  
  created_at      timestamptz not null default now(),
  updated_at      timestamptz not null default now(),
  
  unique(account_id, period)
);

create index if not exists valuation_snapshot_account_period_idx
  on valuation_snapshot(account_id, period desc);

drop trigger if exists valuation_snapshot_set_updated_at on valuation_snapshot;
create trigger valuation_snapshot_set_updated_at
before update on valuation_snapshot
for each row execute function set_updated_at();

-- position_snapshot: состав портфеля за месяц
create table if not exists position_snapshot (
  id              bigserial primary key,
  account_id      bigint not null references account(id),
  period          text not null,  -- "2025-10" format (YYYY-MM)
  
  isin            text not null,
  security_name   text not null,
  
  quantity        numeric(20,6) not null,
  price           numeric(20,2) not null,
  market_value    numeric(20,2) not null,
  
  currency        text not null default 'RUB',
  
  created_at      timestamptz not null default now(),
  updated_at      timestamptz not null default now(),
  
  unique(account_id, period, isin)
);

create index if not exists position_snapshot_account_period_idx
  on position_snapshot(account_id, period desc);

drop trigger if exists position_snapshot_set_updated_at on position_snapshot;
create trigger position_snapshot_set_updated_at
before update on position_snapshot
for each row execute function set_updated_at();

commit;
