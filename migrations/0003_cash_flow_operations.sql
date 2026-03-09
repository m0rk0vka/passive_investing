-- 0003_cash_flow_operations.sql
-- Добавление таблицы для хранения операций движения денежных средств

begin;

-- Тип операции движения денежных средств
do $$
begin
  if not exists (select 1 from pg_type where typname = 'cash_flow_operation_type') then
    create type cash_flow_operation_type as enum ('DEPOSIT', 'WITHDRAWAL', 'SECURITY_PURCHASE', 'SECURITY_SALE', 'DIVIDEND', 'TAX', 'FEE', 'OTHER');
  end if;
end $$;

-- Таблица для хранения операций движения денежных средств
create table if not exists cash_flow_operation (
  id              bigserial primary key,
  account_id      bigint not null references account(id),
  period          text not null,  -- "2025-10" format (YYYY-MM)
  
  operation_date  date not null,
  amount          numeric(20,2) not null,
  currency        text not null default 'RUB',
  operation_type  cash_flow_operation_type not null,
  comment         text,
  
  created_at      timestamptz not null default now(),
  updated_at      timestamptz not null default now(),
  
  unique(account_id, period, operation_date, amount, operation_type)
);

create index if not exists cash_flow_operation_account_period_idx
  on cash_flow_operation(account_id, period desc);

create index if not exists cash_flow_operation_date_idx
  on cash_flow_operation(operation_date desc);

drop trigger if exists cash_flow_operation_set_updated_at on cash_flow_operation;
create trigger cash_flow_operation_set_updated_at
before update on cash_flow_operation
for each row execute function set_updated_at();

commit;