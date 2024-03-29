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

CREATE TABLE desc_categories (
  description varchar(255) not null PRIMARY KEY,
  category varchar(255) not null
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

COMMENT ON TABLE desc_categories IS 'описание категорий';
COMMENT ON COLUMN desc_categories.description IS 'описание';
COMMENT ON COLUMN desc_categories.category IS 'категория';

insert into currencies(id, code, title, format)
values (36, 'AUD', 'Австралийский доллар', '%s'),
       (51, 'AMD', 'Армянских драмов', '%s ֏'),
       (124, 'CAD', 'Канадский доллар', '%s'),
       (156, 'CNY', 'Китайский юань', '%s'),
       (203, 'CZK', 'Чешских крон', '%s'),
       (208, 'DKK', 'Датская крона', '%s'),
       (344, 'HKD', 'Гонконгских долларов', '%s'),
       (348, 'HUF', 'Венгерских форинтов', '%s'),
       (356, 'INR', 'Индийских рупий', '%s'),
       (360, 'IDR', 'Индонезийских рупий', '%s'),
       (392, 'JPY', 'Японских иен', '%s'),
       (398, 'KZT', 'Казахстанских тенге', '%s'),
       (410, 'KRW', 'Вон Республики Корея', '%s'),
       (417, 'KGS', 'Киргизских сомов', '%s'),
       (498, 'MDL', 'Молдавских леев', '%s'),
       (554, 'NZD', 'Новозеландский доллар', '%s'),
       (578, 'NOK', 'Норвежских крон', '%s'),
       (634, 'QAR', 'Катарский риал', '%s'),
       (643, 'RUB', 'Российский рубль', '%s ₽'),
       (702, 'SGD', 'Сингапурский доллар', '%s'),
       (704, 'VND', 'Вьетнамских донгов', '%s'),
       (710, 'ZAR', 'Южноафриканских рэндов', '%s'),
       (752, 'SEK', 'Шведских крон', '%s'),
       (756, 'CHF', 'Швейцарский франк', '%s'),
       (764, 'THB', 'Таиландских батов', '%s'),
       (784, 'AED', 'Дирхам ОАЭ', '%s'),
       (818, 'EGP', 'Египетских фунтов', '%s'),
       (826, 'GBP', 'Фунт стерлингов Соединенного королевства', '%s £'),
       (840, 'USD', 'Доллар США', '$%s'),
       (860, 'UZS', 'Узбекских сумов', '%s'),
       (933, 'BYN', 'Белорусский рубль', '%s'),
       (934, 'TMT', 'Новый туркменский манат', '%s'),
       (941, 'RSD', 'Сербских динаров', '%s'),
       (944, 'AZN', 'Азербайджанский манат', '%s'),
       (946, 'RON', 'Румынский лей', '%s'),
       (949, 'TRY', 'Турецких лир', '%s ₺'),
       (960, 'XDR', 'СДР (специальные права заимствования)', '%s'),
       (972, 'TJS', 'Таджикских сомони', '%s'),
       (975, 'BGN', 'Болгарский лев', '%s'),
       (978, 'EUR', 'Евро', '%s €'),
       (980, 'UAH', 'Украинских гривен', '%s'),
       (981, 'GEL', 'Грузинский лари', '%s'),
       (985, 'PLN', 'Польский злотый', '%s'),
       (986, 'BRL', 'Бразильский реал', '%s')
;

insert into desc_categories(description, category)
values ('автобус', 'Транспорт'),
       ('поезд', 'Транспорт'),
       ('buscart', 'Транспорт'),
       ('проездной', 'Транспорт'),
       ('такси', 'Транспорт'),
       ('штраф', 'Транспорт'),
       ('парикмахерская', 'Красота'),
       ('gym', 'Спорт'),
       ('подшив', 'Одежда'),
       ('футболка', 'Одежда'),
       ('футболки', 'Одежда'),
       ('джинсы', 'Одежда'),
       ('одежда', 'Одежда'),
       ('кроссовки', 'Одежда'),
       ('ботинки', 'Одежда'),
       ('туфли', 'Одежда'),
       ('decathlon', 'Одежда'),
       ('кеды', 'Одежда'),
       ('рюкзак', 'Одежда'),
       ('флиска', 'Одежда'),
       ('подарок', 'Подарки'),
       ('подарки', 'Подарки'),
       ('сувениры', 'Подарки'),
       ('кафе', 'Рестораны'),
       ('mac', 'Рестораны'),
       ('kfc', 'Рестораны'),
       ('ресторан', 'Рестораны'),
       ('обед', 'Рестораны'),
       ('пицерия', 'Рестораны'),
       ('ужин', 'Рестораны'),
       ('ход-дог', 'Рестораны'),
       ('чай', 'Рестораны'),
       ('шаурма', 'Рестораны'),
       ('швепс', 'Рестораны'),
       ('энергетик', 'Рестораны'),
       ('sbb', 'Связь'),
       ('связь', 'Связь'),
       ('телефон', 'Связь'),
       ('интернет', 'Связь'),
       ('yettel', 'Связь'),
       ('мтс', 'Связь'),
       ('вода', 'Продукты'),
       ('лимонад', 'Продукты'),
       ('сок', 'Продукты'),
       ('кола', 'Продукты'),
       ('maxi', 'Продукты'),
       ('булочная', 'Продукты'),
       ('пекарня', 'Продукты'),
       ('рынок', 'Продукты'),
       ('супермаркет', 'Продукты'),
       ('тортик', 'Продукты'),
       ('шоколадка', 'Продукты'),
       ('панда', 'Дом и ремонт'),
       ('элефант', 'Дом и ремонт'),
       ('ikea', 'Дом и ремонт'),
       ('кастрюля', 'Дом и ремонт'),
       ('коврики', 'Дом и ремонт'),
       ('скотч', 'Дом и ремонт'),
       ('стол', 'Дом и ремонт'),
       ('стул', 'Дом и ремонт'),
       ('чайник', 'Дом и ремонт'),
       ('чашка', 'Дом и ремонт'),
       ('ножи', 'Дом и ремонт'),
       ('ручка', 'Дом и ремонт'),
       ('boosty', 'Сервисы'),
       ('protonmail', 'Сервисы'),
       ('youtube', 'Сервисы'),
       ('netflix', 'Сервисы'),
       ('стс', 'Сервисы'),
       ('italki', 'Образование'),
       ('slerm', 'Образование'),
       ('английский', 'Образование'),
       ('книга', 'Образование'),
       ('обучение', 'Образование'),
       ('концерт', 'Развлечения'),
       ('кино', 'Развлечения'),
       ('разное', 'Развлечения'),
       ('канцтовары', 'Развлечения'),
       ('zoovolonter', 'Благотворительность'),
       ('благотворительность', 'Благотворительность'),
       ('донаты', 'Благотворительность'),
       ('автомойка', 'Автомобиль'),
       ('автомобиль', 'Автомобиль'),
       ('стеклоомыватель', 'Автомобиль'),
       ('шиномонтаж', 'Автомобиль'),
       ('аренда', 'Коммуналка'),
       ('коммуналка', 'Коммуналка'),
       ('музей', 'Музеи'),
       ('музеи', 'Музеи'),
       ('автоаренда', 'Туризм'),
       ('визаран', 'Туризм'),
       ('спа-отель', 'Туризм'),
       ('отель', 'Туризм'),
       ('виртуальныйоофис', 'Бизнес'),
       ('бизнес', 'Бизнес'),
       ('документы', 'Документы'),
       ('фотографии', 'Документы'),
       ('массаж', 'Лечение'),
       ('стоматология', 'Лечение'),
       ('поликлиника', 'Лечение'),
       ('пластырь', 'Лечение'),
       ('аптеки', 'Лечение'),
       ('нотариус', 'Нотариус'),
       ('осаго', 'Страхование'),
       ('страхование', 'Страхование')
;
