-- 0004_buying_rules.sql
-- Правила пополнения портфеля: целевая структура по ISIN

begin;

create table if not exists buying_rule (
  id             bigserial primary key,
  portfolio_id   text not null,           -- "real_1" или "virtual_5"
  isin           text not null,
  ticker         text not null default '',
  security_name  text not null default '',
  target_pct     numeric(5,2) not null,   -- целевой процент, например 50.00

  created_at     timestamptz not null default now(),
  updated_at     timestamptz not null default now(),

  unique(portfolio_id, isin)
);

create index if not exists buying_rule_portfolio_idx on buying_rule(portfolio_id);

drop trigger if exists buying_rule_set_updated_at on buying_rule;
create trigger buying_rule_set_updated_at
before update on buying_rule
for each row execute function set_updated_at();

commit;
