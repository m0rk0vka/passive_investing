-- 0002_virtual_portfolios.sql
-- Add support for virtual portfolios (aggregation of multiple accounts)

begin;

-- Virtual Portfolio: aggregates multiple accounts into one view
create table if not exists virtual_portfolio (
  id         bigserial primary key,
  user_id    bigint not null references tg_user(id),
  name       text not null,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  
  unique(user_id, name)
);

create index if not exists virtual_portfolio_user_id_idx on virtual_portfolio(user_id);

drop trigger if exists virtual_portfolio_set_updated_at on virtual_portfolio;
create trigger virtual_portfolio_set_updated_at
before update on virtual_portfolio
for each row execute function set_updated_at();

-- Virtual Portfolio Items: which accounts are included in virtual portfolio
create table if not exists virtual_portfolio_item (
  id                   bigserial primary key,
  virtual_portfolio_id bigint not null references virtual_portfolio(id) on delete cascade,
  account_id           bigint not null references account(id) on delete cascade,
  created_at           timestamptz not null default now(),
  
  unique(virtual_portfolio_id, account_id)
);

create index if not exists virtual_portfolio_item_vp_id_idx 
  on virtual_portfolio_item(virtual_portfolio_id);
create index if not exists virtual_portfolio_item_account_id_idx 
  on virtual_portfolio_item(account_id);

commit;