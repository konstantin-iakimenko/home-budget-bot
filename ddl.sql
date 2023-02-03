CREATE SEQUENCE seq_bill_id START 1001;
CREATE SEQUENCE seq_bill_item_id START 100001;
CREATE SEQUENCE seq_user_id START 101;

CREATE TABLE users (
  id BIGINT NOT NULL DEFAULT nextval('seq_user_id') PRIMARY KEY,
  user_name varchar(255) not null default '',
  first_name varchar(255),
  last_name varchar(255),
  lang varchar(10),
  created_at timestamptz not null default CURRENT_TIMESTAMP
);

CREATE TABLE currencies (
  id BIGINT NOT NULL PRIMARY KEY,
  code varchar(10) not null default '',
  title varchar(255) not null default '',
  format varchar(255) not null default '%s'
);

CREATE TABLE bills (
  id BIGINT NOT NULL DEFAULT nextval('seq_bill_id') PRIMARY KEY,
  user_id bigint not null,
  bought_at timestamp not null,
  description varchar(255),
  category varchar(255),
  amount bigint not null default 0,
  currency bigint not null,
  amount_rub bigint not null default 0,
  amount_usd bigint not null default 0,
  created_at timestamptz not null default CURRENT_TIMESTAMP,
  CONSTRAINT fk_user_id FOREIGN KEY(user_id) REFERENCES users(id),
  CONSTRAINT fk_currency FOREIGN KEY(currency) REFERENCES currencies(id)
);

CREATE TABLE bill_items (
  id BIGINT NOT NULL DEFAULT nextval('seq_bill_item_id') PRIMARY KEY,
  bill_id bigint not null,
  title varchar(255) default null,
  price bigint not null default 0,
  cnt numeric(15, 6) default 1,
  amount bigint not null default 0,
  currency bigint not null,
  amount_rub bigint not null default 0,
  amount_usd bigint not null default 0,
  created_at timestamptz not null default CURRENT_TIMESTAMP,
  CONSTRAINT fk_bill_id FOREIGN KEY(bill_id) REFERENCES bills(id),
  CONSTRAINT fk_currency FOREIGN KEY(currency) REFERENCES currencies(id)
);

CREATE INDEX idx_users_name ON users (user_name);
CREATE INDEX idx_bills_date_category ON bills (bought_at, category);
CREATE INDEX idx_bills_category ON bills (category);
CREATE INDEX idx_bill_items_title ON bill_items (title);

COMMENT ON TABLE currencies IS 'валюты';
COMMENT ON COLUMN currencies.code IS 'код валюты';
COMMENT ON COLUMN currencies.title IS 'наименование валюты';
COMMENT ON COLUMN currencies.format IS 'формат вывода';

COMMENT ON TABLE users IS 'пользователи';
COMMENT ON COLUMN users.user_name IS 'ник пользователя';
COMMENT ON COLUMN users.first_name IS 'имя';
COMMENT ON COLUMN users.last_name IS 'фамилия';
COMMENT ON COLUMN users.lang IS 'язык';

COMMENT ON TABLE bills IS 'счета';
COMMENT ON COLUMN bills.amount IS 'сумма счета';
COMMENT ON COLUMN bills.currency IS 'валюта счета';
COMMENT ON COLUMN bills.amount_rub IS 'сумма счета в рублях';
COMMENT ON COLUMN bills.amount_usd IS 'сумма счета в долларах';
COMMENT ON COLUMN bills.bought_at IS 'дата покупки';

COMMENT ON TABLE bill_items IS 'товары в счете';
COMMENT ON COLUMN bill_items.title IS 'наимнование товара';
COMMENT ON COLUMN bill_items.price IS 'цена за единицу';
COMMENT ON COLUMN bill_items.cnt IS 'кол-во';
COMMENT ON COLUMN bill_items.amount IS 'сумма';
COMMENT ON COLUMN bill_items.currency IS 'валюта';
COMMENT ON COLUMN bill_items.amount_rub IS 'сумма в рублях';
COMMENT ON COLUMN bill_items.amount_usd IS 'сумма в долларах';
